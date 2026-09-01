// Package memory builds and queries private Git-bound repository memory.
package memory

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/orchestrationconfig"
	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
	"github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"
)

const (
	maxGraphBytes   = 64 << 20
	maxVectorBytes  = 256 << 20
	maxIndexBytes   = 128 << 20
	maxSegmentBytes = 16 << 20
	maxNodes        = 1_000_000
	maxEdges        = 4_000_000
	parserRevision  = 1
)

type Embedder interface {
	Embed(context.Context, []string) (int, int, [][]float32, error)
}

type Adapter struct {
	resolve  func(string) (processadapter.Executable, error)
	run      func(context.Context, processadapter.Request) (processadapter.Result, error)
	embedder Embedder
	now      func() time.Time
}

type vectorFile struct {
	Schema    int            `json:"schema"`
	Revision  int            `json:"revision"`
	Dimension int            `json:"dimension"`
	Vectors   []vectorRecord `json:"vectors"`
}

type vectorRecord struct {
	NodeID string    `json:"node_id"`
	Values []float32 `json:"values"`
}

type memoryManifest struct {
	Schema         int            `json:"schema"`
	RepositoryHead string         `json:"repository_head"`
	ParserRevision int            `json:"parser_revision"`
	Files          []manifestFile `json:"files"`
	UpdatedAtUTC   string         `json:"updated_at_utc"`
	Next           string         `json:"next"`
}

type manifestFile struct {
	Path    string `json:"path"`
	Digest  string `json:"digest"`
	Segment string `json:"segment"`
}

type graphSegment struct {
	Schema         int                 `json:"schema"`
	Path           string              `json:"path"`
	Digest         string              `json:"digest"`
	ParserRevision int                 `json:"parser_revision"`
	Nodes          []domain.MemoryNode `json:"nodes"`
	Edges          []domain.MemoryEdge `json:"edges"`
	References     []domain.MemoryEdge `json:"references"`
}

type lexicalFile struct {
	Schema int                 `json:"schema"`
	Terms  map[string][]string `json:"terms"`
}

func New(embedder Embedder) Adapter {
	return NewWith(processadapter.Resolve, (processadapter.Runner{}).Run, embedder, time.Now)
}

func NewWith(resolve func(string) (processadapter.Executable, error), run func(context.Context, processadapter.Request) (processadapter.Result, error), embedder Embedder, now func() time.Time) Adapter {
	if resolve == nil {
		resolve = processadapter.Resolve
	}
	if run == nil {
		run = (processadapter.Runner{}).Run
	}
	if now == nil {
		now = time.Now
	}
	return Adapter{resolve: resolve, run: run, embedder: embedder, now: now}
}

func (adapter Adapter) Sync(ctx context.Context, root, commonDirectory string, policy orchestrationconfig.Memory) (domain.MemoryGraph, error) {
	physicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil || !filepath.IsAbs(physicalRoot) || physicalRoot != filepath.Clean(root) {
		return domain.MemoryGraph{}, errors.New("memory root is unsafe")
	}
	release, err := syncLock(commonDirectory)
	if err != nil {
		return domain.MemoryGraph{}, err
	}
	defer release()
	git, err := adapter.resolve("git")
	if err != nil {
		return domain.MemoryGraph{}, errors.New("Git is unavailable")
	}
	head, err := adapter.gitText(ctx, git, physicalRoot, []string{"rev-parse", "HEAD"}, 4096)
	if err != nil || len(head) != 40 {
		return domain.MemoryGraph{}, errors.New("cannot establish repository head")
	}
	result, err := adapter.run(ctx, processadapter.Request{
		Executable: git.Path, Arguments: []string{"ls-files", "-co", "--exclude-standard", "-z"}, Directory: physicalRoot,
		Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: 64 << 20, Timeout: 2 * time.Minute,
	})
	if err != nil || result.ExitCode != 0 {
		return domain.MemoryGraph{}, errors.New("cannot enumerate repository files")
	}
	paths := splitNUL(result.Stdout)
	if len(paths) > 1_000_000 {
		return domain.MemoryGraph{}, errors.New("repository file count exceeds memory limit")
	}
	sort.Strings(paths)
	builder := graphBuilder{nodes: []domain.MemoryNode{}, edges: []domain.MemoryEdge{}, nodeIDs: make(map[string]bool), symbols: make(map[string]string)}
	manifest := memoryManifest{Schema: domain.OrchestrationSchema, RepositoryHead: head, ParserRevision: parserRevision, Files: []manifestFile{}, UpdatedAtUTC: adapter.now().UTC().Format(time.RFC3339), Next: "run l7 sync --query <text>"}
	commitID := stableID("commit", head)
	builder.addNode(domain.MemoryNode{ID: commitID, Kind: "commit", Name: head, Digest: head, Summary: "repository head"})
	for _, relative := range paths {
		if excluded(relative, policy.Exclude) || secretPath(relative) {
			continue
		}
		absolute, err := safePath(physicalRoot, relative)
		if err != nil {
			continue
		}
		info, err := os.Stat(absolute)
		if err != nil || !info.Mode().IsRegular() || info.Size() < 0 || info.Size() > int64(policy.MaxFileBytes) {
			continue
		}
		data, err := os.ReadFile(absolute)
		if err != nil || bytes.IndexByte(data, 0) >= 0 || !utf8.Valid(data) || likelySecret(data) {
			continue
		}
		digest := fmt.Sprintf("sha256:%x", sha256.Sum256(data))
		segmentID := stableID("segment", relative, digest, fmt.Sprint(parserRevision))
		segmentPath := "memory/segments/" + strings.TrimPrefix(segmentID, "sha256:") + ".json"
		segment, loadErr := loadSegment(commonDirectory, segmentPath, relative, digest)
		if loadErr != nil {
			segment = buildSegment(relative, digest, data)
			if err := saveImmutable(commonDirectory, segmentPath, segment, maxSegmentBytes); err != nil && !errors.Is(err, os.ErrExist) {
				return domain.MemoryGraph{}, err
			}
		}
		builder.addSegment(segment, commitID)
		manifest.Files = append(manifest.Files, manifestFile{Path: relative, Digest: digest, Segment: segmentID})
		if len(builder.nodes) > maxNodes || len(builder.edges) > maxEdges {
			return domain.MemoryGraph{}, errors.New("memory graph exceeds node or edge bounds")
		}
	}
	builder.resolveReferences()
	sort.Slice(builder.nodes, func(i, j int) bool { return builder.nodes[i].ID < builder.nodes[j].ID })
	sort.Slice(builder.edges, func(i, j int) bool {
		if builder.edges[i].From != builder.edges[j].From {
			return builder.edges[i].From < builder.edges[j].From
		}
		if builder.edges[i].Kind != builder.edges[j].Kind {
			return builder.edges[i].Kind < builder.edges[j].Kind
		}
		return builder.edges[i].To < builder.edges[j].To
	})
	graph := domain.MemoryGraph{Schema: domain.OrchestrationSchema, RepositoryHead: head, EmbeddingProvider: "apple_natural_language:unavailable", Nodes: builder.nodes, Edges: builder.edges, UpdatedAtUTC: adapter.now().UTC().Format(time.RFC3339), Next: "run l7 sync --query <text>"}
	vectors := vectorFile{Schema: domain.OrchestrationSchema, Vectors: []vectorRecord{}}
	if adapter.embedder != nil && len(graph.Nodes) > 0 {
		texts := make([]string, len(graph.Nodes))
		for index, node := range graph.Nodes {
			texts[index] = strings.TrimSpace(node.Kind + " " + node.Path + " " + node.Name + " " + node.Summary)
		}
		revision, dimension, values, embedErr := adapter.embedder.Embed(ctx, texts)
		if embedErr == nil && revision > 0 && dimension > 0 && len(values) == len(graph.Nodes) {
			valid := true
			for index, vector := range values {
				if len(vector) != dimension {
					valid = false
					break
				}
				vectors.Vectors = append(vectors.Vectors, vectorRecord{NodeID: graph.Nodes[index].ID, Values: vector})
			}
			if valid {
				vectors.Revision, vectors.Dimension = revision, dimension
				graph.EmbeddingProvider = "apple_natural_language"
				graph.EmbeddingRevision = revision
				graph.EmbeddingDimension = dimension
			} else {
				vectors.Vectors = []vectorRecord{}
			}
		}
	}
	if err := save(commonDirectory, "memory/graph.json", graph, maxGraphBytes); err != nil {
		return domain.MemoryGraph{}, err
	}
	if err := save(commonDirectory, "memory/vectors.json", vectors, maxVectorBytes); err != nil {
		return domain.MemoryGraph{}, err
	}
	if err := save(commonDirectory, "memory/lexical.json", lexicalFile{Schema: domain.OrchestrationSchema, Terms: buildLexicalIndex(graph.Nodes)}, maxIndexBytes); err != nil {
		return domain.MemoryGraph{}, err
	}
	// The atomic manifest is accepted last. Content-addressed segments remain
	// canonical if any derived graph, lexical, or vector index is corrupted.
	if err := save(commonDirectory, "memory/manifest.json", manifest, maxGraphBytes); err != nil {
		return domain.MemoryGraph{}, err
	}
	return graph, nil
}

func (adapter Adapter) Query(ctx context.Context, commonDirectory, query string, limit int) ([]domain.MemoryMatch, error) {
	if query == "" || len(query) > 4096 || strings.ContainsRune(query, 0) || limit < 1 || limit > 100 {
		return nil, errors.New("memory query bounds are invalid")
	}
	var graph domain.MemoryGraph
	if err := load(commonDirectory, "memory/graph.json", &graph, maxGraphBytes); err != nil {
		return nil, err
	}
	if graph.Schema != domain.OrchestrationSchema || len(graph.Nodes) > maxNodes || len(graph.Edges) > maxEdges {
		return nil, errors.New("memory graph is invalid")
	}
	var vectors vectorFile
	vectorAvailable := load(commonDirectory, "memory/vectors.json", &vectors, maxVectorBytes) == nil && vectors.Schema == domain.OrchestrationSchema && vectors.Revision == graph.EmbeddingRevision && vectors.Dimension == graph.EmbeddingDimension
	queryVector := []float32(nil)
	if vectorAvailable && adapter.embedder != nil {
		revision, dimension, values, err := adapter.embedder.Embed(ctx, []string{query})
		if err == nil && revision == vectors.Revision && dimension == vectors.Dimension && len(values) == 1 {
			queryVector = values[0]
		}
	}
	vectorByNode := make(map[string][]float32)
	for _, record := range vectors.Vectors {
		vectorByNode[record.NodeID] = record.Values
	}
	terms := terms(query)
	lexicalHits := make(map[string]int)
	var lexical lexicalFile
	if load(commonDirectory, "memory/lexical.json", &lexical, maxIndexBytes) == nil && lexical.Schema == domain.OrchestrationSchema {
		for _, term := range terms {
			for _, nodeID := range lexical.Terms[term] {
				lexicalHits[nodeID] += 100
			}
		}
	}
	matches := []domain.MemoryMatch{}
	for _, node := range graph.Nodes {
		haystack := strings.ToLower(node.Path + " " + node.Name + " " + node.Summary + " " + node.Kind)
		score := lexicalHits[node.ID]
		why := []string{}
		for _, term := range terms {
			if strings.Contains(haystack, term) {
				if lexicalHits[node.ID] == 0 {
					score += 100
				}
				why = append(why, "lexical:"+term)
			}
		}
		if len(queryVector) > 0 {
			if vector := vectorByNode[node.ID]; len(vector) == len(queryVector) {
				semantic := int((cosine(queryVector, vector) + 1) * 50)
				if semantic > 0 {
					score += semantic
					why = append(why, "local-semantic")
				}
			}
		}
		if score > 0 {
			if node.Kind == "function" || node.Kind == "type" || node.Kind == "class" || node.Kind == "module" {
				score += 25
				why = append(why, "structural")
			}
			if node.Kind == "file" || node.Kind == "test" || node.Kind == "decision" {
				score += 10
				why = append(why, "current-head")
			}
			matches = append(matches, domain.MemoryMatch{Node: node, Score: score, Why: why})
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Score != matches[j].Score {
			return matches[i].Score > matches[j].Score
		}
		return matches[i].Node.ID < matches[j].Node.ID
	})
	if len(matches) > limit {
		matches = matches[:limit]
	}
	return matches, nil
}

// LoadGraph reads and validates the canonical private graph without rebuilding
// it. Onboard uses this to inspect memory availability without mutation.
func LoadGraph(commonDirectory string) (domain.MemoryGraph, error) {
	var graph domain.MemoryGraph
	if err := load(commonDirectory, "memory/graph.json", &graph, maxGraphBytes); err != nil {
		return graph, err
	}
	if graph.Schema != domain.OrchestrationSchema || len(graph.Nodes) > maxNodes || len(graph.Edges) > maxEdges || graph.Next == "" {
		return domain.MemoryGraph{}, errors.New("memory graph is invalid")
	}
	return graph, nil
}

type graphBuilder struct {
	nodes      []domain.MemoryNode
	edges      []domain.MemoryEdge
	nodeIDs    map[string]bool
	symbols    map[string]string
	references []domain.MemoryEdge
}

func buildSegment(relative, digest string, data []byte) graphSegment {
	builder := graphBuilder{nodes: []domain.MemoryNode{}, edges: []domain.MemoryEdge{}, nodeIDs: make(map[string]bool), symbols: make(map[string]string)}
	kind := "file"
	if isTestPath(relative) {
		kind = "test"
	} else if strings.HasPrefix(relative, "docs/artifacts/changes/") {
		kind = "decision"
	}
	fileID := stableID(kind, relative, digest)
	language := languageFor(relative)
	builder.addNode(domain.MemoryNode{ID: fileID, Kind: kind, Path: relative, Name: filepath.Base(relative), Language: language, Digest: digest, Summary: summarize(data, 320)})
	switch language {
	case "go":
		builder.parseGo(relative, fileID, data)
	case "javascript", "typescript":
		builder.parseJSTS(relative, fileID, data, language)
	case "python":
		builder.parsePython(relative, fileID, data)
	}
	return graphSegment{Schema: domain.OrchestrationSchema, Path: relative, Digest: digest, ParserRevision: parserRevision, Nodes: builder.nodes, Edges: builder.edges, References: builder.references}
}

func loadSegment(common, relative, path, digest string) (graphSegment, error) {
	var segment graphSegment
	if err := load(common, relative, &segment, maxSegmentBytes); err != nil {
		return segment, err
	}
	if segment.Schema != domain.OrchestrationSchema || segment.Path != path || segment.Digest != digest || segment.ParserRevision != parserRevision || len(segment.Nodes) == 0 {
		return graphSegment{}, errors.New("memory segment is invalid")
	}
	return segment, nil
}

func (builder *graphBuilder) addSegment(segment graphSegment, commitID string) {
	fileID := ""
	for _, node := range segment.Nodes {
		builder.addNode(node)
		if node.Path == segment.Path && (node.Kind == "file" || node.Kind == "test" || node.Kind == "decision") {
			fileID = node.ID
		}
		if node.Kind == "function" || node.Kind == "type" || node.Kind == "class" || node.Kind == "variable" {
			if _, exists := builder.symbols[node.Name]; !exists {
				builder.symbols[node.Name] = node.ID
			}
		}
	}
	for _, edge := range segment.Edges {
		builder.addEdge(edge)
	}
	builder.references = append(builder.references, segment.References...)
	if fileID != "" {
		builder.addEdge(domain.MemoryEdge{From: commitID, To: fileID, Kind: "contains"})
	}
}

func (builder *graphBuilder) addNode(node domain.MemoryNode) {
	if node.ID != "" && !builder.nodeIDs[node.ID] {
		builder.nodeIDs[node.ID] = true
		builder.nodes = append(builder.nodes, node)
	}
}
func (builder *graphBuilder) addEdge(edge domain.MemoryEdge) {
	if edge.From != "" && edge.To != "" && edge.Kind != "" {
		builder.edges = append(builder.edges, edge)
	}
}

func (builder *graphBuilder) symbol(path, fileID, language, kind, name, summary string, line int) string {
	id := stableID("symbol", path, kind, name)
	builder.addNode(domain.MemoryNode{ID: id, Kind: kind, Path: path, Name: name, Language: language, Summary: summary, Line: line})
	builder.addEdge(domain.MemoryEdge{From: fileID, To: id, Kind: "defines"})
	if _, exists := builder.symbols[name]; !exists {
		builder.symbols[name] = id
	}
	return id
}

func (builder *graphBuilder) parseGo(path, fileID string, data []byte) {
	set := token.NewFileSet()
	file, err := parser.ParseFile(set, path, data, parser.SkipObjectResolution|parser.ParseComments)
	if err != nil {
		return
	}
	for _, declaration := range file.Decls {
		switch value := declaration.(type) {
		case *ast.FuncDecl:
			id := builder.symbol(path, fileID, "go", "function", value.Name.Name, value.Name.Name, set.Position(value.Pos()).Line)
			ast.Inspect(value.Body, func(node ast.Node) bool {
				if call, ok := node.(*ast.CallExpr); ok {
					if name := goCallName(call.Fun); name != "" {
						builder.references = append(builder.references, domain.MemoryEdge{From: id, To: name, Kind: "calls"})
					}
				}
				return true
			})
		case *ast.GenDecl:
			for _, specification := range value.Specs {
				switch spec := specification.(type) {
				case *ast.TypeSpec:
					builder.symbol(path, fileID, "go", "type", spec.Name.Name, spec.Name.Name, set.Position(spec.Pos()).Line)
				case *ast.ValueSpec:
					for _, name := range spec.Names {
						builder.symbol(path, fileID, "go", "variable", name.Name, name.Name, set.Position(name.Pos()).Line)
					}
				case *ast.ImportSpec:
					if spec.Path != nil {
						target := strings.Trim(spec.Path.Value, "\"")
						importID := stableID("module", target)
						builder.addNode(domain.MemoryNode{ID: importID, Kind: "module", Name: target, Language: "go", Summary: target})
						builder.addEdge(domain.MemoryEdge{From: fileID, To: importID, Kind: "imports"})
					}
				}
			}
		}
	}
}

func (builder *graphBuilder) parseJSTS(path, fileID string, data []byte, language string) {
	parsed := builder.parseTreeSitter(path, fileID, data, language)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if !parsed {
			for _, prefix := range []struct{ prefix, kind string }{{"function ", "function"}, {"class ", "class"}, {"interface ", "interface"}, {"type ", "type"}, {"const ", "variable"}, {"export function ", "function"}, {"export class ", "class"}} {
				if strings.HasPrefix(text, prefix.prefix) {
					if name := identifierAfter(text, prefix.prefix); name != "" {
						builder.symbol(path, fileID, language, prefix.kind, name, bounded(text, 240), line)
					}
					break
				}
			}
		}
		if target := quotedModule(text); target != "" && (strings.HasPrefix(text, "import ") || strings.Contains(text, "require(")) {
			id := stableID("module", target)
			builder.addNode(domain.MemoryNode{ID: id, Kind: "module", Name: target, Language: language, Summary: target})
			builder.addEdge(domain.MemoryEdge{From: fileID, To: id, Kind: "imports"})
		}
	}
}

func (builder *graphBuilder) parsePython(path, fileID string, data []byte) {
	parsed := builder.parseTreeSitter(path, fileID, data, "python")
	scanner := bufio.NewScanner(bytes.NewReader(data))
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if !parsed && (strings.HasPrefix(text, "def ") || strings.HasPrefix(text, "async def ")) {
			prefix := "def "
			if strings.HasPrefix(text, "async ") {
				prefix = "async def "
			}
			if name := identifierAfter(text, prefix); name != "" {
				builder.symbol(path, fileID, "python", "function", name, bounded(text, 240), line)
			}
		} else if !parsed && strings.HasPrefix(text, "class ") {
			if name := identifierAfter(text, "class "); name != "" {
				builder.symbol(path, fileID, "python", "class", name, bounded(text, 240), line)
			}
		} else if strings.HasPrefix(text, "import ") || strings.HasPrefix(text, "from ") {
			fields := strings.Fields(text)
			if len(fields) >= 2 {
				target := fields[1]
				id := stableID("module", target)
				builder.addNode(domain.MemoryNode{ID: id, Kind: "module", Name: target, Language: "python", Summary: target})
				builder.addEdge(domain.MemoryEdge{From: fileID, To: id, Kind: "imports"})
			}
		}
	}
}

// parseTreeSitter uses the pinned pure-Go Tree-sitter runtime and embedded
// upstream grammar tables for JavaScript, TypeScript/TSX, and Python. The
// conservative line scanners above remain only as an explicit degraded path
// when a malformed file cannot produce a complete syntax tree.
func (builder *graphBuilder) parseTreeSitter(path, fileID string, data []byte, language string) bool {
	entry := grammars.DetectLanguage(path)
	if entry == nil {
		return false
	}
	grammar := entry.Language()
	if grammar == nil {
		return false
	}
	parser := gotreesitter.NewParser(grammar)
	var tree *gotreesitter.Tree
	var err error
	if entry.TokenSourceFactory != nil {
		tree, err = parser.ParseWithTokenSource(data, entry.TokenSourceFactory(data, grammar))
	} else {
		tree, err = parser.Parse(data)
	}
	if err != nil || tree == nil || tree.RootNode() == nil {
		return false
	}
	defer tree.Release()
	if tree.RootNode().HasError() {
		return false
	}
	definitionIDs := make(map[uint32]string)
	for _, definition := range gotreesitter.ExtractDefinitionSpans(tree) {
		line := 1 + bytes.Count(data[:min(int(definition.StartByte), len(data))], []byte{'\n'})
		summary := definition.Name
		if int(definition.EndByte) <= len(data) && definition.EndByte > definition.StartByte {
			summary = bounded(strings.Join(strings.Fields(string(data[definition.StartByte:definition.EndByte])), " "), 240)
		}
		definitionIDs[definition.StartByte] = builder.symbol(path, fileID, language, definition.Kind, definition.Name, summary, line)
	}
	for _, call := range gotreesitter.ExtractCalls(tree) {
		definition, ok := gotreesitter.EnclosingDefinition(tree, call.StartByte)
		if !ok || call.Name == "" {
			continue
		}
		if source := definitionIDs[definition.StartByte]; source != "" {
			builder.references = append(builder.references, domain.MemoryEdge{From: source, To: call.Name, Kind: "calls"})
		}
	}
	return true
}

func (builder *graphBuilder) resolveReferences() {
	for _, edge := range builder.references {
		if target := builder.symbols[edge.To]; target != "" {
			edge.To = target
			builder.addEdge(edge)
		}
	}
}

func goCallName(expression ast.Expr) string {
	switch value := expression.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.SelectorExpr:
		return value.Sel.Name
	}
	return ""
}
func identifierAfter(value, prefix string) string {
	value = strings.TrimPrefix(value, prefix)
	end := 0
	for _, character := range value {
		if end == 0 && !(unicode.IsLetter(character) || character == '_' || character == '$') {
			return ""
		}
		if !(unicode.IsLetter(character) || unicode.IsDigit(character) || character == '_' || character == '$') {
			break
		}
		end += utf8.RuneLen(character)
	}
	return value[:end]
}
func quotedModule(value string) string {
	for _, quote := range []byte{'\'', '"'} {
		start := strings.IndexByte(value, quote)
		if start >= 0 {
			end := strings.IndexByte(value[start+1:], quote)
			if end >= 0 {
				return value[start+1 : start+1+end]
			}
		}
	}
	return ""
}

func (adapter Adapter) gitText(ctx context.Context, executable processadapter.Executable, root string, arguments []string, limit int) (string, error) {
	result, err := adapter.run(ctx, processadapter.Request{Executable: executable.Path, Arguments: arguments, Directory: root, Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: limit, Timeout: 30 * time.Second})
	if err != nil || result.ExitCode != 0 {
		return "", errors.New("Git command failed")
	}
	value := strings.TrimSpace(string(result.Stdout))
	if value == "" || strings.ContainsAny(value, "\x00\r\n") {
		return "", errors.New("Git output is invalid")
	}
	return value, nil
}
func splitNUL(data []byte) []string {
	values := []string{}
	for _, item := range bytes.Split(data, []byte{0}) {
		if len(item) > 0 && utf8.Valid(item) {
			values = append(values, filepath.ToSlash(string(item)))
		}
	}
	return values
}
func stableID(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return fmt.Sprintf("sha256:%x", sum)
}
func languageFor(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go":
		return "go"
	case ".js", ".jsx", ".mjs", ".cjs":
		return "javascript"
	case ".ts", ".tsx", ".mts", ".cts":
		return "typescript"
	case ".py":
		return "python"
	}
	return ""
}
func isTestPath(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	return strings.HasSuffix(base, "_test.go") || strings.HasSuffix(base, ".test.js") || strings.HasSuffix(base, ".test.ts") || strings.HasPrefix(base, "test_") || strings.Contains(filepath.ToSlash(path), "/tests/")
}
func excluded(relative string, patterns []string) bool {
	lower := strings.ToLower(filepath.ToSlash(relative))
	base := filepath.Base(lower)
	if strings.HasSuffix(base, ".transcript") || strings.HasSuffix(base, ".chatlog") || strings.HasPrefix(lower, "transcripts/") || strings.HasPrefix(lower, ".transcripts/") || strings.Contains(lower, "/transcripts/") || strings.Contains(lower, "/.transcripts/") {
		return true
	}
	for _, pattern := range patterns {
		if pattern == relative {
			return true
		}
		if strings.HasSuffix(pattern, "/**") && strings.HasPrefix(relative, strings.TrimSuffix(pattern, "**")) {
			return true
		}
		if strings.HasSuffix(pattern, ".*") && strings.HasPrefix(relative, strings.TrimSuffix(pattern, "*")) {
			return true
		}
	}
	return false
}
func secretPath(relative string) bool {
	base := strings.ToLower(filepath.Base(relative))
	return base == ".env" || strings.HasPrefix(base, ".env.") || strings.HasSuffix(base, ".pem") || strings.HasSuffix(base, ".key") || strings.HasSuffix(base, ".p12") || strings.Contains(base, "credential")
}
func likelySecret(data []byte) bool {
	upper := strings.ToUpper(string(data))
	for _, marker := range []string{"-----BEGIN PRIVATE KEY-----", "AWS_SECRET_ACCESS_KEY=", "OPENAI_API_KEY=", "ANTHROPIC_API_KEY=", "API_KEY=", "PASSWORD=", "SECRET=", "TOKEN="} {
		if strings.Contains(upper, marker) {
			return true
		}
	}
	return false
}
func safePath(root, relative string) (string, error) {
	relative = filepath.ToSlash(filepath.Clean(relative))
	if relative == "." || relative == ".." || filepath.IsAbs(relative) || strings.HasPrefix(relative, "../") || strings.Contains(relative, "\\") {
		return "", errors.New("unsafe repository path")
	}
	candidate := filepath.Join(root, filepath.FromSlash(relative))
	physical, err := filepath.EvalSymlinks(candidate)
	if err != nil || (physical != root && !strings.HasPrefix(physical, root+string(filepath.Separator))) {
		return "", errors.New("repository path escapes root")
	}
	return physical, nil
}
func summarize(data []byte, limit int) string {
	value := strings.Join(strings.Fields(string(data)), " ")
	return bounded(value, limit)
}
func bounded(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	for limit > 0 && !utf8.RuneStart(value[limit]) {
		limit--
	}
	return value[:limit]
}
func terms(value string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, term := range strings.Fields(strings.ToLower(value)) {
		term = strings.Trim(term, ".,:;()[]{}\"'")
		if len(term) >= 2 && !seen[term] {
			seen[term] = true
			result = append(result, term)
		}
	}
	return result
}

func buildLexicalIndex(nodes []domain.MemoryNode) map[string][]string {
	index := make(map[string][]string)
	for _, node := range nodes {
		for _, term := range terms(node.Path + " " + node.Name + " " + node.Summary + " " + node.Kind) {
			index[term] = append(index[term], node.ID)
		}
	}
	for term := range index {
		sort.Strings(index[term])
	}
	return index
}
func cosine(left, right []float32) float64 {
	var dot, leftNorm, rightNorm float64
	for index := range left {
		l, r := float64(left[index]), float64(right[index])
		dot += l * r
		leftNorm += l * l
		rightNorm += r * r
	}
	if leftNorm == 0 || rightNorm == 0 {
		return 0
	}
	return dot / (math.Sqrt(leftNorm) * math.Sqrt(rightNorm))
}

func save(common, relative string, value any, maximum int) error {
	root, err := memoryRoot(common)
	if err != nil {
		return err
	}
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := localfile.EnsureDirectory(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := localfile.EncodeJSON(value)
	if err != nil || len(data) > maximum {
		return errors.New("memory record is invalid or unbounded")
	}
	if _, err := os.Lstat(path); errors.Is(err, os.ErrNotExist) {
		return localfile.AtomicCreate(path, data, 0o600)
	} else if err != nil {
		return err
	}
	return localfile.AtomicReplace(path, data, 0o600)
}

func saveImmutable(common, relative string, value any, maximum int) error {
	root, err := memoryRoot(common)
	if err != nil {
		return err
	}
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := localfile.EnsureDirectory(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := localfile.EncodeJSON(value)
	if err != nil || len(data) > maximum {
		return errors.New("memory segment is invalid or unbounded")
	}
	return localfile.AtomicCreate(path, data, 0o600)
}

func syncLock(common string) (func(), error) {
	root, err := memoryRoot(common)
	if err != nil {
		return nil, err
	}
	directory := filepath.Join(root, "memory")
	if err := localfile.EnsureDirectory(directory, 0o700); err != nil {
		return nil, err
	}
	path := filepath.Join(directory, ".sync.lock")
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil, errors.New("memory sync is already running; retry after it completes")
		}
		return nil, err
	}
	_ = file.Close()
	return func() { _ = os.Remove(path) }, nil
}
func load(common, relative string, target any, maximum int) error {
	root, err := memoryRoot(common)
	if err != nil {
		return err
	}
	data, err := localfile.Read(filepath.Join(root, filepath.FromSlash(relative)), int64(maximum))
	if err != nil {
		return err
	}
	return localfile.DecodeJSON(data, target)
}
func memoryRoot(common string) (string, error) {
	if !filepath.IsAbs(common) {
		return "", errors.New("Git common directory must be absolute")
	}
	info, err := os.Lstat(filepath.Clean(common))
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return "", errors.New("Git common directory is unsafe")
	}
	return filepath.Join(filepath.Clean(common), "l7"), nil
}

func EncodeMatches(matches []domain.MemoryMatch) ([]byte, error) { return json.Marshal(matches) }

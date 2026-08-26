package render

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"
)

var errJSONBound = errors.New("semantic JSON bound exceeded")

type taxonomyDocument struct {
	SchemaVersion string     `json:"schema_version"`
	Taxonomies    []Taxonomy `json:"taxonomies"`
}

type obligationDocument struct {
	SchemaVersion string       `json:"schema_version"`
	Obligations   []Obligation `json:"obligations"`
}

type guardrailDocument struct {
	SchemaVersion string      `json:"schema_version"`
	Guardrails    []Guardrail `json:"guardrails"`
}

type knowledgeDocument struct {
	SchemaVersion string      `json:"schema_version"`
	Knowledge     []Knowledge `json:"knowledge"`
}

type workflowDocument struct {
	SchemaVersion string           `json:"schema_version"`
	Workflows     []Workflow       `json:"workflows"`
	Budgets       []Budget         `json:"budgets"`
	Delegations   []Delegation     `json:"delegations"`
	Outputs       []OutputContract `json:"outputs"`
}

type profileDocument struct {
	SchemaVersion string    `json:"schema_version"`
	Profiles      []Profile `json:"profiles"`
}

func Decode(files []SourceFile) (Bundle, []Diagnostic) {
	copied := make([]SourceFile, len(files))
	total := 0
	var diagnostics []Diagnostic
	for index, file := range files {
		data := append([]byte(nil), file.Data...)
		copied[index] = SourceFile{Path: file.Path, Data: data}
		if len(data) > MaxAggregateBytes-total {
			diagnostics = addDiagnostic(diagnostics, "SEM-100", file.Path, "semantic input bundle exceeds 2097152 bytes", "narrow the selected semantic bundle")
			break
		}
		total += len(data)
	}
	if len(diagnostics) != 0 {
		return Bundle{}, finishDiagnostics(diagnostics)
	}
	sort.Slice(copied, func(left, right int) bool { return copied[left].Path < copied[right].Path })

	seen := make(map[string]bool, len(copied))
	bundle := Bundle{}
	for _, file := range copied {
		if !safeSourcePath(file.Path) {
			diagnostics = addDiagnostic(diagnostics, "SEM-101", file.Path, "source path is not canonical repository-relative ASCII", "use one exact admitted semantic path")
			continue
		}
		if seen[file.Path] {
			diagnostics = addDiagnostic(diagnostics, "SEM-102", file.Path, "duplicate source path", "supply each source path exactly once")
			continue
		}
		seen[file.Path] = true
		if byteDiagnostics := validateSourceBytes(file.Path, file.Data); len(byteDiagnostics) != 0 {
			diagnostics = appendDiagnostics(diagnostics, byteDiagnostics...)
			continue
		}
		digest := sha256.Sum256(file.Data)
		bundle.SourceDigests = append(bundle.SourceDigests, Digest{Path: file.Path, SHA256: hex.EncodeToString(digest[:])})

		if file.Path == "semantic/workflows/reference/prompt.md.tmpl" {
			if bundle.Template != "" {
				diagnostics = addDiagnostic(diagnostics, "SEM-102", file.Path, "duplicate prompt template", "supply one exact prompt template")
			} else {
				bundle.Template = string(file.Data)
			}
			continue
		}
		if !strings.HasSuffix(file.Path, ".json") {
			diagnostics = addDiagnostic(diagnostics, "SEM-101", file.Path, "unsupported semantic source type", "use an exact JSON descriptor/source or prompt template")
			continue
		}
		if scanDiagnostics := scanJSON(file.Path, file.Data); len(scanDiagnostics) != 0 {
			diagnostics = appendDiagnostics(diagnostics, scanDiagnostics...)
			continue
		}

		switch file.Path {
		case "semantic/taxonomy/registry.json":
			var document taxonomyDocument
			if decodeDiagnostics := decodeExact(file.Path, file.Data, &document); len(decodeDiagnostics) != 0 {
				diagnostics = appendDiagnostics(diagnostics, decodeDiagnostics...)
			} else if document.SchemaVersion != "1.0.0" {
				diagnostics = addDiagnostic(diagnostics, "SEM-111", file.Path, "unsupported taxonomy document schema version", "use schema version 1.0.0")
			} else {
				bundle.Taxonomies = append(bundle.Taxonomies, document.Taxonomies...)
			}
		case "semantic/taxonomy/obligations.json":
			var document obligationDocument
			if decodeDiagnostics := decodeExact(file.Path, file.Data, &document); len(decodeDiagnostics) != 0 {
				diagnostics = appendDiagnostics(diagnostics, decodeDiagnostics...)
			} else if document.SchemaVersion != "1.0.0" {
				diagnostics = addDiagnostic(diagnostics, "SEM-111", file.Path, "unsupported obligation document schema version", "use schema version 1.0.0")
			} else {
				bundle.Obligations = append(bundle.Obligations, document.Obligations...)
			}
		case "semantic/taxonomy/guardrails.json":
			var document guardrailDocument
			if decodeDiagnostics := decodeExact(file.Path, file.Data, &document); len(decodeDiagnostics) != 0 {
				diagnostics = appendDiagnostics(diagnostics, decodeDiagnostics...)
			} else if document.SchemaVersion != "1.0.0" {
				diagnostics = addDiagnostic(diagnostics, "SEM-111", file.Path, "unsupported guardrail document schema version", "use schema version 1.0.0")
			} else {
				bundle.Guardrails = append(bundle.Guardrails, document.Guardrails...)
			}
		case "semantic/taxonomy/knowledge.json":
			var document knowledgeDocument
			if decodeDiagnostics := decodeExact(file.Path, file.Data, &document); len(decodeDiagnostics) != 0 {
				diagnostics = appendDiagnostics(diagnostics, decodeDiagnostics...)
			} else if document.SchemaVersion != "1.0.0" {
				diagnostics = addDiagnostic(diagnostics, "SEM-111", file.Path, "unsupported knowledge document schema version", "use schema version 1.0.0")
			} else {
				bundle.Knowledge = append(bundle.Knowledge, document.Knowledge...)
			}
		case "semantic/workflows/reference/contract.json":
			var document workflowDocument
			if decodeDiagnostics := decodeExact(file.Path, file.Data, &document); len(decodeDiagnostics) != 0 {
				diagnostics = appendDiagnostics(diagnostics, decodeDiagnostics...)
			} else if document.SchemaVersion != "1.0.0" {
				diagnostics = addDiagnostic(diagnostics, "SEM-111", file.Path, "unsupported workflow document schema version", "use schema version 1.0.0")
			} else {
				bundle.Workflows = append(bundle.Workflows, document.Workflows...)
				bundle.Budgets = append(bundle.Budgets, document.Budgets...)
				bundle.Delegations = append(bundle.Delegations, document.Delegations...)
				bundle.Outputs = append(bundle.Outputs, document.Outputs...)
			}
		case "semantic/profiles/generic.json", "semantic/profiles/feature-change.json", "semantic/profiles/behavior-preserving-refactor.json":
			var document profileDocument
			if decodeDiagnostics := decodeExact(file.Path, file.Data, &document); len(decodeDiagnostics) != 0 {
				diagnostics = appendDiagnostics(diagnostics, decodeDiagnostics...)
			} else if document.SchemaVersion != "1.0.0" {
				diagnostics = addDiagnostic(diagnostics, "SEM-111", file.Path, "unsupported profile document schema version", "use schema version 1.0.0")
			} else {
				bundle.Profiles = append(bundle.Profiles, document.Profiles...)
			}
		default:
			if strings.HasPrefix(file.Path, "schemas/semantic/") && strings.HasSuffix(file.Path, ".schema.json") {
				var descriptor SchemaDescriptor
				if decodeDiagnostics := decodeExact(file.Path, file.Data, &descriptor); len(decodeDiagnostics) != 0 {
					diagnostics = appendDiagnostics(diagnostics, decodeDiagnostics...)
				} else {
					bundle.Descriptors = append(bundle.Descriptors, descriptor)
				}
			} else {
				diagnostics = addDiagnostic(diagnostics, "SEM-101", file.Path, "unrecognized semantic source path", "use one path from the approved semantic bundle")
			}
		}
	}
	if len(diagnostics) != 0 {
		return Bundle{}, finishDiagnostics(diagnostics)
	}
	sort.Slice(bundle.Descriptors, func(left, right int) bool { return bundle.Descriptors[left].ID < bundle.Descriptors[right].ID })
	return bundle, nil
}

func validateSourceBytes(path string, data []byte) []Diagnostic {
	var diagnostics []Diagnostic
	if len(data) == 0 || len(data) > MaxFileBytes {
		diagnostics = addDiagnostic(diagnostics, "SEM-100", path, "source is empty or exceeds 262144 bytes", "supply one bounded nonempty source")
	}
	if !utf8.Valid(data) {
		diagnostics = addDiagnostic(diagnostics, "SEM-103", path, "source is not valid UTF-8", "encode the source as UTF-8")
	}
	if bytes.HasPrefix(data, []byte{0xef, 0xbb, 0xbf}) {
		diagnostics = addDiagnostic(diagnostics, "SEM-104", path, "UTF-8 BOM is forbidden", "remove the byte-order mark")
	}
	if len(data) == 0 || data[len(data)-1] != '\n' || (len(data) > 1 && data[len(data)-2] == '\n') {
		diagnostics = addDiagnostic(diagnostics, "SEM-105", path, "source must end in exactly one LF", "normalize the final newline")
	}
	for _, value := range data {
		if value == '\r' {
			diagnostics = addDiagnostic(diagnostics, "SEM-105", path, "carriage returns are forbidden", "normalize line endings to LF")
			break
		}
		if value < 0x20 && value != '\n' {
			diagnostics = addDiagnostic(diagnostics, "SEM-106", path, "raw control bytes are forbidden", "escape data and remove terminal controls")
			break
		}
		if value == 0x7f {
			diagnostics = addDiagnostic(diagnostics, "SEM-106", path, "DEL control byte is forbidden", "remove terminal controls")
			break
		}
	}
	return finishDiagnostics(diagnostics)
}

func scanJSON(path string, data []byte) []Diagnostic {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var diagnostics []Diagnostic
	if err := scanJSONValue(decoder, path, 0, &diagnostics); err != nil && !errors.Is(err, errJSONBound) {
		diagnostics = addDiagnostic(diagnostics, "SEM-107", path, err.Error(), "supply one strict bounded JSON value")
	}
	if len(diagnostics) == 0 {
		if token, err := decoder.Token(); err != io.EOF {
			if err == nil {
				diagnostics = addDiagnostic(diagnostics, "SEM-108", path, fmt.Sprintf("trailing JSON token %v", token), "remove trailing JSON data")
			} else {
				diagnostics = addDiagnostic(diagnostics, "SEM-107", path, err.Error(), "supply one strict bounded JSON value")
			}
		}
	}
	return finishDiagnostics(diagnostics)
}

func scanJSONValue(decoder *json.Decoder, subject string, depth int, diagnostics *[]Diagnostic) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	switch value := token.(type) {
	case json.Delim:
		if depth >= MaxJSONDepth {
			*diagnostics = addDiagnostic(*diagnostics, "SEM-109", subject, "JSON nesting exceeds 32 containers", "flatten the semantic input")
			return errJSONBound
		}
		switch value {
		case '{':
			seen := make(map[string]bool)
			fields := 0
			for decoder.More() {
				keyToken, keyErr := decoder.Token()
				if keyErr != nil {
					return keyErr
				}
				key, ok := keyToken.(string)
				if !ok {
					return fmt.Errorf("object key is not a string")
				}
				fields++
				if fields > MaxObjectFields {
					*diagnostics = addDiagnostic(*diagnostics, "SEM-109", subject, "object contains more than 128 fields", "split or narrow the semantic record")
				}
				if len(key) > MaxStringBytes {
					*diagnostics = addDiagnostic(*diagnostics, "SEM-109", subject, "object key exceeds 65536 bytes", "shorten the object key")
				}
				if seen[key] {
					*diagnostics = addDiagnostic(*diagnostics, "SEM-110", subject+":"+key, "duplicate object key", "retain one unambiguous field")
				}
				seen[key] = true
				if err := scanJSONValue(decoder, subject, depth+1, diagnostics); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
			return err
		case '[':
			items := 0
			for decoder.More() {
				items++
				if items > 2048 {
					*diagnostics = addDiagnostic(*diagnostics, "SEM-109", subject, "array contains more than 2048 items", "narrow the semantic collection")
				}
				if err := scanJSONValue(decoder, subject, depth+1, diagnostics); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
			return err
		default:
			return fmt.Errorf("unexpected JSON delimiter %q", value)
		}
	case string:
		if len(value) > MaxStringBytes {
			*diagnostics = addDiagnostic(*diagnostics, "SEM-109", subject, "string exceeds 65536 bytes", "shorten the semantic value")
		}
	case json.Number:
		text := value.String()
		if strings.ContainsAny(text, ".eE") {
			*diagnostics = addDiagnostic(*diagnostics, "SEM-112", subject, "fractional or exponent-form numbers are forbidden", "use a bounded base-10 integer")
		}
	}
	return nil
}

func decodeExact(path string, data []byte, target any) []Diagnostic {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return []Diagnostic{newDiagnostic("SEM-113", path, err.Error(), "match the exact typed semantic contract")}
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return []Diagnostic{newDiagnostic("SEM-108", path, "JSON contains a trailing value", "retain exactly one JSON value")}
	}
	return nil
}

func safeSourcePath(value string) bool {
	if value == "" || strings.HasPrefix(value, "/") || strings.Contains(value, "\\") || strings.Contains(value, "//") {
		return false
	}
	for _, part := range strings.Split(value, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	for _, character := range value {
		if character < 0x21 || character > 0x7e {
			return false
		}
	}
	return true
}

func newDiagnostic(rule, subject, message, next string) Diagnostic {
	return Diagnostic{Rule: boundedASCII(rule, 32), Subject: boundedASCII(subject, 160), Message: boundedASCII(message, MaxDiagnosticBytes), Next: boundedASCII(next, MaxDiagnosticBytes)}
}

func addDiagnostic(current []Diagnostic, rule, subject, message, next string) []Diagnostic {
	return appendDiagnostics(current, newDiagnostic(rule, subject, message, next))
}

func appendDiagnostics(current []Diagnostic, additional ...Diagnostic) []Diagnostic {
	for _, item := range additional {
		if len(current) < MaxDiagnostics {
			current = append(current, item)
		}
	}
	return current
}

func finishDiagnostics(diagnostics []Diagnostic) []Diagnostic {
	sort.Slice(diagnostics, func(left, right int) bool {
		if diagnostics[left].Rule != diagnostics[right].Rule {
			return diagnostics[left].Rule < diagnostics[right].Rule
		}
		if diagnostics[left].Subject != diagnostics[right].Subject {
			return diagnostics[left].Subject < diagnostics[right].Subject
		}
		return diagnostics[left].Message < diagnostics[right].Message
	})
	return diagnostics
}

func boundedASCII(value string, limit int) string {
	var result strings.Builder
	for _, character := range value {
		if character >= 0x20 && character <= 0x7e {
			result.WriteRune(character)
		} else {
			result.WriteByte('?')
		}
		if result.Len() >= limit {
			break
		}
	}
	return result.String()
}

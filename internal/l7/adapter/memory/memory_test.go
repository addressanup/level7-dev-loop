package memory

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/orchestrationconfig"
)

type fakeEmbedder struct{}

func (fakeEmbedder) Embed(_ context.Context, texts []string) (int, int, [][]float32, error) {
	vectors := make([][]float32, len(texts))
	for index := range vectors {
		vectors[index] = []float32{float32(index + 1), 1}
	}
	return 7, 2, vectors, nil
}

func TestIncrementalSyncReusesContentAddressedSegmentsAndRebuildsIndexes(t *testing.T) {
	root := gitRepository(t)
	write(t, root, "app.ts", "export function first() { return 1 }\n")
	git(t, root, "add", "app.ts")
	git(t, root, "-c", "user.name=L7", "-c", "user.email=l7@example.invalid", "commit", "-q", "-m", "app")
	common := strings.TrimSpace(git(t, root, "rev-parse", "--git-common-dir"))
	if !filepath.IsAbs(common) {
		common = filepath.Join(root, common)
	}
	adapter := New(nil)
	if _, err := adapter.Sync(context.Background(), root, common, orchestrationconfig.Default().Memory); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(common, "l7", "memory", "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest memoryManifest
	if json.Unmarshal(data, &manifest) != nil || len(manifest.Files) < 2 {
		t.Fatalf("invalid manifest: %s", data)
	}
	segmentPath := filepath.Join(common, "l7", "memory", "segments", strings.TrimPrefix(manifest.Files[0].Segment, "sha256:")+".json")
	before, err := os.Stat(segmentPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(common, "l7", "memory", "lexical.json"), []byte(`{"schema":1,"terms":{"broken":`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := adapter.Sync(context.Background(), root, common, orchestrationconfig.Default().Memory); err != nil {
		t.Fatal(err)
	}
	after, err := os.Stat(segmentPath)
	if err != nil || !after.ModTime().Equal(before.ModTime()) {
		t.Fatalf("unchanged segment was rewritten: before=%v after=%v err=%v", before.ModTime(), after.ModTime(), err)
	}
	if matches, err := adapter.Query(context.Background(), common, "first", 5); err != nil || len(matches) == 0 {
		t.Fatalf("rebuilt lexical index is unusable: matches=%#v err=%v", matches, err)
	}
}

func TestConcurrentSyncLockFailsClosedWithoutCorruptingState(t *testing.T) {
	root := gitRepository(t)
	common := strings.TrimSpace(git(t, root, "rev-parse", "--git-common-dir"))
	if !filepath.IsAbs(common) {
		common = filepath.Join(root, common)
	}
	release, err := syncLock(common)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if _, err := New(nil).Sync(context.Background(), root, common, orchestrationconfig.Default().Memory); err == nil || !strings.Contains(err.Error(), "already running") {
		t.Fatalf("concurrent sync did not fail closed: %v", err)
	}
}

func TestSyncBuildsLanguageGraphExcludesSecretsAndQueries(t *testing.T) {
	root := gitRepository(t)
	write(t, root, "main.go", "package sample\nimport \"fmt\"\nfunc Run() { fmt.Println(\"ok\") }\n")
	write(t, root, "web.ts", "import { x } from 'pkg'\nexport function render() { return x }\n")
	write(t, root, "worker.py", "from pkg import item\ndef execute():\n    return item\n")
	write(t, root, ".env", "OPENAI_API_KEY=secret\n")
	write(t, root, "session.transcript", "user and agent transcript\n")
	git(t, root, "add", "main.go", "web.ts", "worker.py", "session.transcript")
	git(t, root, "-c", "user.name=L7", "-c", "user.email=l7@example.invalid", "commit", "-q", "-m", "sources")
	common := strings.TrimSpace(git(t, root, "rev-parse", "--git-common-dir"))
	if !filepath.IsAbs(common) {
		common = filepath.Join(root, common)
	}
	adapter := NewWith(nil, nil, fakeEmbedder{}, func() time.Time { return time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC) })
	graph, err := adapter.Sync(context.Background(), root, common, orchestrationconfig.Default().Memory)
	if err != nil || graph.EmbeddingRevision != 7 || graph.EmbeddingDimension != 2 {
		t.Fatalf("graph=%#v err=%v", graph, err)
	}
	for _, node := range graph.Nodes {
		if node.Path == ".env" || node.Path == "session.transcript" || strings.Contains(node.Summary, "secret") {
			t.Fatalf("secret indexed: %#v", node)
		}
	}
	matches, err := adapter.Query(context.Background(), common, "render web", 10)
	if err != nil || len(matches) == 0 || matches[0].Node.Path != "web.ts" {
		t.Fatalf("matches=%#v err=%v", matches, err)
	}
}

func TestCorruptGraphFailsClosedAndResyncRecovers(t *testing.T) {
	root := gitRepository(t)
	common := strings.TrimSpace(git(t, root, "rev-parse", "--git-common-dir"))
	if !filepath.IsAbs(common) {
		common = filepath.Join(root, common)
	}
	adapter := New(nil)
	if _, err := adapter.Sync(context.Background(), root, common, orchestrationconfig.Default().Memory); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(common, "l7", "memory", "graph.json")
	if err := os.WriteFile(path, []byte(`{"schema":1,"unknown":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := adapter.Query(context.Background(), common, "readme", 10); err == nil {
		t.Fatal("corrupt graph was accepted")
	}
	if _, err := adapter.Sync(context.Background(), root, common, orchestrationconfig.Default().Memory); err != nil {
		t.Fatal(err)
	}
}

func gitRepository(t *testing.T) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	git(t, root, "init", "-q")
	write(t, root, "README.md", "memory fixture\n")
	git(t, root, "add", "README.md")
	git(t, root, "-c", "user.name=L7", "-c", "user.email=l7@example.invalid", "commit", "-q", "-m", "initial")
	return root
}
func write(t *testing.T, root, relative, value string) {
	t.Helper()
	path := filepath.Join(root, relative)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
		t.Fatal(err)
	}
}
func git(t *testing.T, root string, arguments ...string) string {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", root}, arguments...)...)
	command.Env = append(os.Environ(), "LC_ALL=C", "GIT_TERMINAL_PROMPT=0")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", arguments, err, output)
	}
	return string(output)
}

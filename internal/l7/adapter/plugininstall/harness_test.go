package plugininstall

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/pluginhost"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/pluginrelease"
)

// testEnv is one disposable installer world: its own HOME (where the fake
// hosts keep state), a fake bin directory on PATH, a fixture release server,
// and a Manager wired exactly like production except for the source URLs.
type testEnv struct {
	t       *testing.T
	home    string
	l7Home  string
	fakeBin string
	server  *releaseServer
	manager *Manager
	ctx     context.Context
}

func newTestEnv(t *testing.T, versions ...string) *testEnv {
	t.Helper()
	all := fixtures(t)
	if len(versions) == 0 {
		versions = []string{"0.9.0", "1.0.0", "1.0.1", "1.1.0-dev"}
	}
	var releases []*fixtureRelease
	for _, version := range versions {
		release, ok := all[version]
		if !ok {
			t.Fatalf("no fixture release %s", version)
		}
		releases = append(releases, release)
	}
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	env := &testEnv{t: t, home: filepath.Join(base, "home"), l7Home: filepath.Join(base, "home", ".local", "share", "level7"), fakeBin: filepath.Join(base, "bin"), ctx: context.Background()}
	for _, directory := range []string{env.home, env.fakeBin} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeFakes(t, env.fakeBin)
	t.Setenv("HOME", env.home)
	t.Setenv("PATH", env.fakeBin+string(os.PathListSeparator)+"/usr/bin"+string(os.PathListSeparator)+"/bin")
	for _, key := range []string{"CODEX_HOME", "CLAUDE_CONFIG_DIR", "XDG_CONFIG_HOME", "GH_CONFIG_DIR"} {
		t.Setenv(key, "")
		_ = os.Unsetenv(key)
	}
	env.server = newReleaseServer(t, releases...)
	env.manager = env.newManager(t)
	return env
}

func (env *testEnv) newManager(t *testing.T) *Manager {
	t.Helper()
	options := Options{
		Home:       env.l7Home,
		Platform:   Platform{OS: "darwin", Arch: fixtureArch},
		CLIVersion: "test-cli",
		Releases:   pluginrelease.NewClient(nil, env.server.source(), pluginrelease.NewVerifier(nil, nil), "test"),
		Now:        func() time.Time { return time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC) },
	}
	manager, err := New(options)
	if err != nil {
		t.Fatal(err)
	}
	return manager
}

// offline swaps in a manager with no release client, as L7_NETWORK=off does.
func (env *testEnv) offline() *Manager {
	manager := env.newManager(env.t)
	manager.options.Releases = nil
	return manager
}

func (env *testEnv) bootstrapCLI() {
	env.t.Helper()
	link := filepath.Join(env.home, ".local", "bin", "l7")
	if err := os.MkdirAll(filepath.Join(env.l7Home, "cli"), 0o755); err != nil {
		env.t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(env.l7Home, "cli", "link"), []byte(link+"\n"), 0o644); err != nil {
		env.t.Fatal(err)
	}
}

func (env *testEnv) hostState(host string) fakeHostState {
	env.t.Helper()
	var state fakeHostState
	data, err := os.ReadFile(filepath.Join(env.home, ".fake-"+host+".json"))
	if errors.Is(err, os.ErrNotExist) {
		return state
	}
	if err != nil || json.Unmarshal(data, &state) != nil {
		env.t.Fatalf("fake %s state unreadable: %v", host, err)
	}
	return state
}

func (env *testEnv) setHostState(host string, state fakeHostState) {
	env.t.Helper()
	if state.Marketplaces == nil {
		state.Marketplaces = []fakeMarketplace{}
	}
	if state.Plugins == nil {
		state.Plugins = []fakePlugin{}
	}
	data, _ := json.MarshalIndent(state, "", "  ")
	if err := os.WriteFile(filepath.Join(env.home, ".fake-"+host+".json"), data, 0o600); err != nil {
		env.t.Fatal(err)
	}
}

func (env *testEnv) control(host, content string) {
	env.t.Helper()
	path := filepath.Join(env.home, ".fake-"+host+".control")
	if content == "" {
		_ = os.Remove(path)
		return
	}
	if err := os.WriteFile(path, []byte(content+"\n"), 0o600); err != nil {
		env.t.Fatal(err)
	}
}

func (env *testEnv) hostLog(host string) string {
	data, _ := os.ReadFile(filepath.Join(env.home, ".fake-"+host+".log"))
	return string(data)
}

func (env *testEnv) installerState() State {
	env.t.Helper()
	state, _, err := loadState(env.manager.layout)
	if err != nil {
		env.t.Fatal(err)
	}
	return state
}

func (env *testEnv) registered(host, version string) bool {
	state := env.hostState(host)
	root := env.manager.layout.Root(version, host)
	marketplace := state.marketplace(pluginMarketplace)
	if marketplace == nil || marketplace.Root != root {
		return false
	}
	for _, plugin := range state.Plugins {
		if plugin.Name == pluginName && plugin.Marketplace == pluginMarketplace {
			return true
		}
	}
	return false
}

const (
	pluginMarketplace = "level7-engineering-v1"
	pluginName        = "level7-dev-loop"
)

func accepted(request Request) Request {
	request.AcceptUnsigned, request.AcceptUnqualified = true, true
	return request
}

func mustBlocked(t *testing.T, err error, contains string) *Blocked {
	t.Helper()
	var blocked *Blocked
	if !errors.As(err, &blocked) {
		t.Fatalf("expected Blocked containing %q, got %v", contains, err)
	}
	if !strings.Contains(blocked.Reason+" "+blocked.Next, contains) {
		t.Fatalf("blocked %q / %q does not mention %q", blocked.Reason, blocked.Next, contains)
	}
	return blocked
}

func mustFail(t *testing.T, err error, contains string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected failure containing %q, got success", contains)
	}
	var blocked *Blocked
	if errors.As(err, &blocked) {
		t.Fatalf("expected a hard failure containing %q, got Blocked: %v", contains, err)
	}
	if !strings.Contains(err.Error(), contains) {
		t.Fatalf("error %q does not mention %q", err.Error(), contains)
	}
}

func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

func countLines(value, prefix string) int {
	count := 0
	for _, line := range strings.Split(value, "\n") {
		if strings.HasPrefix(line, prefix) {
			count++
		}
	}
	return count
}

var _ = io.EOF
var _ pluginhost.Host

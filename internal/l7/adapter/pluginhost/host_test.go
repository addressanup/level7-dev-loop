package pluginhost

import (
	"context"
	"errors"
	"strings"
	"testing"

	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
)

type scripted struct {
	t        *testing.T
	calls    [][]string
	respond  func(arguments []string) (processadapter.Result, error)
	lastEnv  []string
	lastDir  string
	lastWait int
}

func (fake *scripted) run(_ context.Context, request processadapter.Request) (processadapter.Result, error) {
	fake.calls = append(fake.calls, request.Arguments)
	fake.lastEnv = request.Environment
	fake.lastDir = request.Directory
	if request.Executable != "/opt/host" || request.MaxOutputBytes < 1024 || request.Timeout <= 0 {
		fake.t.Fatalf("unsafe request %+v", request)
	}
	return fake.respond(request.Arguments)
}

func resolveHost(string) (processadapter.Executable, error) {
	return processadapter.Executable{Path: "/opt/host", Digest: strings.Repeat("a", 64)}, nil
}

func TestParseVersionsAndMinimums(t *testing.T) {
	for _, test := range []struct {
		host   Host
		output string
		ok     bool
	}{
		{Codex, "codex-cli 0.162.1\n", true},
		{Codex, "codex-cli 0.149.1\n", true},
		{Codex, "codex-cli 0.149.0\n", false},
		{Codex, "0.162.1\n", false},
		{Codex, "codex-cli garbage\n", false},
		{Claude, "2.1.291 (Claude Code)\n", true},
		{Claude, "2.1.241 (Claude Code)\n", true},
		{Claude, "2.1.240 (Claude Code)\n", false},
		{Claude, "v2.1.300\n", true},
		{Claude, "\n", false},
	} {
		fake := &scripted{t: t, respond: func([]string) (processadapter.Result, error) {
			return processadapter.Result{ExitCode: 0, Stdout: []byte(test.output)}, nil
		}}
		display, err := New(test.host, resolveHost, fake.run).Version(context.Background(), processadapter.Executable{Path: "/opt/host"})
		if (err == nil) != test.ok {
			t.Fatalf("%s %q: display=%q err=%v", test.host, test.output, display, err)
		}
		if test.ok && display != strings.TrimSpace(test.output) {
			t.Fatalf("display %q", display)
		}
	}
	failing := &scripted{t: t, respond: func([]string) (processadapter.Result, error) { return processadapter.Result{ExitCode: 1}, nil }}
	if _, err := New(Codex, resolveHost, failing.run).Version(context.Background(), processadapter.Executable{Path: "/opt/host"}); err == nil {
		t.Fatal("failing version probe accepted")
	}
	if _, err := New(Codex, func(string) (processadapter.Executable, error) {
		return processadapter.Executable{}, errors.New("missing")
	}, failing.run).Executable(); err == nil || !strings.Contains(err.Error(), "never installs hosts") {
		t.Fatalf("missing executable guidance: %v", err)
	}
	if _, err := Parse("browser"); err == nil {
		t.Fatal("unknown host parsed")
	}
}

func TestCommandsUseHostNativeVerbsScopeAndKeepData(t *testing.T) {
	ok := func([]string) (processadapter.Result, error) {
		return processadapter.Result{ExitCode: 0, Stdout: []byte(`{"ok":true}`)}, nil
	}
	executable := processadapter.Executable{Path: "/opt/host"}
	codex := &scripted{t: t, respond: ok}
	adapter := New(Codex, resolveHost, codex.run)
	ctx := context.Background()
	if err := errors.Join(
		adapter.MarketplaceAdd(ctx, executable, "/root/codex"),
		adapter.PluginAdd(ctx, executable, "level7-dev-loop", "level7-engineering-v1"),
		adapter.PluginRemove(ctx, executable, "level7-dev-loop", "level7-engineering-v1"),
		adapter.MarketplaceRemove(ctx, executable, "level7-engineering-v1"),
	); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"plugin marketplace add /root/codex --json",
		"plugin add level7-dev-loop@level7-engineering-v1 --json",
		"plugin remove level7-dev-loop@level7-engineering-v1 --json",
		"plugin marketplace remove level7-engineering-v1 --json",
	}
	for index, call := range codex.calls {
		if strings.Join(call, " ") != want[index] {
			t.Fatalf("codex call %d = %q, want %q", index, strings.Join(call, " "), want[index])
		}
	}
	claude := &scripted{t: t, respond: ok}
	adapter = New(Claude, resolveHost, claude.run)
	if err := errors.Join(
		adapter.MarketplaceAdd(ctx, executable, "/root/claude"),
		adapter.PluginAdd(ctx, executable, "level7-dev-loop", "level7-engineering-v1"),
		adapter.PluginRemove(ctx, executable, "level7-dev-loop", "level7-engineering-v1"),
		adapter.MarketplaceRemove(ctx, executable, "level7-engineering-v1"),
	); err != nil {
		t.Fatal(err)
	}
	want = []string{
		"plugin marketplace add /root/claude --scope user --json",
		"plugin install level7-dev-loop@level7-engineering-v1 --scope user --json",
		"plugin uninstall level7-dev-loop@level7-engineering-v1 --scope user --keep-data --json",
		"plugin marketplace remove level7-engineering-v1 --scope user --json",
	}
	for index, call := range claude.calls {
		if strings.Join(call, " ") != want[index] {
			t.Fatalf("claude call %d = %q, want %q", index, strings.Join(call, " "), want[index])
		}
	}
	if err := adapter.MarketplaceAdd(ctx, executable, "relative/root"); err == nil {
		t.Fatal("relative marketplace root accepted")
	}
	failing := &scripted{t: t, respond: func([]string) (processadapter.Result, error) {
		return processadapter.Result{ExitCode: 1, Stderr: []byte("marketplace \"x\" already exists\n")}, nil
	}}
	if err := New(Codex, resolveHost, failing.run).MarketplaceAdd(ctx, executable, "/root"); err == nil || !strings.Contains(err.Error(), "already exists") || !strings.Contains(err.Error(), "exited 1") {
		t.Fatalf("host diagnostic not propagated: %v", err)
	}
	for _, entry := range claude.lastEnv {
		if strings.HasPrefix(entry, "GH_TOKEN=") || strings.HasPrefix(entry, "ANTHROPIC_API_KEY=") {
			t.Fatalf("secret forwarded to host: %v", claude.lastEnv)
		}
	}
	if claude.lastDir != "/" {
		t.Fatalf("host commands must not run inside a repository: %q", claude.lastDir)
	}
}

func TestObserveParsesRealListingShapes(t *testing.T) {
	codexMarketplaces := `{"marketplaces":[{"name":"personal","root":"/Users/u"},{"name":"level7-engineering","root":"/Users/u/.codex/.tmp/marketplaces/level7-engineering","marketplaceSource":{"sourceType":"git","source":"https://github.com/addressanup/level7-dev-loop.git"}},{"name":"level7-engineering-v1","root":"/Users/u/.local/share/level7/plugin/versions/1.0.0/codex","marketplaceSource":{"sourceType":"local","source":"/Users/u/.local/share/level7/plugin/versions/1.0.0/codex"}}]}`
	codexPlugins := `{"installed":[{"pluginId":"level7-dev-loop@level7-engineering-v1","name":"level7-dev-loop","marketplaceName":"level7-engineering-v1","version":"1.0.0","installed":true,"enabled":true,"source":{"source":"local","path":"/Users/u/.local/share/level7/plugin/versions/1.0.0/codex"},"marketplaceSource":{"sourceType":"local","source":"/Users/u/.local/share/level7/plugin/versions/1.0.0/codex"},"installPolicy":"AVAILABLE","authPolicy":"ON_INSTALL"}],"available":[]}`
	codex := &scripted{t: t, respond: func(arguments []string) (processadapter.Result, error) {
		if strings.Join(arguments, " ") == "plugin marketplace list --json" {
			return processadapter.Result{ExitCode: 0, Stdout: []byte(codexMarketplaces)}, nil
		}
		if strings.Join(arguments, " ") == "plugin list --json --marketplace level7-engineering-v1" {
			return processadapter.Result{ExitCode: 0, Stdout: []byte(codexPlugins)}, nil
		}
		t.Fatalf("unexpected codex query %v", arguments)
		return processadapter.Result{}, nil
	}}
	observation, err := New(Codex, resolveHost, codex.run).Observe(context.Background(), processadapter.Executable{Path: "/opt/host"}, "level7-engineering-v1", "level7-dev-loop")
	if err != nil || !observation.MarketplaceRegistered || !observation.PluginInstalled || observation.PluginVersion != "1.0.0" || !observation.PluginEnabled || observation.MarketplaceRoot != "/Users/u/.local/share/level7/plugin/versions/1.0.0/codex" {
		t.Fatalf("codex observation: %+v %v", observation, err)
	}
	observation, err = New(Codex, resolveHost, codex.run).Observe(context.Background(), processadapter.Executable{Path: "/opt/host"}, "absent", "level7-dev-loop")
	if err != nil || observation.MarketplaceRegistered || observation.PluginInstalled || len(codex.calls) != 3 {
		t.Fatalf("codex absent marketplace must not query plugins: %+v %v calls=%d", observation, err, len(codex.calls))
	}

	claudeMarketplaces := `Loading marketplaces...
[{"name":"claude-plugins-official","source":"github","repo":"anthropics/claude-plugins-official","installLocation":"/Users/u/.claude/plugins/marketplaces/claude-plugins-official"},{"name":"level7-engineering-v1","source":"directory","path":"/Users/u/.local/share/level7/plugin/versions/1.0.0/claude","installLocation":"/Users/u/.local/share/level7/plugin/versions/1.0.0/claude"}]`
	claudePlugins := `[{"id":"level7-dev-loop@level7-engineering","version":"0.1.1","scope":"user","enabled":false,"installPath":"/Users/u/.claude/plugins/cache/level7-engineering/level7-dev-loop/0.1.1"},{"id":"level7-dev-loop@level7-engineering-v1","version":"1.0.0","scope":"user","enabled":true,"installPath":"/Users/u/.claude/plugins/cache/level7-engineering-v1/level7-dev-loop/1.0.0","readFromFolder":"/Users/u/.local/share/level7/plugin/versions/1.0.0/claude","folderVersion":"1.0.0","mcpServers":{"level7":{"type":"stdio","command":"${CLAUDE_PLUGIN_ROOT}/bin/l7","args":["mcp"]}}}]`
	claude := &scripted{t: t, respond: func(arguments []string) (processadapter.Result, error) {
		if strings.Join(arguments, " ") == "plugin marketplace list --json" {
			return processadapter.Result{ExitCode: 0, Stdout: []byte(claudeMarketplaces)}, nil
		}
		return processadapter.Result{ExitCode: 0, Stdout: []byte(claudePlugins)}, nil
	}}
	observation, err = New(Claude, resolveHost, claude.run).Observe(context.Background(), processadapter.Executable{Path: "/opt/host"}, "level7-engineering-v1", "level7-dev-loop")
	if err != nil || !observation.MarketplaceRegistered || !observation.PluginInstalled || observation.PluginVersion != "1.0.0" || observation.MarketplaceRoot != "/Users/u/.local/share/level7/plugin/versions/1.0.0/claude" || !strings.HasSuffix(observation.PluginPath, "/1.0.0") {
		t.Fatalf("claude observation: %+v %v", observation, err)
	}
	malformed := &scripted{t: t, respond: func([]string) (processadapter.Result, error) {
		return processadapter.Result{ExitCode: 0, Stdout: []byte("not json at all")}, nil
	}}
	if _, err := New(Claude, resolveHost, malformed.run).Observe(context.Background(), processadapter.Executable{Path: "/opt/host"}, "x", "y"); err == nil {
		t.Fatal("malformed listing accepted")
	}
	if Codex.ReloadAdvice() == "" || Claude.ReloadAdvice() == "" || !strings.Contains(Claude.ReloadAdvice(), "/reload-plugins") {
		t.Fatal("reload advice missing")
	}
}

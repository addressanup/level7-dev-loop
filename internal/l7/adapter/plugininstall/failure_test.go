package plugininstall

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func (env *testEnv) assertUntouched(t *testing.T, host string) {
	t.Helper()
	if exists(env.manager.layout.StateFile()) {
		t.Fatal("installer state was created by a failed operation")
	}
	if state := env.hostState(host); state.marketplace(pluginMarketplace) != nil || len(state.Plugins) != 0 {
		t.Fatalf("host was mutated by a failed operation: %+v", state)
	}
	if exists(env.manager.layout.Root("1.0.1", host)) || exists(env.manager.layout.Root("1.0.0", host)) {
		t.Fatal("a failed operation left an extraction in place")
	}
	if entries, _ := os.ReadDir(env.manager.layout.Staging()); len(entries) != 0 {
		t.Fatalf("staging was not cleaned: %d entries", len(entries))
	}
	if exists(env.manager.layout.JournalFile()) {
		t.Fatal("a failed verification left a journal")
	}
}

func TestVerificationFailuresFailClosedWithoutMutation(t *testing.T) {
	tests := []struct {
		name     string
		arrange  func(env *testEnv)
		contains string
	}{
		{"corrupt archive bytes", func(env *testEnv) { env.server.corrupt["level7-dev-loop-1.0.1-codex.zip"] = true }, "does not match"},
		{"corrupt checksums", func(env *testEnv) { env.server.corrupt["SHA256SUMS"] = true }, "SHA256SUMS digest"},
		{"corrupt manifest", func(env *testEnv) { env.server.corrupt["RELEASE-MANIFEST.json"] = true }, "RELEASE-MANIFEST.json"},
		{"missing archive asset", func(env *testEnv) { env.server.status["level7-dev-loop-1.0.1-codex.zip"] = 404 }, "status 404"},
		{"server error", func(env *testEnv) { env.server.status["RELEASE-MANIFEST.json"] = 500 }, "status 500"},
		{"no attestation for archive", func(env *testEnv) {
			env.server.noAttestation[fixtures(t)["1.0.1"].Digests["level7-dev-loop-1.0.1-codex.zip"]] = true
		}, "no GitHub attestation"},
		{"no attestation for manifest", func(env *testEnv) {
			env.server.noAttestation[fixtures(t)["1.0.1"].Digests["RELEASE-MANIFEST.json"]] = true
		}, "no GitHub attestation"},
		{"tampered certificate identity accepted by gh", func(env *testEnv) { env.server.tamper["level7-dev-loop-1.0.1-codex.zip"] = "sourceRepositoryURI" }, "trusted identity"},
		{"gh too old", func(env *testEnv) {
			if err := os.WriteFile(filepath.Join(env.home, ".fake-gh.control"), []byte("version=2.60.0\n"), 0o600); err != nil {
				t.Fatal(err)
			}
		}, "too old"},
		{"network unavailable", func(env *testEnv) { env.server.server.Close() }, "release API request failed"},
		{"redirect to foreign host", func(env *testEnv) { env.server.redirectTo = "https://198.51.100.7:1" }, "unexpected host"},
		{"redirect downgrades to http", func(env *testEnv) { env.server.redirectTo = "http://198.51.100.7:1" }, "must stay on HTTPS"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env := newTestEnv(t)
			test.arrange(env)
			_, err := env.manager.Install(env.ctx, accepted(Request{Host: "codex"}))
			mustFail(t, err, test.contains)
			env.assertUntouched(t, "codex")
		})
	}
}

func TestLauncherVerificationRejectsBrokenOrWrongExecutables(t *testing.T) {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	write := func(name, script string) string {
		t.Helper()
		path := filepath.Join(base, name)
		if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
		return path
	}
	good := write("good", "#!/bin/sh\nprintf '%s\\n' '{\"schema\":4,\"outcome\":\"PASS\",\"code\":\"L7-CLI-000\",\"command\":\"version\",\"state\":\"available\",\"version\":\"1.0.1\",\"message\":\"\",\"next\":\"\",\"details\":[]}'\n")
	if err := verifyLauncher(contextBackground(), good, "1.0.1"); err != nil {
		t.Fatal(err)
	}
	for name, script := range map[string]string{
		"crash":         "#!/bin/sh\nkill -9 $$\n",
		"wrong version": "#!/bin/sh\nprintf '%s\\n' '{\"outcome\":\"PASS\",\"command\":\"version\",\"version\":\"9.9.9\"}'\n",
		"not json":      "#!/bin/sh\necho hello\n",
		"failure":       "#!/bin/sh\necho boom >&2\nexit 3\n",
	} {
		if err := verifyLauncher(contextBackground(), write(strings.ReplaceAll(name, " ", "-"), script), "1.0.1"); err == nil {
			t.Fatalf("%s launcher was accepted", name)
		}
	}
	if err := verifyLauncher(contextBackground(), filepath.Join(base, "missing"), "1.0.1"); err == nil {
		t.Fatal("missing launcher was accepted")
	}
}

func TestMaliciousOrDriftedArchivesAreRejected(t *testing.T) {
	base := fixtures(t)["1.0.1"]
	mutations := map[string]func(entries []zipEntry) []zipEntry{
		"traversal": func(entries []zipEntry) []zipEntry {
			return append(entries, zipEntry{Name: "../escape", Mode: 0o644, Data: []byte("x")})
		},
		"absolute": func(entries []zipEntry) []zipEntry {
			return append(entries, zipEntry{Name: "/etc/passwd", Mode: 0o644, Data: []byte("x")})
		},
		"symlink": func(entries []zipEntry) []zipEntry {
			return append(entries, zipEntry{Name: "skills/link", Mode: os.ModeSymlink | 0o777, Data: []byte("/etc")})
		},
		"duplicate": func(entries []zipEntry) []zipEntry {
			return append(entries, zipEntry{Name: "LICENSE", Mode: 0o644, Data: []byte("dup")})
		},
		"unexpected inventory": func(entries []zipEntry) []zipEntry {
			return append(entries, zipEntry{Name: "hooks/evil.sh", Mode: 0o755, Data: []byte("#!/bin/sh\n")})
		},
		"missing mcp": func(entries []zipEntry) []zipEntry {
			return without(entries, ".mcp.json")
		},
		"executable mode on data": func(entries []zipEntry) []zipEntry {
			for index := range entries {
				if entries[index].Name == "README.md" {
					entries[index].Mode = 0o755
				}
			}
			return entries
		},
		"swapped host manifest": func(entries []zipEntry) []zipEntry {
			for index := range entries {
				if entries[index].Name == ".codex-plugin/plugin.json" {
					entries[index].Data = []byte(`{"name":"level7-dev-loop","version":"1.0.0"}`)
				}
			}
			return entries
		},
		"substituted binary": func(entries []zipEntry) []zipEntry {
			for index := range entries {
				if entries[index].Name == "bin/darwin-"+fixtureArch+"/l7-embed" {
					entries[index].Data = machOStub(fixtureArch, "substitute")
				}
			}
			return entries
		},
		"wrong architecture binary": func(entries []zipEntry) []zipEntry {
			for index := range entries {
				if entries[index].Name == "bin/darwin-amd64/l7-embed" {
					entries[index].Data = machOStub("arm64", "embed-1.0.1")
				}
			}
			return entries
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			release := *base
			release.Assets = map[string][]byte{}
			release.Digests = map[string]string{}
			for asset, data := range base.Assets {
				release.Assets[asset], release.Digests[asset] = data, base.Digests[asset]
			}
			archiveName := "level7-dev-loop-1.0.1-codex.zip"
			original := release.Assets[archiveName]
			entries := readZipEntries(t, original)
			mutated := writeZip(t, mutate(entries))
			release.Assets[archiveName] = mutated
			// The attacker also publishes consistent digests, so only package
			// validation stands between the archive and the host.
			release.Digests[archiveName] = digestHex(mutated)
			release.Assets["SHA256SUMS"] = bytes.ReplaceAll(release.Assets["SHA256SUMS"], []byte(base.Digests[archiveName]), []byte(release.Digests[archiveName]))
			release.Digests["SHA256SUMS"] = digestHex(release.Assets["SHA256SUMS"])
			manifest := bytes.ReplaceAll(release.Assets["RELEASE-MANIFEST.json"], []byte(base.Digests[archiveName]), []byte(release.Digests[archiveName]))
			manifest = bytes.ReplaceAll(manifest, []byte(base.Digests["SHA256SUMS"]), []byte(release.Digests["SHA256SUMS"]))
			manifest = bytes.ReplaceAll(manifest, []byte(`"size": `+itoa(len(original))+`,`), []byte(`"size": `+itoa(len(mutated))+`,`))
			release.Assets["RELEASE-MANIFEST.json"] = manifest
			release.Digests["RELEASE-MANIFEST.json"] = digestHex(manifest)
			env := newTestEnv(t, "1.0.0")
			env.server.releases = append(env.server.releases, &release)
			_, err := env.manager.Install(env.ctx, accepted(Request{Host: "codex"}))
			if err == nil {
				t.Fatal("mutated archive was installed")
			}
			env.assertUntouched(t, "codex")
		})
	}
}

func TestIncompatiblePlatformAndHostPrerequisitesBlock(t *testing.T) {
	env := newTestEnv(t)
	linux := env.newManager(t)
	linux.options.Platform = Platform{OS: "linux", Arch: "amd64"}
	_, err := linux.Install(env.ctx, accepted(Request{Host: "codex"}))
	mustBlocked(t, err, "macOS arm64 and amd64")

	env.control("codex", "version=codex-cli 0.120.0")
	_, err = env.manager.Install(env.ctx, accepted(Request{Host: "codex"}))
	mustBlocked(t, err, "older than the minimum")

	env.control("codex", "version=garbage")
	_, err = env.manager.Install(env.ctx, accepted(Request{Host: "codex"}))
	mustBlocked(t, err, "unrecognized")

	if err := os.Remove(filepath.Join(env.fakeBin, "claude")); err != nil {
		t.Fatal(err)
	}
	_, err = env.manager.Install(env.ctx, accepted(Request{Host: "claude"}))
	mustBlocked(t, err, "was not found on PATH")
	if err := os.Remove(filepath.Join(env.fakeBin, "gh")); err != nil {
		t.Fatal(err)
	}
	env.control("codex", "")
	_, err = env.manager.Install(env.ctx, accepted(Request{Host: "codex"}))
	mustFail(t, err, "GitHub CLI (gh) is required")
	env.assertUntouched(t, "codex")
	_, err = env.manager.Install(env.ctx, accepted(Request{Host: "browser"}))
	mustBlocked(t, err, "codex|claude")
	_, err = env.manager.Update(env.ctx, accepted(Request{Host: "codex", All: true}))
	mustBlocked(t, err, "exactly one")
	_, err = env.manager.Update(env.ctx, accepted(Request{Host: "codex"}))
	mustBlocked(t, err, "not managed")
	report, err := env.manager.Update(env.ctx, accepted(Request{All: true}))
	if err != nil || report.State != "nothing-managed" {
		t.Fatalf("update --all with nothing managed: %+v error=%v", report, err)
	}
}

func TestCorruptStateAndUnexpectedManagedPathsFailClosed(t *testing.T) {
	env := newTestEnv(t)
	if _, err := env.manager.Install(env.ctx, accepted(Request{Host: "codex", Version: "1.0.0"})); err != nil {
		t.Fatal(err)
	}
	statePath := env.manager.layout.StateFile()
	original, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{
		"garbage":        []byte("not json"),
		"tampered":       bytes.Replace(original, []byte(`"version": "1.0.0"`), []byte(`"version": "1.0.9"`), 1),
		"foreign schema": bytes.Replace(original, []byte(`"schema": 1`), []byte(`"schema": 7`), 1),
	} {
		t.Run(name, func(t *testing.T) {
			if err := os.WriteFile(statePath, data, 0o600); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = os.WriteFile(statePath, original, 0o600) }()
			_, err := env.manager.Status(env.ctx, Request{})
			mustFail(t, err, "move it away")
			_, err = env.manager.Update(env.ctx, accepted(Request{Host: "codex", Unpin: true}))
			mustFail(t, err, "move it away")
			if !env.registered("codex", "1.0.0") {
				t.Fatal("corrupt state led to a host mutation")
			}
		})
	}

	t.Run("modified installation", func(t *testing.T) {
		root := env.installerState().Hosts["codex"].Root
		target := filepath.Join(root, "skills", "l7-next", "SKILL.md")
		if err := os.WriteFile(target, []byte("tampered\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		status, err := env.manager.Status(env.ctx, Request{})
		if err != nil || status.Outcome != OutcomeFailed || !strings.Contains(status.Status.Hosts[1].Verification.Tree, "FAILED") {
			t.Fatalf("status did not report the tampered tree: %+v error=%v", status, err)
		}
		_, err = env.manager.Install(env.ctx, accepted(Request{Host: "claude", Version: "1.0.0"}))
		if err != nil {
			t.Fatal(err)
		}
		_, err = env.offline().Uninstall(env.ctx, Request{Host: "claude", Yes: true})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := env.manager.Update(env.ctx, accepted(Request{Host: "codex", Unpin: true})); err != nil {
			t.Fatal(err)
		}
		if !exists(root) {
			t.Fatal("a tampered retained tree was deleted instead of kept for inspection")
		}
		_, err = env.offline().Rollback(env.ctx, Request{Host: "codex"})
		mustFail(t, err, "failed verification")
	})

	t.Run("unmanaged directory at a version root", func(t *testing.T) {
		fresh := newTestEnv(t)
		root := fresh.manager.layout.Root("1.0.1", "codex")
		if err := os.MkdirAll(filepath.Join(root, "skills"), 0o755); err != nil {
			t.Fatal(err)
		}
		_, err := fresh.manager.Install(fresh.ctx, accepted(Request{Host: "codex"}))
		mustFail(t, err, "not an installer-managed extraction")
		if !exists(filepath.Join(root, "skills")) {
			t.Fatal("foreign directory was removed")
		}
	})

	t.Run("orphan receipt is redone", func(t *testing.T) {
		fresh := newTestEnv(t)
		if err := os.MkdirAll(fresh.manager.layout.VersionDir("1.0.1"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fresh.manager.layout.Receipt("1.0.1", "codex"), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := fresh.manager.Install(fresh.ctx, accepted(Request{Host: "codex"})); err != nil {
			t.Fatal(err)
		}
	})
}

func TestCLILinkIsNeverForcedOverForeignFiles(t *testing.T) {
	env := newTestEnv(t)
	env.bootstrapCLI()
	link := filepath.Join(env.home, ".local", "bin", "l7")
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(link, []byte("#!/bin/sh\necho someone else\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	report, err := env.manager.Install(env.ctx, accepted(Request{Host: "codex"}))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(report.Notes, " "), "refusing to replace") {
		t.Fatalf("foreign link file was not reported: %+v", report.Notes)
	}
	data, _ := os.ReadFile(link)
	if !strings.Contains(string(data), "someone else") {
		t.Fatal("foreign file at the link path was replaced")
	}
	if !env.registered("codex", "1.0.1") {
		t.Fatal("the plugin installation itself must still succeed")
	}
}

type zipEntry struct {
	Name string
	Mode os.FileMode
	Data []byte
}

func readZipEntries(t *testing.T, data []byte) []zipEntry {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	entries := []zipEntry{}
	for _, file := range reader.File {
		input, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		content, err := io.ReadAll(input)
		_ = input.Close()
		if err != nil {
			t.Fatal(err)
		}
		entries = append(entries, zipEntry{Name: file.Name, Mode: file.Mode(), Data: content})
	}
	return entries
}

func writeZip(t *testing.T, entries []zipEntry) []byte {
	t.Helper()
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	for _, entry := range entries {
		header := &zip.FileHeader{Name: entry.Name, Method: zip.Store}
		header.SetMode(entry.Mode)
		file, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.Write(entry.Data); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func without(entries []zipEntry, name string) []zipEntry {
	kept := entries[:0]
	for _, entry := range entries {
		if entry.Name != name {
			kept = append(kept, entry)
		}
	}
	return kept
}

func itoa(value int) string { return strconv.Itoa(value) }

func contextBackground() context.Context { return context.Background() }

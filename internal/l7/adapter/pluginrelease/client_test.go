package pluginrelease

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
)

func loopbackSource(server *httptest.Server) Source {
	return Source{APIBase: server.URL, DownloadBase: server.URL, AllowedHosts: []string{"127.0.0.1"}, InsecureLoopback: true}
}

func TestDefaultSourceIsPinnedToGitHubOverHTTPS(t *testing.T) {
	source := DefaultSource()
	if source.APIBase != "https://api.github.com" || source.DownloadBase != "https://github.com" || source.InsecureLoopback {
		t.Fatalf("default source drifted: %+v", source)
	}
	for _, target := range []string{"http://github.com/x", "https://evil.example/x", "https://github.com.evil.example/x", "http://127.0.0.1/x"} {
		if err := source.allow(mustURL(t, target)); err == nil {
			t.Fatalf("%s allowed", target)
		}
	}
	for _, target := range []string{"https://github.com/x", "https://objects.githubusercontent.com/x", "https://release-assets.githubusercontent.com/x", "https://api.github.com/x"} {
		if err := source.allow(mustURL(t, target)); err != nil {
			t.Fatalf("%s refused: %v", target, err)
		}
	}
}

func TestDownloadsNeverCarryCredentialsAndStayBounded(t *testing.T) {
	var mu sync.Mutex
	seen := []string{}
	payload := bytes.Repeat([]byte("x"), 1024)
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		mu.Lock()
		seen = append(seen, request.URL.Path)
		mu.Unlock()
		if request.Header.Get("Authorization") != "" || request.Header.Get("Cookie") != "" {
			http.Error(response, "credentials leaked", http.StatusBadRequest)
			return
		}
		switch request.URL.Path {
		case "/addressanup/level7-dev-loop/releases/download/v1.0.0/asset.bin":
			http.Redirect(response, request, server.URL+"/moved/asset.bin", http.StatusFound)
		case "/moved/asset.bin":
			_, _ = response.Write(payload)
		case "/addressanup/level7-dev-loop/releases/download/v1.0.0/short.bin":
			_, _ = response.Write(payload[:10])
		case "/addressanup/level7-dev-loop/releases/download/v1.0.0/long.bin":
			_, _ = response.Write(append(payload, 'y'))
		case "/addressanup/level7-dev-loop/releases/download/v1.0.0/loop.bin":
			http.Redirect(response, request, request.URL.String(), http.StatusFound)
		default:
			http.NotFound(response, request)
		}
	}))
	t.Cleanup(server.Close)
	client := NewClient(nil, loopbackSource(server), Verifier{}, "test")
	directory := t.TempDir()
	path, digest, err := client.Download(context.Background(), "v1.0.0", "asset.bin", int64(len(payload)), directory)
	if err != nil || digest != digestOf(payload) || filepath.Base(path) != "asset.bin" {
		t.Fatalf("redirected download: path=%q digest=%q err=%v", path, digest, err)
	}
	if _, _, err := client.Download(context.Background(), "v1.0.0", "asset.bin", int64(len(payload)), directory); err == nil {
		t.Fatal("download overwrote an existing staged file")
	}
	for _, asset := range []string{"short.bin", "long.bin", "loop.bin", "missing.bin"} {
		if _, _, err := client.Download(context.Background(), "v1.0.0", asset, int64(len(payload)), directory); err == nil {
			t.Fatalf("%s download succeeded", asset)
		}
		if _, statErr := os.Lstat(filepath.Join(directory, asset)); statErr == nil {
			t.Fatalf("%s left a partial file", asset)
		}
	}
	if _, _, err := client.Download(context.Background(), "v1.0.0", "../escape", 10, directory); err == nil {
		t.Fatal("asset path traversal accepted")
	}
	if _, _, err := client.Download(context.Background(), "v1.0.0", "huge.bin", maxArchiveBytes+1, directory); err == nil {
		t.Fatal("unbounded size accepted")
	}
}

func TestAPIResponsesAreValidatedAndRateLimitsExplained(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Header.Get("User-Agent") == "" || request.Header.Get("Accept") != "application/vnd.github+json" {
			http.Error(response, "bad headers", http.StatusBadRequest)
			return
		}
		switch {
		case request.URL.Path == "/repos/addressanup/level7-dev-loop/releases":
			_, _ = response.Write([]byte(`[{"tag_name":"v1.0.0","draft":false,"prerelease":false,"assets":[{"name":"RELEASE-MANIFEST.json","size":3},{"name":"SHA256SUMS","size":4}]},{"tag_name":"junk","assets":[]},{"tag_name":"v2.0.0","draft":true,"assets":[{"name":"RELEASE-MANIFEST.json","size":3}]}]`))
		case strings.HasPrefix(request.URL.Path, "/repos/addressanup/level7-dev-loop/releases/tags/v1.0.0"):
			_, _ = response.Write([]byte(`{"tag_name":"v1.0.0","draft":false,"prerelease":false,"assets":[{"name":"RELEASE-MANIFEST.json","size":3}]}`))
		case strings.HasPrefix(request.URL.Path, "/repos/addressanup/level7-dev-loop/releases/tags/v3.0.0"):
			http.Error(response, `{"message":"API rate limit exceeded"}`, http.StatusForbidden)
		case strings.HasPrefix(request.URL.Path, "/repos/addressanup/level7-dev-loop/attestations/sha256:"):
			_, _ = response.Write([]byte(`{"attestations":[{"bundle":{"mediaType":"application/vnd.dev.sigstore.bundle.v0.3+json","a":1},"repository_id":1},{"bundle":{"b":[1,2]}}]}`))
		default:
			_, _ = response.Write([]byte(`{"tag_name":"v1.0.0","tag_name":"v1.0.0"}`))
		}
	}))
	t.Cleanup(server.Close)
	client := NewClient(nil, loopbackSource(server), Verifier{}, "1.0.0-dev")
	releases, err := client.Candidates(context.Background())
	if err != nil || len(releases) != 1 || releases[0].Tag != "v1.0.0" || releases[0].Assets["SHA256SUMS"] != 4 {
		t.Fatalf("candidates: %+v %v", releases, err)
	}
	if release, err := client.ReleaseByTag(context.Background(), "v1.0.0"); err != nil || release.Tag != "v1.0.0" {
		t.Fatalf("release by tag: %+v %v", release, err)
	}
	if _, err := client.ReleaseByTag(context.Background(), "v3.0.0"); err == nil || !strings.Contains(err.Error(), "rate limit") {
		t.Fatalf("rate limit explanation missing: %v", err)
	}
	if _, err := client.ReleaseByTag(context.Background(), "v9.9.9"); err == nil {
		t.Fatal("duplicate-key JSON accepted")
	}
	if _, err := client.ReleaseByTag(context.Background(), "main"); err == nil {
		t.Fatal("non-version tag accepted")
	}
	bundlePath := filepath.Join(t.TempDir(), "bundles.jsonl")
	count, err := client.AttestationBundles(context.Background(), strings.Repeat("a", 64), bundlePath)
	if err != nil || count != 2 {
		t.Fatalf("bundles: %d %v", count, err)
	}
	data, _ := os.ReadFile(bundlePath)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 || !json.Valid([]byte(lines[0])) || !json.Valid([]byte(lines[1])) || strings.Contains(lines[0], "\n") {
		t.Fatalf("bundle JSONL: %q", data)
	}
}

func TestMatchVerificationRequiresExactIdentity(t *testing.T) {
	output := testdata(t, "v1.0.0-SHA256SUMS-gh-verify.json")
	workflow, _ := WorkflowForManifest("RELEASE-MANIFEST.json")
	expect := Expectation{Workflow: workflow, SourceCommit: "ed1457f2c54e2308f757d7ea3bf4fcfa113575ec", SubjectName: "SHA256SUMS", SubjectSHA256: "633b7d9ffdbf7a704cccb1427f9beeeb31c41d37fcc9e2ef72951dccf32f309f"}
	evidence, err := matchVerification(output, expect, "/usr/local/bin/gh")
	if err != nil || evidence.SourceCommit != expect.SourceCommit || !strings.Contains(evidence.SignerIdentity, "release.yml@refs/heads/main") || evidence.Timestamp == "" || evidence.RunInvocation == "" {
		t.Fatalf("evidence: %+v error=%v", evidence, err)
	}
	mutations := map[string]func() ([]byte, Expectation){
		"other commit": func() ([]byte, Expectation) {
			changed := expect
			changed.SourceCommit = "bb1457f2c54e2308f757d7ea3bf4fcfa113575ec"
			return output, changed
		},
		"other subject digest": func() ([]byte, Expectation) {
			changed := expect
			changed.SubjectSHA256 = strings.Repeat("0", 64)
			return output, changed
		},
		"other subject name": func() ([]byte, Expectation) {
			changed := expect
			changed.SubjectName = "level7-dev-loop-1.0.0-evil.zip"
			return output, changed
		},
		"other workflow": func() ([]byte, Expectation) {
			changed := expect
			changed.Workflow, _ = WorkflowForManifest("UNSIGNED-PRERELEASE-MANIFEST.json")
			return output, changed
		},
		"self-hosted runner": func() ([]byte, Expectation) {
			return bytes.Replace(output, []byte(`"runnerEnvironment": "github-hosted"`), []byte(`"runnerEnvironment": "self-hosted"`), 1), expect
		},
		"other repository": func() ([]byte, Expectation) {
			return bytes.Replace(output, []byte(`"sourceRepositoryURI": "https://github.com/addressanup/level7-dev-loop"`), []byte(`"sourceRepositoryURI": "https://github.com/evil/level7-dev-loop"`), 1), expect
		},
		"other ref": func() ([]byte, Expectation) {
			return bytes.Replace(output, []byte(`"sourceRepositoryRef": "refs/heads/main"`), []byte(`"sourceRepositoryRef": "refs/heads/feature"`), 1), expect
		},
		"other issuer": func() ([]byte, Expectation) {
			return bytes.Replace(output, []byte(`"issuer": "https://token.actions.githubusercontent.com"`), []byte(`"issuer": "https://accounts.example"`), 1), expect
		},
		"other predicate": func() ([]byte, Expectation) {
			return bytes.Replace(output, []byte(`"predicateType": "https://slsa.dev/provenance/v1"`), []byte(`"predicateType": "https://example/other"`), 1), expect
		},
		"no timestamps": func() ([]byte, Expectation) {
			return bytes.Replace(output, []byte(`"verifiedTimestamps": [`), []byte(`"verifiedTimestamps": [], "ignored": [`), 1), expect
		},
		"empty result": func() ([]byte, Expectation) { return []byte(`[]`), expect },
		"malformed":    func() ([]byte, Expectation) { return []byte(`{"verificationResult": `), expect },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			data, expectation := mutate()
			if _, err := matchVerification(data, expectation, "gh"); err == nil {
				t.Fatal("mismatched verification accepted")
			}
		})
	}
}

func TestVerifierPrerequisiteAndInvocation(t *testing.T) {
	calls := [][]string{}
	run := func(_ context.Context, request processadapter.Request) (processadapter.Result, error) {
		calls = append(calls, request.Arguments)
		for _, entry := range request.Environment {
			if strings.HasPrefix(entry, "GH_TOKEN=") || strings.HasPrefix(entry, "GITHUB_TOKEN=") {
				t.Fatalf("token forwarded to gh: %v", request.Environment)
			}
		}
		if len(request.Arguments) == 1 && request.Arguments[0] == "--version" {
			return processadapter.Result{ExitCode: 0, Stdout: []byte("gh version 2.101.0 (2026-09-15)\nhttps://github.com/cli/cli/releases/tag/v2.101.0\n")}, nil
		}
		return processadapter.Result{ExitCode: 0, Stdout: testdata(t, "v1.0.0-SHA256SUMS-gh-verify.json")}, nil
	}
	t.Setenv("GH_TOKEN", "secret-must-not-cross")
	resolve := func(name string) (processadapter.Executable, error) {
		return processadapter.Executable{Path: "/usr/local/bin/gh", Digest: strings.Repeat("a", 64)}, nil
	}
	verifier := NewVerifier(resolve, run)
	gh, version, err := verifier.Prerequisite(context.Background())
	if err != nil || version != "2.101.0" || gh.Path != "/usr/local/bin/gh" {
		t.Fatalf("prerequisite: %q %v", version, err)
	}
	directory := t.TempDir()
	artifact := filepath.Join(directory, "SHA256SUMS")
	bundle := filepath.Join(directory, "bundle.jsonl")
	if err := os.WriteFile(artifact, testdata(t, "v1.0.0-SHA256SUMS"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bundle, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	workflow, _ := WorkflowForManifest("RELEASE-MANIFEST.json")
	expect := Expectation{Workflow: workflow, SourceCommit: "ed1457f2c54e2308f757d7ea3bf4fcfa113575ec", SubjectName: "SHA256SUMS", SubjectSHA256: "633b7d9ffdbf7a704cccb1427f9beeeb31c41d37fcc9e2ef72951dccf32f309f"}
	if _, err := verifier.Verify(context.Background(), gh, artifact, bundle, expect); err != nil {
		t.Fatal(err)
	}
	arguments := strings.Join(calls[len(calls)-1], " ")
	for _, required := range []string{"attestation verify " + artifact, "--bundle " + bundle, "--repo addressanup/level7-dev-loop", "--signer-workflow addressanup/level7-dev-loop/.github/workflows/release.yml", "--source-ref refs/heads/main", "--source-digest ed1457f2c54e2308f757d7ea3bf4fcfa113575ec", "--predicate-type https://slsa.dev/provenance/v1", "--deny-self-hosted-runners", "--format json"} {
		if !strings.Contains(arguments, required) {
			t.Fatalf("gh invocation %q lacks %q", arguments, required)
		}
	}
	if _, err := verifier.Verify(context.Background(), gh, artifact, bundle, Expectation{}); err == nil {
		t.Fatal("incomplete expectation accepted")
	}
	failing := NewVerifier(resolve, func(_ context.Context, request processadapter.Request) (processadapter.Result, error) {
		if len(request.Arguments) == 1 {
			return processadapter.Result{ExitCode: 0, Stdout: []byte("gh version 2.101.0\n")}, nil
		}
		return processadapter.Result{ExitCode: 1, Stderr: []byte("✗ Verification failed")}, nil
	})
	if _, err := failing.Verify(context.Background(), gh, artifact, bundle, expect); err == nil || !strings.Contains(err.Error(), "rejected SHA256SUMS") {
		t.Fatalf("gh rejection not propagated: %v", err)
	}
	for name, output := range map[string]string{"too old": "gh version 2.66.1 (2025-01-01)\n", "garbage": "something else\n", "no version": "gh version\n"} {
		old := NewVerifier(resolve, func(context.Context, processadapter.Request) (processadapter.Result, error) {
			return processadapter.Result{ExitCode: 0, Stdout: []byte(output)}, nil
		})
		if _, _, err := old.Prerequisite(context.Background()); err == nil {
			t.Fatalf("%s gh accepted", name)
		}
	}
	missing := NewVerifier(func(string) (processadapter.Executable, error) {
		return processadapter.Executable{}, errors.New("missing")
	}, run)
	if _, _, err := missing.Prerequisite(context.Background()); err == nil || !strings.Contains(err.Error(), "no GitHub login is needed") {
		t.Fatalf("missing gh guidance: %v", err)
	}
}

func TestResolveBindsManifestChecksumsAndArchiveToOneCommit(t *testing.T) {
	manifest := testdata(t, "v1.0.0-RELEASE-MANIFEST.json")
	sums := testdata(t, "v1.0.0-SHA256SUMS")
	verified := map[string]bool{}
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch {
		case strings.HasSuffix(request.URL.Path, "/RELEASE-MANIFEST.json"):
			_, _ = response.Write(manifest)
		case strings.HasSuffix(request.URL.Path, "/SHA256SUMS"):
			_, _ = response.Write(sums)
		case strings.Contains(request.URL.Path, "/attestations/sha256:"):
			_, _ = response.Write([]byte(`{"attestations":[{"bundle":{"x":1}}]}`))
		default:
			http.NotFound(response, request)
		}
	}))
	t.Cleanup(server.Close)
	run := func(_ context.Context, request processadapter.Request) (processadapter.Result, error) {
		if len(request.Arguments) == 1 {
			return processadapter.Result{ExitCode: 0, Stdout: []byte("gh version 2.101.0\n")}, nil
		}
		subject := filepath.Base(request.Arguments[2])
		verified[subject] = true
		digest, _, err := FileDigest(request.Arguments[2], maxArchiveBytes)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(strings.Join(request.Arguments, " "), "--source-digest ed1457f2c54e2308f757d7ea3bf4fcfa113575ec") {
			t.Fatalf("unexpected gh arguments %v", request.Arguments)
		}
		result := bytes.Replace(testdata(t, "v1.0.0-SHA256SUMS-gh-verify.json"), []byte(`"name": "SHA256SUMS"`), []byte(`"name": "`+subject+`"`), 1)
		result = bytes.Replace(result, []byte(`633b7d9ffdbf7a704cccb1427f9beeeb31c41d37fcc9e2ef72951dccf32f309f`), []byte(digest), 1)
		return processadapter.Result{ExitCode: 0, Stdout: result}, nil
	}
	resolve := func(string) (processadapter.Executable, error) {
		return processadapter.Executable{Path: "/usr/local/bin/gh"}, nil
	}
	client := NewClient(nil, loopbackSource(server), NewVerifier(resolve, run), "test")
	release := Release{Tag: "v1.0.0", Version: mustVersion("1.0.0"), Assets: map[string]int64{"RELEASE-MANIFEST.json": int64(len(manifest)), "SHA256SUMS": int64(len(sums)), "level7-dev-loop-1.0.0-codex.zip": 22735431, "level7-dev-loop-1.0.0-claude.zip": 22735243}}
	release.Manifest, _ = WorkflowForManifest("RELEASE-MANIFEST.json")
	stage := t.TempDir()
	resolved, err := client.Resolve(context.Background(), release, stage)
	if err != nil {
		t.Fatal(err)
	}
	if !verified["RELEASE-MANIFEST.json"] || !verified["SHA256SUMS"] || len(resolved.Evidence) != 2 || resolved.GHVersion != "2.101.0" {
		t.Fatalf("resolution evidence: %+v verified=%v", resolved.Evidence, verified)
	}
	if _, _, err := client.FetchArchive(context.Background(), resolved, "codex", stage, ""); err == nil {
		t.Fatal("missing archive asset fetched")
	}
	if _, _, err := client.FetchArchive(context.Background(), Resolved{}, "codex", stage, ""); err == nil {
		t.Fatal("unresolved release fetched")
	}
	mismatched := release
	mismatched.Tag = "v1.0.1"
	mismatched.Version = mustVersion("1.0.1")
	if _, err := client.Resolve(context.Background(), mismatched, t.TempDir()); err == nil {
		t.Fatal("manifest describing another tag accepted")
	}
	timeout, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	if _, err := client.Resolve(timeout, release, t.TempDir()); err == nil {
		t.Fatal("expired context ignored")
	}
}

func mustURL(t *testing.T, value string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

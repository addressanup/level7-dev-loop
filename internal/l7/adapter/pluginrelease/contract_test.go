package pluginrelease

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseVersionAndCompare(t *testing.T) {
	ordered := []string{"0.9.0", "1.0.0-alpha", "1.0.0-alpha.1", "1.0.0-alpha.beta", "1.0.0-beta", "1.0.0-beta.2", "1.0.0-beta.11", "1.0.0-dev", "1.0.0-rc.1", "1.0.0", "1.0.1", "1.1.0-dev", "1.1.0", "2.0.0"}
	for index := 1; index < len(ordered); index++ {
		left, err := ParseVersion(ordered[index-1])
		if err != nil {
			t.Fatal(err)
		}
		right, err := ParseVersion(ordered[index])
		if err != nil {
			t.Fatal(err)
		}
		if left.Compare(right) >= 0 || right.Compare(left) <= 0 || left.Compare(left) != 0 {
			t.Fatalf("%s should sort before %s", ordered[index-1], ordered[index])
		}
		if right.String() != ordered[index] {
			t.Fatalf("round trip %q became %q", ordered[index], right.String())
		}
	}
	for _, invalid := range []string{"", "v1.0.0", "1.0", "1.0.0.0", "01.0.0", "1.0.0-", "1.0.0-01", "1.0.0+build", "1.0.0-a b", "1.0.0-ünicode", strings.Repeat("1", 70)} {
		if _, err := ParseVersion(invalid); err == nil {
			t.Fatalf("%q parsed", invalid)
		}
	}
	version, _ := ParseVersion("1.0.0-dev")
	if !version.Prerelease() {
		t.Fatal("1.0.0-dev is a prerelease")
	}
}

func testdata(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// The real v1.0.0 assets are the first contract this installer consumes; the
// manifest bytes are fixed here so the decoder is tested against publication,
// not against a synthetic shape.
func TestDecodeRealStableManifest(t *testing.T) {
	contract, err := DecodeManifest("RELEASE-MANIFEST.json", testdata(t, "v1.0.0-RELEASE-MANIFEST.json"))
	if err != nil {
		t.Fatal(err)
	}
	if contract.Version.String() != "1.0.0" || contract.Tag != "v1.0.0" || contract.Channel != ChannelStable || contract.Workflow.Path != ".github/workflows/release.yml" {
		t.Fatalf("identity: %+v", contract)
	}
	if contract.SourceCommit != "ed1457f2c54e2308f757d7ea3bf4fcfa113575ec" || contract.SourceTree != "cb596fe99648ec545bb924258896434bc1c87e73" || contract.RunID != "38042885014" {
		t.Fatalf("source identity: %+v", contract)
	}
	if contract.Manifest.SHA256 != "9e62aa6cab359f5e8c1f53a1eef16d9b27ade809442af5637bd7974cc775be76" || contract.Manifest.Size != 3013 {
		t.Fatalf("manifest digest: %+v", contract.Manifest)
	}
	if contract.Archives["claude"].SHA256 != "6c48280e078429ecef61815186042efe404345e97b0a9c0b7b1f77c6e0bf111c" || contract.Archives["claude"].Size != 22735243 ||
		contract.Archives["codex"].SHA256 != "37702406afcdd77b39ee3b58c99de96e8497ece4a7d3ca87b20f54abda038dea" || contract.Archives["codex"].Name != "level7-dev-loop-1.0.0-codex.zip" {
		t.Fatalf("archives: %+v", contract.Archives)
	}
	if contract.Checksums.SHA256 != "633b7d9ffdbf7a704cccb1427f9beeeb31c41d37fcc9e2ef72951dccf32f309f" || contract.Checksums.Size != 197 {
		t.Fatalf("checksums: %+v", contract.Checksums)
	}
	if contract.Inputs["darwin-arm64/l7"] != "6148c531f6fdd397a0d2891770573af7b5e3d53d134deee8878ecb1cdc47f572" || len(contract.Inputs) != 4 {
		t.Fatalf("inputs: %+v", contract.Inputs)
	}
	qualification := contract.Qualification
	if !qualification.Unsigned() || !qualification.Unqualified() || qualification.Policy != "owner-authorized-unsigned-stable-v1" || qualification.FormalSupport != "WITHHELD" || len(qualification.Waivers) != 4 {
		t.Fatalf("qualification: %+v", qualification)
	}
	warnings := strings.Join(qualification.Warnings(), "\n")
	if !strings.Contains(warnings, "UNSIGNED and NOT NOTARIZED") || !strings.Contains(warnings, "WITHHELD") || !strings.Contains(warnings, "NOT_RUN") || !strings.Contains(warnings, "never disable Gatekeeper") {
		t.Fatalf("warnings: %q", warnings)
	}
	if err := ParseChecksums(contract, testdata(t, "v1.0.0-SHA256SUMS")); err != nil {
		t.Fatal(err)
	}
}

func TestDecodeRealPrereleaseManifest(t *testing.T) {
	contract, err := DecodeManifest("UNSIGNED-PRERELEASE-MANIFEST.json", testdata(t, "v1.0.0-dev-UNSIGNED-PRERELEASE-MANIFEST.json"))
	if err != nil {
		t.Fatal(err)
	}
	if contract.Version.String() != "1.0.0-dev" || contract.Channel != ChannelPrerelease || contract.Workflow.Path != ".github/workflows/unsigned-prerelease.yml" || contract.SourceCommit != "245d5140f6dea7cc062c11bd51d499b22077c5f6" {
		t.Fatalf("identity: %+v", contract)
	}
	if contract.Archives["codex"].SHA256 != "305dfd213ddcdcf749bb8f427ec539cdd01f6392ea6140b8b9732ec7a4e172e3" || contract.Inputs["darwin-amd64/l7"] != "412a1d09f657066aab55fdab33c4b111e1a64f3cdbf731ed11b00cc45c5babb4" {
		t.Fatalf("assets: %+v %+v", contract.Archives, contract.Inputs)
	}
	if !contract.Qualification.Unsigned() || !contract.Qualification.Unqualified() || contract.Qualification.Policy != "l7-v1-explicitly-unsigned-prerelease" {
		t.Fatalf("qualification: %+v", contract.Qualification)
	}
}

func TestDecodeManifestRejectsDriftAndUnknownShapes(t *testing.T) {
	stable := testdata(t, "v1.0.0-RELEASE-MANIFEST.json")
	prerelease := testdata(t, "v1.0.0-dev-UNSIGNED-PRERELEASE-MANIFEST.json")
	replace := func(data []byte, old, new string) []byte { return bytes.Replace(data, []byte(old), []byte(new), 1) }
	tests := map[string]struct {
		name string
		data []byte
	}{
		"unknown manifest name": {"MANIFEST.json", stable},
		"other repository":      {"RELEASE-MANIFEST.json", replace(stable, `"repository": "addressanup/level7-dev-loop"`, `"repository": "evil/level7-dev-loop"`)},
		"other workflow":        {"RELEASE-MANIFEST.json", replace(stable, `"path": ".github/workflows/release.yml"`, `"path": ".github/workflows/evil.yml"`)},
		"tag drift":             {"RELEASE-MANIFEST.json", replace(stable, `"tag": "v1.0.0"`, `"tag": "v1.0.1"`)},
		"channel drift":         {"RELEASE-MANIFEST.json", replace(stable, `"release_channel": "stable"`, `"release_channel": "beta"`)},
		"schema drift":          {"RELEASE-MANIFEST.json", replace(stable, `"schema": 1`, `"schema": 2`)},
		"duplicate key":         {"RELEASE-MANIFEST.json", replace(stable, `"version": "1.0.0"`, `"version": "1.0.0", "version": "1.0.0"`)},
		"package digest disagreement": {"RELEASE-MANIFEST.json", replace(stable,
			`"level7-dev-loop-1.0.0-codex.zip": "37702406afcdd77b39ee3b58c99de96e8497ece4a7d3ca87b20f54abda038dea"`,
			`"level7-dev-loop-1.0.0-codex.zip": "0000000000000000000000000000000000000000000000000000000000000000"`)},
		"missing input digest":            {"RELEASE-MANIFEST.json", replace(stable, `"darwin-arm64/l7": "6148c531f6fdd397a0d2891770573af7b5e3d53d134deee8878ecb1cdc47f572"`, `"darwin-arm64/other": "6148c531f6fdd397a0d2891770573af7b5e3d53d134deee8878ecb1cdc47f572"`)},
		"stable manifest as prerelease":   {"UNSIGNED-PRERELEASE-MANIFEST.json", stable},
		"prerelease with release version": {"UNSIGNED-PRERELEASE-MANIFEST.json", replace(replace(prerelease, `"version": "1.0.0-dev"`, `"version": "1.0.0"`), `"tag": "v1.0.0-dev"`, `"tag": "v1.0.0"`)},
		"trailing data":                   {"RELEASE-MANIFEST.json", append(append([]byte{}, stable...), []byte("{}")...)},
		"empty":                           {"RELEASE-MANIFEST.json", nil},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeManifest(test.name, test.data); err == nil {
				t.Fatal("drifted manifest decoded")
			}
		})
	}
}

func TestParseChecksumsRequiresExactArchiveSet(t *testing.T) {
	contract, err := DecodeManifest("RELEASE-MANIFEST.json", testdata(t, "v1.0.0-RELEASE-MANIFEST.json"))
	if err != nil {
		t.Fatal(err)
	}
	valid := string(testdata(t, "v1.0.0-SHA256SUMS"))
	for name, data := range map[string]string{
		"missing archive":  strings.SplitN(valid, "\n", 2)[0] + "\n",
		"extra asset":      valid + strings.Repeat("a", 64) + "  extra.zip\n",
		"wrong digest":     strings.Replace(valid, "6c48280e", "00000000", 1),
		"single space":     strings.Replace(valid, "  level7", " level7", 1),
		"path in name":     strings.Replace(valid, "  level7", "  ../level7", 1),
		"duplicate asset":  valid + strings.SplitN(valid, "\n", 2)[0] + "\n",
		"uppercase digest": strings.ToUpper(valid),
	} {
		if err := ParseChecksums(contract, []byte(data)); err == nil {
			t.Fatalf("%s accepted", name)
		}
	}
}

func TestQualificationFailsClosedOnUnknownValues(t *testing.T) {
	signed := Qualification{DeveloperIDSigning: "PASS", Notarization: "ACCEPTED", FormalSupport: "SUPPORTED", HoldoutEvaluation: "PASS", ProviderTrials: "PASS"}
	if signed.Unsigned() || signed.Unqualified() || len(signed.Warnings()) != 0 {
		t.Fatalf("fully qualified release flagged: %+v", signed.Warnings())
	}
	for _, partial := range []Qualification{{}, {DeveloperIDSigning: "PASS"}, {DeveloperIDSigning: "PASS", Notarization: "ACCEPTED"}, {DeveloperIDSigning: "pass", Notarization: "ACCEPTED", FormalSupport: "SUPPORTED", HoldoutEvaluation: "PASS", ProviderTrials: "PASS"}} {
		if !partial.Unsigned() && !partial.Unqualified() {
			t.Fatalf("partial qualification passed: %+v", partial)
		}
	}
}

func TestSelectHonoursChannelsAndPins(t *testing.T) {
	releases := []Release{
		{Tag: "v0.9.0", Version: mustVersion("0.9.0")},
		{Tag: "v1.0.0", Version: mustVersion("1.0.0")},
		{Tag: "v1.1.0-dev", Version: mustVersion("1.1.0-dev"), Prerelease: true},
	}
	if release, err := Select(releases, Selection{Channel: ChannelStable}); err != nil || release.Tag != "v1.0.0" {
		t.Fatalf("stable selection: %+v %v", release, err)
	}
	if release, err := Select(releases, Selection{Channel: ChannelPrerelease}); err != nil || release.Tag != "v1.1.0-dev" {
		t.Fatalf("prerelease selection: %+v %v", release, err)
	}
	if _, err := Select(releases, Selection{Channel: ChannelStable, Exact: "1.1.0-dev"}); err == nil || !strings.Contains(err.Error(), "--channel prerelease") {
		t.Fatalf("pinned prerelease without channel: %v", err)
	}
	if release, err := Select(releases, Selection{Channel: ChannelPrerelease, Exact: "1.1.0-dev"}); err != nil || release.Tag != "v1.1.0-dev" {
		t.Fatalf("explicit pinned prerelease: %v", err)
	}
	if release, err := Select(releases, Selection{Channel: ChannelStable, Exact: "0.9.0"}); err != nil || release.Tag != "v0.9.0" {
		t.Fatalf("pinned older stable: %v", err)
	}
	if _, err := Select(releases, Selection{Channel: ChannelStable, Exact: "3.0.0"}); err == nil || !strings.Contains(err.Error(), "v0.9.0, v1.0.0, v1.1.0-dev") {
		t.Fatalf("missing pin should list releases: %v", err)
	}
	if _, err := Select(releases[2:], Selection{Channel: ChannelStable}); err == nil {
		t.Fatal("stable channel selected a prerelease")
	}
	if _, err := Select(releases, Selection{Channel: "nightly"}); err == nil {
		t.Fatal("invalid channel accepted")
	}
}

func TestNormalizeReleaseRejectsInconsistentPublication(t *testing.T) {
	valid := apiRelease{TagName: "v1.0.0", Assets: []struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	}{{"RELEASE-MANIFEST.json", 10}, {"SHA256SUMS", 10}}}
	if release, ok := normalizeRelease(valid); !ok || release.Manifest.Channel != ChannelStable {
		t.Fatalf("valid release rejected: %+v", release)
	}
	drafted := valid
	drafted.Draft = true
	if _, ok := normalizeRelease(drafted); ok {
		t.Fatal("draft accepted")
	}
	flagged := valid
	flagged.Prerelease = true
	if _, ok := normalizeRelease(flagged); ok {
		t.Fatal("stable manifest on a prerelease-flagged release accepted")
	}
	twoManifests := valid
	twoManifests.Assets = append(twoManifests.Assets, struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	}{"UNSIGNED-PRERELEASE-MANIFEST.json", 10})
	if _, ok := normalizeRelease(twoManifests); ok {
		t.Fatal("two manifests accepted")
	}
	badTag := valid
	badTag.TagName = "release-1"
	if _, ok := normalizeRelease(badTag); ok {
		t.Fatal("non-version tag accepted")
	}
}

func mustVersion(value string) Version {
	version, err := ParseVersion(value)
	if err != nil {
		panic(err)
	}
	return version
}

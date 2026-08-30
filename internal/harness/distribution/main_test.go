package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCurrentRepositoryDistributionCheck(t *testing.T) {
	root := distributionRepositoryRoot(t)
	inputs, err := loadInputs(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkGeneratedFiles(inputs); err != nil {
		t.Fatal(err)
	}
	packages, err := buildPackages(inputs)
	if err != nil {
		t.Fatal(err)
	}
	if len(packages) != 2 || packages[0].Host != "codex" || packages[1].Host != "claude" || packages[0].ArchiveDigest == packages[1].ArchiveDigest {
		t.Fatalf("unexpected packages: %+v", packages)
	}
	for _, built := range packages {
		if err := validateArchive(built.Archive, built.Entries); err != nil {
			t.Fatalf("%s archive: %v", built.Host, err)
		}
		if !bytes.Contains(built.Catalog, []byte("level7-engineering")) || bytes.Contains(built.Catalog, []byte("level7-engineering-development")) {
			t.Fatalf("%s catalog is not stable-bound", built.Host)
		}
	}

	var stdout, stderr bytes.Buffer
	if code := run([]string{"--root", root, "--check"}, &stdout, &stderr); code != 0 {
		t.Fatalf("run code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "distribution-check: PASS") ||
		!strings.Contains(stdout.String(), "actual_host_this_run=NOT_RUN") ||
		!strings.Contains(stdout.String(), "recorded_host_gate=SMOKE_TESTED") || stderr.Len() != 0 {
		t.Fatalf("unexpected check output stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestDistributionMetadataRemainsBound(t *testing.T) {
	inputs, err := loadInputs(distributionRepositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	packages, err := buildPackages(inputs)
	if err != nil {
		t.Fatal(err)
	}
	expectedDigests := map[string]string{
		"codex":  "58ec422efd1b672f3c5d2aa6d1e7672077fb7741c68abcc548179c188f329dba",
		"claude": "0a589d5566ffb6498f0501f76cd198ac0100edc3570a07f094fe1de595241c49",
	}
	for _, built := range packages {
		t.Run(built.Host, func(t *testing.T) {
			if err := validateBuiltPackage(built); err != nil {
				t.Fatal(err)
			}
			if built.ArchiveDigest != expectedDigests[built.Host] {
				t.Fatalf("archive digest changed: %s", built.ArchiveDigest)
			}

			distributionData := requirePackageEntry(t, built, "DISTRIBUTION.json")
			var metadata distributionMetadata
			if err := decodeStrict(distributionData, &metadata); err != nil {
				t.Fatal(err)
			}
			_, host, err := hostInputs(inputs, built.Host)
			if err != nil {
				t.Fatal(err)
			}
			want := distributionMetadata{
				Schema: 2, Name: inputs.Descriptor.Name, Version: inputs.Descriptor.Version,
				Channel: inputs.Descriptor.Channel, Host: built.Host, ManifestPath: host.ManifestPath,
				CatalogPath: host.CatalogPath, CatalogSHA256: built.CatalogDigest,
				SourceDigest: metadata.SourceDigest, Builder: builderVersion,
				SupportClaim: "WITHHELD", ActualHostGate: "SMOKE_TESTED",
			}
			if metadata != want || metadata.SourceDigest != built.SourceDigest ||
				!sha256Pattern.MatchString(metadata.SourceDigest) || sha256Hex(built.Catalog) != built.CatalogDigest {
				t.Fatalf("distribution metadata is not exactly bound: got=%+v want=%+v", metadata, want)
			}

			var provenance provenanceInput
			if err := decodeStrict(requirePackageEntry(t, built, "PROVENANCE.input.json"), &provenance); err != nil {
				t.Fatal(err)
			}
			if provenance.Package != metadata.Name || provenance.Version != metadata.Version || provenance.Host != metadata.Host ||
				provenance.SourceDigest != metadata.SourceDigest || provenance.Builder != metadata.Builder {
				t.Fatalf("provenance identity diverged from distribution metadata: %+v", provenance)
			}

			var sbom sbomDocument
			if err := decodeStrict(requirePackageEntry(t, built, "SBOM.spdx.json"), &sbom); err != nil {
				t.Fatal(err)
			}
			if len(sbom.Packages) != 1 || sbom.Packages[0].VersionInfo != metadata.Version || len(sbom.Packages[0].Checksums) != 1 ||
				sbom.Packages[0].Checksums[0].Algorithm != "SHA256" || sbom.Packages[0].Checksums[0].ChecksumValue != metadata.SourceDigest {
				t.Fatalf("SBOM identity diverged from distribution metadata: %+v", sbom.Packages)
			}

			var manifest inventory
			if err := decodeStrict(requirePackageEntry(t, built, "INVENTORY.json"), &manifest); err != nil {
				t.Fatal(err)
			}
			matches := 0
			for _, file := range manifest.Files {
				if file.Path == "DISTRIBUTION.json" {
					matches++
					if file.Mode != "0644" || file.Size != len(distributionData) || file.SHA256 != sha256Hex(distributionData) {
						t.Fatalf("distribution inventory binding is invalid: %+v", file)
					}
				}
			}
			if matches != 1 {
				t.Fatalf("inventory contains %d DISTRIBUTION.json entries", matches)
			}
		})
	}
}

func TestCatalogBytesParticipateInHostSourceIdentity(t *testing.T) {
	inputs, err := loadInputs(distributionRepositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, host := range []string{"codex", "claude"} {
		_, descriptor, err := hostInputs(inputs, host)
		if err != nil {
			t.Fatal(err)
		}
		catalog, err := renderCatalog(inputs.Descriptor, host)
		if err != nil {
			t.Fatal(err)
		}
		mutated := append([]byte{}, catalog...)
		mutated[len(mutated)-2] ^= 1
		files := map[string][]byte{"fixture": []byte("same package inputs\n")}
		first := digestSource(inputs, host, descriptor.CatalogPath, catalog, files)
		second := digestSource(inputs, host, descriptor.CatalogPath, mutated, files)
		if first == second || !sha256Pattern.MatchString(first) || !sha256Pattern.MatchString(second) {
			t.Fatalf("%s catalog bytes did not change source identity: %s %s", host, first, second)
		}
	}
}

func TestCommittedMarketplaceProjectionRemainsExact(t *testing.T) {
	inputs, err := loadInputs(distributionRepositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	expected, err := renderCommittedMarketplaceFiles(inputs)
	if err != nil {
		t.Fatal(err)
	}
	pluginRoot := filepath.ToSlash(filepath.Join("plugins", inputs.Descriptor.Name))
	payloadFiles := 0
	for relative := range expected {
		if strings.HasPrefix(relative, pluginRoot+"/") {
			payloadFiles++
		}
	}
	if payloadFiles != len(inputs.Descriptor.Skills)+5 {
		t.Fatalf("payload files=%d want=%d", payloadFiles, len(inputs.Descriptor.Skills)+5)
	}

	var codex codexCatalog
	if err := decodeStrict(expected[inputs.Descriptor.Hosts.Codex.CatalogPath], &codex); err != nil {
		t.Fatal(err)
	}
	if codex.Name != "level7-engineering" || codex.Interface.DisplayName != "Level 7 Engineering" || len(codex.Plugins) != 1 ||
		codex.Plugins[0].Name != inputs.Descriptor.Name || codex.Plugins[0].Source.Source != "local" ||
		codex.Plugins[0].Source.Path != "./plugins/"+inputs.Descriptor.Name ||
		codex.Plugins[0].Policy.Installation != "AVAILABLE" || codex.Plugins[0].Policy.Authentication != "ON_INSTALL" ||
		codex.Plugins[0].Category != inputs.Descriptor.Hosts.Codex.Category {
		t.Fatalf("unexpected committed Codex catalog: %+v", codex)
	}

	var claude claudeCatalog
	if err := decodeStrict(expected[inputs.Descriptor.Hosts.Claude.CatalogPath], &claude); err != nil {
		t.Fatal(err)
	}
	if claude.Name != "level7-engineering" || claude.Owner != inputs.Descriptor.Author || len(claude.Plugins) != 1 ||
		claude.Plugins[0].Name != inputs.Descriptor.Name || claude.Plugins[0].Source != "./plugins/"+inputs.Descriptor.Name ||
		claude.Plugins[0].Version != inputs.Descriptor.Version || claude.Plugins[0].License != inputs.Descriptor.License ||
		claude.Plugins[0].Category != inputs.Descriptor.Hosts.Claude.Category || !claude.Plugins[0].Strict {
		t.Fatalf("unexpected committed Claude catalog: %+v", claude)
	}

	rootFiles, err := renderRootFiles(inputs.Descriptor)
	if err != nil {
		t.Fatal(err)
	}
	for _, manifest := range []string{inputs.Descriptor.Hosts.Codex.ManifestPath, inputs.Descriptor.Hosts.Claude.ManifestPath} {
		if !bytes.Equal(expected[filepath.ToSlash(filepath.Join(pluginRoot, manifest))], rootFiles[manifest]) {
			t.Fatalf("committed payload manifest drift: %s", manifest)
		}
	}
	for relative, data := range inputs.Skills {
		if !bytes.Equal(expected[filepath.ToSlash(filepath.Join(pluginRoot, relative))], data) {
			t.Fatalf("committed payload skill drift: %s", relative)
		}
	}
	if _, present := expected[filepath.ToSlash(filepath.Join(pluginRoot, "plugin.json"))]; present {
		t.Fatal("committed payload contains a redundant legacy manifest")
	}
	readme := expected[filepath.ToSlash(filepath.Join(pluginRoot, "README.md"))]
	if !bytes.Contains(readme, []byte("`"+inputs.Descriptor.Version+"`")) ||
		!bytes.Contains(readme, []byte("adds no executable")) {
		t.Fatalf("committed payload README is not release-bound: %q", readme)
	}
}

func TestCommittedMarketplaceProjectionFailsClosed(t *testing.T) {
	inputs, err := loadInputs(distributionRepositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	expected, err := renderCommittedMarketplaceFiles(inputs)
	if err != nil {
		t.Fatal(err)
	}
	pluginRoot := filepath.Join("plugins", inputs.Descriptor.Name)
	tests := []struct {
		name   string
		mutate func(*testing.T, string)
	}{
		{name: "stale catalog version", mutate: func(t *testing.T, root string) {
			t.Helper()
			name := filepath.Join(root, filepath.FromSlash(inputs.Descriptor.Hosts.Claude.CatalogPath))
			data, readErr := os.ReadFile(name)
			if readErr != nil {
				t.Fatal(readErr)
			}
			data = bytes.Replace(data, []byte(`"version": "`+inputs.Descriptor.Version+`"`), []byte(`"version": "9.9.9"`), 1)
			if writeErr := os.WriteFile(name, data, 0o644); writeErr != nil {
				t.Fatal(writeErr)
			}
		}},
		{name: "catalog source escape", mutate: func(t *testing.T, root string) {
			t.Helper()
			name := filepath.Join(root, filepath.FromSlash(inputs.Descriptor.Hosts.Codex.CatalogPath))
			data, readErr := os.ReadFile(name)
			if readErr != nil {
				t.Fatal(readErr)
			}
			data = bytes.Replace(data, []byte("./plugins/"+inputs.Descriptor.Name), []byte("../outside"), 1)
			if writeErr := os.WriteFile(name, data, 0o644); writeErr != nil {
				t.Fatal(writeErr)
			}
		}},
		{name: "skill substitution", mutate: func(t *testing.T, root string) {
			t.Helper()
			name := filepath.Join(root, pluginRoot, "skills", inputs.Descriptor.Skills[0], "SKILL.md")
			if writeErr := os.WriteFile(name, []byte("substituted\n"), 0o644); writeErr != nil {
				t.Fatal(writeErr)
			}
		}},
		{name: "extra file", mutate: func(t *testing.T, root string) {
			t.Helper()
			if writeErr := os.WriteFile(filepath.Join(root, pluginRoot, "extra.txt"), []byte("extra\n"), 0o644); writeErr != nil {
				t.Fatal(writeErr)
			}
		}},
		{name: "extra directory", mutate: func(t *testing.T, root string) {
			t.Helper()
			if mkdirErr := os.Mkdir(filepath.Join(root, pluginRoot, "empty"), 0o755); mkdirErr != nil {
				t.Fatal(mkdirErr)
			}
		}},
		{name: "symlink", mutate: func(t *testing.T, root string) {
			t.Helper()
			if symlinkErr := os.Symlink("README.md", filepath.Join(root, pluginRoot, "README.link")); symlinkErr != nil {
				t.Fatal(symlinkErr)
			}
		}},
		{name: "executable mode", mutate: func(t *testing.T, root string) {
			t.Helper()
			if chmodErr := os.Chmod(filepath.Join(root, pluginRoot, "README.md"), 0o755); chmodErr != nil {
				t.Fatal(chmodErr)
			}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := writeCommittedMarketplaceFixture(t, expected)
			test.mutate(t, root)
			fixture := inputs
			fixture.Root = root
			if err := checkGeneratedFiles(fixture); err == nil {
				t.Fatal("mutated committed marketplace passed")
			}
		})
	}
}

func TestGeneratedManifestsUseOneStableIdentity(t *testing.T) {
	inputs, err := loadInputs(distributionRepositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := renderRootFiles(inputs.Descriptor)
	if err != nil {
		t.Fatal(err)
	}
	for relative, data := range rendered {
		if !bytes.Contains(data, []byte(inputs.Descriptor.Version)) {
			t.Fatalf("%s does not bind version %s", relative, inputs.Descriptor.Version)
		}
		if !bytes.Contains(data, []byte(`"version": "`+inputs.Descriptor.Version+`"`)) || bytes.Contains(data, []byte("-dev.")) ||
			bytes.Contains(data, []byte("mcpServers")) || bytes.Contains(data, []byte(`"hooks"`)) {
			t.Fatalf("%s contains unstable or effectful metadata", relative)
		}
	}
	var codex codexManifest
	if err := decodeStrict(rendered[".codex-plugin/plugin.json"], &codex); err != nil {
		t.Fatal(err)
	}
	if codex.Skills != "./skills/" ||
		codex.Interface.LongDescription != "Solo-first development conductor for inspect, implement, test, repair, self-review, and review-ready handoff. Team assurance is opt-in. The instruction-only package does not add executables, hooks, MCP servers, or network access." ||
		codex.Interface.Capabilities == nil || len(codex.Interface.Capabilities) != 0 ||
		len(codex.Interface.DefaultPrompt) != 3 ||
		!strings.Contains(codex.Interface.DefaultPrompt[0], "l7-next") ||
		!strings.Contains(codex.Interface.DefaultPrompt[1], "l7-next") ||
		strings.Contains(strings.ToLower(strings.Join(codex.Interface.DefaultPrompt, "\n")), "audit") {
		t.Fatalf("unexpected Codex manifest: %+v", codex)
	}
	var claude claudeManifest
	if err := decodeStrict(rendered[".claude-plugin/plugin.json"], &claude); err != nil {
		t.Fatal(err)
	}
	if claude.Schema == "" || claude.DisplayName == "" {
		t.Fatalf("unexpected Claude manifest: %+v", claude)
	}
}

func TestStableVersionAndChangelogIdentityAreCanonical(t *testing.T) {
	for _, version := range []string{"0.0.0", "0.1.0", "12.34.56"} {
		if !versionPattern.MatchString(version) {
			t.Fatalf("canonical stable version rejected: %s", version)
		}
	}
	for _, version := range []string{
		"1.0.0-rc.1", "1.0.0-preview.1", "1.0.0-dev.1", "1.0.0+build.1",
		"01.0.0", "1.00.0", "1.0.00", "v1.0.0", "1.0", "1.0.0.0",
	} {
		if versionPattern.MatchString(version) {
			t.Fatalf("noncanonical stable version passed: %s", version)
		}
	}
	if next, err := nextStablePatchVersion("0.1.0"); err != nil || next != "0.1.1" {
		t.Fatalf("next stable patch version=%q error=%v", next, err)
	}
	large := "0.1." + strings.Repeat("9", 128)
	if next, err := nextStablePatchVersion(large); err != nil || !versionPattern.MatchString(next) || next == large {
		t.Fatalf("large stable patch version did not advance canonically: next=%q error=%v", next, err)
	}

	license := []byte("MIT License\n")
	canonical := []byte("# Changelog\n\n## 0.1.0\n")
	if err := validatePackageDocuments(license, canonical, "0.1.0"); err != nil {
		t.Fatal(err)
	}
	for _, changelog := range [][]byte{
		[]byte("# 0.1.0\n"),
		[]byte("## 0.1.1\n"),
		[]byte("# Changelog\n\nprefix ## 0.1.0\n"),
		[]byte("# Changelog\n\n## 0.1.0 — Unreleased\n"),
		append(append([]byte{}, canonical...), []byte("## 0.1.0\n")...),
		append(append([]byte{}, canonical...), []byte("## 0.1.0 — Unreleased\n")...),
	} {
		if err := validatePackageDocuments(license, changelog, "0.1.0"); err == nil {
			t.Fatalf("noncanonical changelog identity passed: %q", changelog)
		}
	}
}

func TestDescriptorAndCompatibilityFailClosed(t *testing.T) {
	inputs, err := loadInputs(distributionRepositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		mutate func(*packageDescriptor)
	}{
		{name: "development version", mutate: func(value *packageDescriptor) { value.Version = "0.1.0-dev.6" }},
		{name: "release candidate", mutate: func(value *packageDescriptor) { value.Version = "0.1.0-rc.1" }},
		{name: "build suffix", mutate: func(value *packageDescriptor) { value.Version = "0.1.0+build.1" }},
		{name: "leading-zero patch", mutate: func(value *packageDescriptor) { value.Version = "0.1.00" }},
		{name: "development channel", mutate: func(value *packageDescriptor) { value.Channel = "development" }},
		{name: "unknown release channel", mutate: func(value *packageDescriptor) { value.Channel = "release" }},
		{name: "network", mutate: func(value *packageDescriptor) { value.Permissions.Level7Network = true }},
		{name: "hook", mutate: func(value *packageDescriptor) { value.Permissions.Hook = true }},
		{name: "catalog path", mutate: func(value *packageDescriptor) { value.Hosts.Codex.CatalogPath = "marketplace.json" }},
		{name: "duplicate skill", mutate: func(value *packageDescriptor) { value.Skills[1] = value.Skills[0] }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := inputs.Descriptor
			value.Skills = append([]string{}, value.Skills...)
			test.mutate(&value)
			if err := validateDescriptor(value); err == nil {
				t.Fatalf("unsafe descriptor passed: %+v", value)
			}
		})
	}

	if err := validateCompatibility(inputs.Compatibility); err != nil {
		t.Fatalf("repository compatibility matrix failed: %v", err)
	}
	compatibilityTests := []struct {
		name   string
		mutate func(*compatibilityMatrix)
	}{
		{name: "schema", mutate: func(value *compatibilityMatrix) { value.Schema = 2 }},
		{name: "host order", mutate: func(value *compatibilityMatrix) {
			value.Entries[0], value.Entries[1] = value.Entries[1], value.Entries[0]
		}},
		{name: "host product", mutate: func(value *compatibilityMatrix) { value.Entries[0].HostProduct = "Another Codex" }},
		{name: "declared surface", mutate: func(value *compatibilityMatrix) { value.Entries[1].DeclaredSurface = "another surface" }},
		{name: "codex fixture version", mutate: func(value *compatibilityMatrix) { value.Entries[0].AdapterFixtureVersion = "codex-cli 9.9.9" }},
		{name: "claude fixture version", mutate: func(value *compatibilityMatrix) { value.Entries[1].AdapterFixtureVersion = "9.9.9" }},
		{name: "codex qualification target", mutate: func(value *compatibilityMatrix) { value.Entries[0].QualificationTarget = "codex-cli 9.9.9" }},
		{name: "claude qualification target", mutate: func(value *compatibilityMatrix) { value.Entries[1].QualificationTarget = "9.9.9 (Claude Code)" }},
		{name: "platform order", mutate: func(value *compatibilityMatrix) {
			value.Entries[0].OperatingSystems[0], value.Entries[0].OperatingSystems[1] = value.Entries[0].OperatingSystems[1], value.Entries[0].OperatingSystems[0]
		}},
		{name: "arm host runtime downgrade", mutate: func(value *compatibilityMatrix) { value.Entries[0].OperatingSystems[0].HostRuntime = "NOT_RUN" }},
		{name: "host runtime promotion", mutate: func(value *compatibilityMatrix) { value.Entries[1].OperatingSystems[1].HostRuntime = "PASS" }},
		{name: "required capability", mutate: func(value *compatibilityMatrix) { value.Entries[0].RequiredCapabilities[0] = "plugin.json" }},
		{name: "required capability order", mutate: func(value *compatibilityMatrix) {
			value.Entries[1].RequiredCapabilities[0], value.Entries[1].RequiredCapabilities[1] = value.Entries[1].RequiredCapabilities[1], value.Entries[1].RequiredCapabilities[0]
		}},
		{name: "optional capability", mutate: func(value *compatibilityMatrix) { value.Entries[0].OptionalCapabilities = []string{"network"} }},
		{name: "null optional capabilities", mutate: func(value *compatibilityMatrix) { value.Entries[1].OptionalCapabilities = nil }},
		{name: "degraded behavior", mutate: func(value *compatibilityMatrix) { value.Entries[0].DegradedBehavior = "continue anyway" }},
		{name: "unsupported provider smoke promotion", mutate: func(value *compatibilityMatrix) { value.Entries[0].ProviderExecution = "SMOKE_TESTED" }},
		{name: "provider execution promotion", mutate: func(value *compatibilityMatrix) { value.Entries[0].ProviderExecution = "PASS" }},
		{name: "actual lifecycle downgrade", mutate: func(value *compatibilityMatrix) { value.Entries[1].ActualHostLifecycle = "NOT_RUN" }},
		{name: "actual lifecycle promotion", mutate: func(value *compatibilityMatrix) { value.Entries[1].ActualHostLifecycle = "PASS" }},
		{name: "rollback", mutate: func(value *compatibilityMatrix) { value.Entries[1].Rollback = "delete everything" }},
		{name: "support promotion", mutate: func(value *compatibilityMatrix) { value.Entries[0].SupportClaim = "SUPPORTED" }},
	}
	for _, test := range compatibilityTests {
		t.Run("compatibility "+test.name, func(t *testing.T) {
			matrix := cloneCompatibilityMatrix(inputs.Compatibility)
			test.mutate(&matrix)
			if err := validateCompatibility(matrix); err == nil {
				t.Fatalf("mutated compatibility matrix passed: %+v", matrix)
			}
		})
	}

	unknown := append([]byte{}, inputs.DescriptorBytes...)
	unknown = bytes.Replace(unknown, []byte("\n  \"name\""), []byte("\n  \"unknown\": true,\n  \"name\""), 1)
	var descriptor packageDescriptor
	if err := decodeStrict(unknown, &descriptor); err == nil {
		t.Fatal("unknown descriptor field passed")
	}
	trailing := append(append([]byte{}, inputs.DescriptorBytes...), []byte("{}\n")...)
	if err := decodeStrict(trailing, &descriptor); err == nil {
		t.Fatal("trailing JSON passed")
	}
}

func TestWriteOutputsCreatesOnlyBoundedStableLayout(t *testing.T) {
	inputs, err := loadInputs(distributionRepositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	packages, err := buildPackages(inputs)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	output := filepath.Join(root, "build", "distributions")
	if err := writeOutputs(root, output, packages); err != nil {
		t.Fatal(err)
	}
	for _, built := range packages {
		archive, err := os.ReadFile(filepath.Join(output, built.ArchiveName))
		if err != nil || !bytes.Equal(archive, built.Archive) {
			t.Fatalf("%s archive output error=%v", built.Host, err)
		}
		checksum, err := os.ReadFile(filepath.Join(output, built.ArchiveName+".sha256"))
		if err != nil || string(checksum) != built.ArchiveDigest+"  "+built.ArchiveName+"\n" {
			t.Fatalf("%s checksum=%q error=%v", built.Host, checksum, err)
		}
		manifest := filepath.Join(output, built.Host+"-marketplace", "plugins", inputs.Descriptor.Name, filepath.FromSlash(built.Entries[0].Name))
		if _, err := os.Stat(manifest); err != nil {
			t.Fatalf("%s marketplace package missing: %v", built.Host, err)
		}
		catalog, err := os.ReadFile(filepath.Join(output, built.Host+"-marketplace", filepath.FromSlash(built.CatalogPath)))
		if err != nil || !bytes.Equal(catalog, built.Catalog) || sha256Hex(catalog) != built.CatalogDigest {
			t.Fatalf("%s marketplace catalog is not bound: digest=%s error=%v", built.Host, sha256Hex(catalog), err)
		}
	}
	unsafe := cloneOfflineQualificationPackages(packages)
	unsafe[0].Catalog = append(unsafe[0].Catalog, '\n')
	unsafe[0].CatalogDigest = sha256Hex(unsafe[0].Catalog)
	if err := writeOutputs(root, output, unsafe); err == nil {
		t.Fatal("catalog-substituted output passed")
	}
	for _, mutation := range []struct {
		name   string
		mutate func([]builtPackage)
	}{
		{name: "archive relabel", mutate: func(values []builtPackage) { values[0].ArchiveName = "renamed.zip" }},
		{name: "archive escape", mutate: func(values []builtPackage) { values[0].ArchiveName = "../renamed.zip" }},
		{name: "archive collision", mutate: func(values []builtPackage) { values[1].ArchiveName = values[0].ArchiveName }},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			values := cloneOfflineQualificationPackages(packages)
			mutation.mutate(values)
			if _, err := validateOutputPackageSet(values); err == nil {
				t.Fatal("unsafe output identity passed")
			}
		})
	}
	mixed := cloneOfflineQualificationPackages(packages)
	mixed[1], err = syntheticLifecycleVariant(mixed[1])
	if err != nil {
		t.Fatal(err)
	}
	mixed[1].ArchiveName = "level7-dev-loop-claude-" + mixed[1].Version + ".zip"
	if _, err := validateOutputPackageSet(mixed); err == nil {
		t.Fatal("mixed-version output passed")
	}
	for _, built := range packages {
		archive, err := os.ReadFile(filepath.Join(output, built.ArchiveName))
		if err != nil || !bytes.Equal(archive, built.Archive) {
			t.Fatalf("rejected output changed existing %s archive: %v", built.Host, err)
		}
	}
	if err := writeOutputs(root, filepath.Join(root, "outside"), packages); err == nil {
		t.Fatal("output outside exact build/distributions passed")
	}

	symlinkRoot := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(symlinkRoot, "build")); err != nil {
		t.Fatal(err)
	}
	if err := writeOutputs(symlinkRoot, filepath.Join(symlinkRoot, "build", "distributions"), packages); err == nil {
		t.Fatal("symlinked build root passed")
	}
}

func TestReadRegularBoundedRejectsSymlinkedParent(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "package.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "distribution")); err != nil {
		t.Fatal(err)
	}
	if _, err := readRegularBounded(root, "distribution/package.json"); err == nil {
		t.Fatal("source path with a symlinked parent passed")
	}
}

func TestRunRejectsAmbiguousModes(t *testing.T) {
	root := distributionRepositoryRoot(t)
	for _, arguments := range [][]string{
		nil,
		{"--root", root, "--check", "--output", "build/distributions"},
		{"--root", root, "--check", "extra"},
		{"--root", root, "--json"},
		{"--root", root, "--json", "--output", "build/distributions"},
	} {
		var stdout, stderr bytes.Buffer
		if code := run(arguments, &stdout, &stderr); code != 2 || stdout.Len() != 0 {
			t.Fatalf("arguments=%v code=%d stdout=%q stderr=%q", arguments, code, stdout.String(), stderr.String())
		}
	}
}

func TestRunEmitsCanonicalOfflineQualificationJSON(t *testing.T) {
	root := distributionRepositoryRoot(t)
	var stdout, stderr bytes.Buffer
	if code := run([]string{"--root", root, "--check", "--json"}, &stdout, &stderr); code != 0 || stderr.Len() != 0 {
		t.Fatalf("run code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	var report offlineQualificationReport
	if err := decodeStrict(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	canonical, err := jsonBytes(report)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(stdout.Bytes(), canonical) || report.ReleaseReady || report.Authority != qualificationUnevaluated ||
		report.Signing != qualificationNotRun || report.Publication != qualificationNotRun || len(report.Packages) != 2 {
		t.Fatalf("unexpected offline qualification output: %s", stdout.Bytes())
	}
}

func TestPackageMetadataJSONRemainsStrict(t *testing.T) {
	inputs, err := loadInputs(distributionRepositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	packages, err := buildPackages(inputs)
	if err != nil {
		t.Fatal(err)
	}
	for _, built := range packages {
		for _, entry := range built.Entries {
			if !strings.HasSuffix(entry.Name, ".json") {
				continue
			}
			var value any
			decoder := json.NewDecoder(bytes.NewReader(entry.Data))
			if err := decoder.Decode(&value); err != nil {
				t.Fatalf("%s %s: %v", built.Host, entry.Name, err)
			}
			if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
				t.Fatalf("%s %s contains trailing JSON", built.Host, entry.Name)
			}
		}
	}
}

func distributionRepositoryRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err == nil {
		root, err = filepath.EvalSymlinks(root)
	}
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func requirePackageEntry(t *testing.T, built builtPackage, name string) []byte {
	t.Helper()
	var data []byte
	matches := 0
	for _, entry := range built.Entries {
		if entry.Name == name {
			matches++
			data = entry.Data
		}
	}
	if matches != 1 {
		t.Fatalf("%s package contains %d %s entries", built.Host, matches, name)
	}
	return data
}

func cloneCompatibilityMatrix(matrix compatibilityMatrix) compatibilityMatrix {
	clone := matrix
	clone.Entries = append([]compatibilityEntry{}, matrix.Entries...)
	for index := range clone.Entries {
		clone.Entries[index].OperatingSystems = append([]operatingSystem{}, matrix.Entries[index].OperatingSystems...)
		clone.Entries[index].RequiredCapabilities = append([]string{}, matrix.Entries[index].RequiredCapabilities...)
		clone.Entries[index].OptionalCapabilities = append([]string{}, matrix.Entries[index].OptionalCapabilities...)
	}
	return clone
}

func writeCommittedMarketplaceFixture(t *testing.T, expected map[string][]byte) string {
	t.Helper()
	root := t.TempDir()
	for relative, data := range expected {
		if err := writeRegular(root, relative, data); err != nil {
			t.Fatalf("write fixture %s: %v", relative, err)
		}
	}
	return root
}

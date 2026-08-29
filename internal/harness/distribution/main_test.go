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
		if !bytes.Contains(built.Catalog, []byte("level7-engineering-development")) {
			t.Fatalf("%s catalog is not development-bound", built.Host)
		}
	}

	var stdout, stderr bytes.Buffer
	if code := run([]string{"--root", root, "--check"}, &stdout, &stderr); code != 0 {
		t.Fatalf("run code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "distribution-check: PASS") || !strings.Contains(stdout.String(), "actual_host=NOT_RUN") || stderr.Len() != 0 {
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
		"codex":  "9e54fff83a4ef3812bcfeb8737ec095305c828c7fd33e35926ae54588df39fd0",
		"claude": "718ea9366ac6d286a954e655275f994de9d6e9fd2679123efda903c8f6881acb",
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
				Schema: 1, Name: inputs.Descriptor.Name, Version: inputs.Descriptor.Version,
				Channel: inputs.Descriptor.Channel, Host: built.Host, ManifestPath: host.ManifestPath,
				CatalogPath: host.CatalogPath, SourceDigest: metadata.SourceDigest, Builder: builderVersion,
				SupportClaim: "WITHHELD", ActualHostGate: "NOT_RUN",
			}
			if metadata != want || !sha256Pattern.MatchString(metadata.SourceDigest) {
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

func TestGeneratedManifestsUseOnePrereleaseIdentity(t *testing.T) {
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
		if bytes.Contains(data, []byte(`"version": "1.0.0"`)) || bytes.Contains(data, []byte("mcpServers")) || bytes.Contains(data, []byte(`"hooks"`)) {
			t.Fatalf("%s contains promoted or effectful metadata", relative)
		}
	}
	var codex codexManifest
	if err := decodeStrict(rendered[".codex-plugin/plugin.json"], &codex); err != nil {
		t.Fatal(err)
	}
	if codex.Skills != "./skills/" ||
		!strings.Contains(codex.Interface.LongDescription, "Solo-first development conductor") ||
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

func TestDescriptorAndCompatibilityFailClosed(t *testing.T) {
	inputs, err := loadInputs(distributionRepositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		mutate func(*packageDescriptor)
	}{
		{name: "stable version", mutate: func(value *packageDescriptor) { value.Version = "1.0.0" }},
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
		{name: "host runtime promotion", mutate: func(value *compatibilityMatrix) { value.Entries[1].OperatingSystems[1].HostRuntime = "PASS" }},
		{name: "required capability", mutate: func(value *compatibilityMatrix) { value.Entries[0].RequiredCapabilities[0] = "plugin.json" }},
		{name: "required capability order", mutate: func(value *compatibilityMatrix) {
			value.Entries[1].RequiredCapabilities[0], value.Entries[1].RequiredCapabilities[1] = value.Entries[1].RequiredCapabilities[1], value.Entries[1].RequiredCapabilities[0]
		}},
		{name: "optional capability", mutate: func(value *compatibilityMatrix) { value.Entries[0].OptionalCapabilities = []string{"network"} }},
		{name: "null optional capabilities", mutate: func(value *compatibilityMatrix) { value.Entries[1].OptionalCapabilities = nil }},
		{name: "degraded behavior", mutate: func(value *compatibilityMatrix) { value.Entries[0].DegradedBehavior = "continue anyway" }},
		{name: "provider execution promotion", mutate: func(value *compatibilityMatrix) { value.Entries[0].ProviderExecution = "PASS" }},
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

func TestWriteOutputsCreatesOnlyBoundedDevelopmentLayout(t *testing.T) {
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
	if err := writeOutputs(root, output, inputs.Descriptor.Name, packages); err != nil {
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
		if _, err := os.Stat(filepath.Join(output, built.Host+"-marketplace", filepath.FromSlash(built.CatalogPath))); err != nil {
			t.Fatalf("%s marketplace catalog missing: %v", built.Host, err)
		}
	}
	if err := writeOutputs(root, filepath.Join(root, "outside"), inputs.Descriptor.Name, packages); err == nil {
		t.Fatal("output outside exact build/distributions passed")
	}

	symlinkRoot := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(symlinkRoot, "build")); err != nil {
		t.Fatal(err)
	}
	if err := writeOutputs(symlinkRoot, filepath.Join(symlinkRoot, "build", "distributions"), inputs.Descriptor.Name, packages); err == nil {
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
	} {
		var stdout, stderr bytes.Buffer
		if code := run(arguments, &stdout, &stderr); code != 2 || stdout.Len() != 0 {
			t.Fatalf("arguments=%v code=%d stdout=%q stderr=%q", arguments, code, stdout.String(), stderr.String())
		}
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

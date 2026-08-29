package main

import (
	"bytes"
	"encoding/json"
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

func TestGeneratedManifestsUseOnePrerelaseIdentity(t *testing.T) {
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
	if codex.Skills != "./skills/" || !strings.Contains(codex.Interface.LongDescription, "Development-only") {
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

	matrix := inputs.Compatibility
	matrix.Entries = append([]compatibilityEntry{}, matrix.Entries...)
	matrix.Entries[0].ProviderExecution = "PASS"
	if err := validateCompatibility(matrix); err == nil {
		t.Fatal("provider execution promotion passed")
	}
	matrix = inputs.Compatibility
	matrix.Entries = append([]compatibilityEntry{}, matrix.Entries...)
	matrix.Entries[1].OperatingSystems = append([]operatingSystem{}, matrix.Entries[1].OperatingSystems...)
	matrix.Entries[1].OperatingSystems[1].HostRuntime = "PASS"
	if err := validateCompatibility(matrix); err == nil {
		t.Fatal("actual-host promotion passed")
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
			if decoder.More() {
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

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
)

const (
	builderVersion  = "wave5-v1"
	maximumFileSize = 1 << 20
)

var (
	packageNamePattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	versionPattern     = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+-[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*$`)
	sha256Pattern      = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

type author struct {
	Name string `json:"name"`
}

type permissions struct {
	Level7Network     bool   `json:"level7_network"`
	BundledExecutable bool   `json:"bundled_executable"`
	MCPServer         bool   `json:"mcp_server"`
	Hook              bool   `json:"hook"`
	HostSetting       bool   `json:"host_setting"`
	Telemetry         bool   `json:"telemetry"`
	WorkspaceBoundary string `json:"workspace_boundary"`
}

type hostDescriptor struct {
	ManifestPath string `json:"manifest_path"`
	CatalogPath  string `json:"catalog_path"`
	DisplayName  string `json:"display_name"`
	Category     string `json:"category"`
	Invocation   string `json:"invocation"`
}

type hosts struct {
	Codex  hostDescriptor `json:"codex"`
	Claude hostDescriptor `json:"claude"`
}

type packageDescriptor struct {
	Schema      int         `json:"schema"`
	Name        string      `json:"name"`
	Version     string      `json:"version"`
	Channel     string      `json:"channel"`
	Description string      `json:"description"`
	Author      author      `json:"author"`
	Homepage    string      `json:"homepage"`
	Repository  string      `json:"repository"`
	License     string      `json:"license"`
	Keywords    []string    `json:"keywords"`
	Skills      []string    `json:"skills"`
	Permissions permissions `json:"permissions"`
	Hosts       hosts       `json:"hosts"`
}

type operatingSystem struct {
	GOOS         string `json:"goos"`
	GOARCH       string `json:"goarch"`
	PackageBuild string `json:"package_build"`
	HostRuntime  string `json:"host_runtime"`
}

type compatibilityEntry struct {
	Host                  string            `json:"host"`
	HostProduct           string            `json:"host_product"`
	DeclaredSurface       string            `json:"declared_surface"`
	AdapterFixtureVersion string            `json:"adapter_fixture_version"`
	QualificationTarget   string            `json:"qualification_target"`
	OperatingSystems      []operatingSystem `json:"operating_systems"`
	RequiredCapabilities  []string          `json:"required_capabilities"`
	OptionalCapabilities  []string          `json:"optional_capabilities"`
	DegradedBehavior      string            `json:"degraded_behavior"`
	ProviderExecution     string            `json:"provider_execution"`
	ActualHostLifecycle   string            `json:"actual_host_lifecycle"`
	Rollback              string            `json:"rollback"`
	SupportClaim          string            `json:"support_claim"`
}

type compatibilityMatrix struct {
	Schema         int                  `json:"schema"`
	ArtifactSchema string               `json:"artifact_schema"`
	Entries        []compatibilityEntry `json:"entries"`
}

type loadedInputs struct {
	Root             string
	Descriptor       packageDescriptor
	DescriptorBytes  []byte
	Compatibility    compatibilityMatrix
	CompatibilityRaw []byte
	License          []byte
	Changelog        []byte
	Skills           map[string][]byte
}

type builtPackage struct {
	Host          string
	Version       string
	Entries       []archiveEntry
	Archive       []byte
	ArchiveDigest string
	ArchiveName   string
	CatalogPath   string
	Catalog       []byte
}

type codexInterface struct {
	DisplayName      string   `json:"displayName"`
	ShortDescription string   `json:"shortDescription"`
	LongDescription  string   `json:"longDescription"`
	DeveloperName    string   `json:"developerName"`
	Category         string   `json:"category"`
	DefaultPrompt    []string `json:"defaultPrompt"`
}

type codexManifest struct {
	Name        string         `json:"name"`
	Version     string         `json:"version"`
	Description string         `json:"description"`
	Author      author         `json:"author"`
	Homepage    string         `json:"homepage"`
	Repository  string         `json:"repository"`
	License     string         `json:"license"`
	Keywords    []string       `json:"keywords"`
	Skills      string         `json:"skills"`
	Interface   codexInterface `json:"interface"`
}

type claudeManifest struct {
	Schema      string   `json:"$schema"`
	Name        string   `json:"name"`
	DisplayName string   `json:"displayName"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      author   `json:"author"`
	Homepage    string   `json:"homepage"`
	Repository  string   `json:"repository"`
	License     string   `json:"license"`
	Keywords    []string `json:"keywords"`
}

type rootManifest struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      author   `json:"author"`
	Homepage    string   `json:"homepage"`
	Repository  string   `json:"repository"`
	License     string   `json:"license"`
	Keywords    []string `json:"keywords"`
	Skills      string   `json:"skills"`
}

type legacyMarketplacePlugin struct {
	Name        string   `json:"name"`
	Source      string   `json:"source"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Author      author   `json:"author"`
	Repository  string   `json:"repository"`
	License     string   `json:"license"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Strict      bool     `json:"strict"`
}

type legacyMarketplace struct {
	Name    string                    `json:"name"`
	Owner   author                    `json:"owner"`
	Plugins []legacyMarketplacePlugin `json:"plugins"`
}

type compatibilityProjection struct {
	Schema         int                `json:"schema"`
	PackageVersion string             `json:"package_version"`
	ArtifactSchema string             `json:"artifact_schema"`
	Entry          compatibilityEntry `json:"entry"`
}

type permissionsProjection struct {
	Schema         int         `json:"schema"`
	PackageVersion string      `json:"package_version"`
	Host           string      `json:"host"`
	Permissions    permissions `json:"permissions"`
	SupportClaim   string      `json:"support_claim"`
}

type distributionMetadata struct {
	Schema         int    `json:"schema"`
	Name           string `json:"name"`
	Version        string `json:"version"`
	Channel        string `json:"channel"`
	Host           string `json:"host"`
	ManifestPath   string `json:"manifest_path"`
	CatalogPath    string `json:"catalog_path"`
	SourceDigest   string `json:"source_digest"`
	Builder        string `json:"builder"`
	SupportClaim   string `json:"support_claim"`
	ActualHostGate string `json:"actual_host_gate"`
}

type inventoryEntry struct {
	Path   string `json:"path"`
	Mode   string `json:"mode"`
	Size   int    `json:"size"`
	SHA256 string `json:"sha256"`
}

type inventory struct {
	Schema int              `json:"schema"`
	Scope  string           `json:"scope"`
	Files  []inventoryEntry `json:"files"`
}

type sbomChecksum struct {
	Algorithm     string `json:"algorithm"`
	ChecksumValue string `json:"checksumValue"`
}

type sbomPackage struct {
	Name             string         `json:"name"`
	SPDXID           string         `json:"SPDXID"`
	VersionInfo      string         `json:"versionInfo"`
	DownloadLocation string         `json:"downloadLocation"`
	FilesAnalyzed    bool           `json:"filesAnalyzed"`
	LicenseConcluded string         `json:"licenseConcluded"`
	LicenseDeclared  string         `json:"licenseDeclared"`
	CopyrightText    string         `json:"copyrightText"`
	Checksums        []sbomChecksum `json:"checksums"`
}

type sbomCreationInfo struct {
	Created  string   `json:"created"`
	Creators []string `json:"creators"`
}

type sbomDocument struct {
	SPDXVersion       string           `json:"spdxVersion"`
	DataLicense       string           `json:"dataLicense"`
	SPDXID            string           `json:"SPDXID"`
	Name              string           `json:"name"`
	DocumentNamespace string           `json:"documentNamespace"`
	CreationInfo      sbomCreationInfo `json:"creationInfo"`
	Packages          []sbomPackage    `json:"packages"`
}

type provenanceInput struct {
	Schema         int      `json:"schema"`
	Unsigned       bool     `json:"unsigned"`
	Package        string   `json:"package"`
	Version        string   `json:"version"`
	Host           string   `json:"host"`
	SourceDigest   string   `json:"source_digest"`
	Builder        string   `json:"builder"`
	Recipe         string   `json:"recipe"`
	ExternalInputs []string `json:"external_inputs"`
	Claim          string   `json:"claim"`
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(arguments []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("distribution", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "repository root")
	output := flags.String("output", "", "exact build/distributions output path")
	check := flags.Bool("check", false, "verify source drift, archives, and fixture lifecycle without persistent output")
	if err := flags.Parse(arguments); err != nil {
		return 2
	}
	if flags.NArg() != 0 || (*check && *output != "") || (!*check && *output == "") {
		fmt.Fprintln(stderr, "distribution: use exactly one of --check or --output build/distributions")
		return 2
	}

	inputs, err := loadInputs(*root)
	if err != nil {
		fmt.Fprintf(stderr, "distribution: %v\n", err)
		return 1
	}
	if err := checkGeneratedFiles(inputs); err != nil {
		fmt.Fprintf(stderr, "distribution: %v\n", err)
		return 1
	}
	packages, err := buildPackages(inputs)
	if err != nil {
		fmt.Fprintf(stderr, "distribution: %v\n", err)
		return 1
	}

	if *check {
		second, secondErr := buildPackages(inputs)
		if secondErr != nil {
			fmt.Fprintf(stderr, "distribution: second build: %v\n", secondErr)
			return 1
		}
		if err := compareBuilds(packages, second); err != nil {
			fmt.Fprintf(stderr, "distribution: reproducibility: %v\n", err)
			return 1
		}
		if err := qualifyLifecycleSet(packages); err != nil {
			fmt.Fprintf(stderr, "distribution: lifecycle qualification: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "distribution-check: PASS packages=%d version=%s codex=%s claude=%s actual_host=NOT_RUN\n",
			len(packages), inputs.Descriptor.Version, packages[0].ArchiveDigest, packages[1].ArchiveDigest)
		return 0
	}

	if err := writeOutputs(inputs.Root, *output, inputs.Descriptor.Name, packages); err != nil {
		fmt.Fprintf(stderr, "distribution: write output: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "distribution-build: PASS output=build/distributions packages=%d version=%s actual_host=NOT_RUN\n", len(packages), inputs.Descriptor.Version)
	return 0
}

func loadInputs(root string) (loadedInputs, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return loadedInputs{}, fmt.Errorf("resolve root: %w", err)
	}
	absolute, err = filepath.EvalSymlinks(absolute)
	if err != nil {
		return loadedInputs{}, fmt.Errorf("resolve physical root: %w", err)
	}
	info, err := os.Lstat(absolute)
	if err != nil || !info.IsDir() {
		return loadedInputs{}, errors.New("repository root is not a physical directory")
	}

	descriptorRaw, err := readRegularBounded(absolute, "distribution/package.json")
	if err != nil {
		return loadedInputs{}, err
	}
	var descriptor packageDescriptor
	if err := decodeStrict(descriptorRaw, &descriptor); err != nil {
		return loadedInputs{}, fmt.Errorf("decode distribution/package.json: %w", err)
	}
	if err := validateDescriptor(descriptor); err != nil {
		return loadedInputs{}, err
	}
	canonicalDescriptor, err := jsonBytes(descriptor)
	if err != nil || !bytes.Equal(descriptorRaw, canonicalDescriptor) {
		return loadedInputs{}, errors.New("distribution/package.json is not canonical generated JSON")
	}

	compatibilityRaw, err := readRegularBounded(absolute, "distribution/compatibility.json")
	if err != nil {
		return loadedInputs{}, err
	}
	var compatibility compatibilityMatrix
	if err := decodeStrict(compatibilityRaw, &compatibility); err != nil {
		return loadedInputs{}, fmt.Errorf("decode distribution/compatibility.json: %w", err)
	}
	if err := validateCompatibility(compatibility); err != nil {
		return loadedInputs{}, err
	}
	canonicalCompatibility, err := jsonBytes(compatibility)
	if err != nil || !bytes.Equal(compatibilityRaw, canonicalCompatibility) {
		return loadedInputs{}, errors.New("distribution/compatibility.json is not canonical generated JSON")
	}

	license, err := readRegularBounded(absolute, "LICENSE")
	if err != nil {
		return loadedInputs{}, err
	}
	changelog, err := readRegularBounded(absolute, "CHANGELOG.md")
	if err != nil {
		return loadedInputs{}, err
	}
	if !bytes.Contains(license, []byte("MIT License")) || !bytes.Contains(changelog, []byte(descriptor.Version)) || !bytes.Contains(changelog, []byte("Unreleased")) {
		return loadedInputs{}, errors.New("license or changelog does not bind the development package identity")
	}

	skillBytes := make(map[string][]byte, len(descriptor.Skills))
	for _, skill := range descriptor.Skills {
		relative := filepath.ToSlash(filepath.Join("skills", skill, "SKILL.md"))
		data, readErr := readRegularBounded(absolute, relative)
		if readErr != nil {
			return loadedInputs{}, readErr
		}
		if !bytes.Contains(data, []byte("name: "+skill+"\n")) || !bytes.Contains(data, []byte("user-invocable: true\n")) {
			return loadedInputs{}, fmt.Errorf("skill %s lacks its exact public identity", skill)
		}
		skillBytes[relative] = data
	}

	return loadedInputs{
		Root: absolute, Descriptor: descriptor, DescriptorBytes: descriptorRaw,
		Compatibility: compatibility, CompatibilityRaw: compatibilityRaw,
		License: license, Changelog: changelog, Skills: skillBytes,
	}, nil
}

func validateDescriptor(descriptor packageDescriptor) error {
	if descriptor.Schema != 1 || !packageNamePattern.MatchString(descriptor.Name) || !versionPattern.MatchString(descriptor.Version) || descriptor.Channel != "development" {
		return errors.New("package schema, name, prerelease version, or channel is invalid")
	}
	if descriptor.Description == "" || descriptor.Author.Name == "" || descriptor.License != "MIT" ||
		!strings.HasPrefix(descriptor.Homepage, "https://github.com/addressanup/level7-dev-loop") || descriptor.Repository != "https://github.com/addressanup/level7-dev-loop" {
		return errors.New("package publisher metadata is incomplete or unexpected")
	}
	if len(descriptor.Keywords) == 0 || !slices.IsSorted(descriptor.Skills) || len(descriptor.Skills) != 12 {
		return errors.New("package keyword or skill inventory is invalid")
	}
	seen := make(map[string]bool, len(descriptor.Skills))
	for _, skill := range descriptor.Skills {
		if !packageNamePattern.MatchString(skill) || seen[skill] {
			return errors.New("package skill inventory contains an invalid or duplicate name")
		}
		seen[skill] = true
	}
	permission := descriptor.Permissions
	if permission.Level7Network || permission.BundledExecutable || permission.MCPServer || permission.Hook || permission.HostSetting || permission.Telemetry || strings.TrimSpace(permission.WorkspaceBoundary) == "" {
		return errors.New("development package permissions exceed the approved inert boundary")
	}
	if err := validateHost("codex", descriptor.Hosts.Codex, ".codex-plugin/plugin.json", ".agents/plugins/marketplace.json"); err != nil {
		return err
	}
	return validateHost("claude", descriptor.Hosts.Claude, ".claude-plugin/plugin.json", ".claude-plugin/marketplace.json")
}

func validateHost(name string, host hostDescriptor, manifest, catalog string) error {
	if host.ManifestPath != manifest || host.CatalogPath != catalog || host.DisplayName == "" || host.Category == "" || host.Invocation == "" || strings.ContainsAny(host.Invocation, "\r\n\x00") {
		return fmt.Errorf("%s host descriptor is invalid", name)
	}
	return nil
}

func validateCompatibility(matrix compatibilityMatrix) error {
	if matrix.Schema != 1 || matrix.ArtifactSchema != "lean-risk-v1" || len(matrix.Entries) != 2 || matrix.Entries[0].Host != "codex" || matrix.Entries[1].Host != "claude" {
		return errors.New("compatibility matrix identity or host ordering is invalid")
	}
	for _, entry := range matrix.Entries {
		if entry.HostProduct == "" || entry.DeclaredSurface == "" || entry.AdapterFixtureVersion == "" || entry.QualificationTarget == "" ||
			len(entry.OperatingSystems) != 2 || len(entry.RequiredCapabilities) == 0 || entry.ProviderExecution != "NOT_RUN" ||
			entry.ActualHostLifecycle != "NOT_RUN" || entry.SupportClaim != "WITHHELD" || entry.DegradedBehavior == "" || entry.Rollback == "" {
			return fmt.Errorf("compatibility entry %s makes an incomplete or promoted claim", entry.Host)
		}
		if entry.OperatingSystems[0].GOOS != "darwin" || entry.OperatingSystems[0].GOARCH != "arm64" ||
			entry.OperatingSystems[1].GOOS != "darwin" || entry.OperatingSystems[1].GOARCH != "amd64" {
			return fmt.Errorf("compatibility entry %s has an unexpected platform matrix", entry.Host)
		}
		for _, operatingSystem := range entry.OperatingSystems {
			if operatingSystem.PackageBuild != "TESTED" || operatingSystem.HostRuntime != "NOT_RUN" {
				return fmt.Errorf("compatibility entry %s has an invalid evidence state", entry.Host)
			}
		}
	}
	return nil
}

func renderRootFiles(descriptor packageDescriptor) (map[string][]byte, error) {
	interfaceMetadata := codexInterface{
		DisplayName: descriptor.Hosts.Codex.DisplayName, ShortDescription: "Risk-tiered build, test, review, and ship loop",
		LongDescription: "Development-only skills for a lean risk-tiered build, test, review, and merge workflow. No provider support, installation, or release claim is implied.",
		DeveloperName:   descriptor.Author.Name, Category: descriptor.Hosts.Codex.Category,
		DefaultPrompt: []string{
			"Run l7-next and tell me the current project phase",
			"Start a post-launch feature through l7-change",
			"Audit the current release candidate with l7-release",
		},
	}
	codex := codexManifest{
		Name: descriptor.Name, Version: descriptor.Version, Description: descriptor.Description, Author: descriptor.Author,
		Homepage: descriptor.Homepage, Repository: descriptor.Repository, License: descriptor.License,
		Keywords: append([]string{}, descriptor.Keywords...), Skills: "./skills/", Interface: interfaceMetadata,
	}
	claude := claudeManifest{
		Schema: "https://json.schemastore.org/claude-code-plugin-manifest.json", Name: descriptor.Name,
		DisplayName: descriptor.Hosts.Claude.DisplayName, Version: descriptor.Version, Description: descriptor.Description,
		Author: descriptor.Author, Homepage: descriptor.Homepage, Repository: descriptor.Repository,
		License: descriptor.License, Keywords: append([]string{}, descriptor.Keywords...),
	}
	root := rootManifest{
		Name: descriptor.Name, Version: descriptor.Version, Description: descriptor.Description, Author: descriptor.Author,
		Homepage: descriptor.Homepage, Repository: descriptor.Repository, License: descriptor.License,
		Keywords: append([]string{}, descriptor.Keywords...), Skills: "./skills/",
	}
	marketplace := legacyMarketplace{
		Name: "level7-engineering-development", Owner: descriptor.Author,
		Plugins: []legacyMarketplacePlugin{{
			Name: descriptor.Name, Source: ".", Description: descriptor.Description, Version: descriptor.Version,
			Author: descriptor.Author, Repository: descriptor.Repository, License: descriptor.License,
			Category: descriptor.Hosts.Claude.Category, Tags: []string{"workflow", "risk-tier"}, Strict: true,
		}},
	}
	values := map[string]any{
		".codex-plugin/plugin.json":  codex,
		".claude-plugin/plugin.json": claude,
		"plugin.json":                root,
		"marketplace.json":           marketplace,
	}
	result := make(map[string][]byte, len(values))
	for path, value := range values {
		data, err := jsonBytes(value)
		if err != nil {
			return nil, fmt.Errorf("render %s: %w", path, err)
		}
		result[path] = data
	}
	return result, nil
}

func checkGeneratedFiles(inputs loadedInputs) error {
	expected, err := renderRootFiles(inputs.Descriptor)
	if err != nil {
		return err
	}
	for relative, want := range expected {
		got, readErr := readRegularBounded(inputs.Root, relative)
		if readErr != nil {
			return readErr
		}
		if !bytes.Equal(got, want) {
			return fmt.Errorf("generated file drift: %s", relative)
		}
	}
	return nil
}

func buildPackages(inputs loadedInputs) ([]builtPackage, error) {
	rendered, err := renderRootFiles(inputs.Descriptor)
	if err != nil {
		return nil, err
	}
	packages := make([]builtPackage, 0, 2)
	for _, host := range []string{"codex", "claude"} {
		entry, descriptor, err := hostInputs(inputs, host)
		if err != nil {
			return nil, err
		}
		files := map[string][]byte{
			descriptor.ManifestPath: rendered[descriptor.ManifestPath],
			"LICENSE":               append([]byte{}, inputs.License...),
			"CHANGELOG.md":          append([]byte{}, inputs.Changelog...),
		}
		for relative, data := range inputs.Skills {
			files[relative] = append([]byte{}, data...)
		}
		compatibility, err := jsonBytes(compatibilityProjection{
			Schema: 1, PackageVersion: inputs.Descriptor.Version,
			ArtifactSchema: inputs.Compatibility.ArtifactSchema, Entry: entry,
		})
		if err != nil {
			return nil, err
		}
		permissionData, err := jsonBytes(permissionsProjection{
			Schema: 1, PackageVersion: inputs.Descriptor.Version, Host: host,
			Permissions: inputs.Descriptor.Permissions, SupportClaim: "WITHHELD",
		})
		if err != nil {
			return nil, err
		}
		files["COMPATIBILITY.json"] = compatibility
		files["PERMISSIONS.json"] = permissionData

		sourceDigest := digestSource(inputs, host, files)
		distribution, err := jsonBytes(distributionMetadata{
			Schema: 1, Name: inputs.Descriptor.Name, Version: inputs.Descriptor.Version,
			Channel: inputs.Descriptor.Channel, Host: host, ManifestPath: descriptor.ManifestPath,
			CatalogPath: descriptor.CatalogPath, SourceDigest: sourceDigest, Builder: builderVersion,
			SupportClaim: "WITHHELD", ActualHostGate: "NOT_RUN",
		})
		if err != nil {
			return nil, err
		}
		files["DISTRIBUTION.json"] = distribution
		sbom, err := jsonBytes(makeSBOM(inputs.Descriptor, host, sourceDigest))
		if err != nil {
			return nil, err
		}
		files["SBOM.spdx.json"] = sbom
		provenance, err := jsonBytes(provenanceInput{
			Schema: 1, Unsigned: true, Package: inputs.Descriptor.Name, Version: inputs.Descriptor.Version,
			Host: host, SourceDigest: sourceDigest, Builder: builderVersion,
			Recipe:         "offline deterministic standard-library package assembly",
			ExternalInputs: []string{}, Claim: "development input only; authenticity and release promotion are not established",
		})
		if err != nil {
			return nil, err
		}
		files["PROVENANCE.input.json"] = provenance

		fileEntries := entriesFromMap(files)
		inventoryData, err := jsonBytes(makeInventory(fileEntries))
		if err != nil {
			return nil, err
		}
		fileEntries = append(fileEntries, archiveEntry{Name: "INVENTORY.json", Data: inventoryData, Mode: 0o644})
		sort.Slice(fileEntries, func(i, j int) bool { return fileEntries[i].Name < fileEntries[j].Name })
		archive, err := createArchive(fileEntries)
		if err != nil {
			return nil, fmt.Errorf("build %s archive: %w", host, err)
		}
		if err := validateArchive(archive, fileEntries); err != nil {
			return nil, fmt.Errorf("validate %s archive: %w", host, err)
		}
		catalog, err := renderCatalog(inputs.Descriptor, host)
		if err != nil {
			return nil, err
		}
		digest := sha256Hex(archive)
		packages = append(packages, builtPackage{
			Host: host, Version: inputs.Descriptor.Version, Entries: fileEntries, Archive: archive,
			ArchiveDigest: digest,
			ArchiveName:   fmt.Sprintf("%s-%s-%s.zip", inputs.Descriptor.Name, host, inputs.Descriptor.Version),
			CatalogPath:   descriptor.CatalogPath, Catalog: catalog,
		})
	}
	return packages, nil
}

func hostInputs(inputs loadedInputs, host string) (compatibilityEntry, hostDescriptor, error) {
	for _, entry := range inputs.Compatibility.Entries {
		if entry.Host != host {
			continue
		}
		if host == "codex" {
			return entry, inputs.Descriptor.Hosts.Codex, nil
		}
		if host == "claude" {
			return entry, inputs.Descriptor.Hosts.Claude, nil
		}
	}
	return compatibilityEntry{}, hostDescriptor{}, fmt.Errorf("missing host %s", host)
}

func digestSource(inputs loadedInputs, host string, files map[string][]byte) string {
	values := map[string][]byte{
		"distribution/package.json":       inputs.DescriptorBytes,
		"distribution/compatibility.json": inputs.CompatibilityRaw,
		"host":                            []byte(host + "\n"),
	}
	for name, data := range files {
		values[name] = data
	}
	return digestNamedBytes(values)
}

func makeInventory(entries []archiveEntry) inventory {
	files := make([]inventoryEntry, 0, len(entries))
	for _, entry := range entries {
		files = append(files, inventoryEntry{
			Path: entry.Name, Mode: "0644", Size: len(entry.Data), SHA256: sha256Hex(entry.Data),
		})
	}
	return inventory{Schema: 1, Scope: "all archive files except INVENTORY.json", Files: files}
}

func makeSBOM(descriptor packageDescriptor, host, sourceDigest string) sbomDocument {
	return sbomDocument{
		SPDXVersion: "SPDX-2.3", DataLicense: "CC0-1.0", SPDXID: "SPDXRef-DOCUMENT",
		Name:              descriptor.Name + "-" + host + "-" + descriptor.Version,
		DocumentNamespace: descriptor.Repository + "/sbom/" + host + "/" + sourceDigest,
		CreationInfo:      sbomCreationInfo{Created: "1980-01-01T00:00:00Z", Creators: []string{"Tool: " + builderVersion}},
		Packages: []sbomPackage{{
			Name: descriptor.Name + "-" + host, SPDXID: "SPDXRef-Package", VersionInfo: descriptor.Version,
			DownloadLocation: "NOASSERTION", FilesAnalyzed: false, LicenseConcluded: descriptor.License,
			LicenseDeclared: descriptor.License, CopyrightText: "Copyright (c) 2026 Level 7 Engineering",
			Checksums: []sbomChecksum{{Algorithm: "SHA256", ChecksumValue: sourceDigest}},
		}},
	}
}

func renderCatalog(descriptor packageDescriptor, host string) ([]byte, error) {
	if host == "codex" {
		type source struct {
			Source string `json:"source"`
			Path   string `json:"path"`
		}
		type policy struct {
			Installation   string `json:"installation"`
			Authentication string `json:"authentication"`
		}
		type plugin struct {
			Name     string `json:"name"`
			Source   source `json:"source"`
			Policy   policy `json:"policy"`
			Category string `json:"category"`
		}
		type interfaceMetadata struct {
			DisplayName string `json:"displayName"`
		}
		type catalog struct {
			Name      string            `json:"name"`
			Interface interfaceMetadata `json:"interface"`
			Plugins   []plugin          `json:"plugins"`
		}
		return jsonBytes(catalog{
			Name: "level7-engineering-development", Interface: interfaceMetadata{DisplayName: "Level 7 Engineering (Development)"},
			Plugins: []plugin{{
				Name: descriptor.Name, Source: source{Source: "local", Path: "./plugins/" + descriptor.Name},
				Policy: policy{Installation: "AVAILABLE", Authentication: "ON_INSTALL"}, Category: descriptor.Hosts.Codex.Category,
			}},
		})
	}
	if host == "claude" {
		type plugin struct {
			Name        string `json:"name"`
			Source      string `json:"source"`
			Description string `json:"description"`
			Version     string `json:"version"`
			License     string `json:"license"`
			Category    string `json:"category"`
			Strict      bool   `json:"strict"`
		}
		type catalog struct {
			Name    string   `json:"name"`
			Owner   author   `json:"owner"`
			Plugins []plugin `json:"plugins"`
		}
		return jsonBytes(catalog{
			Name: "level7-engineering-development", Owner: descriptor.Author,
			Plugins: []plugin{{
				Name: descriptor.Name, Source: "./plugins/" + descriptor.Name, Description: descriptor.Description,
				Version: descriptor.Version, License: descriptor.License, Category: descriptor.Hosts.Claude.Category, Strict: true,
			}},
		})
	}
	return nil, fmt.Errorf("unknown host %q", host)
}

func compareBuilds(first, second []builtPackage) error {
	if len(first) != len(second) {
		return errors.New("package count changed")
	}
	for index := range first {
		if first[index].Host != second[index].Host || first[index].ArchiveDigest != second[index].ArchiveDigest ||
			!bytes.Equal(first[index].Archive, second[index].Archive) || !bytes.Equal(first[index].Catalog, second[index].Catalog) {
			return fmt.Errorf("%s package bytes changed between clean builds", first[index].Host)
		}
	}
	return nil
}

func writeOutputs(root, requested, name string, packages []builtPackage) error {
	requestedAbsolute, err := filepath.Abs(requested)
	if err != nil {
		return err
	}
	expected := filepath.Join(root, "build", "distributions")
	if filepath.Clean(requestedAbsolute) != expected {
		return fmt.Errorf("output must be exact %s", expected)
	}
	buildRoot := filepath.Dir(expected)
	if err := ensurePhysicalDirectory(root, "build", true); err != nil {
		return err
	}
	temporary, err := os.MkdirTemp(buildRoot, ".distributions-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temporary)

	for _, built := range packages {
		marketplaceRoot := filepath.Join(temporary, built.Host+"-marketplace")
		pluginRoot := filepath.Join(marketplaceRoot, "plugins", name)
		for _, entry := range built.Entries {
			if err := writeRegular(pluginRoot, entry.Name, entry.Data); err != nil {
				return err
			}
		}
		if err := writeRegular(marketplaceRoot, built.CatalogPath, built.Catalog); err != nil {
			return err
		}
		if err := writeRegular(temporary, built.ArchiveName, built.Archive); err != nil {
			return err
		}
		checksum := []byte(built.ArchiveDigest + "  " + built.ArchiveName + "\n")
		if err := writeRegular(temporary, built.ArchiveName+".sha256", checksum); err != nil {
			return err
		}
	}
	if info, statErr := os.Lstat(expected); statErr == nil {
		if !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
			return errors.New("existing distribution output is not a physical directory")
		}
		if err := os.RemoveAll(expected); err != nil {
			return err
		}
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return statErr
	}
	if err := os.Rename(temporary, expected); err != nil {
		return err
	}
	return nil
}

func writeRegular(root, relative string, data []byte) error {
	clean, err := cleanRelativePath(relative)
	if err != nil {
		return fmt.Errorf("unsafe output path %q: %w", relative, err)
	}
	target := filepath.Join(root, clean)
	within, err := filepath.Rel(root, target)
	if err != nil || within == ".." || strings.HasPrefix(within, ".."+string(filepath.Separator)) {
		return fmt.Errorf("output path escapes root: %q", relative)
	}
	if err := ensurePhysicalDirectory(root, filepath.Dir(clean), true); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(target), ".write-")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Chmod(temporaryName, 0o644); err != nil {
		return err
	}
	return os.Rename(temporaryName, target)
}

func readRegularBounded(root, relative string) ([]byte, error) {
	clean, err := cleanRelativePath(relative)
	if err != nil {
		return nil, fmt.Errorf("unsafe source path %q: %w", relative, err)
	}
	if err := ensurePhysicalDirectory(root, filepath.Dir(clean), false); err != nil {
		return nil, fmt.Errorf("read %s: %w", relative, err)
	}
	name := filepath.Join(root, clean)
	info, err := os.Lstat(name)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", relative, err)
	}
	if !info.Mode().IsRegular() || info.Mode()&fs.ModeSymlink != 0 || info.Size() < 1 || info.Size() > maximumFileSize {
		return nil, fmt.Errorf("source %s is not a bounded regular file", relative)
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", relative, err)
	}
	if len(data) == 0 || len(data) > maximumFileSize || bytes.IndexByte(data, 0) >= 0 {
		return nil, fmt.Errorf("source %s has invalid content", relative)
	}
	return data, nil
}

func cleanRelativePath(relative string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(relative))
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || strings.ContainsRune(clean, 0) {
		return "", errors.New("path is not a bounded relative name")
	}
	return clean, nil
}

func ensurePhysicalDirectory(root, relative string, create bool) error {
	if create {
		if err := os.MkdirAll(root, 0o755); err != nil {
			return err
		}
	}
	rootInfo, err := os.Lstat(root)
	if err != nil {
		return err
	}
	if !rootInfo.IsDir() || rootInfo.Mode()&fs.ModeSymlink != 0 {
		return errors.New("path root is not a physical directory")
	}
	if relative == "" || relative == "." {
		return nil
	}
	clean, err := cleanRelativePath(relative)
	if err != nil {
		return err
	}
	current := root
	for _, element := range strings.Split(clean, string(filepath.Separator)) {
		current = filepath.Join(current, element)
		info, statErr := os.Lstat(current)
		if errors.Is(statErr, os.ErrNotExist) && create {
			if mkdirErr := os.Mkdir(current, 0o755); mkdirErr != nil {
				return mkdirErr
			}
			info, statErr = os.Lstat(current)
		}
		if statErr != nil {
			return statErr
		}
		if !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
			return fmt.Errorf("path component is not a physical directory: %s", current)
		}
	}
	return nil
}

func decodeStrict(data []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("JSON contains trailing content")
	}
	return nil
}

func jsonBytes(value any) ([]byte, error) {
	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func entriesFromMap(files map[string][]byte) []archiveEntry {
	entries := make([]archiveEntry, 0, len(files))
	for name, data := range files {
		entries = append(entries, archiveEntry{Name: name, Data: append([]byte{}, data...), Mode: 0o644})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	return entries
}

func digestNamedBytes(values map[string][]byte) string {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	hash := sha256.New()
	for _, name := range names {
		hash.Write([]byte(name))
		hash.Write([]byte{0})
		hash.Write(values[name])
		hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func sha256Hex(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

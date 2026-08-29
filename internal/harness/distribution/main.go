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
	packageNamePattern     = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	versionPattern         = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+-[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*$`)
	sha256Pattern          = regexp.MustCompile(`^[0-9a-f]{64}$`)
	lifecycleSyncDirectory = syncPhysicalDirectory
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
	expected := expectedCompatibilityEntries()
	if matrix.Schema != 1 || matrix.ArtifactSchema != "lean-risk-v1" || len(matrix.Entries) != len(expected) {
		return errors.New("compatibility matrix identity or host ordering is invalid")
	}
	for index := range expected {
		if !equalCompatibilityEntry(matrix.Entries[index], expected[index]) {
			return fmt.Errorf("compatibility entry %d does not match the exact %s development boundary", index, expected[index].Host)
		}
	}
	return nil
}

func expectedCompatibilityEntries() []compatibilityEntry {
	platforms := []operatingSystem{
		{GOOS: "darwin", GOARCH: "arm64", PackageBuild: "TESTED", HostRuntime: "NOT_RUN"},
		{GOOS: "darwin", GOARCH: "amd64", PackageBuild: "TESTED", HostRuntime: "NOT_RUN"},
	}
	const (
		declaredSurface = "development package layout and local marketplace metadata"
		degraded        = "Missing, mismatched, or unobserved host behavior withholds support and permits no provider execution claim."
		rollback        = "Preview package-manager-owned state and remove only exact Level 7-owned development package bytes."
	)
	return []compatibilityEntry{
		{
			Host: "codex", HostProduct: "Codex CLI", DeclaredSurface: declaredSurface,
			AdapterFixtureVersion: "codex-cli 0.149.1", QualificationTarget: "codex-cli 0.150.1",
			OperatingSystems:     append([]operatingSystem{}, platforms...),
			RequiredCapabilities: []string{".codex-plugin/plugin.json", "skills/<name>/SKILL.md"},
			OptionalCapabilities: []string{}, DegradedBehavior: degraded, ProviderExecution: "NOT_RUN",
			ActualHostLifecycle: "NOT_RUN", Rollback: rollback, SupportClaim: "WITHHELD",
		},
		{
			Host: "claude", HostProduct: "Claude Code", DeclaredSurface: declaredSurface,
			AdapterFixtureVersion: "2.1.241", QualificationTarget: "2.1.247 (Claude Code)",
			OperatingSystems:     append([]operatingSystem{}, platforms...),
			RequiredCapabilities: []string{".claude-plugin/plugin.json", "skills/<name>/SKILL.md"},
			OptionalCapabilities: []string{}, DegradedBehavior: degraded, ProviderExecution: "NOT_RUN",
			ActualHostLifecycle: "NOT_RUN", Rollback: rollback, SupportClaim: "WITHHELD",
		},
	}
}

func equalCompatibilityEntry(first, second compatibilityEntry) bool {
	return first.Host == second.Host &&
		first.HostProduct == second.HostProduct &&
		first.DeclaredSurface == second.DeclaredSurface &&
		first.AdapterFixtureVersion == second.AdapterFixtureVersion &&
		first.QualificationTarget == second.QualificationTarget &&
		slices.Equal(first.OperatingSystems, second.OperatingSystems) &&
		slices.Equal(first.RequiredCapabilities, second.RequiredCapabilities) &&
		(first.OptionalCapabilities == nil) == (second.OptionalCapabilities == nil) &&
		slices.Equal(first.OptionalCapabilities, second.OptionalCapabilities) &&
		first.DegradedBehavior == second.DegradedBehavior &&
		first.ProviderExecution == second.ProviderExecution &&
		first.ActualHostLifecycle == second.ActualHostLifecycle &&
		first.Rollback == second.Rollback &&
		first.SupportClaim == second.SupportClaim
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

func writeDurableRegular(root, relative string, data []byte) (result error) {
	clean, err := cleanRelativePath(relative)
	if err != nil {
		return fmt.Errorf("unsafe output path %q: %w", relative, err)
	}
	target := filepath.Join(root, clean)
	within, err := filepath.Rel(root, target)
	if err != nil || within == ".." || strings.HasPrefix(within, ".."+string(filepath.Separator)) {
		return fmt.Errorf("output path escapes root: %q", relative)
	}
	if err := ensureDurableDirectory(root, filepath.Dir(clean)); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(target), ".write-")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	temporaryOpen := true
	temporaryLinked := true
	defer func() {
		if temporaryOpen {
			result = errors.Join(result, temporary.Close())
		}
		if temporaryLinked {
			temporaryRelative, relErr := filepath.Rel(root, temporaryName)
			if relErr == nil {
				relErr = removeDurably(root, temporaryRelative, true)
			}
			result = errors.Join(result, relErr)
		}
	}()
	if _, err := temporary.Write(data); err != nil {
		return err
	}
	if err := temporary.Chmod(0o644); err != nil {
		return err
	}
	if err := temporary.Sync(); err != nil {
		return err
	}
	if err := temporary.Close(); err != nil {
		temporaryOpen = false
		return err
	}
	temporaryOpen = false
	if err := os.Rename(temporaryName, target); err != nil {
		return err
	}
	temporaryLinked = false
	return lifecycleSyncDirectory(filepath.Dir(target))
}

func syncPhysicalDirectory(name string) error {
	before, err := os.Lstat(name)
	if err != nil {
		return fmt.Errorf("inspect directory for sync: %w", err)
	}
	if !before.IsDir() || before.Mode()&fs.ModeSymlink != 0 {
		return errors.New("sync target is not a physical directory")
	}
	root, err := os.OpenRoot(name)
	if err != nil {
		return fmt.Errorf("anchor directory for sync: %w", err)
	}
	opened, err := root.Stat(".")
	if err != nil || !opened.IsDir() || !os.SameFile(before, opened) {
		_ = root.Close()
		return errors.New("directory identity changed while anchoring sync")
	}
	directory, err := root.Open(".")
	if err != nil {
		_ = root.Close()
		return fmt.Errorf("open directory for sync: %w", err)
	}
	syncErr := directory.Sync()
	closeErr := directory.Close()
	linked, linkedErr := os.Lstat(name)
	current, currentErr := root.Stat(".")
	rootCloseErr := root.Close()
	if linkedErr != nil || currentErr != nil || linked.Mode()&fs.ModeSymlink != 0 || !linked.IsDir() ||
		!current.IsDir() || !os.SameFile(opened, linked) || !os.SameFile(opened, current) {
		err = errors.New("directory identity changed during sync")
	}
	if joined := errors.Join(syncErr, closeErr, err, rootCloseErr); joined != nil {
		return fmt.Errorf("sync physical directory %s: %w", name, joined)
	}
	return nil
}

func syncDirectoryTree(root, relative string) error {
	root = filepath.Clean(root)
	current := root
	if relative != "" && relative != "." {
		clean, err := cleanRelativePath(filepath.ToSlash(relative))
		if err != nil {
			return err
		}
		current = filepath.Join(root, clean)
	}
	within, err := filepath.Rel(root, current)
	if err != nil || within == ".." || strings.HasPrefix(within, ".."+string(filepath.Separator)) {
		return errors.New("directory sync tree escapes its root")
	}
	if err := ensurePhysicalDirectory(root, within, false); err != nil {
		return err
	}
	for {
		if err := lifecycleSyncDirectory(current); err != nil {
			return err
		}
		if current == root {
			return nil
		}
		current = filepath.Dir(current)
	}
}

func ensureDurableRoot(root string) error {
	root = filepath.Clean(root)
	current := root
	missing := make([]string, 0)
	for {
		info, err := os.Lstat(current)
		if err == nil {
			if !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
				return errors.New("path root ancestor is not a physical directory")
			}
			break
		}
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		parent := filepath.Dir(current)
		if parent == current {
			return errors.New("no physical ancestor exists for path root")
		}
		missing = append(missing, current)
		current = parent
	}
	if err := lifecycleSyncDirectory(current); err != nil {
		return err
	}
	anchorParent := filepath.Dir(current)
	if anchorParent != current {
		if err := lifecycleSyncDirectory(anchorParent); err != nil {
			return err
		}
	}
	for index := len(missing) - 1; index >= 0; index-- {
		created := missing[index]
		parent := filepath.Dir(created)
		if err := os.Mkdir(created, 0o755); err != nil {
			return err
		}
		info, err := os.Lstat(created)
		if err != nil || !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
			return errors.Join(errors.New("created path root is not a physical directory"), err)
		}
		if err := lifecycleSyncDirectory(created); err != nil {
			return err
		}
		if err := lifecycleSyncDirectory(parent); err != nil {
			return err
		}
	}
	return nil
}

func ensureDurableDirectory(root, relative string) error {
	root = filepath.Clean(root)
	if err := ensureDurableRoot(root); err != nil {
		return err
	}
	if relative == "" || relative == "." {
		return nil
	}
	clean, err := cleanRelativePath(filepath.ToSlash(relative))
	if err != nil {
		return err
	}
	current := root
	for _, element := range strings.Split(clean, string(filepath.Separator)) {
		current = filepath.Join(current, element)
		info, statErr := os.Lstat(current)
		if errors.Is(statErr, os.ErrNotExist) {
			if err := os.Mkdir(current, 0o755); err != nil {
				return err
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
	return syncDirectoryTree(root, clean)
}

func makeDurableDirectory(root, relative string) error {
	clean, err := cleanRelativePath(filepath.ToSlash(relative))
	if err != nil {
		return err
	}
	parent := filepath.Dir(clean)
	if err := ensureDurableDirectory(root, parent); err != nil {
		return err
	}
	name := filepath.Join(root, clean)
	if err := os.Mkdir(name, 0o755); err != nil {
		return err
	}
	return syncDirectoryTree(root, clean)
}

func syncRenameDirectories(root, sourceRelative, destinationRelative string) error {
	source, err := cleanRelativePath(filepath.ToSlash(sourceRelative))
	if err != nil {
		return err
	}
	destination, err := cleanRelativePath(filepath.ToSlash(destinationRelative))
	if err != nil {
		return err
	}
	destinationParent := filepath.Dir(destination)
	if err := syncDirectoryTree(root, destinationParent); err != nil {
		return err
	}
	sourceParent := filepath.Dir(source)
	if sourceParent != destinationParent {
		return syncDirectoryTree(root, sourceParent)
	}
	return nil
}

func renameDurably(root, sourceRelative, destinationRelative string) error {
	source, err := cleanRelativePath(filepath.ToSlash(sourceRelative))
	if err != nil {
		return err
	}
	destination, err := cleanRelativePath(filepath.ToSlash(destinationRelative))
	if err != nil {
		return err
	}
	if err := os.Rename(filepath.Join(root, source), filepath.Join(root, destination)); err != nil {
		return err
	}
	return syncRenameDirectories(root, source, destination)
}

func removeDurably(root, relative string, allowMissing bool) error {
	clean, err := cleanRelativePath(filepath.ToSlash(relative))
	if err != nil {
		return err
	}
	root = filepath.Clean(root)
	target := filepath.Join(root, clean)
	parent := filepath.Dir(target)
	info, err := os.Lstat(target)
	if errors.Is(err, os.ErrNotExist) {
		if !allowMissing {
			return err
		}
		return syncNearestExistingDirectory(root, parent)
	}
	if err != nil {
		return err
	}
	if info.Mode()&fs.ModeSymlink != 0 || (!info.Mode().IsRegular() && !info.IsDir()) {
		return errors.New("durable removal target is not a physical file or directory")
	}
	parentRelative, err := filepath.Rel(root, parent)
	if err != nil {
		return err
	}
	if err := ensurePhysicalDirectory(root, parentRelative, false); err != nil {
		return err
	}
	if info.IsDir() {
		if err := lifecycleSyncDirectory(target); err != nil {
			return err
		}
	}
	if err := os.Remove(target); err != nil {
		return err
	}
	return lifecycleSyncDirectory(parent)
}

func syncNearestExistingDirectory(root, directory string) error {
	root = filepath.Clean(root)
	current := filepath.Clean(directory)
	within, err := filepath.Rel(root, current)
	if err != nil || within == ".." || strings.HasPrefix(within, ".."+string(filepath.Separator)) {
		return errors.New("directory absence sync escapes its root")
	}
	for {
		info, statErr := os.Lstat(current)
		if statErr == nil {
			if !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
				return errors.New("directory absence sync reached a non-physical path")
			}
			relative, relErr := filepath.Rel(root, current)
			if relErr != nil {
				return relErr
			}
			return syncDirectoryTree(root, relative)
		}
		if !errors.Is(statErr, os.ErrNotExist) {
			return statErr
		}
		if current == root {
			return statErr
		}
		current = filepath.Dir(current)
	}
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

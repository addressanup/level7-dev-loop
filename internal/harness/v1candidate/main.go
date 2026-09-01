// Command v1candidate qualifies the offline, disposable package lifecycle for
// both Level 7 v1 host archives. It never reads or writes a real host root.
package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"debug/buildinfo"
	"debug/macho"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	stableVersion      = "0.1.1"
	candidateVersion   = "1.0.0-dev"
	releaseVersion     = "1.0.0"
	candidateChannel   = "development-candidate"
	releaseChannel     = "stable"
	v1MarketplaceName  = "level7-engineering-v1"
	maxArchiveBytes    = 256 << 20
	maxArchiveFiles    = 512
	maxFileBytes       = 128 << 20
	maxExpandedBytes   = 512 << 20
	maxProcessOutput   = 4 << 20
	networkDenyProfile = "(version 1)(allow default)(deny network*)"
)

type candidateIdentity struct {
	Version string
	Channel string
}

var canonicalSkills = []string{
	"l7-build", "l7-change", "l7-constitution", "l7-cyber", "l7-deploy", "l7-experience", "l7-geometry", "l7-greenfield",
	"l7-headless", "l7-next", "l7-onboard", "l7-ops", "l7-release", "l7-review", "l7-storybook", "l7-sync",
}

type archiveFile struct {
	Path   string
	Data   []byte
	Mode   fs.FileMode
	SHA256 string
}

type archivePackage struct {
	Host    string
	Version string
	Digest  string
	Files   []archiveFile
}

type declaredFile struct {
	Path   string `json:"path"`
	Mode   string `json:"mode,omitempty"`
	Size   int    `json:"size"`
	SHA256 string `json:"sha256"`
}

type ownedFile struct {
	Path   string `json:"path"`
	Mode   string `json:"mode"`
	Size   int    `json:"size"`
	SHA256 string `json:"sha256"`
}

type installedPackage struct {
	Version string      `json:"version"`
	Digest  string      `json:"digest"`
	Files   []ownedFile `json:"files"`
}

type lifecycleReceipt struct {
	Schema      int              `json:"schema"`
	Host        string           `json:"host"`
	Active      installedPackage `json:"active"`
	Previous    installedPackage `json:"previous"`
	StateSHA256 string           `json:"state_sha256"`
}

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(arguments []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("v1candidate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	candidateDirectory := flags.String("candidate", "", "directory containing v1 candidate archives")
	stableDirectory := flags.String("stable", "", "directory containing frozen v0.1.1 archives")
	workDirectory := flags.String("work", "", "empty disposable work root")
	candidateVersionFlag := flags.String("candidate-version", candidateVersion, "exact candidate package version")
	candidateChannelFlag := flags.String("candidate-channel", candidateChannel, "exact candidate package channel")
	if flags.Parse(arguments) != nil || flags.NArg() != 0 || *candidateDirectory == "" || *stableDirectory == "" || *workDirectory == "" {
		return 2
	}
	identity, err := validateCandidateIdentity(*candidateVersionFlag, *candidateChannelFlag)
	if err != nil {
		fmt.Fprintln(stderr, "v1candidate: candidate identity must be exactly 1.0.0-dev/development-candidate or 1.0.0/stable")
		return 2
	}
	if runtime.GOOS != "darwin" || (runtime.GOARCH != "arm64" && runtime.GOARCH != "amd64") {
		fmt.Fprintln(stderr, "v1candidate: native package execution requires macOS arm64 or amd64")
		return 1
	}
	candidateRoot, err := physicalDirectory(*candidateDirectory)
	if err != nil {
		fmt.Fprintln(stderr, "v1candidate: candidate directory is unsafe")
		return 1
	}
	stableRoot, err := physicalDirectory(*stableDirectory)
	if err != nil {
		fmt.Fprintln(stderr, "v1candidate: stable directory is unsafe")
		return 1
	}
	workRoot, err := physicalDirectory(*workDirectory)
	if err != nil || workRoot == candidateRoot || workRoot == stableRoot || isWithin(workRoot, candidateRoot) || isWithin(workRoot, stableRoot) {
		fmt.Fprintln(stderr, "v1candidate: disposable work directory is unsafe")
		return 1
	}

	for _, host := range []string{"codex", "claude"} {
		lifecycleRoot := filepath.Join(workRoot, "lifecycle-"+host)
		executionRoot := filepath.Join(workRoot, "execution-"+host)
		if pathExists(lifecycleRoot) || pathExists(executionRoot) {
			fmt.Fprintf(stderr, "v1candidate: %s disposable root is not empty\n", host)
			return 1
		}
		cleanup := func() error {
			return errors.Join(removeOwnedRoot(workRoot, lifecycleRoot), removeOwnedRoot(workRoot, executionRoot))
		}

		stableName := fmt.Sprintf("level7-dev-loop-%s-%s.zip", host, stableVersion)
		stableArchive, err := loadArchive(filepath.Join(stableRoot, stableName), host, stableVersion)
		if err == nil {
			err = validateStable(stableArchive)
		}
		if err == nil {
			err = verifySHA256Sidecar(filepath.Join(stableRoot, stableName+".sha256"), stableName, stableArchive.Digest)
		}
		candidateName := fmt.Sprintf("level7-dev-loop-%s-%s.zip", identity.Version, host)
		var candidateArchive archivePackage
		if err == nil {
			candidateArchive, err = loadArchive(filepath.Join(candidateRoot, candidateName), host, identity.Version)
		}
		if err == nil {
			err = validateCandidate(candidateArchive, identity)
		}
		if err == nil {
			err = installInitial(lifecycleRoot, stableArchive)
		}
		if err == nil {
			err = upgradeInstallation(lifecycleRoot, stableArchive.Digest, candidateArchive)
		}
		if err == nil {
			err = executeCandidate(lifecycleRoot, executionRoot, host, identity.Version)
		}
		if err == nil {
			err = rollbackInstallation(lifecycleRoot, candidateArchive.Digest, stableArchive.Digest)
		}
		if err == nil {
			err = verifyTree(filepath.Join(lifecycleRoot, "active"), receiptPackage(stableArchive))
		}
		if err == nil {
			err = removeInstallation(lifecycleRoot)
		}
		cleanupErr := cleanup()
		if err = errors.Join(err, cleanupErr); err != nil {
			fmt.Fprintf(stderr, "v1candidate: %s: %v\n", host, err)
			return 1
		}
		fmt.Fprintf(stdout, "%s stable_sha256=%s candidate_sha256=%s native=darwin/%s lifecycle=PASS cli=PASS mcp=PASS\n", host, stableArchive.Digest, candidateArchive.Digest, runtime.GOARCH)
	}
	return 0
}

func validateCandidateIdentity(version, channel string) (candidateIdentity, error) {
	identity := candidateIdentity{Version: version, Channel: channel}
	if identity == (candidateIdentity{Version: candidateVersion, Channel: candidateChannel}) ||
		identity == (candidateIdentity{Version: releaseVersion, Channel: releaseChannel}) {
		return identity, nil
	}
	return candidateIdentity{}, errors.New("unsupported candidate identity")
}

func loadArchive(name, host, version string) (archivePackage, error) {
	info, err := os.Lstat(name)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Size() < 1 || info.Size() > maxArchiveBytes {
		return archivePackage{}, errors.New("archive is not a bounded regular file")
	}
	archiveBytes, err := os.ReadFile(name)
	if err != nil {
		return archivePackage{}, errors.New("cannot read archive")
	}
	digest := sha256.Sum256(archiveBytes)
	reader, err := zip.NewReader(bytes.NewReader(archiveBytes), int64(len(archiveBytes)))
	if err != nil || len(reader.File) == 0 || len(reader.File) > maxArchiveFiles {
		return archivePackage{}, errors.New("archive directory is invalid or unbounded")
	}
	result := archivePackage{Host: host, Version: version, Digest: hex.EncodeToString(digest[:]), Files: make([]archiveFile, 0, len(reader.File))}
	var expanded uint64
	previous := ""
	for _, entry := range reader.File {
		if !safeArchivePath(entry.Name) || (previous != "" && entry.Name <= previous) || strings.HasPrefix(entry.Name, previous+"/") {
			return archivePackage{}, errors.New("archive paths are unsafe, duplicated, unsorted, or colliding")
		}
		previous = entry.Name
		mode := entry.Mode()
		if !mode.IsRegular() || mode&os.ModeSymlink != 0 || (mode.Perm() != 0o644 && mode.Perm() != 0o755) || entry.UncompressedSize64 > maxFileBytes {
			return archivePackage{}, fmt.Errorf("archive entry %q has an unsafe type, mode, or size", entry.Name)
		}
		expanded += entry.UncompressedSize64
		if expanded > maxExpandedBytes {
			return archivePackage{}, errors.New("expanded archive exceeds its bound")
		}
		input, err := entry.Open()
		if err != nil {
			return archivePackage{}, errors.New("cannot open archive entry")
		}
		data, readErr := io.ReadAll(io.LimitReader(input, maxFileBytes+1))
		closeErr := input.Close()
		if readErr != nil || closeErr != nil || uint64(len(data)) != entry.UncompressedSize64 {
			return archivePackage{}, fmt.Errorf("archive entry %q did not decode exactly", entry.Name)
		}
		sum := sha256.Sum256(data)
		result.Files = append(result.Files, archiveFile{Path: entry.Name, Data: data, Mode: mode.Perm(), SHA256: hex.EncodeToString(sum[:])})
	}
	return result, nil
}

func validateStable(pkg archivePackage) error {
	manifest := ".codex-plugin/plugin.json"
	if pkg.Host == "claude" {
		manifest = ".claude-plugin/plugin.json"
	}
	if err := validateManifest(pkg, manifest); err != nil {
		return err
	}
	var inventory struct {
		Schema int            `json:"schema"`
		Scope  string         `json:"scope"`
		Files  []declaredFile `json:"files"`
	}
	if err := decodeArchiveJSON(pkg, "INVENTORY.json", &inventory); err != nil || inventory.Schema != 1 || inventory.Scope != "all archive files except INVENTORY.json" {
		return errors.New("stable inventory metadata is invalid")
	}
	if err := validateDeclaredFiles(pkg, inventory.Files, map[string]bool{"INVENTORY.json": true}, true); err != nil {
		return fmt.Errorf("stable inventory: %w", err)
	}
	var metadata struct {
		Schema         int    `json:"schema"`
		Name           string `json:"name"`
		Version        string `json:"version"`
		Channel        string `json:"channel"`
		Host           string `json:"host"`
		ManifestPath   string `json:"manifest_path"`
		CatalogPath    string `json:"catalog_path"`
		CatalogSHA256  string `json:"catalog_sha256"`
		SourceDigest   string `json:"source_digest"`
		Builder        string `json:"builder"`
		SupportClaim   string `json:"support_claim"`
		ActualHostGate string `json:"actual_host_gate"`
	}
	if err := decodeArchiveJSON(pkg, "DISTRIBUTION.json", &metadata); err != nil || metadata.Schema != 2 || metadata.Name != "level7-dev-loop" ||
		metadata.Version != stableVersion || metadata.Channel != "stable" || metadata.Host != pkg.Host || metadata.ManifestPath != manifest || metadata.SupportClaim != "WITHHELD" {
		return errors.New("stable distribution identity is invalid")
	}
	if archiveFileNamed(pkg, "bin/l7") != nil || archiveFileNamed(pkg, ".mcp.json") != nil {
		return errors.New("frozen stable package unexpectedly contains executable behavior")
	}
	return nil
}

func validateCandidate(pkg archivePackage, identity candidateIdentity) error {
	if observed, err := validateCandidateIdentity(pkg.Version, identity.Channel); err != nil || observed != identity {
		return errors.New("candidate version and channel are invalid")
	}
	manifest := ".codex-plugin/plugin.json"
	if pkg.Host == "claude" {
		manifest = ".claude-plugin/plugin.json"
	}
	if err := validateManifest(pkg, manifest); err != nil {
		return err
	}
	expected := []string{
		manifest, candidateMarketplacePath(pkg.Host), ".mcp.json", "CHANGELOG.md", "CHECKSUMS.json", "LICENSE", "PERMISSIONS.json", "PROVENANCE.input.json", "README.md", "SBOM.spdx.json",
		"bin/darwin-amd64/l7", "bin/darwin-amd64/l7-embed", "bin/darwin-arm64/l7", "bin/darwin-arm64/l7-embed", "bin/l7",
	}
	for _, skill := range canonicalSkills {
		expected = append(expected, "skills/"+skill+"/SKILL.md")
	}
	sort.Strings(expected)
	if len(pkg.Files) != len(expected) {
		return fmt.Errorf("candidate inventory has %d files, want %d", len(pkg.Files), len(expected))
	}
	for index, file := range pkg.Files {
		if file.Path != expected[index] {
			return fmt.Errorf("candidate inventory substitution at %d: %q != %q", index, file.Path, expected[index])
		}
		wantMode := fs.FileMode(0o644)
		if strings.HasPrefix(file.Path, "bin/") {
			wantMode = 0o755
		}
		if file.Mode.Perm() != wantMode {
			return fmt.Errorf("candidate mode for %q is %04o, want %04o", file.Path, file.Mode.Perm(), wantMode)
		}
	}

	var checksums struct {
		Schema  int            `json:"schema"`
		Version string         `json:"version"`
		Files   []declaredFile `json:"files"`
	}
	if err := decodeArchiveJSON(pkg, "CHECKSUMS.json", &checksums); err != nil || checksums.Schema != 1 || checksums.Version != pkg.Version {
		return errors.New("candidate checksum metadata is invalid")
	}
	if err := validateDeclaredFiles(pkg, checksums.Files, map[string]bool{"CHECKSUMS.json": true, "SBOM.spdx.json": true}, false); err != nil {
		return fmt.Errorf("candidate checksums: %w", err)
	}
	if err := validateCandidateSBOM(pkg, checksums.Files); err != nil {
		return err
	}
	if err := validateCandidateProvenance(pkg, identity); err != nil {
		return err
	}
	if err := validateCandidateMarketplace(pkg); err != nil {
		return err
	}
	if err := validateMCPConfiguration(pkg); err != nil {
		return err
	}
	for _, architecture := range []string{"arm64", "amd64"} {
		for _, executable := range []string{"l7", "l7-embed"} {
			entry := archiveFileNamed(pkg, "bin/darwin-"+architecture+"/"+executable)
			if entry == nil || validateMachO(entry.Data, architecture) != nil {
				return fmt.Errorf("candidate darwin/%s %s is invalid", architecture, executable)
			}
			if executable == "l7" && validateGoBinaryIdentity(entry.Data, identity.Version) != nil {
				return fmt.Errorf("candidate darwin/%s Level 7 build identity is invalid", architecture)
			}
		}
	}
	launcher := archiveFileNamed(pkg, "bin/l7")
	if launcher == nil || !bytes.HasPrefix(launcher.Data, []byte("#!/bin/sh\n")) || !bytes.Contains(launcher.Data, []byte(`exec "$platform_dir/l7" "$@"`)) {
		return errors.New("candidate native launcher is invalid")
	}
	return nil
}

func candidateMarketplacePath(host string) string {
	if host == "claude" {
		return ".claude-plugin/marketplace.json"
	}
	return ".agents/plugins/marketplace.json"
}

func validateCandidateProvenance(pkg archivePackage, identity candidateIdentity) error {
	var provenance struct {
		Schema         int    `json:"schema"`
		Version        string `json:"version"`
		Host           string `json:"host"`
		Channel        string `json:"channel"`
		ArtifactState  string `json:"artifact_state"`
		Authority      string `json:"authority"`
		ReleaseBlocked bool   `json:"release_blocked"`
		Next           string `json:"next"`
	}
	if err := decodeArchiveJSON(pkg, "PROVENANCE.input.json", &provenance); err != nil || provenance.Schema != 2 || provenance.Version != identity.Version ||
		provenance.Host != pkg.Host || provenance.Channel != identity.Channel || provenance.ArtifactState != "package-input" ||
		provenance.Authority != "external-only" || !provenance.ReleaseBlocked || provenance.Next == "" {
		return errors.New("candidate release-blocked input provenance is invalid")
	}
	return nil
}

func validateCandidateMarketplace(pkg archivePackage) error {
	if pkg.Host == "codex" {
		var catalog struct {
			Name      string `json:"name"`
			Interface struct {
				DisplayName string `json:"displayName"`
			} `json:"interface"`
			Plugins []struct {
				Name   string `json:"name"`
				Source struct {
					Source string `json:"source"`
					Path   string `json:"path"`
				} `json:"source"`
				Policy struct {
					Installation   string `json:"installation"`
					Authentication string `json:"authentication"`
				} `json:"policy"`
				Category string `json:"category"`
			} `json:"plugins"`
		}
		if err := decodeArchiveJSON(pkg, candidateMarketplacePath(pkg.Host), &catalog); err != nil || catalog.Name != v1MarketplaceName || catalog.Interface.DisplayName != "Level 7 Engineering v1" || len(catalog.Plugins) != 1 {
			return errors.New("candidate Codex marketplace is invalid")
		}
		plugin := catalog.Plugins[0]
		if plugin.Name != "level7-dev-loop" || plugin.Source.Source != "local" || plugin.Source.Path != "." ||
			plugin.Policy.Installation != "AVAILABLE" || plugin.Policy.Authentication != "ON_INSTALL" || plugin.Category != "Developer Tools" {
			return errors.New("candidate Codex marketplace source is invalid")
		}
		return nil
	}
	var catalog struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Owner       struct {
			Name string `json:"name"`
		} `json:"owner"`
		Plugins []struct {
			Name        string `json:"name"`
			Source      string `json:"source"`
			Description string `json:"description"`
			Version     string `json:"version"`
			License     string `json:"license"`
			Category    string `json:"category"`
			Strict      bool   `json:"strict"`
		} `json:"plugins"`
	}
	if err := decodeArchiveJSON(pkg, candidateMarketplacePath(pkg.Host), &catalog); err != nil || catalog.Name != v1MarketplaceName || catalog.Description == "" || catalog.Owner.Name != "Level 7 Engineering" || len(catalog.Plugins) != 1 {
		return errors.New("candidate Claude marketplace is invalid")
	}
	plugin := catalog.Plugins[0]
	if plugin.Name != "level7-dev-loop" || plugin.Source != "." || plugin.Description == "" || plugin.Version != pkg.Version ||
		plugin.License != "MIT" || plugin.Category != "Development Tools" || !plugin.Strict {
		return errors.New("candidate Claude marketplace source is invalid")
	}
	return nil
}

func validateManifest(pkg archivePackage, name string) error {
	entry := archiveFileNamed(pkg, name)
	if entry == nil {
		return errors.New("host manifest is missing")
	}
	var manifest map[string]any
	if json.Unmarshal(entry.Data, &manifest) != nil || manifest["name"] != "level7-dev-loop" || manifest["version"] != pkg.Version {
		return errors.New("host manifest identity is invalid")
	}
	return nil
}

func validateDeclaredFiles(pkg archivePackage, declared []declaredFile, excluded map[string]bool, requireMode bool) error {
	expected := make([]archiveFile, 0, len(pkg.Files))
	for _, file := range pkg.Files {
		if !excluded[file.Path] {
			expected = append(expected, file)
		}
	}
	if len(declared) != len(expected) {
		return fmt.Errorf("declares %d files, want %d", len(declared), len(expected))
	}
	for index, current := range declared {
		file := expected[index]
		if current.Path != file.Path || current.Size != len(file.Data) || current.SHA256 != file.SHA256 || !validDigest(current.SHA256) {
			return fmt.Errorf("file %d does not bind %q", index, file.Path)
		}
		if requireMode && current.Mode != fmt.Sprintf("%04o", file.Mode.Perm()) {
			return fmt.Errorf("file %q mode is not bound", file.Path)
		}
		if !requireMode && current.Mode != "" {
			return fmt.Errorf("file %q has unexpected checksum mode", file.Path)
		}
	}
	return nil
}

func validateCandidateSBOM(pkg archivePackage, checks []declaredFile) error {
	var document struct {
		SPDXVersion       string          `json:"spdxVersion"`
		DataLicense       string          `json:"dataLicense"`
		SPDXID            string          `json:"SPDXID"`
		Name              string          `json:"name"`
		DocumentNamespace string          `json:"documentNamespace"`
		CreationInfo      json.RawMessage `json:"creationInfo"`
		Packages          json.RawMessage `json:"packages"`
		Relationships     json.RawMessage `json:"relationships"`
		Files             []struct {
			FileName         string `json:"fileName"`
			SPDXID           string `json:"SPDXID"`
			LicenseConcluded string `json:"licenseConcluded"`
			CopyrightText    string `json:"copyrightText"`
			Checksums        []struct {
				Algorithm     string `json:"algorithm"`
				ChecksumValue string `json:"checksumValue"`
			} `json:"checksums"`
		} `json:"files"`
	}
	if err := decodeArchiveJSON(pkg, "SBOM.spdx.json", &document); err != nil || document.SPDXVersion != "SPDX-2.3" || document.DataLicense != "CC0-1.0" || document.SPDXID != "SPDXRef-DOCUMENT" ||
		document.Name != "level7-dev-loop-"+pkg.Version+"-"+pkg.Host || !strings.HasSuffix(document.DocumentNamespace, "/"+pkg.Version+"/"+pkg.Host) || len(document.Files) != len(checks) {
		return errors.New("candidate SBOM identity is invalid")
	}
	for index, file := range document.Files {
		if file.FileName != "./"+checks[index].Path || len(file.Checksums) != 1 || file.Checksums[0].Algorithm != "SHA256" || file.Checksums[0].ChecksumValue != checks[index].SHA256 {
			return fmt.Errorf("candidate SBOM file %d does not bind checksums", index)
		}
	}
	return nil
}

func validateMCPConfiguration(pkg archivePackage) error {
	var configuration struct {
		Servers map[string]struct {
			Type    string   `json:"type"`
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}
	if err := decodeArchiveJSON(pkg, ".mcp.json", &configuration); err != nil || len(configuration.Servers) != 1 {
		return errors.New("candidate MCP configuration is invalid")
	}
	server, ok := configuration.Servers["level7"]
	wantCommand := "./bin/l7"
	if pkg.Host == "claude" {
		wantCommand = "${CLAUDE_PLUGIN_ROOT}/bin/l7"
	}
	if !ok || server.Type != "stdio" || server.Command != wantCommand || len(server.Args) != 1 || server.Args[0] != "mcp" {
		return errors.New("candidate MCP command is invalid")
	}
	return nil
}

func installInitial(root string, pkg archivePackage) error {
	if !safeLifecycleRoot(root) || pathExists(root) {
		return errors.New("initial lifecycle root is unsafe or occupied")
	}
	if err := os.Mkdir(root, 0o700); err != nil {
		return err
	}
	if err := extractPackage(filepath.Join(root, "active"), pkg); err != nil {
		_ = os.RemoveAll(root)
		return err
	}
	receipt := lifecycleReceipt{Schema: 1, Host: pkg.Host, Active: receiptPackage(pkg)}
	if err := writeReceipt(root, receipt); err != nil {
		_ = os.RemoveAll(root)
		return err
	}
	return verifyTree(filepath.Join(root, "active"), receipt.Active)
}

func upgradeInstallation(root, expectedDigest string, next archivePackage) error {
	receipt, err := loadReceipt(root)
	if err != nil || receipt.Host != next.Host || receipt.Active.Digest != expectedDigest || receipt.Previous.Digest != "" {
		return errors.New("upgrade receipt does not bind the expected active package")
	}
	if err := verifyTree(filepath.Join(root, "active"), receipt.Active); err != nil {
		return fmt.Errorf("upgrade preflight: %w", err)
	}
	stage := filepath.Join(root, "stage")
	if pathExists(stage) || pathExists(filepath.Join(root, "previous")) {
		return errors.New("upgrade transaction paths are occupied")
	}
	if err := extractPackage(stage, next); err != nil {
		return err
	}
	if err := verifyTree(stage, receiptPackage(next)); err != nil {
		_ = os.RemoveAll(stage)
		return err
	}
	active := filepath.Join(root, "active")
	previous := filepath.Join(root, "previous")
	if err := os.Rename(active, previous); err != nil {
		_ = os.RemoveAll(stage)
		return err
	}
	if err := os.Rename(stage, active); err != nil {
		_ = os.Rename(previous, active)
		_ = os.RemoveAll(stage)
		return err
	}
	receipt.Previous = receipt.Active
	receipt.Active = receiptPackage(next)
	if err := writeReceipt(root, receipt); err != nil {
		return err
	}
	return errors.Join(verifyTree(active, receipt.Active), verifyTree(previous, receipt.Previous))
}

func rollbackInstallation(root, expectedActive, expectedPrevious string) error {
	receipt, err := loadReceipt(root)
	if err != nil || receipt.Active.Digest != expectedActive || receipt.Previous.Digest != expectedPrevious {
		return errors.New("rollback receipt does not bind the expected package pair")
	}
	active := filepath.Join(root, "active")
	previous := filepath.Join(root, "previous")
	if err := errors.Join(verifyTree(active, receipt.Active), verifyTree(previous, receipt.Previous)); err != nil {
		return fmt.Errorf("rollback preflight: %w", err)
	}
	swap := filepath.Join(root, "swap")
	if pathExists(swap) {
		return errors.New("rollback transaction path is occupied")
	}
	if err := os.Rename(active, swap); err != nil {
		return err
	}
	if err := os.Rename(previous, active); err != nil {
		_ = os.Rename(swap, active)
		return err
	}
	if err := os.Rename(swap, previous); err != nil {
		return err
	}
	receipt.Active, receipt.Previous = receipt.Previous, receipt.Active
	if err := writeReceipt(root, receipt); err != nil {
		return err
	}
	return errors.Join(verifyTree(active, receipt.Active), verifyTree(previous, receipt.Previous))
}

func removeInstallation(root string) error {
	if !safeLifecycleRoot(root) {
		return errors.New("removal root is unsafe")
	}
	receipt, err := loadReceipt(root)
	if err != nil {
		return err
	}
	if err := verifyTree(filepath.Join(root, "active"), receipt.Active); err != nil {
		return fmt.Errorf("active removal conflict: %w", err)
	}
	if receipt.Previous.Digest != "" {
		if err := verifyTree(filepath.Join(root, "previous"), receipt.Previous); err != nil {
			return fmt.Errorf("previous removal conflict: %w", err)
		}
	}
	allowed := map[string]bool{"active": true, "receipt.json": true}
	if receipt.Previous.Digest != "" {
		allowed["previous"] = true
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !allowed[entry.Name()] {
			return fmt.Errorf("unowned lifecycle entry %q", entry.Name())
		}
	}
	return os.RemoveAll(root)
}

func extractPackage(root string, pkg archivePackage) (returnErr error) {
	if pathExists(root) {
		return errors.New("package extraction root is occupied")
	}
	if err := os.Mkdir(root, 0o700); err != nil {
		return err
	}
	defer func() {
		if returnErr != nil {
			_ = os.RemoveAll(root)
		}
	}()
	for _, file := range pkg.Files {
		name := filepath.Join(root, filepath.FromSlash(file.Path))
		if !isWithin(name, root) {
			return errors.New("package path escaped extraction root")
		}
		if err := os.MkdirAll(filepath.Dir(name), 0o700); err != nil {
			return err
		}
		output, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, file.Mode.Perm())
		if err != nil {
			return err
		}
		written, writeErr := output.Write(file.Data)
		closeErr := output.Close()
		if writeErr != nil || closeErr != nil || written != len(file.Data) {
			return errors.New("package file write was incomplete")
		}
	}
	return nil
}

func verifyTree(root string, installed installedPackage) error {
	if installed.Digest == "" || !validDigest(installed.Digest) || len(installed.Files) == 0 {
		return errors.New("ownership receipt is empty or invalid")
	}
	expected := make(map[string]ownedFile, len(installed.Files))
	for _, file := range installed.Files {
		if !safeArchivePath(file.Path) || expected[file.Path].Path != "" || !validDigest(file.SHA256) || (file.Mode != "0644" && file.Mode != "0755") || file.Size < 0 || file.Size > maxFileBytes {
			return errors.New("ownership receipt file set is invalid")
		}
		expected[file.Path] = file
	}
	seen := make(map[string]bool, len(expected))
	err := filepath.WalkDir(root, func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if name == root {
			return nil
		}
		relative, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if entry.IsDir() {
			return nil
		}
		wanted, ok := expected[relative]
		if !ok {
			return fmt.Errorf("unowned file %q", relative)
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != parseMode(wanted.Mode) || info.Size() != int64(wanted.Size) {
			return fmt.Errorf("owned file %q type, mode, or size changed", relative)
		}
		input, err := os.Open(name)
		if err != nil {
			return err
		}
		hash := sha256.New()
		written, copyErr := io.Copy(hash, io.LimitReader(input, maxFileBytes+1))
		closeErr := input.Close()
		if copyErr != nil || closeErr != nil || written != info.Size() || hex.EncodeToString(hash.Sum(nil)) != wanted.SHA256 {
			return fmt.Errorf("owned file %q content changed", relative)
		}
		seen[relative] = true
		return nil
	})
	if err != nil {
		return err
	}
	if len(seen) != len(expected) {
		return errors.New("owned file is missing")
	}
	return nil
}

func receiptPackage(pkg archivePackage) installedPackage {
	files := make([]ownedFile, 0, len(pkg.Files))
	for _, file := range pkg.Files {
		files = append(files, ownedFile{Path: file.Path, Mode: fmt.Sprintf("%04o", file.Mode.Perm()), Size: len(file.Data), SHA256: file.SHA256})
	}
	return installedPackage{Version: pkg.Version, Digest: pkg.Digest, Files: files}
}

func writeReceipt(root string, receipt lifecycleReceipt) error {
	receipt.StateSHA256 = ""
	digest, err := receiptDigest(receipt)
	if err != nil {
		return err
	}
	receipt.StateSHA256 = digest
	data, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	temporary := filepath.Join(root, ".receipt.tmp")
	_ = os.Remove(temporary)
	file, err := os.OpenFile(temporary, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	written, writeErr := file.Write(data)
	syncErr := file.Sync()
	closeErr := file.Close()
	if err := errors.Join(writeErr, syncErr, closeErr); err != nil || written != len(data) {
		_ = os.Remove(temporary)
		return errors.Join(err, errors.New("receipt write was incomplete"))
	}
	return os.Rename(temporary, filepath.Join(root, "receipt.json"))
}

func loadReceipt(root string) (lifecycleReceipt, error) {
	var receipt lifecycleReceipt
	data, err := os.ReadFile(filepath.Join(root, "receipt.json"))
	if err != nil || len(data) == 0 || len(data) > 2<<20 || decodeStrict(data, &receipt) != nil {
		return lifecycleReceipt{}, errors.New("lifecycle receipt is missing or invalid")
	}
	if receipt.Schema != 1 || (receipt.Host != "codex" && receipt.Host != "claude") || receipt.Active.Version == "" || receipt.Active.Digest == "" {
		return lifecycleReceipt{}, errors.New("lifecycle receipt identity is invalid")
	}
	want := receipt.StateSHA256
	receipt.StateSHA256 = ""
	observed, err := receiptDigest(receipt)
	if err != nil || !validDigest(want) || observed != want {
		return lifecycleReceipt{}, errors.New("lifecycle receipt state binding is stale")
	}
	receipt.StateSHA256 = want
	return receipt, nil
}

func receiptDigest(receipt lifecycleReceipt) (string, error) {
	data, err := json.Marshal(receipt)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func executeCandidate(lifecycleRoot, executionRoot, host, expectedVersion string) error {
	if err := os.Mkdir(executionRoot, 0o700); err != nil {
		return err
	}
	for _, directory := range []string{"home", "tmp", "repo"} {
		if err := os.Mkdir(filepath.Join(executionRoot, directory), 0o700); err != nil {
			return err
		}
	}
	launcher := filepath.Join(lifecycleRoot, "active", "bin", "l7")
	versionOutput, versionError, err := boundedCommand(executionRoot, launcher, []string{"--version", "--json"}, nil)
	if err != nil || len(versionError) != 0 {
		return fmt.Errorf("native JSON version command failed: %w stderr=%q", err, versionError)
	}
	var version struct {
		Schema  int    `json:"schema"`
		Outcome string `json:"outcome"`
		Code    string `json:"code"`
		Command string `json:"command"`
		State   string `json:"state"`
		Version string `json:"version"`
		Message string `json:"message"`
		Next    string `json:"next"`
		Details []any  `json:"details"`
	}
	if decodeStrict(versionOutput, &version) != nil || version.Schema != 4 || version.Outcome != "PASS" || version.Code != "L7-CLI-000" || version.Command != "version" || version.Version != expectedVersion {
		return errors.New("native JSON CLI contract is invalid")
	}
	requests := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"v1candidate","version":"1"}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"l7_v1_invalid","arguments":{}}}`,
	}, "\n") + "\n"
	mcpOutput, mcpError, err := boundedCommand(executionRoot, launcher, []string{"mcp"}, []byte(requests))
	if err != nil || len(mcpError) != 0 {
		return fmt.Errorf("native MCP command failed: %w stderr=%q", err, mcpError)
	}
	responses, err := decodeMCP(mcpOutput)
	if err != nil || len(responses) != 3 {
		return errors.New("native MCP framing is invalid")
	}
	if err := validateMCPResponses(responses, host, expectedVersion); err != nil {
		return err
	}
	return nil
}

func boundedCommand(executionRoot, executable string, arguments []string, input []byte) ([]byte, []byte, error) {
	sandbox := "/usr/bin/sandbox-exec"
	info, err := os.Lstat(sandbox)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Mode()&0o111 == 0 {
		return nil, nil, errors.New("macOS network-denial sandbox is unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	sandboxArguments := append([]string{"-p", networkDenyProfile, executable}, arguments...)
	command := exec.CommandContext(ctx, sandbox, sandboxArguments...)
	command.Dir = filepath.Join(executionRoot, "repo")
	command.Env = []string{
		"PATH=/usr/bin:/bin", "HOME=" + filepath.Join(executionRoot, "home"), "TMPDIR=" + filepath.Join(executionRoot, "tmp"),
		"LC_ALL=C", "LANG=C", "L7_NETWORK=off", "L7_TELEMETRY=off", "GIT_TERMINAL_PROMPT=0", "NO_PROXY=", "no_proxy=",
		"HTTP_PROXY=http://127.0.0.1:1", "HTTPS_PROXY=http://127.0.0.1:1", "ALL_PROXY=http://127.0.0.1:1",
	}
	command.Stdin = bytes.NewReader(input)
	stdout := &limitBuffer{remaining: maxProcessOutput}
	stderr := &limitBuffer{remaining: maxProcessOutput}
	command.Stdout = stdout
	command.Stderr = stderr
	err = command.Run()
	if ctx.Err() != nil {
		return stdout.Bytes(), stderr.Bytes(), ctx.Err()
	}
	if stdout.exceeded || stderr.exceeded {
		return stdout.Bytes(), stderr.Bytes(), errors.New("native process output exceeded bound")
	}
	return stdout.Bytes(), stderr.Bytes(), err
}

type limitBuffer struct {
	buffer    bytes.Buffer
	remaining int
	exceeded  bool
}

func (buffer *limitBuffer) Write(data []byte) (int, error) {
	if len(data) > buffer.remaining {
		accepted := max(0, buffer.remaining)
		_, _ = buffer.buffer.Write(data[:accepted])
		buffer.remaining = 0
		buffer.exceeded = true
		return len(data), nil
	}
	buffer.remaining -= len(data)
	return buffer.buffer.Write(data)
}

func (buffer *limitBuffer) Bytes() []byte { return append([]byte{}, buffer.buffer.Bytes()...) }

type mcpEnvelope struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func decodeMCP(data []byte) ([]mcpEnvelope, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 64<<10), 16<<20)
	responses := []mcpEnvelope{}
	for scanner.Scan() {
		var response mcpEnvelope
		if decodeStrict(scanner.Bytes(), &response) != nil || response.JSONRPC != "2.0" {
			return nil, errors.New("invalid MCP response")
		}
		responses = append(responses, response)
	}
	return responses, scanner.Err()
}

func validateMCPResponses(responses []mcpEnvelope, host, expectedVersion string) error {
	var initialize struct {
		ProtocolVersion string `json:"protocolVersion"`
		Capabilities    struct {
			Tools struct {
				ListChanged bool `json:"listChanged"`
			} `json:"tools"`
		} `json:"capabilities"`
		ServerInfo struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"serverInfo"`
		Instructions string `json:"instructions"`
	}
	if string(responses[0].ID) != "1" || responses[0].Error != nil || json.Unmarshal(responses[0].Result, &initialize) != nil ||
		initialize.ProtocolVersion != "2025-11-25" || initialize.ServerInfo.Name != "level7" || initialize.ServerInfo.Version != expectedVersion {
		return fmt.Errorf("%s MCP initialize contract is invalid", host)
	}
	var listing struct {
		Tools []struct {
			Name        string         `json:"name"`
			Description string         `json:"description"`
			InputSchema map[string]any `json:"inputSchema"`
		} `json:"tools"`
	}
	if string(responses[1].ID) != "2" || responses[1].Error != nil || json.Unmarshal(responses[1].Result, &listing) != nil || len(listing.Tools) != 6 {
		return fmt.Errorf("%s MCP tools/list contract is invalid", host)
	}
	want := []string{"l7_v1_cyber", "l7_v1_headless", "l7_v1_memory", "l7_v1_onboard", "l7_v1_provider_discovery", "l7_v1_route_explain"}
	observed := make([]string, 0, len(listing.Tools))
	for _, tool := range listing.Tools {
		if tool.Description == "" || tool.InputSchema["type"] != "object" || tool.InputSchema["additionalProperties"] != false {
			return fmt.Errorf("%s MCP tool schema is invalid", host)
		}
		observed = append(observed, tool.Name)
	}
	sort.Strings(observed)
	if strings.Join(observed, "\n") != strings.Join(want, "\n") {
		return fmt.Errorf("%s MCP tool inventory is invalid", host)
	}
	if string(responses[2].ID) != "3" || responses[2].Error == nil || responses[2].Error.Code != -32602 || len(responses[2].Result) != 0 {
		return fmt.Errorf("%s MCP tools/call rejection contract is invalid", host)
	}
	return nil
}

func decodeArchiveJSON(pkg archivePackage, name string, target any) error {
	entry := archiveFileNamed(pkg, name)
	if entry == nil {
		return fmt.Errorf("archive entry %q is missing", name)
	}
	return decodeStrict(entry.Data, target)
}

func decodeStrict(data []byte, target any) error {
	if err := validateJSON(data); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return errors.New("JSON has trailing data")
	}
	return nil
}

func validateJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value func(int) error
	value = func(depth int) error {
		if depth > 64 {
			return errors.New("JSON nesting exceeds bound")
		}
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		delimiter, ok := token.(json.Delim)
		if !ok {
			return nil
		}
		switch delimiter {
		case '{':
			seen := map[string]bool{}
			for decoder.More() {
				keyToken, err := decoder.Token()
				key, keyOK := keyToken.(string)
				if err != nil || !keyOK || seen[key] {
					return errors.New("JSON object key is invalid or duplicated")
				}
				seen[key] = true
				if err := value(depth + 1); err != nil {
					return err
				}
			}
			closing, err := decoder.Token()
			if err != nil || closing != json.Delim('}') {
				return errors.New("JSON object is unterminated")
			}
		case '[':
			for decoder.More() {
				if err := value(depth + 1); err != nil {
					return err
				}
			}
			closing, err := decoder.Token()
			if err != nil || closing != json.Delim(']') {
				return errors.New("JSON array is unterminated")
			}
		default:
			return errors.New("JSON delimiter is invalid")
		}
		return nil
	}
	if err := value(0); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return errors.New("JSON has trailing data")
	}
	return nil
}

func archiveFileNamed(pkg archivePackage, name string) *archiveFile {
	index := sort.Search(len(pkg.Files), func(index int) bool { return pkg.Files[index].Path >= name })
	if index == len(pkg.Files) || pkg.Files[index].Path != name {
		return nil
	}
	return &pkg.Files[index]
}

func validateMachO(data []byte, architecture string) error {
	file, err := macho.NewFile(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer file.Close()
	want := macho.CpuArm64
	if architecture == "amd64" {
		want = macho.CpuAmd64
	}
	if file.Cpu != want {
		return errors.New("Mach-O architecture mismatch")
	}
	return nil
}

func validateGoBinaryIdentity(data []byte, expectedVersion string) error {
	info, err := buildinfo.Read(bytes.NewReader(data))
	if err != nil {
		return errors.New("Go build identity is invalid")
	}
	version, err := readMachOStringSymbol(data, "main.version")
	if err != nil {
		return errors.New("Go binary version is invalid")
	}
	return validateDecodedGoIdentity(info.Path, info.Main.Path, version, expectedVersion)
}

func validateDecodedGoIdentity(path, module, version, expectedVersion string) error {
	if path != "github.com/addressanup/level7-dev-loop/cmd/l7" || module != "github.com/addressanup/level7-dev-loop" || version != expectedVersion {
		return errors.New("Go binary path, module, or version is invalid")
	}
	return nil
}

func readMachOStringSymbol(data []byte, name string) (string, error) {
	value, err := macho.NewFile(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer value.Close()
	if value.Symtab == nil {
		return "", errors.New("Mach-O symbol table is missing")
	}
	address := uint64(0)
	matches := 0
	for _, symbol := range value.Symtab.Syms {
		if symbol.Name == name {
			address = symbol.Value
			matches++
		}
	}
	if matches != 1 {
		return "", errors.New("Mach-O string symbol is missing or duplicated")
	}
	header, err := machODataAt(value, data, address, 16)
	if err != nil {
		return "", err
	}
	stringAddress := value.ByteOrder.Uint64(header[:8])
	stringLength := value.ByteOrder.Uint64(header[8:])
	if stringLength < 1 || stringLength > 64 {
		return "", errors.New("Mach-O string symbol length is invalid")
	}
	content, err := machODataAt(value, data, stringAddress, stringLength)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func machODataAt(value *macho.File, data []byte, address, size uint64) ([]byte, error) {
	for _, section := range value.Sections {
		if address < section.Addr || address-section.Addr > section.Size || size > section.Size-(address-section.Addr) {
			continue
		}
		offset := uint64(section.Offset) + address - section.Addr
		if offset <= uint64(len(data)) && size <= uint64(len(data))-offset {
			return data[offset : offset+size], nil
		}
	}
	return nil, errors.New("Mach-O address is outside file-backed sections")
}

func verifySHA256Sidecar(name, archiveName, digest string) error {
	info, err := os.Lstat(name)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Size() > 1024 {
		return errors.New("stable SHA-256 sidecar is invalid")
	}
	data, err := os.ReadFile(name)
	if err != nil || string(data) != digest+"  "+archiveName+"\n" {
		return errors.New("stable SHA-256 sidecar does not bind the archive")
	}
	return nil
}

func safeArchivePath(value string) bool {
	if value == "" || len(value) > 4096 || !utf8.ValidString(value) || strings.Contains(value, "\\") || path.IsAbs(value) || path.Clean(value) != value || value == "." || value == ".." || strings.HasPrefix(value, "../") {
		return false
	}
	for _, character := range value {
		if character < 0x20 || character == 0x7f {
			return false
		}
	}
	for _, part := range strings.Split(value, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}

func validDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil && value == strings.ToLower(value)
}

func parseMode(value string) fs.FileMode {
	if value == "0755" {
		return 0o755
	}
	return 0o644
}

func physicalDirectory(value string) (string, error) {
	absolute, err := filepath.Abs(value)
	if err != nil || absolute != filepath.Clean(value) {
		return "", errors.New("directory path is not physical and absolute")
	}
	physical, err := filepath.EvalSymlinks(absolute)
	if err != nil || physical != absolute {
		return "", errors.New("directory path contains a symlink")
	}
	info, err := os.Stat(physical)
	if err != nil || !info.IsDir() {
		return "", errors.New("path is not a directory")
	}
	return physical, nil
}

func safeLifecycleRoot(root string) bool {
	return filepath.IsAbs(root) && filepath.Clean(root) == root && strings.HasPrefix(filepath.Base(root), "lifecycle-") && filepath.Dir(root) != root
}

func isWithin(name, root string) bool {
	return name != root && strings.HasPrefix(filepath.Clean(name), filepath.Clean(root)+string(filepath.Separator))
}

func removeOwnedRoot(workRoot, target string) error {
	if !isWithin(target, workRoot) || (!strings.HasPrefix(filepath.Base(target), "lifecycle-") && !strings.HasPrefix(filepath.Base(target), "execution-")) {
		return errors.New("cleanup target is unsafe")
	}
	return os.RemoveAll(target)
}

func pathExists(name string) bool {
	_, err := os.Lstat(name)
	return err == nil || !errors.Is(err, os.ErrNotExist)
}

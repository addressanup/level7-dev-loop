// Package pluginpackage validates, extracts, and verifies Level 7 v1 host
// archives. It accepts only the closed package layout l7pack produces and
// proves installer ownership of every extracted tree through receipts.
package pluginpackage

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"debug/macho"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
)

const (
	MarketplaceName = "level7-engineering-v1"
	PluginName      = "level7-dev-loop"

	maxArchiveBytes  = 256 << 20
	maxArchiveFiles  = 512
	maxFileBytes     = 128 << 20
	maxExpandedBytes = 512 << 20
	maxReceiptBytes  = 4 << 20
)

var Architectures = []string{"arm64", "amd64"}

type File struct {
	Path   string
	Mode   fs.FileMode
	Data   []byte
	SHA256 string
}

// Package is one fully read archive. Digest is the SHA-256 of the archive
// bytes, so it matches the release manifest and SHA256SUMS directly.
type Package struct {
	Host    string
	Version string
	Digest  string
	Files   []File
}

func ValidHost(host string) bool { return host == "codex" || host == "claude" }

// Load reads a bounded archive and rejects traversal, absolute paths,
// duplicates, directory collisions, links, special files, unexpected modes,
// oversized members, and decompression beyond the expanded bound.
func Load(name, host, version string) (Package, error) {
	if !ValidHost(host) || version == "" {
		return Package{}, errors.New("package host or version is invalid")
	}
	info, err := os.Lstat(name)
	if err != nil || !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > maxArchiveBytes {
		return Package{}, errors.New("archive is not a bounded regular file")
	}
	archiveBytes, err := os.ReadFile(name)
	if err != nil {
		return Package{}, errors.New("cannot read archive")
	}
	digest := sha256.Sum256(archiveBytes)
	reader, err := zip.NewReader(bytes.NewReader(archiveBytes), int64(len(archiveBytes)))
	if err != nil || len(reader.File) == 0 || len(reader.File) > maxArchiveFiles {
		return Package{}, errors.New("archive directory is invalid or unbounded")
	}
	result := Package{Host: host, Version: version, Digest: hex.EncodeToString(digest[:]), Files: make([]File, 0, len(reader.File))}
	var expanded uint64
	previous := ""
	for _, entry := range reader.File {
		if !SafePath(entry.Name) || (previous != "" && entry.Name <= previous) || strings.HasPrefix(entry.Name, previous+"/") {
			return Package{}, errors.New("archive paths are unsafe, duplicated, unsorted, or colliding")
		}
		previous = entry.Name
		mode := entry.Mode()
		if !mode.IsRegular() || mode&^fs.ModePerm != 0 || (mode.Perm() != 0o644 && mode.Perm() != 0o755) || entry.UncompressedSize64 > maxFileBytes {
			return Package{}, fmt.Errorf("archive entry %q has an unsafe type, mode, or size", entry.Name)
		}
		expanded += entry.UncompressedSize64
		if expanded > maxExpandedBytes {
			return Package{}, errors.New("expanded archive exceeds its bound")
		}
		input, err := entry.Open()
		if err != nil {
			return Package{}, errors.New("cannot open archive entry")
		}
		data, readErr := io.ReadAll(io.LimitReader(input, maxFileBytes+1))
		closeErr := input.Close()
		if readErr != nil || closeErr != nil || uint64(len(data)) != entry.UncompressedSize64 {
			return Package{}, fmt.Errorf("archive entry %q did not decode exactly", entry.Name)
		}
		sum := sha256.Sum256(data)
		result.Files = append(result.Files, File{Path: entry.Name, Data: data, Mode: mode.Perm(), SHA256: hex.EncodeToString(sum[:])})
	}
	return result, nil
}

// Identity is what a validated package declares about itself.
type Identity struct {
	Marketplace string
	Plugin      string
	MCPCommand  string
}

// Validate enforces the closed v1 inventory for the host, binds the running
// architecture's executables to the attested input digests when provided, and
// returns the catalog identity the host must be told about.
func Validate(pkg Package, arch string, inputs map[string]string) (Identity, error) {
	if !ValidHost(pkg.Host) || (arch != "arm64" && arch != "amd64") {
		return Identity{}, errors.New("package host or architecture is unsupported")
	}
	manifestPath := ".codex-plugin/plugin.json"
	marketplacePath := ".agents/plugins/marketplace.json"
	if pkg.Host == "claude" {
		manifestPath = ".claude-plugin/plugin.json"
		marketplacePath = ".claude-plugin/marketplace.json"
	}
	fixed := map[string]bool{
		manifestPath: true, marketplacePath: true, ".mcp.json": true, "CHANGELOG.md": true, "CHECKSUMS.json": true, "LICENSE": true,
		"PERMISSIONS.json": true, "PROVENANCE.input.json": true, "README.md": true, "SBOM.spdx.json": true,
		"bin/darwin-amd64/l7": true, "bin/darwin-amd64/l7-embed": true, "bin/darwin-arm64/l7": true, "bin/darwin-arm64/l7-embed": true, "bin/l7": true,
	}
	seen := make(map[string]bool, len(pkg.Files))
	skills := 0
	for _, file := range pkg.Files {
		switch {
		case fixed[file.Path]:
		case isSkillPath(file.Path):
			skills++
		default:
			return Identity{}, fmt.Errorf("package contains unexpected entry %q", file.Path)
		}
		if seen[file.Path] {
			return Identity{}, fmt.Errorf("package repeats %q", file.Path)
		}
		seen[file.Path] = true
		wantMode := fs.FileMode(0o644)
		if strings.HasPrefix(file.Path, "bin/") {
			wantMode = 0o755
		}
		if file.Mode.Perm() != wantMode {
			return Identity{}, fmt.Errorf("package mode for %q is %04o, want %04o", file.Path, file.Mode.Perm(), wantMode)
		}
	}
	for name := range fixed {
		if !seen[name] {
			return Identity{}, fmt.Errorf("package lacks required entry %q", name)
		}
	}
	if skills == 0 {
		return Identity{}, errors.New("package contains no skills")
	}
	if err := validateManifest(pkg, manifestPath); err != nil {
		return Identity{}, err
	}
	identity, err := validateMarketplace(pkg, marketplacePath)
	if err != nil {
		return Identity{}, err
	}
	if err := validateChecksums(pkg); err != nil {
		return Identity{}, err
	}
	command, err := validateMCP(pkg)
	if err != nil {
		return Identity{}, err
	}
	identity.MCPCommand = command
	launcher := fileNamed(pkg, "bin/l7")
	if launcher == nil || !bytes.HasPrefix(launcher.Data, []byte("#!/bin/sh\n")) || !bytes.Contains(launcher.Data, []byte(`exec "$platform_dir/l7" "$@"`)) {
		return Identity{}, errors.New("package launcher is invalid")
	}
	for _, architecture := range Architectures {
		for _, executable := range []string{"l7", "l7-embed"} {
			relative := "darwin-" + architecture + "/" + executable
			entry := fileNamed(pkg, "bin/"+relative)
			if entry == nil || validateMachO(entry.Data, architecture) != nil {
				return Identity{}, fmt.Errorf("package %s is not a darwin/%s Mach-O executable", relative, architecture)
			}
			if inputs != nil && inputs[relative] != entry.SHA256 {
				return Identity{}, fmt.Errorf("package %s digest does not match the attested release input", relative)
			}
		}
	}
	return identity, nil
}

func isSkillPath(value string) bool {
	rest, ok := strings.CutPrefix(value, "skills/")
	if !ok {
		return false
	}
	name, file, ok := strings.Cut(rest, "/")
	if !ok || file != "SKILL.md" || name == "" || len(name) > 64 {
		return false
	}
	for index, character := range name {
		lower := character >= 'a' && character <= 'z'
		digit := character >= '0' && character <= '9'
		if !(lower || digit || (character == '-' && index > 0)) {
			return false
		}
	}
	return true
}

func validateManifest(pkg Package, name string) error {
	var manifest struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	if err := decodeLenient(pkg, name, &manifest); err != nil || manifest.Name != PluginName || manifest.Version != pkg.Version {
		return errors.New("package host manifest identity is invalid")
	}
	return nil
}

func validateMarketplace(pkg Package, name string) (Identity, error) {
	if pkg.Host == "codex" {
		var catalog struct {
			Name    string `json:"name"`
			Plugins []struct {
				Name   string `json:"name"`
				Source struct {
					Source string `json:"source"`
					Path   string `json:"path"`
				} `json:"source"`
			} `json:"plugins"`
		}
		if err := decodeLenient(pkg, name, &catalog); err != nil || catalog.Name != MarketplaceName || len(catalog.Plugins) != 1 ||
			catalog.Plugins[0].Name != PluginName || catalog.Plugins[0].Source.Source != "local" || catalog.Plugins[0].Source.Path != "." {
			return Identity{}, errors.New("package Codex marketplace catalog is invalid")
		}
		return Identity{Marketplace: catalog.Name, Plugin: catalog.Plugins[0].Name}, nil
	}
	var catalog struct {
		Name    string `json:"name"`
		Plugins []struct {
			Name    string `json:"name"`
			Source  string `json:"source"`
			Version string `json:"version"`
		} `json:"plugins"`
	}
	if err := decodeLenient(pkg, name, &catalog); err != nil || catalog.Name != MarketplaceName || len(catalog.Plugins) != 1 ||
		catalog.Plugins[0].Name != PluginName || catalog.Plugins[0].Source != "." || catalog.Plugins[0].Version != pkg.Version {
		return Identity{}, errors.New("package Claude marketplace catalog is invalid")
	}
	return Identity{Marketplace: catalog.Name, Plugin: catalog.Plugins[0].Name}, nil
}

func validateChecksums(pkg Package) error {
	var checksums struct {
		Schema  int    `json:"schema"`
		Version string `json:"version"`
		Files   []struct {
			Path   string `json:"path"`
			SHA256 string `json:"sha256"`
			Size   int    `json:"size"`
		} `json:"files"`
	}
	if err := decodeLenient(pkg, "CHECKSUMS.json", &checksums); err != nil || checksums.Schema != 1 || checksums.Version != pkg.Version {
		return errors.New("package checksum metadata is invalid")
	}
	expected := make([]File, 0, len(pkg.Files))
	for _, file := range pkg.Files {
		if file.Path != "CHECKSUMS.json" && file.Path != "SBOM.spdx.json" {
			expected = append(expected, file)
		}
	}
	if len(checksums.Files) != len(expected) {
		return errors.New("package checksums do not cover the inventory")
	}
	for index, declared := range checksums.Files {
		file := expected[index]
		if declared.Path != file.Path || declared.Size != len(file.Data) || declared.SHA256 != file.SHA256 {
			return fmt.Errorf("package checksum for %q does not bind its bytes", file.Path)
		}
	}
	return nil
}

func validateMCP(pkg Package) (string, error) {
	var configuration struct {
		Servers map[string]struct {
			Type    string   `json:"type"`
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}
	if err := decodeLenient(pkg, ".mcp.json", &configuration); err != nil || len(configuration.Servers) != 1 {
		return "", errors.New("package MCP configuration is invalid")
	}
	server, ok := configuration.Servers["level7"]
	wantCommand := "./bin/l7"
	if pkg.Host == "claude" {
		wantCommand = "${CLAUDE_PLUGIN_ROOT}/bin/l7"
	}
	if !ok || server.Type != "stdio" || server.Command != wantCommand || len(server.Args) != 1 || server.Args[0] != "mcp" {
		return "", errors.New("package MCP command is invalid")
	}
	return server.Command, nil
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

func decodeLenient(pkg Package, name string, target any) error {
	entry := fileNamed(pkg, name)
	if entry == nil {
		return fmt.Errorf("package entry %q is missing", name)
	}
	if !json.Valid(entry.Data) {
		return fmt.Errorf("package entry %q is not JSON", name)
	}
	return json.Unmarshal(entry.Data, target)
}

func fileNamed(pkg Package, name string) *File {
	index := sort.Search(len(pkg.Files), func(index int) bool { return pkg.Files[index].Path >= name })
	if index == len(pkg.Files) || pkg.Files[index].Path != name {
		return nil
	}
	return &pkg.Files[index]
}

// SafePath accepts only clean, relative, slash-separated member paths.
func SafePath(value string) bool {
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

// Extract writes the package into a new directory. Nothing existing is ever
// overwritten; a failure removes only what this call created.
func Extract(root string, pkg Package) (returnErr error) {
	if !filepath.IsAbs(root) {
		return errors.New("extraction root must be absolute")
	}
	if _, err := os.Lstat(root); err == nil || !errors.Is(err, os.ErrNotExist) {
		return errors.New("extraction root is occupied")
	}
	if err := os.Mkdir(root, 0o755); err != nil {
		return fmt.Errorf("create extraction root: %w", err)
	}
	defer func() {
		if returnErr != nil {
			_ = os.RemoveAll(root)
		}
	}()
	for _, file := range pkg.Files {
		if !SafePath(file.Path) {
			return errors.New("package path is unsafe")
		}
		name := filepath.Join(root, filepath.FromSlash(file.Path))
		if !within(name, root) {
			return errors.New("package path escaped the extraction root")
		}
		if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
			return err
		}
		output, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, file.Mode.Perm())
		if err != nil {
			return err
		}
		written, writeErr := output.Write(file.Data)
		syncErr := output.Sync()
		closeErr := output.Close()
		if err := errors.Join(writeErr, syncErr, closeErr); err != nil || written != len(file.Data) {
			return errors.New("package file write was incomplete")
		}
	}
	return nil
}

func within(name, root string) bool {
	return name != root && strings.HasPrefix(filepath.Clean(name), filepath.Clean(root)+string(filepath.Separator))
}

// OwnedFile is one entry of an installer ownership receipt.
type OwnedFile struct {
	Path   string `json:"path"`
	Mode   string `json:"mode"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

// Receipt proves that a versioned directory was written by this installer
// from one verified archive, and records the trust facts it carried.
type Receipt struct {
	Schema        int         `json:"schema"`
	Host          string      `json:"host"`
	Version       string      `json:"version"`
	Tag           string      `json:"tag"`
	Channel       string      `json:"channel"`
	ArchiveName   string      `json:"archive_name"`
	ArchiveSHA256 string      `json:"archive_sha256"`
	SourceCommit  string      `json:"source_commit"`
	Workflow      string      `json:"workflow"`
	Policy        string      `json:"policy"`
	Unsigned      bool        `json:"unsigned"`
	Unqualified   bool        `json:"unqualified"`
	Warnings      []string    `json:"warnings"`
	VerifiedAtUTC string      `json:"verified_at_utc"`
	Marketplace   string      `json:"marketplace"`
	Plugin        string      `json:"plugin"`
	Files         []OwnedFile `json:"files"`
	StateSHA256   string      `json:"state_sha256"`
}

// Owned lists the package inventory for a receipt.
func Owned(pkg Package) []OwnedFile {
	files := make([]OwnedFile, 0, len(pkg.Files))
	for _, file := range pkg.Files {
		files = append(files, OwnedFile{Path: file.Path, Mode: fmt.Sprintf("%04o", file.Mode.Perm()), Size: int64(len(file.Data)), SHA256: file.SHA256})
	}
	return files
}

func WriteReceipt(name string, receipt Receipt) error {
	receipt.StateSHA256 = ""
	if receipt.Warnings == nil {
		receipt.Warnings = []string{}
	}
	digest, err := receiptDigest(receipt)
	if err != nil {
		return err
	}
	receipt.StateSHA256 = digest
	data, err := localfile.EncodeJSON(receipt)
	if err != nil {
		return err
	}
	return localfile.AtomicCreate(name, data, 0o644)
}

func LoadReceipt(name string) (Receipt, error) {
	data, err := localfile.Read(name, maxReceiptBytes)
	if err != nil {
		return Receipt{}, fmt.Errorf("ownership receipt is unreadable: %w", err)
	}
	var receipt Receipt
	if err := localfile.DecodeJSON(data, &receipt); err != nil {
		return Receipt{}, fmt.Errorf("ownership receipt is malformed: %w", err)
	}
	if receipt.Schema != 1 || !ValidHost(receipt.Host) || receipt.Version == "" || !validDigest(receipt.ArchiveSHA256) || len(receipt.Files) == 0 {
		return Receipt{}, errors.New("ownership receipt identity is invalid")
	}
	want := receipt.StateSHA256
	receipt.StateSHA256 = ""
	observed, err := receiptDigest(receipt)
	if err != nil || !validDigest(want) || observed != want {
		return Receipt{}, errors.New("ownership receipt state binding is stale")
	}
	receipt.StateSHA256 = want
	return receipt, nil
}

func receiptDigest(receipt Receipt) (string, error) {
	data, err := json.Marshal(receipt)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// VerifyTree proves a directory still holds exactly the receipted files.
func VerifyTree(root string, files []OwnedFile) error {
	if len(files) == 0 {
		return errors.New("ownership receipt is empty")
	}
	expected := make(map[string]OwnedFile, len(files))
	for _, file := range files {
		if !SafePath(file.Path) || expected[file.Path].Path != "" || !validDigest(file.SHA256) || (file.Mode != "0644" && file.Mode != "0755") || file.Size < 0 || file.Size > maxFileBytes {
			return errors.New("ownership receipt file set is invalid")
		}
		expected[file.Path] = file
	}
	if info, err := os.Lstat(root); err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("managed directory %s is missing or not a directory", root)
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
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("unexpected symlink %q in managed directory", relative)
		}
		if entry.IsDir() {
			return nil
		}
		wanted, ok := expected[relative]
		if !ok {
			return fmt.Errorf("unexpected file %q in managed directory", relative)
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != parseMode(wanted.Mode) || info.Size() != wanted.Size {
			return fmt.Errorf("managed file %q type, mode, or size changed", relative)
		}
		input, err := os.Open(name)
		if err != nil {
			return err
		}
		hash := sha256.New()
		written, copyErr := io.Copy(hash, io.LimitReader(input, maxFileBytes+1))
		closeErr := input.Close()
		if copyErr != nil || closeErr != nil || written != info.Size() || hex.EncodeToString(hash.Sum(nil)) != wanted.SHA256 {
			return fmt.Errorf("managed file %q content changed", relative)
		}
		seen[relative] = true
		return nil
	})
	if err != nil {
		return err
	}
	if len(seen) != len(expected) {
		return errors.New("managed directory is missing receipted files")
	}
	return nil
}

// RemoveTree deletes a directory only after proving it still matches the
// receipt, so unrelated or modified content is never destroyed.
func RemoveTree(root string, files []OwnedFile) error {
	if err := VerifyTree(root, files); err != nil {
		return fmt.Errorf("refusing to remove %s: %w", root, err)
	}
	return os.RemoveAll(root)
}

func parseMode(value string) fs.FileMode {
	if value == "0755" {
		return 0o755
	}
	return 0o644
}

func validDigest(value string) bool {
	if len(value) != 64 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

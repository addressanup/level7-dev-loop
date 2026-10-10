// Package pluginrelease resolves, downloads, and verifies Level 7 plugin
// releases from the one trusted GitHub repository. It owns the installer's
// only network path and the attestation trust boundary.
package pluginrelease

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
)

const (
	// Repository is the only release source. It is fixed on purpose: a
	// manifest, release listing, or environment variable can never redirect
	// installation to another repository.
	Repository = "addressanup/level7-dev-loop"
	// SourceRef is the only Git ref whose workflow runs may sign releases.
	SourceRef = "refs/heads/main"
	// PredicateType is the attestation predicate every release asset carries.
	PredicateType = "https://slsa.dev/provenance/v1"
	// ChecksumsAsset is the release-level checksum file name.
	ChecksumsAsset = "SHA256SUMS"
	// PackageName is the plugin identity inside every host archive.
	PackageName = "level7-dev-loop"

	maxManifestBytes = 1 << 20
	maxChecksumBytes = 64 << 10
	maxArchiveBytes  = 256 << 20
	maxJSONDepth     = 32
)

type Channel string

const (
	ChannelStable     Channel = "stable"
	ChannelPrerelease Channel = "prerelease"
)

func (channel Channel) Valid() bool { return channel == ChannelStable || channel == ChannelPrerelease }

// Workflow is one allowlisted release workflow and the manifest asset it
// publishes. Only these two identities may sign installable releases.
type Workflow struct {
	Path          string
	ManifestAsset string
	Channel       Channel
}

var Workflows = []Workflow{
	{Path: ".github/workflows/release.yml", ManifestAsset: "RELEASE-MANIFEST.json", Channel: ChannelStable},
	{Path: ".github/workflows/unsigned-prerelease.yml", ManifestAsset: "UNSIGNED-PRERELEASE-MANIFEST.json", Channel: ChannelPrerelease},
}

// WorkflowForManifest maps a release asset name to its allowlisted workflow.
func WorkflowForManifest(name string) (Workflow, bool) {
	for _, workflow := range Workflows {
		if workflow.ManifestAsset == name {
			return workflow, true
		}
	}
	return Workflow{}, false
}

// SignerIdentity is the exact certificate identity a workflow run must carry.
func (workflow Workflow) SignerIdentity() string {
	return "https://github.com/" + Repository + "/" + workflow.Path + "@" + SourceRef
}

type Asset struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// Qualification repeats the release's own disclosures. Anything that is not
// an explicit positive result is treated as absent, so a missing field can
// never read as signed or supported.
type Qualification struct {
	Policy             string   `json:"policy"`
	DeveloperIDSigning string   `json:"developer_id_signing"`
	Notarization       string   `json:"notarization"`
	FormalSupport      string   `json:"formal_support"`
	HoldoutEvaluation  string   `json:"holdout_evaluation"`
	ProviderTrials     string   `json:"provider_trials"`
	Gatekeeper         string   `json:"gatekeeper"`
	Waivers            []string `json:"waivers"`
}

// Unsigned reports whether the binaries lack a verified Developer ID
// signature or accepted notarization.
func (qualification Qualification) Unsigned() bool {
	return qualification.DeveloperIDSigning != "PASS" || qualification.Notarization != "ACCEPTED"
}

// Unqualified reports whether formal support is withheld or any release gate
// was not run.
func (qualification Qualification) Unqualified() bool {
	return qualification.FormalSupport != "SUPPORTED" || qualification.HoldoutEvaluation != "PASS" || qualification.ProviderTrials != "PASS"
}

// Warnings lists the user-facing facts a user accepts when installing.
func (qualification Qualification) Warnings() []string {
	var warnings []string
	if qualification.Unsigned() {
		warnings = append(warnings, "binaries are UNSIGNED and NOT NOTARIZED (developer_id_signing="+valueOr(qualification.DeveloperIDSigning)+", notarization="+valueOr(qualification.Notarization)+"); macOS Gatekeeper may block them; never disable Gatekeeper globally")
	}
	if qualification.Unqualified() {
		warnings = append(warnings, "formal support is "+valueOr(qualification.FormalSupport)+"; protected holdout evaluation="+valueOr(qualification.HoldoutEvaluation)+", exact-asset provider trials="+valueOr(qualification.ProviderTrials))
	}
	return warnings
}

// Contract is the minimal versioned update contract: everything the installer
// needs from one release, normalized across manifest shapes.
type Contract struct {
	Version       Version
	Tag           string
	Channel       Channel
	Workflow      Workflow
	SourceCommit  string
	SourceTree    string
	RunID         string
	Manifest      Asset
	Checksums     Asset
	Archives      map[string]Asset
	Inputs        map[string]string
	Qualification Qualification
}

// Archive returns the host archive recorded by the contract.
func (contract Contract) Archive(host string) (Asset, error) {
	asset, ok := contract.Archives[host]
	if !ok {
		return Asset{}, fmt.Errorf("release %s has no %s archive", contract.Tag, host)
	}
	return asset, nil
}

// DecodeManifest normalizes one allowlisted manifest asset. The bytes are not
// trusted yet: callers must verify the returned Manifest digest through its
// attestation before acting on anything else in the contract.
func DecodeManifest(name string, data []byte) (Contract, error) {
	workflow, ok := WorkflowForManifest(name)
	if !ok {
		return Contract{}, fmt.Errorf("%q is not an allowlisted release manifest", name)
	}
	if len(data) == 0 || len(data) > maxManifestBytes {
		return Contract{}, errors.New("release manifest is empty or exceeds its size bound")
	}
	if err := validateJSONShape(data); err != nil {
		return Contract{}, fmt.Errorf("release manifest JSON is invalid: %w", err)
	}
	sum := digestOf(data)
	var contract Contract
	var err error
	switch workflow.Channel {
	case ChannelStable:
		contract, err = decodeStableManifest(data)
	default:
		contract, err = decodePrereleaseManifest(data)
	}
	if err != nil {
		return Contract{}, err
	}
	contract.Workflow = workflow
	contract.Channel = workflow.Channel
	contract.Manifest = Asset{Name: name, SHA256: sum, Size: int64(len(data))}
	if err := contract.validate(); err != nil {
		return Contract{}, err
	}
	return contract, nil
}

type stableManifest struct {
	Schema         int    `json:"schema"`
	Version        string `json:"version"`
	Tag            string `json:"tag"`
	ReleaseChannel string `json:"release_channel"`
	Policy         string `json:"policy"`
	Gatekeeper     string `json:"gatekeeper"`
	Candidate      struct {
		Commit string `json:"commit"`
		Tree   string `json:"tree"`
	} `json:"candidate"`
	Workflow struct {
		Repository string `json:"repository"`
		Path       string `json:"path"`
		RunID      string `json:"run_id"`
	} `json:"workflow"`
	Qualification struct {
		DeveloperIDSigning string `json:"developer_id_signing"`
		AppleNotarization  string `json:"apple_notarization"`
		Holdout            string `json:"protected_holdout_evaluation"`
		ProviderTrials     string `json:"exact_asset_provider_trials"`
		FormalSupport      string `json:"formal_support"`
	} `json:"qualification"`
	Waivers          map[string]string `json:"waivers"`
	Artifacts        []Asset           `json:"artifacts"`
	Checksums        Asset             `json:"checksums"`
	UnsignedInputs   map[string]string `json:"unsigned_inputs"`
	UnsignedPackages map[string]string `json:"unsigned_packages"`
}

func decodeStableManifest(data []byte) (Contract, error) {
	var manifest stableManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Contract{}, errors.New("stable release manifest does not match the expected shape")
	}
	if manifest.Schema != 1 || manifest.ReleaseChannel != "stable" || manifest.Policy == "" {
		return Contract{}, errors.New("stable release manifest schema, channel, or policy is unsupported")
	}
	if stable, ok := WorkflowForManifest("RELEASE-MANIFEST.json"); !ok || manifest.Workflow.Repository != Repository || manifest.Workflow.Path != stable.Path {
		return Contract{}, errors.New("stable release manifest names a non-allowlisted repository or workflow")
	}
	waivers := make([]string, 0, len(manifest.Waivers))
	for gate, decision := range manifest.Waivers {
		if decision == "WAIVED" {
			waivers = append(waivers, gate)
		}
	}
	sort.Strings(waivers)
	contract := Contract{
		Tag: manifest.Tag, SourceCommit: manifest.Candidate.Commit, SourceTree: manifest.Candidate.Tree, RunID: manifest.Workflow.RunID,
		Checksums: manifest.Checksums, Inputs: manifest.UnsignedInputs,
		Qualification: Qualification{
			Policy: manifest.Policy, DeveloperIDSigning: manifest.Qualification.DeveloperIDSigning, Notarization: manifest.Qualification.AppleNotarization,
			FormalSupport: manifest.Qualification.FormalSupport, HoldoutEvaluation: manifest.Qualification.Holdout, ProviderTrials: manifest.Qualification.ProviderTrials,
			Gatekeeper: manifest.Gatekeeper, Waivers: waivers,
		},
	}
	version, err := ParseVersion(manifest.Version)
	if err != nil {
		return Contract{}, fmt.Errorf("stable release manifest version: %w", err)
	}
	contract.Version = version
	if contract.Archives, err = archivesFor(version, manifest.Artifacts, manifest.UnsignedPackages); err != nil {
		return Contract{}, err
	}
	return contract, nil
}

type prereleaseManifest struct {
	Schema    int    `json:"schema"`
	Version   string `json:"version"`
	Tag       string `json:"tag"`
	Channel   string `json:"channel"`
	ChangeID  string `json:"change_id"`
	Candidate struct {
		Commit string `json:"commit"`
		Tree   string `json:"tree"`
	} `json:"candidate"`
	Workflow struct {
		Repository string `json:"repository"`
		RunID      string `json:"run_id"`
	} `json:"workflow"`
	DeveloperIDSigning string  `json:"developer_id_signing"`
	Notarization       string  `json:"notarization"`
	Support            string  `json:"support"`
	ProviderExecution  string  `json:"provider_execution"`
	Gatekeeper         string  `json:"gatekeeper"`
	Artifacts          []Asset `json:"artifacts"`
	Checksums          Asset   `json:"checksums"`
	UnsignedInputs     []struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
	} `json:"unsigned_inputs"`
	UnsignedPackages map[string]string `json:"unsigned_packages"`
}

func decodePrereleaseManifest(data []byte) (Contract, error) {
	var manifest prereleaseManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Contract{}, errors.New("prerelease manifest does not match the expected shape")
	}
	if manifest.Schema != 1 || manifest.Channel != "development-candidate" || manifest.ChangeID == "" {
		return Contract{}, errors.New("prerelease manifest schema, channel, or change identity is unsupported")
	}
	if manifest.Workflow.Repository != Repository {
		return Contract{}, errors.New("prerelease manifest names a non-allowlisted repository")
	}
	inputs := make(map[string]string, len(manifest.UnsignedInputs))
	for _, input := range manifest.UnsignedInputs {
		if _, duplicate := inputs[input.Path]; duplicate {
			return Contract{}, errors.New("prerelease manifest repeats an unsigned input")
		}
		inputs[input.Path] = input.SHA256
	}
	contract := Contract{
		Tag: manifest.Tag, SourceCommit: manifest.Candidate.Commit, SourceTree: manifest.Candidate.Tree, RunID: manifest.Workflow.RunID,
		Checksums: manifest.Checksums, Inputs: inputs,
		Qualification: Qualification{
			Policy: manifest.ChangeID, DeveloperIDSigning: manifest.DeveloperIDSigning, Notarization: manifest.Notarization,
			FormalSupport: manifest.Support, HoldoutEvaluation: "NOT_RUN", ProviderTrials: manifest.ProviderExecution, Gatekeeper: manifest.Gatekeeper,
		},
	}
	version, err := ParseVersion(manifest.Version)
	if err != nil {
		return Contract{}, fmt.Errorf("prerelease manifest version: %w", err)
	}
	if !version.Prerelease() {
		return Contract{}, errors.New("prerelease manifest carries a release version")
	}
	contract.Version = version
	if contract.Archives, err = archivesFor(version, manifest.Artifacts, manifest.UnsignedPackages); err != nil {
		return Contract{}, err
	}
	return contract, nil
}

func archivesFor(version Version, artifacts []Asset, packages map[string]string) (map[string]Asset, error) {
	if len(artifacts) != 2 {
		return nil, errors.New("release manifest must record exactly two host archives")
	}
	archives := make(map[string]Asset, 2)
	for _, host := range []string{"codex", "claude"} {
		name := ArchiveName(version, host)
		var found *Asset
		for index := range artifacts {
			if artifacts[index].Name == name {
				found = &artifacts[index]
			}
		}
		if found == nil || !ValidDigest(found.SHA256) || found.Size < 1 || found.Size > maxArchiveBytes {
			return nil, fmt.Errorf("release manifest does not record a bounded %s archive", host)
		}
		if packages != nil && packages[name] != found.SHA256 {
			return nil, fmt.Errorf("release manifest package digests disagree for %s", name)
		}
		archives[host] = *found
	}
	return archives, nil
}

// ArchiveName is the fixed host archive naming scheme.
func ArchiveName(version Version, host string) string {
	return PackageName + "-" + version.String() + "-" + host + ".zip"
}

func (contract Contract) validate() error {
	if contract.Tag != "v"+contract.Version.String() {
		return errors.New("release manifest tag does not match its version")
	}
	if !validCommit(contract.SourceCommit) || !validCommit(contract.SourceTree) {
		return errors.New("release manifest source commit or tree is invalid")
	}
	if contract.RunID == "" || !numeric(contract.RunID) {
		return errors.New("release manifest workflow run is invalid")
	}
	if contract.Checksums.Name != ChecksumsAsset || !ValidDigest(contract.Checksums.SHA256) || contract.Checksums.Size < 1 || contract.Checksums.Size > maxChecksumBytes {
		return errors.New("release manifest checksum asset is invalid")
	}
	for _, input := range []string{"darwin-arm64/l7", "darwin-arm64/l7-embed", "darwin-amd64/l7", "darwin-amd64/l7-embed"} {
		if !ValidDigest(contract.Inputs[input]) {
			return fmt.Errorf("release manifest lacks a digest for input %s", input)
		}
	}
	if contract.Qualification.Policy == "" {
		return errors.New("release manifest lacks a policy identity")
	}
	return nil
}

// ParseChecksums parses a SHA256SUMS document and requires it to list exactly
// the contract's host archives with the contract's digests.
func ParseChecksums(contract Contract, data []byte) error {
	if len(data) == 0 || len(data) > maxChecksumBytes {
		return errors.New("SHA256SUMS is empty or exceeds its size bound")
	}
	seen := make(map[string]string)
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		digest, name, ok := strings.Cut(line, "  ")
		if !ok || !ValidDigest(digest) || name == "" || strings.ContainsAny(name, "/\\ \t\r") {
			return errors.New("SHA256SUMS line is malformed")
		}
		if _, duplicate := seen[name]; duplicate {
			return errors.New("SHA256SUMS repeats an asset")
		}
		seen[name] = digest
	}
	if len(seen) != len(contract.Archives) {
		return errors.New("SHA256SUMS does not list exactly the release archives")
	}
	for _, archive := range contract.Archives {
		if seen[archive.Name] != archive.SHA256 {
			return fmt.Errorf("SHA256SUMS disagrees with the manifest for %s", archive.Name)
		}
	}
	return nil
}

func ValidDigest(value string) bool {
	if len(value) != 64 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func validCommit(value string) bool {
	if len(value) != 40 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func valueOr(value string) string {
	if value == "" {
		return "UNKNOWN"
	}
	return value
}

// validateJSONShape rejects duplicate keys, unbounded nesting, and trailing
// data before any lenient decode.
func validateJSONShape(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value func(int) error
	value = func(depth int) error {
		if depth > maxJSONDepth {
			return errors.New("nesting exceeds bound")
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
					return errors.New("object key is invalid or duplicated")
				}
				seen[key] = true
				if err := value(depth + 1); err != nil {
					return err
				}
			}
			closing, err := decoder.Token()
			if err != nil || closing != json.Delim('}') {
				return errors.New("object is unterminated")
			}
		case '[':
			for decoder.More() {
				if err := value(depth + 1); err != nil {
					return err
				}
			}
			closing, err := decoder.Token()
			if err != nil || closing != json.Delim(']') {
				return errors.New("array is unterminated")
			}
		default:
			return errors.New("delimiter is invalid")
		}
		return nil
	}
	if err := value(0); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return errors.New("trailing data")
	}
	return nil
}

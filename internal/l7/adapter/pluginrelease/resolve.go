package pluginrelease

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
)

// Selection names which release to install: the channel (stable unless the
// user explicitly asks for prereleases) and an optional exact version.
type Selection struct {
	Channel Channel
	Exact   string
}

// Select picks one release. Stable ignores prerelease-flagged releases; an
// exact version is honored only when its channel was selected explicitly.
func Select(releases []Release, selection Selection) (Release, error) {
	if !selection.Channel.Valid() {
		return Release{}, errors.New("release channel must be stable or prerelease")
	}
	if selection.Exact != "" {
		exact, err := ParseVersion(selection.Exact)
		if err != nil {
			return Release{}, fmt.Errorf("pinned version: %w", err)
		}
		for _, release := range releases {
			if release.Version.Compare(exact) == 0 {
				if release.Prerelease && selection.Channel != ChannelPrerelease {
					return Release{}, fmt.Errorf("%s is a prerelease; pass --channel prerelease to select it explicitly", release.Tag)
				}
				return release, nil
			}
		}
		return Release{}, fmt.Errorf("release v%s is not available; installable releases: %s", exact.String(), describeReleases(releases))
	}
	var chosen *Release
	for index := range releases {
		release := &releases[index]
		if release.Prerelease && selection.Channel != ChannelPrerelease {
			continue
		}
		if chosen == nil || release.Version.Compare(chosen.Version) > 0 {
			chosen = release
		}
	}
	if chosen == nil {
		return Release{}, fmt.Errorf("no %s release is available", selection.Channel)
	}
	return *chosen, nil
}

func describeReleases(releases []Release) string {
	tags := make([]string, 0, len(releases))
	for _, release := range releases {
		tags = append(tags, release.Tag)
	}
	sort.Strings(tags)
	if len(tags) == 0 {
		return "none"
	}
	return strings.Join(tags, ", ")
}

// Resolved is one release whose manifest and checksums are verified.
type Resolved struct {
	Release   Release
	Contract  Contract
	GHVersion string
	Evidence  map[string]Evidence
	gh        processadapter.Executable
}

// Resolve downloads the release manifest and checksums into stage and binds
// both to the trusted workflow and exact source commit before returning.
func (client Client) Resolve(ctx context.Context, release Release, stage string) (Resolved, error) {
	if err := ensureStage(stage); err != nil {
		return Resolved{}, err
	}
	gh, ghVersion, err := client.verifier.Prerequisite(ctx)
	if err != nil {
		return Resolved{}, err
	}
	manifestName := release.Manifest.ManifestAsset
	size, ok := release.Assets[manifestName]
	if !ok || size < 1 || size > maxManifestBytes {
		return Resolved{}, fmt.Errorf("release %s does not attach a bounded %s", release.Tag, manifestName)
	}
	manifestPath, _, err := client.Download(ctx, release.Tag, manifestName, size, stage)
	if err != nil {
		return Resolved{}, err
	}
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return Resolved{}, fmt.Errorf("cannot read downloaded manifest: %w", err)
	}
	contract, err := DecodeManifest(manifestName, data)
	if err != nil {
		return Resolved{}, err
	}
	if contract.Tag != release.Tag || contract.Workflow != release.Manifest {
		return Resolved{}, fmt.Errorf("manifest inside %s describes %s; refusing the mismatch", release.Tag, contract.Tag)
	}
	resolved := Resolved{Release: release, Contract: contract, GHVersion: ghVersion, Evidence: make(map[string]Evidence, 3), gh: gh}
	evidence, err := client.attest(ctx, gh, manifestPath, stage, Expectation{Workflow: contract.Workflow, SourceCommit: contract.SourceCommit, SubjectName: manifestName, SubjectSHA256: contract.Manifest.SHA256})
	if err != nil {
		return Resolved{}, err
	}
	resolved.Evidence[manifestName] = evidence

	checksumSize, ok := release.Assets[ChecksumsAsset]
	if !ok || checksumSize != contract.Checksums.Size {
		return Resolved{}, fmt.Errorf("release %s does not attach the %s the manifest describes", release.Tag, ChecksumsAsset)
	}
	checksumPath, checksumDigest, err := client.Download(ctx, release.Tag, ChecksumsAsset, contract.Checksums.Size, stage)
	if err != nil {
		return Resolved{}, err
	}
	if checksumDigest != contract.Checksums.SHA256 {
		return Resolved{}, errors.New("SHA256SUMS digest does not match the attested manifest")
	}
	checksumData, err := os.ReadFile(checksumPath)
	if err != nil {
		return Resolved{}, fmt.Errorf("cannot read downloaded checksums: %w", err)
	}
	if err := ParseChecksums(contract, checksumData); err != nil {
		return Resolved{}, err
	}
	evidence, err = client.attest(ctx, gh, checksumPath, stage, Expectation{Workflow: contract.Workflow, SourceCommit: contract.SourceCommit, SubjectName: ChecksumsAsset, SubjectSHA256: checksumDigest})
	if err != nil {
		return Resolved{}, err
	}
	resolved.Evidence[ChecksumsAsset] = evidence
	return resolved, nil
}

// FetchArchive obtains the verified host archive, reusing a cached copy only
// when its bytes still match the attested digest.
func (client Client) FetchArchive(ctx context.Context, resolved Resolved, host, stage, cache string) (string, Evidence, error) {
	if resolved.gh.Path == "" {
		return "", Evidence{}, errors.New("release was not resolved through the verifier")
	}
	asset, err := resolved.Contract.Archive(host)
	if err != nil {
		return "", Evidence{}, err
	}
	if size, ok := resolved.Release.Assets[asset.Name]; !ok || size != asset.Size {
		return "", Evidence{}, fmt.Errorf("release %s does not attach %s with the manifest's size", resolved.Release.Tag, asset.Name)
	}
	if err := ensureStage(stage); err != nil {
		return "", Evidence{}, err
	}
	path := ""
	cachedPath := filepath.Join(cache, resolved.Release.Tag, asset.Name)
	if cache != "" {
		if digest, size, err := FileDigest(cachedPath, maxArchiveBytes); err == nil && digest == asset.SHA256 && size == asset.Size {
			path = cachedPath
		}
	}
	if path == "" {
		downloaded, digest, err := client.Download(ctx, resolved.Release.Tag, asset.Name, asset.Size, stage)
		if err != nil {
			return "", Evidence{}, err
		}
		if digest != asset.SHA256 {
			_ = os.Remove(downloaded)
			return "", Evidence{}, fmt.Errorf("%s digest %s does not match the attested manifest digest %s", asset.Name, digest, asset.SHA256)
		}
		path = downloaded
		if cache != "" {
			// A stale or corrupt cache entry is replaced by the verified bytes;
			// anything that is not a regular file is left alone.
			if err := os.MkdirAll(filepath.Dir(cachedPath), 0o700); err == nil {
				if info, statErr := os.Lstat(cachedPath); (statErr != nil || info.Mode().IsRegular()) && os.Rename(downloaded, cachedPath) == nil {
					path = cachedPath
				}
			}
		}
	}
	evidence, err := client.attest(ctx, resolved.gh, path, stage, Expectation{Workflow: resolved.Contract.Workflow, SourceCommit: resolved.Contract.SourceCommit, SubjectName: asset.Name, SubjectSHA256: asset.SHA256})
	if err != nil {
		return "", Evidence{}, err
	}
	return path, evidence, nil
}

func (client Client) attest(ctx context.Context, gh processadapter.Executable, artifactPath, stage string, expect Expectation) (Evidence, error) {
	bundlePath := filepath.Join(stage, filepath.Base(artifactPath)+".attestations.jsonl")
	if _, err := client.AttestationBundles(ctx, expect.SubjectSHA256, bundlePath); err != nil {
		return Evidence{}, fmt.Errorf("%s: %w", expect.SubjectName, err)
	}
	return client.verifier.Verify(ctx, gh, artifactPath, bundlePath, expect)
}

func ensureStage(stage string) error {
	if !filepath.IsAbs(stage) {
		return errors.New("staging directory must be absolute")
	}
	info, err := os.Lstat(stage)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("staging directory is missing or unsafe")
	}
	return nil
}

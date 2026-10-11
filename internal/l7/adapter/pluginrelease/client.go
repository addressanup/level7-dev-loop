package pluginrelease

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	maxAPIBytes         = 4 << 20
	maxAttestationBytes = 16 << 20
	maxRedirects        = 5
	requestTimeout      = 10 * time.Minute
)

// Doer is the bounded HTTP port; production uses NewHTTPClient.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// Source pins where releases are listed and downloaded. Only tests override
// it; production always uses DefaultSource.
type Source struct {
	APIBase      string
	DownloadBase string
	// AllowedHosts bounds every redirect destination.
	AllowedHosts []string
	// InsecureLoopback admits plain HTTP to loopback addresses for fixture
	// servers only. Production leaves it false.
	InsecureLoopback bool
}

func DefaultSource() Source {
	return Source{
		APIBase:      "https://api.github.com",
		DownloadBase: "https://github.com",
		AllowedHosts: []string{"api.github.com", "github.com", "objects.githubusercontent.com", "release-assets.githubusercontent.com"},
	}
}

// NewHTTPClient builds the installer's HTTPS client: no proxy, TLS 1.2+, a
// bounded redirect chain, and destinations restricted to the source hosts.
func NewHTTPClient(source Source, transport http.RoundTripper) *http.Client {
	if transport == nil {
		transport = &http.Transport{
			Proxy:                 nil,
			TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
			DialContext:           (&net.Dialer{Timeout: 30 * time.Second}).DialContext,
			TLSHandshakeTimeout:   30 * time.Second,
			ResponseHeaderTimeout: 60 * time.Second,
			MaxIdleConns:          4,
		}
	}
	return &http.Client{
		Transport: transport,
		Timeout:   requestTimeout,
		CheckRedirect: func(request *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return errors.New("release download exceeded the redirect bound")
			}
			if err := source.allow(request.URL); err != nil {
				return err
			}
			// Credentials are never attached, and a redirect must not acquire any.
			request.Header.Del("Authorization")
			request.Header.Del("Cookie")
			return nil
		},
	}
}

func (source Source) allow(target *url.URL) error {
	host := target.Hostname()
	if target.Scheme != "https" {
		if !(source.InsecureLoopback && target.Scheme == "http" && isLoopback(host)) {
			return errors.New("release download must stay on HTTPS")
		}
	}
	for _, allowed := range source.AllowedHosts {
		if host == allowed {
			return nil
		}
	}
	return fmt.Errorf("release download redirected to unexpected host %q", host)
}

func isLoopback(host string) bool {
	if host == "localhost" {
		return true
	}
	address := net.ParseIP(host)
	return address != nil && address.IsLoopback()
}

// Client talks to the trusted repository only. Every response is bounded and
// decoded strictly before use; nothing it returns is trusted until the
// attestation verifier has bound it to the exact source commit.
type Client struct {
	http      Doer
	source    Source
	verifier  Verifier
	userAgent string
}

func NewClient(doer Doer, source Source, verifier Verifier, cliVersion string) Client {
	if doer == nil {
		doer = NewHTTPClient(source, nil)
	}
	return Client{http: doer, source: source, verifier: verifier, userAgent: "level7-dev-loop-plugin/" + strings.Map(safeAgentRune, cliVersion)}
}

func safeAgentRune(character rune) rune {
	if (character >= '0' && character <= '9') || (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') || character == '.' || character == '-' {
		return character
	}
	return -1
}

// Release is one GitHub release as listed by the API.
type Release struct {
	Tag        string
	Version    Version
	Draft      bool
	Prerelease bool
	Immutable  bool
	Assets     map[string]int64
	Manifest   Workflow
}

type apiRelease struct {
	TagName    string `json:"tag_name"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
	Immutable  bool   `json:"immutable"`
	Assets     []struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	} `json:"assets"`
}

// Candidates lists the repository's installable releases: non-draft releases
// whose tag is a strict version and which carry exactly one allowlisted
// manifest asset. Others are ignored, never guessed about.
func (client Client) Candidates(ctx context.Context) ([]Release, error) {
	data, err := client.api(ctx, "/repos/"+Repository+"/releases?per_page=50", maxAPIBytes)
	if err != nil {
		return nil, err
	}
	var listed []apiRelease
	if err := json.Unmarshal(data, &listed); err != nil {
		return nil, errors.New("release listing does not match the expected shape")
	}
	releases := make([]Release, 0, len(listed))
	for _, entry := range listed {
		release, ok := normalizeRelease(entry)
		if ok {
			releases = append(releases, release)
		}
	}
	return releases, nil
}

// ReleaseByTag fetches one exact release.
func (client Client) ReleaseByTag(ctx context.Context, tag string) (Release, error) {
	if !validTag(tag) {
		return Release{}, errors.New("release tag is invalid")
	}
	data, err := client.api(ctx, "/repos/"+Repository+"/releases/tags/"+tag, maxAPIBytes)
	if err != nil {
		return Release{}, err
	}
	var entry apiRelease
	if err := json.Unmarshal(data, &entry); err != nil {
		return Release{}, errors.New("release record does not match the expected shape")
	}
	release, ok := normalizeRelease(entry)
	if !ok || release.Tag != tag {
		return Release{}, fmt.Errorf("release %s is not an installable Level 7 release", tag)
	}
	return release, nil
}

func normalizeRelease(entry apiRelease) (Release, bool) {
	if entry.Draft || !validTag(entry.TagName) {
		return Release{}, false
	}
	version, err := ParseVersion(strings.TrimPrefix(entry.TagName, "v"))
	if err != nil {
		return Release{}, false
	}
	release := Release{Tag: entry.TagName, Version: version, Draft: entry.Draft, Prerelease: entry.Prerelease, Immutable: entry.Immutable, Assets: make(map[string]int64, len(entry.Assets))}
	manifests := 0
	for _, asset := range entry.Assets {
		if asset.Name == "" || len(asset.Name) > 255 || strings.ContainsAny(asset.Name, "/\\\x00") {
			return Release{}, false
		}
		release.Assets[asset.Name] = asset.Size
		if workflow, ok := WorkflowForManifest(asset.Name); ok {
			release.Manifest = workflow
			manifests++
		}
	}
	if manifests != 1 {
		return Release{}, false
	}
	// A stable manifest on a prerelease-flagged release, or the reverse, is a
	// publication mistake the installer refuses to interpret.
	if (release.Manifest.Channel == ChannelStable) == release.Prerelease {
		return Release{}, false
	}
	return release, true
}

func validTag(tag string) bool {
	if len(tag) < 2 || len(tag) > 65 || tag[0] != 'v' {
		return false
	}
	_, err := ParseVersion(tag[1:])
	return err == nil
}

func (client Client) api(ctx context.Context, path string, limit int64) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, client.source.APIBase+path, nil)
	if err != nil {
		return nil, errors.New("release API request could not be created")
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	request.Header.Set("User-Agent", client.userAgent)
	response, err := client.http.Do(request)
	if err != nil {
		return nil, fmt.Errorf("release API request failed: %w", sanitizeTransportError(err))
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return nil, errors.New("release API returned 404 (release or asset does not exist)")
	}
	if response.StatusCode == http.StatusForbidden || response.StatusCode == http.StatusTooManyRequests {
		return nil, errors.New("GitHub API rate limit or access limit reached; retry later (unauthenticated installs use the public API)")
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("release API returned status %d", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil || int64(len(data)) > limit || len(data) == 0 {
		return nil, errors.New("release API response is empty or exceeds its size bound")
	}
	if err := validateJSONShape(data); err != nil {
		return nil, fmt.Errorf("release API JSON is invalid: %w", err)
	}
	return data, nil
}

// Download fetches one release asset to a new file in directory and returns
// its path and digest. The response is bounded by the expected size.
func (client Client) Download(ctx context.Context, tag, asset string, expectedSize int64, directory string) (string, string, error) {
	if !validTag(tag) || asset == "" || strings.ContainsAny(asset, "/\\\x00") || expectedSize < 1 || expectedSize > maxArchiveBytes {
		return "", "", errors.New("release asset request is invalid")
	}
	target := filepath.Join(directory, asset)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, client.source.DownloadBase+"/"+Repository+"/releases/download/"+tag+"/"+url.PathEscape(asset), nil)
	if err != nil {
		return "", "", errors.New("release download request could not be created")
	}
	request.Header.Set("Accept", "application/octet-stream")
	request.Header.Set("User-Agent", client.userAgent)
	response, err := client.http.Do(request)
	if err != nil {
		return "", "", fmt.Errorf("release download failed: %w", sanitizeTransportError(err))
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("release download of %s returned status %d", asset, response.StatusCode)
	}
	if response.ContentLength > expectedSize {
		return "", "", fmt.Errorf("release asset %s is larger than the manifest declares", asset)
	}
	output, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", "", fmt.Errorf("cannot create download file: %w", err)
	}
	hash := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(output, hash), io.LimitReader(response.Body, expectedSize+1))
	syncErr := output.Sync()
	closeErr := output.Close()
	if err := errors.Join(copyErr, syncErr, closeErr); err != nil {
		_ = os.Remove(target)
		return "", "", fmt.Errorf("release download of %s was interrupted: %w", asset, sanitizeTransportError(err))
	}
	if written != expectedSize {
		_ = os.Remove(target)
		return "", "", fmt.Errorf("release asset %s has %d bytes, manifest declares %d", asset, written, expectedSize)
	}
	return target, hex.EncodeToString(hash.Sum(nil)), nil
}

// AttestationBundles fetches the Sigstore bundles GitHub stores for one
// subject digest and writes them as JSON lines for gh to verify offline.
func (client Client) AttestationBundles(ctx context.Context, digest, path string) (int, error) {
	if !ValidDigest(digest) {
		return 0, errors.New("attestation subject digest is invalid")
	}
	data, err := client.api(ctx, "/repos/"+Repository+"/attestations/sha256:"+digest+"?per_page=30", maxAttestationBytes)
	if err != nil {
		return 0, fmt.Errorf("attestation lookup: %w", err)
	}
	var document struct {
		Attestations []struct {
			Bundle json.RawMessage `json:"bundle"`
		} `json:"attestations"`
	}
	if err := json.Unmarshal(data, &document); err != nil || len(document.Attestations) == 0 {
		return 0, errors.New("no GitHub attestation exists for this asset digest")
	}
	// Bundles are compacted byte-for-byte (no re-encoding) so signatures and
	// large integers reach gh exactly as GitHub stored them.
	var lines bytes.Buffer
	for _, attestation := range document.Attestations {
		if len(attestation.Bundle) == 0 || attestation.Bundle[0] != '{' {
			return 0, errors.New("attestation record lacks a bundle object")
		}
		if err := json.Compact(&lines, attestation.Bundle); err != nil {
			return 0, errors.New("attestation bundle is malformed")
		}
		lines.WriteByte('\n')
	}
	if err := os.WriteFile(path, lines.Bytes(), 0o600); err != nil {
		return 0, fmt.Errorf("cannot write attestation bundle: %w", err)
	}
	return len(document.Attestations), nil
}

// sanitizeTransportError keeps transport failures informative without echoing
// response bodies or query strings.
func sanitizeTransportError(err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return fmt.Errorf("%s %s: %s", urlErr.Op, redactURL(urlErr.URL), urlErr.Err)
	}
	return err
}

func redactURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "<url>"
	}
	parsed.RawQuery = ""
	parsed.User = nil
	return parsed.String()
}

func digestOf(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// FileDigest hashes a bounded regular file.
func FileDigest(path string, limit int64) (string, int64, error) {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() > limit {
		return "", 0, errors.New("file is missing, not regular, or exceeds its size bound")
	}
	file, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()
	hash := sha256.New()
	written, err := io.Copy(hash, io.LimitReader(file, limit+1))
	if err != nil || written != info.Size() {
		return "", 0, errors.New("file changed while hashing")
	}
	return hex.EncodeToString(hash.Sum(nil)), written, nil
}

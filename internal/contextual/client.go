package contextual

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"scout/internal/domain"
)

const (
	defaultTimeout      = 45 * time.Second
	maximumTimeout      = 2 * time.Minute
	defaultContextBytes = 96 * 1024
	defaultFileBytes    = 24 * 1024
	defaultPromptSlack  = 64 * 1024
	maximumContextBytes = 512 * 1024
	maximumPromptBytes  = 1024 * 1024
)

// Config describes a configured JSON/HTTP contextual provider. Authorization is
// intentionally separate from credentials: a credential alone never authorizes source egress.
type Config struct {
	Endpoint          string
	Model             string
	PromptVersion     string
	CredentialEnv     string
	AuthorizationHead string
	AuthorizationPref string
	EgressAuthorized  bool
	Timeout           time.Duration
	MaxContextBytes   int
	MaxFileBytes      int
	MaxPromptBytes    int
	HTTPClient        *http.Client
}

// Client sends only redacted, bounded, explicitly authorized source excerpts.
type Client struct{ config Config }

// NewClient normalizes safe limits. It permits an unavailable provider so a
// caller can obtain a truthful blocked coverage row from Review.
func NewClient(config Config) (*Client, error) {
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}
	if config.Timeout < time.Second || config.Timeout > maximumTimeout {
		return nil, fmt.Errorf("contextual timeout must be between %s and %s", time.Second, maximumTimeout)
	}
	if config.MaxContextBytes == 0 {
		config.MaxContextBytes = defaultContextBytes
	}
	if config.MaxFileBytes == 0 {
		config.MaxFileBytes = defaultFileBytes
	}
	if config.MaxContextBytes < 1 || config.MaxContextBytes > maximumContextBytes {
		return nil, fmt.Errorf("contextual max context must be between 1 and %d bytes", maximumContextBytes)
	}
	if config.MaxFileBytes < 1 || config.MaxFileBytes > config.MaxContextBytes {
		return nil, errors.New("contextual max file bytes must be positive and no greater than max context")
	}
	if config.MaxPromptBytes == 0 {
		config.MaxPromptBytes = config.MaxContextBytes + defaultPromptSlack
	}
	if config.MaxPromptBytes < 1 || config.MaxPromptBytes > maximumPromptBytes {
		return nil, fmt.Errorf("contextual max prompt must be between 1 and %d bytes", maximumPromptBytes)
	}
	if config.AuthorizationHead == "" {
		config.AuthorizationHead = "Authorization"
	}
	if config.AuthorizationPref == "" {
		config.AuthorizationPref = "Bearer "
	}
	return &Client{config: config}, nil
}

// ReviewRequest contains the canonical feature and only already-sanitized tool
// results and observations relevant to one feature/category review.
type ReviewRequest struct {
	Root              string
	Snapshot          string
	RunID             string
	Feature           domain.Feature
	Family            domain.Family
	Category          string
	PromptVersion     string
	ToolResults       []domain.AdapterResult
	PriorObservations []domain.Observation
}

// ManifestEntry identifies a bounded source snapshot sent to the provider.
// SourceDigest covers the full, unredacted file read before egress; content is
// deliberately excluded so the manifest can be persisted safely.
type ManifestEntry struct {
	Range        domain.SourceRange `json:"range"`
	Reason       string             `json:"reason"`
	Bytes        int                `json:"bytes"`
	SourceDigest string             `json:"sourceDigest"`
}

// ManifestOmission makes source that could not be faithfully included visible.
type ManifestOmission struct {
	Path   string `json:"path,omitempty"`
	Reason string `json:"reason"`
}

// Manifest records actual egress input rather than trusting a model assertion.
type Manifest struct {
	Inspected []ManifestEntry    `json:"inspected,omitempty"`
	Omissions []ManifestOmission `json:"omissions,omitempty"`
	Truncated bool               `json:"truncated"`
}

// ReviewMetadata retains the safe contextual details not represented by domain
// coverage or observations. Reproductions are keyed only by validated observation IDs.
type ReviewMetadata struct {
	DeterministicBlindSpot       string            `json:"deterministicBlindSpot"`
	ReproductionByObservationID map[string]string `json:"reproductionByObservationId,omitempty"`
}

// ReviewResult retains usable, validated observations when later source or provider
// failures make the category incomplete.
type ReviewResult struct {
	Coverage     domain.Coverage      `json:"coverage"`
	Observations []domain.Observation `json:"observations,omitempty"`
	Manifest     Manifest             `json:"manifest"`
	Metadata     *ReviewMetadata      `json:"metadata,omitempty"`
}

// Review performs one feature × category provider assessment. It never calls a
// provider for a category the integration elected to reuse.
func (c *Client) Review(ctx context.Context, request ReviewRequest) (ReviewResult, error) {
	if err := validateRequest(request); err != nil {
		return ReviewResult{}, err
	}
	coverage := domain.Coverage{Key: domain.CoverageKey{FeatureID: request.Feature.FeatureID, Family: request.Family, Category: request.Category}}
	credential, err := c.prerequisiteCredential()
	if err != nil {
		coverage.Status, coverage.Reason = domain.CoverageBlocked, err.Error()
		return ReviewResult{Coverage: coverage}, err
	}

	redactor := secretRedactor{}
	manifest, sources := readFeatureSources(request.Root, request.Feature, c.config.MaxContextBytes, c.config.MaxFileBytes)
	snapshotSources := append([]promptSource(nil), sources...)
	safeManifest := redactManifest(manifest, redactor)
	if len(manifest.Inspected) == 0 {
		coverage.Status, coverage.Reason = domain.CoverageIncomplete, "no readable feature source rows"
		coverage.OmittedPaths = omissionPaths(manifest.Omissions)
		return ReviewResult{Coverage: coverage, Manifest: safeManifest}, errors.New(coverage.Reason)
	}
	if request.PromptVersion != "" && request.PromptVersion != c.config.PromptVersion {
		return ReviewResult{}, errors.New("contextual request prompt version differs from configured provider policy")
	}
	request.PromptVersion = c.config.PromptVersion
	for index := range sources {
		sources[index].Reason = redactor.redact(sources[index].Reason)
		sources[index].Text = redactor.redact(sources[index].Text)
	}
	prompt, err := makePrompt(request, sources, safeManifest, redactor)
	if err != nil {
		coverage.Status, coverage.Reason = domain.CoverageIncomplete, "could not construct contextual provider request"
		coverage.OmittedPaths = omissionPaths(manifest.Omissions)
		return ReviewResult{Coverage: coverage, Manifest: safeManifest}, fmt.Errorf("%s: %w", coverage.Reason, err)
	}
	payload, err := marshalProviderRequest(c.config.Model, effectivePromptVersion(c.config.PromptVersion, request.PromptVersion), prompt)
	if err != nil {
		coverage.Status, coverage.Reason = domain.CoverageIncomplete, "could not encode contextual provider request"
		return ReviewResult{Coverage: coverage, Manifest: safeManifest}, fmt.Errorf("%s: %w", coverage.Reason, err)
	}
	if len(payload) > c.config.MaxPromptBytes {
		manifest.Truncated = true
		manifest.Omissions = append(manifest.Omissions, ManifestOmission{Reason: "contextual request exceeded the configured byte limit"})
		safeManifest = redactManifest(manifest, redactor)
		coverage.Status, coverage.Reason = domain.CoverageIncomplete, "contextual provider request exceeds prompt size limit"
		coverage.OmittedPaths = omissionPaths(manifest.Omissions)
		return ReviewResult{Coverage: coverage, Manifest: safeManifest}, errors.New(coverage.Reason)
	}

	response, err := c.call(ctx, credential, payload)
	if err != nil {
		coverage.Status, coverage.Reason = domain.CoverageIncomplete, "contextual provider request failed"
		coverage.OmittedPaths = omissionPaths(manifest.Omissions)
		return ReviewResult{Coverage: coverage, Manifest: safeManifest}, fmt.Errorf("%s: %w", coverage.Reason, err)
	}
	result, validationErr := validateProviderResponse(request, response, manifest, snapshotSources, c.config.Model, effectivePromptVersion(c.config.PromptVersion, request.PromptVersion))
	result.Manifest = safeManifest
	if validationErr != nil {
		result.Coverage.Status, result.Coverage.Reason = domain.CoverageIncomplete, "contextual provider response failed evidence validation"
		result.Coverage.OmittedPaths = omissionPaths(manifest.Omissions)
		return result, fmt.Errorf("%s: %w", result.Coverage.Reason, validationErr)
	}
	if drifted, driftErr := manifestSourceDrift(request.Root, manifest); driftErr != nil || len(drifted) != 0 {
		result.Coverage.Status, result.Coverage.Reason = domain.CoverageIncomplete, "feature source changed or became unreadable during contextual review"
		result.Coverage.OmittedPaths = appendUniquePaths(omissionPaths(manifest.Omissions), drifted)
		if driftErr != nil {
			return result, fmt.Errorf("%s: %w", result.Coverage.Reason, driftErr)
		}
		return result, errors.New(result.Coverage.Reason)
	}
	if manifest.Truncated || len(manifest.Omissions) != 0 {
		result.Coverage.Status, result.Coverage.Reason = domain.CoverageIncomplete, "feature context was unreadable or truncated"
		result.Coverage.OmittedPaths = omissionPaths(manifest.Omissions)
		return result, errors.New(result.Coverage.Reason)
	}
	return result, nil
}


// CheckPrerequisites reports whether the configured provider may receive source.
// It only validates configuration, explicit egress authorization, and credential
// presence; it never reads source or contacts the provider.
func CheckPrerequisites(config Config) error {
	client, err := NewClient(config)
	if err != nil {
		return err
	}
	_, err = client.prerequisiteCredential()
	return err
}

func (c *Client) prerequisiteCredential() (string, error) {
	if reason := c.unavailableReason(); reason != "" {
		return "", errors.New(reason)
	}
	credential, present := os.LookupEnv(c.config.CredentialEnv)
	if !present || credential == "" {
		return "", errors.New("contextual provider credential environment variable is unavailable")
	}
	return credential, nil
}
func (c *Client) unavailableReason() string {
	if !c.config.EgressAuthorized {
		return "contextual provider source egress is not authorized"
	}
	if c.config.Endpoint == "" || c.config.Model == "" || c.config.PromptVersion == "" || c.config.CredentialEnv == "" {
		return "contextual provider is not fully configured"
	}
	endpoint, err := url.Parse(c.config.Endpoint)
	if err != nil || endpoint.Scheme == "" || endpoint.Host == "" || endpoint.User != nil || (endpoint.Scheme != "https" && endpoint.Scheme != "http") {
		return "contextual provider endpoint is invalid"
	}
	return ""
}

func (c *Client) call(parent context.Context, credential string, payload []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(parent, c.config.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.Endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(c.config.AuthorizationHead, c.config.AuthorizationPref+credential)
	httpClient := c.config.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	response, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 2*1024*1024+1))
	if err != nil {
		return nil, err
	}
	if len(body) > 2*1024*1024 {
		return nil, errors.New("contextual provider response exceeds safe size limit")
	}
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return nil, fmt.Errorf("contextual provider returned HTTP %d", response.StatusCode)
	}
	return body, nil
}

func validateRequest(request ReviewRequest) error {
	if request.Root == "" || request.Snapshot == "" || request.RunID == "" {
		return errors.New("contextual review requires root, source snapshot, and run ID")
	}
	if err := request.Feature.Validate(); err != nil {
		return fmt.Errorf("contextual feature: %w", err)
	}
	if !validCategory(request.Family, request.Category) {
		return errors.New("contextual category is not defined for its family")
	}
	for _, result := range request.ToolResults {
		if err := result.Validate(); err != nil {
			return fmt.Errorf("contextual tool result: %w", err)
		}
	}
	for _, observation := range request.PriorObservations {
		if err := observation.Validate(); err != nil {
			return fmt.Errorf("contextual prior observation: %w", err)
		}
	}
	return nil
}

func effectivePromptVersion(configured, requested string) string {
	if requested != "" {
		return requested
	}
	return configured
}

func omissionPaths(omissions []ManifestOmission) []string {
	paths := make([]string, 0, len(omissions))
	seen := map[string]bool{}
	for _, omission := range omissions {
		if omission.Path != "" && !seen[omission.Path] {
			paths, seen[omission.Path] = append(paths, omission.Path), true
		}
	}
	return paths
}

func appendUniquePaths(paths, additional []string) []string {
	seen := make(map[string]bool, len(paths)+len(additional))
	for _, path := range paths {
		seen[path] = true
	}
	for _, path := range additional {
		if path != "" && !seen[path] {
			paths, seen[path] = append(paths, path), true
		}
	}
	return paths
}

func readFeatureSources(root string, feature domain.Feature, maxContext, maxFile int) (Manifest, []promptSource) {
	refs := collectSourceRefs(feature)
	manifest := Manifest{}
	sources := make([]promptSource, 0, len(refs))
	rootResolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		return Manifest{Omissions: []ManifestOmission{{Reason: "assessment root is unreadable"}}}, nil
	}
	rootResolved, err = filepath.Abs(rootResolved)
	if err != nil {
		return Manifest{Omissions: []ManifestOmission{{Reason: "assessment root is invalid"}}}, nil
	}
	remaining := maxContext
	for _, ref := range refs {
		if remaining == 0 {
			manifest.Truncated = true
			manifest.Omissions = append(manifest.Omissions, ManifestOmission{Path: ref.path, Reason: "context byte limit reached"})
			continue
		}
		if ref.path == ".scout" || strings.HasPrefix(ref.path, ".scout/") {
			manifest.Omissions = append(manifest.Omissions, ManifestOmission{Path: ref.path, Reason: "Scout state is excluded from assessment inputs"})
			continue
		}
		absolute, err := safeSourcePath(rootResolved, ref.path)
		if err != nil {
			manifest.Omissions = append(manifest.Omissions, ManifestOmission{Path: ref.path, Reason: "source path is unsafe or unreadable"})
			continue
		}
		content, err := os.ReadFile(absolute)
		if err != nil {
			manifest.Omissions = append(manifest.Omissions, ManifestOmission{Path: ref.path, Reason: "source file is unreadable"})
			continue
		}
		limit := min(maxFile, remaining)
		excerpt, startLine, endLine, truncated := lineRange(content, limit)
		if len(excerpt) == 0 {
			manifest.Omissions = append(manifest.Omissions, ManifestOmission{Path: ref.path, Reason: "source file is empty"})
			continue
		}
		sourceRange := domain.SourceRange{Path: ref.path, StartLine: startLine, EndLine: endLine, StartByte: 1, EndByte: len(excerpt)}
		digest := sourceDigest(content)
		manifest.Inspected = append(manifest.Inspected, ManifestEntry{Range: sourceRange, Reason: ref.reason, Bytes: len(excerpt), SourceDigest: digest})
		sources = append(sources, promptSource{promptRange: promptRange{Path: ref.path, StartLine: startLine, EndLine: endLine, StartByte: 1, EndByte: len(excerpt)}, Reason: ref.reason, Text: string(excerpt)})
		remaining -= len(excerpt)
		if truncated {
			manifest.Truncated = true
			manifest.Omissions = append(manifest.Omissions, ManifestOmission{Path: ref.path, Reason: "source file exceeded context limit"})
		}
	}
	return manifest, sources
}

func sourceDigest(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func manifestSourceDrift(root string, manifest Manifest) ([]string, error) {
	rootResolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, fmt.Errorf("assessment root is unreadable: %w", err)
	}
	rootResolved, err = filepath.Abs(rootResolved)
	if err != nil {
		return nil, fmt.Errorf("assessment root is invalid: %w", err)
	}
	var drifted []string
	for _, entry := range manifest.Inspected {
		absolute, err := safeSourcePath(rootResolved, entry.Range.Path)
		if err != nil {
			drifted = append(drifted, entry.Range.Path)
			continue
		}
		content, err := os.ReadFile(absolute)
		if err != nil || sourceDigest(content) != entry.SourceDigest {
			drifted = append(drifted, entry.Range.Path)
		}
	}
	return drifted, nil
}

type sourceRef struct{ path, reason string }

func collectSourceRefs(feature domain.Feature) []sourceRef {
	refs := make([]sourceRef, 0, len(feature.OwnedFiles)+len(feature.ContextFiles)+len(feature.Tests))
	seen := map[string]bool{}
	appendRef := func(path, reason string) {
		if !seen[path] {
			seen[path] = true
			refs = append(refs, sourceRef{path: path, reason: reason})
		}
	}
	for _, ref := range feature.OwnedFiles {
		appendRef(ref.Path, "owned: "+ref.Reason)
	}
	for _, ref := range feature.ContextFiles {
		appendRef(ref.Path, "context: "+ref.Reason)
	}
	for _, test := range feature.Tests {
		appendRef(test.Path, "feature test")
	}
	return refs
}

func safeSourcePath(root, scoped string) (string, error) {
	candidate := filepath.Join(root, filepath.FromSlash(scoped))
	resolved, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(root, resolved)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", errors.New("source path escapes root")
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.Mode().IsRegular() {
		return "", errors.New("source path is not a regular file")
	}
	return resolved, nil
}

// Package web exposes a local, read-only view of persisted Scout state.
package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"scout/internal/clawpatch"
	"scout/internal/domain"
	"scout/internal/state"
)

var pageTemplate = template.Must(template.New("page").Parse(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>{{.Title}}</title></head>
<body><main><h1>{{.Title}}</h1>{{if .Links}}<nav><ul>{{range .Links}}<li><a href="{{.Path}}">{{.Label}}</a></li>{{end}}</ul></nav>{{end}}{{if .Body}}<pre wrap>{{.Body}}</pre>{{end}}</main></body></html>`))

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?is)-----BEGIN [A-Z0-9 ]*PRIVATE KEY-----.*?-----END [A-Z0-9 ]*PRIVATE KEY-----`),
	regexp.MustCompile(`(?i)\b(gh[pousr]_[A-Za-z0-9_]{20,}|github_pat_[A-Za-z0-9_]{20,}|AKIA[0-9A-Z]{16}|AIza[0-9A-Za-z_-]{20,})\b`),
	regexp.MustCompile(`(?i)((?:api[_-]?key|secret|token|password|passwd|authorization)\s*[:=]\s*["']?)([^\s"'\x60,;}{]{6,})`),
	regexp.MustCompile(`(?i)(bearer\s+)[A-Za-z0-9._~+/-]{12,}`),
}

// New returns a handler backed by the supplied store. It only reads persisted
// Scout state and never serves target files or artifact contents.
func New(store *state.Store) http.Handler {
	return handler{store: store}
}

type handler struct {
	store *state.Store
}

type link struct {
	Path  string
	Label string
}

type page struct {
	Title string
	Links []link
	Body  string
}

type runDetails struct {
	Run          domain.Run           `json:"run"`
	Observations []domain.Observation `json:"observations"`
}

type issueEvidence struct {
	Issue        domain.Issue         `json:"issue"`
	Run          domain.Run           `json:"run"`
	Observations []domain.Observation `json:"observations"`
}

type dashboard struct {
	Issues    []domain.Issue `json:"issues"`
	LatestRun *domain.Run    `json:"latestRun,omitempty"`
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if h.store == nil {
		h.writeError(w, http.StatusServiceUnavailable, "state store is unavailable")
		return
	}

	switch {
	case r.URL.Path == "/":
		h.dashboard(w, r)
	case r.URL.Path == "/issues":
		h.issueList(w, r, false)
	case r.URL.Path == "/features":
		h.features(w, r, false)
	case r.URL.Path == "/runs/latest":
		h.runLatest(w, r, false)
	case issueEvidencePath(r.URL.Path, "/issues/"):
		h.issueEvidence(w, r, strings.TrimPrefix(r.URL.Path, "/issues/"), false)
	case strings.HasPrefix(r.URL.Path, "/issues/"):
		h.issue(w, r, strings.TrimPrefix(r.URL.Path, "/issues/"), false)
	case strings.HasPrefix(r.URL.Path, "/runs/"):
		h.run(w, r, strings.TrimPrefix(r.URL.Path, "/runs/"), false)
	case r.URL.Path == "/api/issues":
		h.issueList(w, r, true)
	case r.URL.Path == "/api/features":
		h.features(w, r, true)
	case r.URL.Path == "/api/runs/latest":
		h.runLatest(w, r, true)
	case issueEvidencePath(r.URL.Path, "/api/issues/"):
		h.issueEvidence(w, r, strings.TrimPrefix(r.URL.Path, "/api/issues/"), true)
	case strings.HasPrefix(r.URL.Path, "/api/issues/"):
		h.issue(w, r, strings.TrimPrefix(r.URL.Path, "/api/issues/"), true)
	case strings.HasPrefix(r.URL.Path, "/api/runs/"):
		h.run(w, r, strings.TrimPrefix(r.URL.Path, "/api/runs/"), true)
	default:
		h.writeError(w, http.StatusNotFound, "route not found")
	}
}

func (h handler) dashboard(w http.ResponseWriter, r *http.Request) {
	issues, err := h.store.ListIssues(r.Context())
	if err != nil {
		h.writeStoreError(w, err)
		return
	}
	view := dashboard{Issues: issues}
	latest, err := h.store.LatestRecordedRun(r.Context())
	if err == nil {
		view.LatestRun = &latest
	} else if !errors.Is(err, state.ErrNotFound) {
		h.writeStoreError(w, err)
		return
	}
	h.writeHTML(w, "Scout state", view, []link{
		{Path: "/issues", Label: "Issues"},
		{Path: "/runs/latest", Label: "Latest run"},
		{Path: "/features", Label: "Canonical feature mappings"},
		{Path: "/api/issues", Label: "Issues JSON"},
		{Path: "/api/runs/latest", Label: "Latest run JSON"},
		{Path: "/api/features", Label: "Features JSON"},
	})
}

func (h handler) issueList(w http.ResponseWriter, r *http.Request, asJSON bool) {
	issues, err := h.store.ListIssues(r.Context())
	if err != nil {
		h.writeStoreError(w, err)
		return
	}
	if asJSON {
		h.writeJSON(w, http.StatusOK, struct {
			Issues []domain.Issue `json:"issues"`
		}{Issues: issues})
		return
	}
	h.writeHTML(w, "Issues", struct {
		Issues []domain.Issue `json:"issues"`
	}{Issues: issues}, nil)
}

func (h handler) issue(w http.ResponseWriter, r *http.Request, issueID string, asJSON bool) {
	if !validRouteID(issueID) {
		h.writeError(w, http.StatusBadRequest, "invalid issue ID")
		return
	}
	issue, err := h.store.LoadIssue(r.Context(), issueID)
	if err != nil {
		h.writeStoreError(w, err)
		return
	}
	if asJSON {
		h.writeJSON(w, http.StatusOK, struct {
			Issue domain.Issue `json:"issue"`
		}{Issue: issue})
		return
	}
	h.writeHTML(w, "Issue "+issueID, struct {
		Issue domain.Issue `json:"issue"`
	}{Issue: issue}, nil)
}

func (h handler) runLatest(w http.ResponseWriter, r *http.Request, asJSON bool) {
	run, err := h.store.LatestRecordedRun(r.Context())
	if err != nil {
		h.writeStoreError(w, err)
		return
	}
	h.writeRun(w, r.Context(), run, asJSON)
}

func (h handler) issueEvidence(w http.ResponseWriter, r *http.Request, suffix string, asJSON bool) {
	parts := strings.Split(suffix, "/runs/")
	if len(parts) != 2 || !validRouteID(parts[0]) || !validRouteID(parts[1]) {
		h.writeError(w, http.StatusBadRequest, "invalid issue or run ID")
		return
	}
	issue, err := h.store.LoadIssue(r.Context(), parts[0])
	if err != nil {
		h.writeStoreError(w, err)
		return
	}
	run, err := h.store.LoadRun(r.Context(), parts[1])
	if err != nil {
		h.writeStoreError(w, err)
		return
	}
	observations, err := h.store.LoadObservations(r.Context(), run.RunID)
	if err != nil {
		h.writeStoreError(w, err)
		return
	}
	linked := make(map[string]struct{}, len(issue.ObservationIDs))
	for _, observationID := range issue.ObservationIDs {
		linked[observationID] = struct{}{}
	}
	evidence := make([]domain.Observation, 0)
	for _, observation := range observations {
		if _, found := linked[observation.ObservationID]; found {
			evidence = append(evidence, observation)
		}
	}
	view := issueEvidence{Issue: issue, Run: run, Observations: evidence}
	if asJSON {
		h.writeJSON(w, http.StatusOK, view)
		return
	}
	h.writeHTML(w, "Issue evidence "+issue.IssueID, view, nil)
}

func (h handler) run(w http.ResponseWriter, r *http.Request, runID string, asJSON bool) {
	if !validRouteID(runID) {
		h.writeError(w, http.StatusBadRequest, "invalid run ID")
		return
	}
	run, err := h.store.LoadRun(r.Context(), runID)
	if err != nil {
		h.writeStoreError(w, err)
		return
	}
	h.writeRun(w, r.Context(), run, asJSON)
}

func (h handler) writeRun(w http.ResponseWriter, ctx context.Context, run domain.Run, asJSON bool) {
	observations, err := h.store.LoadObservations(ctx, run.RunID)
	if err != nil {
		h.writeStoreError(w, err)
		return
	}
	view := runDetails{Run: run, Observations: observations}
	if asJSON {
		h.writeJSON(w, http.StatusOK, view)
		return
	}
	h.writeHTML(w, "Run "+run.RunID, view, nil)
}

func (h handler) features(w http.ResponseWriter, r *http.Request, asJSON bool) {
	features, err := clawpatch.LoadFeatures(filepath.Join(h.store.Root(), ".scout", "clawpatch"))
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "could not load canonical feature mappings")
		return
	}
	for index := range features {
		features[index].RawOriginal = nil
	}
	if asJSON {
		h.writeJSON(w, http.StatusOK, struct {
			Features []domain.Feature `json:"features"`
		}{Features: features})
		return
	}
	h.writeHTML(w, "Canonical feature mappings", struct {
		Features []domain.Feature `json:"features"`
	}{Features: features}, nil)
}

func (h handler) writeJSON(w http.ResponseWriter, status int, value any) {
	safe, err := sanitize(value)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "could not render state")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(safe); err != nil {
		return
	}
}

func (h handler) writeHTML(w http.ResponseWriter, title string, value any, links []link) {
	safe, err := sanitize(value)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "could not render state")
		return
	}
	body, err := json.MarshalIndent(safe, "", "  ")
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "could not render state")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := pageTemplate.Execute(w, page{Title: title, Links: links, Body: string(body)}); err != nil {
		return
	}
}

func (h handler) writeStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, state.ErrNotFound) {
		h.writeError(w, http.StatusNotFound, "state record not found")
		return
	}
	h.writeError(w, http.StatusInternalServerError, "could not read state")
}

func (h handler) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = fmt.Fprintf(w, `{"error":%q}`+"\n", message)
}

func setSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'")
	w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
}

func issueEvidencePath(path, prefix string) bool {
	return strings.HasPrefix(path, prefix) && strings.Contains(strings.TrimPrefix(path, prefix), "/runs/")
}

func validRouteID(value string) bool {
	if value == "" || len(value) > 256 || !utf8.ValidString(value) {
		return false
	}
	for _, r := range value {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.' || r == ':') {
			return false
		}
	}
	return true
}

func sanitize(value any) (any, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var decoded any
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	if err := decoder.Decode(&decoded); err != nil {
		return nil, err
	}
	return sanitizeJSON(decoded), nil
}

func sanitizeJSON(value any) any {
	switch value := value.(type) {
	case string:
		return redact(value)
	case []any:
		for index := range value {
			value[index] = sanitizeJSON(value[index])
		}
	case map[string]any:
		for key, child := range value {
			value[key] = sanitizeJSON(child)
		}
	}
	return value
}

func redact(value string) string {
	for _, pattern := range secretPatterns {
		value = pattern.ReplaceAllStringFunc(value, func(match string) string {
			lower := strings.ToLower(match)
			switch {
			case strings.HasPrefix(lower, "bearer "):
				return match[:7] + "[REDACTED]"
			case strings.Contains(lower, "private key"):
				return "[REDACTED]"
			case strings.Contains(match, "="):
				return match[:strings.Index(match, "=")+1] + "[REDACTED]"
			case strings.Contains(match, ":"):
				return match[:strings.Index(match, ":")+1] + "[REDACTED]"
			default:
				return "[REDACTED]"
			}
		})
	}
	return value
}

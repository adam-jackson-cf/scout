package main

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"scout/internal/domain"
	"scout/internal/state"
)

const p22Check = "security.ci-mutable-action-ref"

func p22Day(n int) time.Time {
	return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(n) * 24 * time.Hour)
}

// This is a real persisted publication, not a command/evaluator substitute.
// A second exact cohort proves explicit historical-version selection independently
// of the default pin. The old and current versions have distinct observations.
func p22Store(t *testing.T, ready bool) (string, domain.ShadowLedger) {
	t.Helper()
	root := t.TempDir()
	s, err := state.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	stamp := p22Day(30)
	p := domain.ShadowPublication{ID: "sealed", RepositoryID: "42", RecordedAt: stamp}
	for v := 1; v <= 2; v++ {
		version := fmt.Sprintf("rule-%d", v)
		minDays := 14
		if !ready {
			minDays = 31
		}
		p.Frontiers = append(p.Frontiers, domain.ShadowFrontier{RepositoryID: "42", CheckID: p22Check, RuleVersion: version, CompleteThrough: stamp, PolicyRevision: "protected-policy", MinDays: minDays, MinObservations: 5})
		p.Intervals = append(p.Intervals, domain.ShadowInterval{ID: version, RepositoryID: "42", CheckID: p22Check, RuleVersion: version, After: p22Day(-1), Through: stamp})
		for i := 1; i <= 5; i++ {
			start := p22Day(v*5 + i)
			id := fmt.Sprintf("%d", v*10+i)
			p.Facts = append(p.Facts, domain.PRRunFact{ID: domain.AttemptID{RepositoryID: "42", RunID: id, JobID: "1", Attempt: "1"}, Scope: domain.ShadowScopeCI, PRNumber: id, HeadCommit: "head-" + id, HeadTree: "tree-" + id, BaseCommit: "base", MergeBase: "merge-base", AssessedRevision: "merge-" + id, CheckID: p22Check, RuleVersion: version, CreatedAt: start, StartedAt: start, FinishedAt: start.Add(time.Hour), Status: "assessed", Comparison: "comparable", Inspection: "inspected", Pins: domain.ShadowPins{Scout: "scout", Tools: "tools", Catalogue: "catalogue", Policy: "policy", Recipe: "guardian"}, RecordedAt: &stamp})
		}
	}
	if err := s.ApplyShadowPublication(t.Context(), p); err != nil {
		t.Fatal(err)
	}
	ledger, err := s.ReadShadowLedger(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	return root, ledger
}

type p22File struct {
	Mode     fs.FileMode
	Size     int64
	Modified time.Time
	Digest   string
}

func p22Snapshot(t *testing.T, root string) map[string]p22File {
	t.Helper()
	files := map[string]p22File{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		name, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		f := p22File{Mode: info.Mode(), Size: info.Size(), Modified: info.ModTime()}
		if info.Mode().IsRegular() {
			raw, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			digest := sha256.Sum256(raw)
			f.Digest = hex.EncodeToString(digest[:])
		}
		files[name] = f
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

// Calls the real dispatcher. Capture both channels without pipe-size deadlocks.
func p22Run(t *testing.T, root string, extra ...string) (int, string, string) {
	t.Helper()
	out, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	errout, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatal(err)
	}
	defer errout.Close()
	oldOut, oldErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = out, errout
	defer func() { os.Stdout, os.Stderr = oldOut, oldErr }()
	args := append([]string{"readiness", "--root", root}, extra...)
	code := run(boundedTestContext(t), args)
	if _, err := out.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	if _, err := errout.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	stdout, err := io.ReadAll(out)
	if err != nil {
		t.Fatal(err)
	}
	stderr, err := io.ReadAll(errout)
	if err != nil {
		t.Fatal(err)
	}
	return code, string(stdout), string(stderr)
}

type p22Transport struct{ calls *atomic.Int64 }

func (s p22Transport) RoundTrip(*http.Request) (*http.Response, error) {
	s.calls.Add(1)
	return nil, fmt.Errorf("P22 forbids network access")
}

func p22Hostile(t *testing.T) {
	t.Helper()
	bin := t.TempDir()
	log := filepath.Join(t.TempDir(), "tool-invocations")
	for _, name := range []string{"git", "node", "npm", "npx", "clawpatch", "gitleaks", "semgrep", "claude", "codex", "curl", "wget"} {
		if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\nprintf '%s\\n' \"$0\" >> \"$P22_TOOL_LOG\"\nexit 97\n"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin)
	t.Setenv("P22_TOOL_LOG", log)
	t.Setenv("SCOUT_MODEL_CLI", filepath.Join(bin, "claude"))
	t.Setenv("SCOUT_CLAWPATCH_BINARY", filepath.Join(bin, "clawpatch"))
	var calls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(server.Close)
	for _, name := range []string{"HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "http_proxy", "https_proxy", "all_proxy", "OPENAI_BASE_URL", "ANTHROPIC_BASE_URL", "OPENROUTER_BASE_URL", "GITHUB_API_URL"} {
		t.Setenv(name, server.URL)
	}
	t.Setenv("NO_PROXY", "")
	previous := http.DefaultTransport
	http.DefaultTransport = p22Transport{calls: &calls}
	t.Cleanup(func() {
		http.DefaultTransport = previous
		if n := calls.Load(); n != 0 {
			t.Errorf("shadow readiness made %d network calls", n)
		}
		if raw, err := os.ReadFile(log); !os.IsNotExist(err) {
			t.Errorf("shadow readiness invoked a tool/model: %q (read error %v)", raw, err)
		}
	})
}

type p22Reply struct {
	Status   string `json:"status"`
	ExitCode int    `json:"exitCode"`
	Error    string `json:"error"`
	Result   struct {
		Mode    string                         `json:"queryMode"`
		Dossier json.RawMessage                `json:"dossier"`
		Digest  string                         `json:"dossierDigest"`
		Live    struct{ Certification string } `json:"live"`
	} `json:"result"`
}

// Decode only the caller-visible fields asserted here; no test-local evaluator
// or canonical encoder stands in for production.
func p22Dossier(t *testing.T, raw json.RawMessage) (string, bool) {
	t.Helper()
	var d struct {
		EncodingVersion, RepositoryID, CheckID, RuleVersion string
		AsOf                                                time.Time
		Ready                                               bool
		Reason                                              string
		Stats                                               struct{ PRRuns, TotalRuns int }
	}
	if err := json.Unmarshal(raw, &d); err != nil {
		t.Fatal(err)
	}
	if d.EncodingVersion != "scout.shadow/v1" || d.RepositoryID != "42" || d.CheckID != p22Check || !d.AsOf.Equal(p22Day(30)) || d.Stats.PRRuns != 5 || d.Stats.TotalRuns != 5 || d.Reason == "" {
		t.Errorf("CLI did not resolve persisted dossier fields: %+v", d)
	}
	return d.RuleVersion, d.Ready
}

func p22JSON(t *testing.T, raw string, code int) p22Reply {
	t.Helper()
	var reply p22Reply
	decoder := json.NewDecoder(strings.NewReader(raw))
	if err := decoder.Decode(&reply); err != nil {
		t.Fatalf("shadow CLI must emit a JSON outcome (exit %d): %v; stdout=%q", code, err, raw)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		t.Fatalf("extra JSON/output after shadow outcome: %q", raw)
	}
	if reply.ExitCode != code {
		t.Fatalf("JSON exitCode=%d, process exit=%d", reply.ExitCode, code)
	}
	return reply
}

// P22 owns CLI parsing, output, exit mapping and the read-only boundary. It does
// not duplicate S4 aggregation/visibility proofs or S3 provider authentication.
func TestADR012ShadowCLIReadOnlyAndExits(t *testing.T) {
	t.Run("no-store-no-initialization-tools-models-or-network", func(t *testing.T) {
		root := t.TempDir()
		p22Hostile(t)
		before := p22Snapshot(t, root)
		code, stdout, stderr := p22Run(t, root, "--shadow", "--check", p22Check, "--json")
		if after := p22Snapshot(t, root); !reflect.DeepEqual(before, after) {
			t.Errorf("missing-store query initialized or mutated root: before=%v after=%v", before, after)
		}
		if code != 2 {
			t.Fatalf("readiness --shadow without a store exit = %d, want 2; stdout=%q stderr=%q", code, stdout, stderr)
		}
		reply := p22JSON(t, stdout, code)
		if reply.Status != "unmeasured" {
			t.Errorf("missing-store status=%q, want unmeasured", reply.Status)
		}
	})

	for _, ready := range []bool{true, false} {
		t.Run(fmt.Sprintf("computed-ready-%t-read-only-while-writer-locked", ready), func(t *testing.T) {
			root, _ := p22Store(t, ready)
			// A live writer owns the exclusive lock. Open(), unlike
			// OpenExistingReadOnly(), cannot serve this command successfully.
			writer, err := state.Open(root)
			if err != nil {
				t.Fatal(err)
			}
			defer writer.Close()
			if err := os.WriteFile(filepath.Join(root, ".scout", "policy.yaml"), []byte("invalid: ["), 0600); err != nil {
				t.Fatal(err)
			}
			p22Hostile(t)
			before := p22Snapshot(t, root)
			code, stdout, stderr := p22Run(t, root, "--shadow", "--check", p22Check, "--json")
			if code != 0 {
				t.Fatalf("computed shadow exit=%d, want 0; stdout=%q stderr=%q", code, stdout, stderr)
			}
			reply := p22JSON(t, stdout, code)
			version, gotReady := p22Dossier(t, reply.Result.Dossier)
			if version != "rule-2" || gotReady != ready {
				t.Errorf("default version/readiness=%s/%t, want rule-2/%t", version, gotReady, ready)
			}
			sum := sha256.Sum256(reply.Result.Dossier)
			digest := "sha256:" + hex.EncodeToString(sum[:])
			if reply.Result.Digest != digest {
				t.Errorf("dossierDigest=%q, want SHA-256 of exact canonical dossier bytes %q", reply.Result.Digest, digest)
			}
			if reply.Result.Mode != "current_window" || reply.Result.Live.Certification != "current" {
				t.Errorf("default is not current certification: %+v", reply.Result)
			}
			cutoff := p22Day(30).Format(time.RFC3339)
			code, raw, stderr := p22Run(t, root, "--shadow", "--check", p22Check, "--as-of", cutoff, "--json")
			if code != 0 {
				t.Fatalf("explicit C exit=%d, want 0: %s %s", code, raw, stderr)
			}
			historical := p22JSON(t, raw, code)
			if historical.Result.Mode != "historical_reproduction" || historical.Result.Live.Certification != "historical" || historical.Result.Digest != digest {
				t.Errorf("explicit C must reproduce digest but never activation certification: %+v", historical.Result)
			}
			if !bytes.Equal(historical.Result.Dossier, reply.Result.Dossier) {
				t.Error("explicit C changed canonical dossier bytes")
			}
			code, text, stderr := p22Run(t, root, "--shadow", "--check", p22Check)
			if code != 0 {
				t.Fatalf("text shadow exit=%d: %s %s", code, text, stderr)
			}
			for _, token := range []string{p22Check, "rule-2", cutoff, digest, "current_window", "current"} {
				if !strings.Contains(text, token) {
					t.Errorf("text output missing %q: %s", token, text)
				}
			}
			if after := p22Snapshot(t, root); !reflect.DeepEqual(before, after) {
				t.Errorf("query modified store/policy while writer held lock: before=%v after=%v", before, after)
			}
		})
	}

	t.Run("pending-late-default-refused-explicit-C-historical", func(t *testing.T) {
		root, ledger := p22Store(t, true)
		s, err := state.Open(root)
		if err != nil {
			t.Fatal(err)
		}
		fact := ledger.Facts[len(ledger.Facts)-1]
		fact.ID.RunID, fact.PRNumber = "999", "999"
		fact.HeadCommit, fact.HeadTree = "late-head", "late-tree"
		fact.CreatedAt, fact.StartedAt, fact.FinishedAt = p22Day(20), p22Day(20), p22Day(21)
		fact.RecordedAt = nil
		if err := s.RecordPRRun(t.Context(), fact, nil, nil); err != nil {
			t.Fatal(err)
		}
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
		p22Hostile(t)
		code, raw, stderr := p22Run(t, root, "--shadow", "--check", p22Check, "--json")
		if code != 2 {
			t.Fatalf("default pending-late certification exit=%d, want 2: %s %s", code, raw, stderr)
		}
		current := p22JSON(t, raw, code)
		if current.Result.Mode != "current_window" || current.Result.Live.Certification != "refused" {
			t.Errorf("default must refuse current certification: %+v", current.Result)
		}
		code, raw, stderr = p22Run(t, root, "--shadow", "--check", p22Check, "--as-of", p22Day(30).Format(time.RFC3339), "--json")
		if code != 0 {
			t.Fatalf("explicit C historical exit=%d, want 0: %s %s", code, raw, stderr)
		}
		historical := p22JSON(t, raw, code)
		if historical.Result.Mode != "historical_reproduction" || historical.Result.Live.Certification != "historical" || historical.Result.Digest != current.Result.Digest || !bytes.Equal(historical.Result.Dossier, current.Result.Dossier) {
			t.Errorf("historical reproduction must retain refused current dossier without becoming activation evidence: current=%+v historical=%+v", current.Result, historical.Result)
		}
	})

	t.Run("existing-store-without-frontier-is-unmeasured", func(t *testing.T) {
		root := t.TempDir()
		s, err := state.Open(root)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
		p22Hostile(t)
		code, raw, stderr := p22Run(t, root, "--shadow", "--check", p22Check, "--json")
		if code != 2 {
			t.Fatalf("no complete frontier exit=%d, want 2: %s %s", code, raw, stderr)
		}
		p22JSON(t, raw, code)
	})

	t.Run("cutoffs-versions-and-exact-flags", func(t *testing.T) {
		root, _ := p22Store(t, true)
		p22Hostile(t)
		for _, tc := range []struct {
			name    string
			args    []string
			code    int
			version string
		}{
			{"known-historical-version", []string{"--version", "rule-1", "--as-of", p22Day(30).Format(time.RFC3339)}, 0, "rule-1"},
			{"unknown-exact-version", []string{"--version", "unknown"}, 3, ""},
			{"invalid-time", []string{"--as-of", "not-a-time"}, 3, ""},
			{"after-C-before-reader-now", []string{"--as-of", p22Day(31).Format(time.RFC3339)}, 2, ""},
			{"after-both-C-and-reader-now", []string{"--as-of", "9999-01-01T00:00:00Z"}, 3, ""},
			{"check-required", nil, 3, ""},
			{"unknown-flag", []string{"--not-a-shadow-flag"}, 3, ""},
			{"supported-encoding", []string{"--encoding", "scout.shadow/v1", "--version", "rule-1", "--as-of", p22Day(30).Format(time.RFC3339)}, 0, "rule-1"},
			{"unsupported-encoding", []string{"--encoding", "scout.shadow/v999"}, 3, ""},
		} {
			t.Run(tc.name, func(t *testing.T) {
				args := []string{"--shadow", "--json"}
				if tc.name != "check-required" {
					args = append(args, "--check", p22Check)
				}
				args = append(args, tc.args...)
				code, raw, stderr := p22Run(t, root, args...)
				if code != tc.code {
					t.Fatalf("exit=%d, want %d; stdout=%q stderr=%q", code, tc.code, raw, stderr)
				}
				reply := p22JSON(t, raw, code)
				if tc.version != "" {
					version, _ := p22Dossier(t, reply.Result.Dossier)
					if version != tc.version || reply.Result.Live.Certification != "historical" {
						t.Errorf("explicit historical cohort ignored: %+v", reply.Result)
					}
				}
			})
		}
	})

	for _, version := range []int{13, 999} {
		t.Run(fmt.Sprintf("unsupported-schema-%d", version), func(t *testing.T) {
			root := t.TempDir()
			if err := os.Mkdir(filepath.Join(root, ".scout"), 0700); err != nil {
				t.Fatal(err)
			}
			// Version preflight must refuse before reading tables or migrating.
			db, err := sql.Open("sqlite", filepath.Join(root, ".scout", "state.sqlite"))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(fmt.Sprintf("PRAGMA user_version=%d", version)); err != nil {
				t.Fatal(err)
			}
			if err := db.Close(); err != nil {
				t.Fatal(err)
			}
			p22Hostile(t)
			before := p22Snapshot(t, root)
			code, raw, stderr := p22Run(t, root, "--shadow", "--check", p22Check, "--json")
			if code != 3 {
				t.Fatalf("schema %d exit=%d, want 3: %s %s", version, code, raw, stderr)
			}
			reply := p22JSON(t, raw, code)
			if version == 13 && (!strings.Contains(reply.Error, "writable upgrade") || !strings.Contains(reply.Error, "preserve unpublished history") || strings.Contains(reply.Error, "re-baseline")) {
				t.Errorf("v13 must name preservation-safe writable upgrade: %q", reply.Error)
			}
			if after := p22Snapshot(t, root); !reflect.DeepEqual(before, after) {
				t.Errorf("schema refusal mutated store: before=%v after=%v", before, after)
			}
		})
	}

	t.Run("invalid-storage", func(t *testing.T) {
		root := t.TempDir()
		if err := os.Mkdir(filepath.Join(root, ".scout"), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, ".scout", "state.sqlite"), []byte("not a SQLite store"), 0600); err != nil {
			t.Fatal(err)
		}
		p22Hostile(t)
		before := p22Snapshot(t, root)
		code, raw, stderr := p22Run(t, root, "--shadow", "--check", p22Check, "--json")
		if code != 3 {
			t.Fatalf("invalid storage exit=%d, want 3: %s %s", code, raw, stderr)
		}
		p22JSON(t, raw, code)
		if after := p22Snapshot(t, root); !reflect.DeepEqual(before, after) {
			t.Error("invalid-storage refusal mutated root")
		}
	})

	t.Run("readiness-matrix-remains-compatible", func(t *testing.T) {
		// Existing matrix owner exercises the real matrix path and REQ-18 rows.
		// Do not put it under the shadow-only hostile PATH: git is legitimate here.
		TestReadinessMatrixPrintsSection53Rows(t)
	})
}

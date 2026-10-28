package state

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"scout/internal/domain"
)

// v1StoreDDL is the schema of the Scout v1 implementation's last store
// (user_version 3), written out as that implementation created it (initial
// import 41ea1eb): the v1 tables, the v2 inventory_snapshots table and the v3
// ALTER, not the current schema and not a current store rewound.
var v1StoreDDL = []string{
	`CREATE TABLE metadata (key TEXT PRIMARY KEY, value TEXT NOT NULL)`,
	`CREATE TABLE plans (plan_id TEXT PRIMARY KEY, run_id TEXT NOT NULL UNIQUE, repository_id TEXT NOT NULL, created_at INTEGER NOT NULL, payload BLOB NOT NULL)`,
	`CREATE TABLE runs (run_id TEXT PRIMARY KEY, plan_id TEXT NOT NULL UNIQUE, repository_id TEXT NOT NULL, state TEXT NOT NULL, started_at INTEGER NOT NULL, finished_at INTEGER, payload BLOB NOT NULL, FOREIGN KEY(plan_id) REFERENCES plans(plan_id))`,
	`CREATE TABLE coverage (run_id TEXT NOT NULL, coverage_key TEXT NOT NULL, status TEXT NOT NULL, payload BLOB NOT NULL, PRIMARY KEY(run_id, coverage_key), FOREIGN KEY(run_id) REFERENCES runs(run_id) ON DELETE CASCADE)`,
	`CREATE TABLE observations (observation_id TEXT PRIMARY KEY, run_id TEXT NOT NULL, observed_at INTEGER NOT NULL, payload BLOB NOT NULL, FOREIGN KEY(run_id) REFERENCES runs(run_id) ON DELETE RESTRICT)`,
	`CREATE TABLE artifacts (digest TEXT PRIMARY KEY, artifact_id TEXT NOT NULL UNIQUE, content_type TEXT NOT NULL, size_bytes INTEGER NOT NULL, created_at INTEGER NOT NULL)`,
	`CREATE TABLE issues (issue_id TEXT PRIMARY KEY, state TEXT NOT NULL, updated_at INTEGER NOT NULL, payload BLOB NOT NULL)`,
	`CREATE TABLE issue_observations (issue_id TEXT NOT NULL, observation_id TEXT NOT NULL, PRIMARY KEY(issue_id, observation_id), FOREIGN KEY(issue_id) REFERENCES issues(issue_id) ON DELETE CASCADE, FOREIGN KEY(observation_id) REFERENCES observations(observation_id) ON DELETE RESTRICT)`,
	`CREATE TABLE issue_history (issue_id TEXT NOT NULL, ordinal INTEGER NOT NULL, payload BLOB NOT NULL, PRIMARY KEY(issue_id, ordinal), FOREIGN KEY(issue_id) REFERENCES issues(issue_id) ON DELETE CASCADE)`,
	`CREATE INDEX coverage_success_lookup ON coverage(coverage_key, status, run_id)`,
	`CREATE INDEX runs_success_lookup ON runs(state, finished_at DESC)`,
	`CREATE INDEX observations_by_run ON observations(run_id, observed_at)`,
	`CREATE INDEX issues_by_state ON issues(state, updated_at DESC)`,
	`CREATE TABLE inventory_snapshots (run_id TEXT PRIMARY KEY, payload BLOB NOT NULL, FOREIGN KEY(run_id) REFERENCES runs(run_id))`,
	`ALTER TABLE runs ADD COLUMN finalized INTEGER NOT NULL DEFAULT 1`,
	`CREATE TABLE run_job_commits (run_id TEXT NOT NULL, job_id TEXT NOT NULL, committed_at INTEGER NOT NULL, payload BLOB NOT NULL, PRIMARY KEY(run_id, job_id), FOREIGN KEY(run_id) REFERENCES runs(run_id) ON DELETE CASCADE)`,
	`CREATE INDEX run_job_commits_by_run ON run_job_commits(run_id, committed_at, job_id)`,
}

// v1Store is a state.sqlite built from v1StoreDDL at user_version 3.
type v1Store struct {
	root string
	db   *sql.DB
}

func newV1Store(t *testing.T) *v1Store {
	t.Helper()
	root := t.TempDir()
	db := openFixtureDB(t, root, "wal")
	for _, statement := range v1StoreDDL {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("v1 DDL %q: %v", statement, err)
		}
	}
	if _, err := db.Exec(`PRAGMA user_version=3`); err != nil {
		t.Fatal(err)
	}
	return &v1Store{root: root, db: db}
}

func (v *v1Store) close(t *testing.T) {
	t.Helper()
	if err := v.db.Close(); err != nil {
		t.Fatal(err)
	}
}

// insertRun writes a plan, its run and observations as the v1 implementation
// stored them: plain rows, no plan_jobs.
func (v *v1Store) insertRun(t *testing.T, plan domain.Plan, run domain.Run, observations []domain.Observation) {
	t.Helper()
	planPayload, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.db.Exec(`INSERT INTO plans(plan_id, run_id, repository_id, created_at, payload) VALUES (?, ?, ?, ?, ?)`, plan.PlanID, plan.RunID, plan.Snapshot.RepositoryID, unixMillis(plan.CreatedAt), planPayload); err != nil {
		t.Fatal(err)
	}
	runPayload, err := json.Marshal(run)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.db.Exec(`INSERT INTO runs(run_id, plan_id, repository_id, state, started_at, finished_at, payload) VALUES (?, ?, ?, ?, ?, NULL, ?)`, run.RunID, run.PlanID, run.Snapshot.RepositoryID, run.State, unixMillis(run.StartedAt), runPayload); err != nil {
		t.Fatal(err)
	}
	for _, observation := range observations {
		payload, err := json.Marshal(observation)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := v.db.Exec(`INSERT INTO observations(observation_id, run_id, observed_at, payload) VALUES (?, ?, ?, ?)`, observation.ObservationID, run.RunID, unixMillis(observation.ObservedAt), payload); err != nil {
			t.Fatal(err)
		}
	}
}

// openFixtureDB creates root/.scout/state.sqlite in the given journal mode.
func openFixtureDB(t *testing.T, root, journalMode string) *sql.DB {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, ".scout"), 0o700); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", filepath.Join(root, ".scout", "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`PRAGMA journal_mode=` + journalMode); err != nil {
		t.Fatal(err)
	}
	return db
}

// describeSchema renders the schema facts a re-anchored store must share with a
// new one: sqlite_schema objects by type and name; per table its columns (name,
// type, not-null, default, primary key, in order), indexes (name, uniqueness,
// origin, partial) with their columns (sort order, collation) and foreign keys.
// Stored CREATE text is deliberately not compared (v3 used ALTER TABLE).
func describeSchema(t *testing.T, db *sql.DB) []string {
	t.Helper()
	rows := func(query string, drop ...string) [][]string {
		result, err := db.Query(query)
		if err != nil {
			t.Fatalf("%s: %v", query, err)
		}
		defer result.Close()
		columns, err := result.Columns()
		if err != nil {
			t.Fatal(err)
		}
		var out [][]string
		for result.Next() {
			values := make([]sql.NullString, len(columns))
			targets := make([]any, len(columns))
			for index := range values {
				targets[index] = &values[index]
			}
			if err := result.Scan(targets...); err != nil {
				t.Fatal(err)
			}
			var fields []string
			for index, value := range values {
				if slices.Contains(drop, columns[index]) {
					continue
				}
				fields = append(fields, columns[index]+"="+value.String)
			}
			out = append(out, fields)
		}
		if err := result.Err(); err != nil {
			t.Fatal(err)
		}
		return out
	}
	var lines []string
	add := func(prefix string, fields [][]string) {
		for _, row := range fields {
			lines = append(lines, prefix+" "+strings.Join(row, " "))
		}
	}
	add("object", rows(`SELECT type, name FROM sqlite_schema ORDER BY type, name`))
	var tables []string
	for _, row := range rows(`SELECT name FROM sqlite_schema WHERE type = 'table' ORDER BY name`) {
		tables = append(tables, strings.TrimPrefix(row[0], "name="))
	}
	for _, table := range tables {
		add("column "+table, rows(fmt.Sprintf(`PRAGMA table_info(%q)`, table)))
		indexes := rows(fmt.Sprintf(`PRAGMA index_list(%q)`, table), "seq")
		slices.SortFunc(indexes, func(a, b []string) int { return strings.Compare(a[0], b[0]) })
		add("index "+table, indexes)
		for _, index := range indexes {
			name := strings.TrimPrefix(index[0], "name=")
			add("indexcolumn "+name, rows(fmt.Sprintf(`PRAGMA index_xinfo(%q)`, name)))
		}
		add("foreignkey "+table, rows(fmt.Sprintf(`PRAGMA foreign_key_list(%q)`, table)))
	}
	return lines
}

// A re-anchored v1 store (built from the v1 DDL) and a new store have the same
// schema, so the single current DDL list cannot drift from the migrated shape.
func TestReanchoredV1StoreHasTheSchemaOfANewStore(t *testing.T) {
	fresh, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer fresh.Close()

	v1 := newV1Store(t)
	v1.close(t)
	migrated, err := Open(v1.root)
	if err != nil {
		t.Fatal(err)
	}
	defer migrated.Close()

	want, got := describeSchema(t, fresh.db), describeSchema(t, migrated.db)
	if len(want) < 100 {
		t.Fatalf("schema description looks empty (%d lines)", len(want))
	}
	if !slices.Equal(want, got) {
		for _, line := range want {
			if !slices.Contains(got, line) {
				t.Errorf("missing from re-anchored store: %s", line)
			}
		}
		for _, line := range got {
			if !slices.Contains(want, line) {
				t.Errorf("only in re-anchored store: %s", line)
			}
		}
		t.Fatal("re-anchored v1 store differs from a new store")
	}
	for _, store := range []*Store{fresh, migrated} {
		var version int
		if err := store.db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil || version != schemaVersion {
			t.Fatalf("user_version = %d, %v", version, err)
		}
	}
}

// The schema description must see a difference in an index's column order or
// sort direction, or the equality test above would pass a drifted DDL.
func TestDescribeSchemaSeesIndexColumnAndOrderDrift(t *testing.T) {
	describe := func(statement string) []string {
		db, err := sql.Open("sqlite", ":memory:")
		if err != nil {
			t.Fatal(err)
		}
		db.SetMaxOpenConns(1)
		defer db.Close()
		for _, ddl := range []string{`CREATE TABLE t (a INTEGER, b INTEGER)`, statement} {
			if _, err := db.Exec(ddl); err != nil {
				t.Fatal(err)
			}
		}
		return describeSchema(t, db)
	}
	base := describe(`CREATE INDEX i ON t(a, b DESC)`)
	for _, drifted := range []string{`CREATE INDEX i ON t(a, b)`, `CREATE INDEX i ON t(a)`, `CREATE UNIQUE INDEX i ON t(a, b DESC)`, `CREATE INDEX i ON t(b DESC, a)`} {
		if slices.Equal(base, describe(drifted)) {
			t.Fatalf("describeSchema missed the drift %q", drifted)
		}
	}
}

// writeVersionedDB writes root/.scout/state.sqlite holding only a user_version,
// in the given journal mode, and closes it.
func writeVersionedDB(t *testing.T, root string, version int, journalMode string) string {
	t.Helper()
	db := openFixtureDB(t, root, journalMode)
	if _, err := db.Exec(fmt.Sprintf(`PRAGMA user_version=%d`, version)); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(root, ".scout", "state.sqlite")
}

func TestOpenRefusesUnsupportedVersionsWithoutTouchingTheStore(t *testing.T) {
	for _, mode := range []string{"wal", "delete"} {
		for _, version := range []int{1, 2, 4, 7, 11, 13, 99} {
			t.Run(fmt.Sprintf("%s/v%d", mode, version), func(t *testing.T) {
				root := t.TempDir()
				path := writeVersionedDB(t, root, version, mode)
				before, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				store, err := Open(root)
				if err == nil {
					store.Close()
					t.Fatalf("Open accepted user_version %d", version)
				}
				want := "re-baseline"
				if version > schemaVersion {
					want = "newer than supported version"
				}
				if !strings.Contains(err.Error(), want) || !strings.Contains(err.Error(), fmt.Sprintf("version %d", version)) {
					t.Fatalf("error %q does not name version %d and %q", err, version, want)
				}
				after, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(before, after) {
					t.Fatal("refused Open wrote the database file (or changed its journal mode)")
				}
				for _, name := range []string{"writer.lock", "artifacts"} {
					if _, err := os.Stat(filepath.Join(root, ".scout", name)); !errors.Is(err, os.ErrNotExist) {
						t.Fatalf("refused Open created %s: %v", name, err)
					}
				}
			})
		}
	}
}

func TestOpenAcceptsCurrentAndV1ReanchorVersions(t *testing.T) {
	current, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := current.Close(); err != nil {
		t.Fatal(err)
	}
	v1 := newV1Store(t)
	v1.close(t)
	migrated, err := Open(v1.root)
	if err != nil {
		t.Fatalf("v1 store (user_version 3) refused: %v", err)
	}
	defer migrated.Close()
}

func TestOpenExistingReadOnlyVersionRules(t *testing.T) {
	for _, version := range []int{1, 2, 4, 7, 11} {
		root := t.TempDir()
		writeVersionedDB(t, root, version, "wal")
		store, err := OpenExistingReadOnly(root)
		if err == nil {
			store.Close()
			t.Fatalf("read-only open accepted user_version %d", version)
		}
		if !strings.Contains(err.Error(), "re-baseline") || !strings.Contains(err.Error(), fmt.Sprintf("version %d", version)) {
			t.Fatalf("version %d: %v", version, err)
		}
	}

	root := t.TempDir()
	writeVersionedDB(t, root, 13, "wal")
	if store, err := OpenExistingReadOnly(root); err == nil {
		store.Close()
		t.Fatal("read-only open accepted a newer store")
	} else if !strings.Contains(err.Error(), "newer than supported version") {
		t.Fatalf("newer store: %v", err)
	}

	v1 := newV1Store(t)
	v1.close(t)
	if store, err := OpenExistingReadOnly(v1.root); err == nil {
		store.Close()
		t.Fatal("read-only open accepted a v1 store")
	} else if !strings.Contains(err.Error(), "requires writable re-anchor") || strings.Contains(err.Error(), "re-baseline") {
		t.Fatalf("v1 store: %v", err)
	}

	writable, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	root = writable.Root()
	if err := writable.Close(); err != nil {
		t.Fatal(err)
	}
	readOnly, err := OpenExistingReadOnly(root)
	if err != nil {
		t.Fatalf("current store refused read-only: %v", err)
	}
	defer readOnly.Close()
	if _, err := readOnly.ListIssues(context.Background()); err != nil {
		t.Fatal(err)
	}
}

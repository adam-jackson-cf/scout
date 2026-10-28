// Package state persists sanitized Scout assessment state under a target's .scout directory.
package state

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"syscall"
	"time"

	"scout/internal/domain"
	"scout/internal/redact"

	_ "modernc.org/sqlite"
)

const schemaVersion = 12

var (
	// ErrStoreLocked means another process already owns the target state directory.
	ErrStoreLocked = errors.New("scout state directory is locked by another process")
	// ErrRepositoryIdentityMismatch prevents state from one target being reused for another.
	ErrRepositoryIdentityMismatch = errors.New("repository identity does not match existing Scout state")
	// ErrNotFound reports a missing durable state record.
	ErrNotFound = errors.New("state record not found")
	// ErrImmutable reports an attempt to replace an immutable plan, run, or observation.
	ErrImmutable = errors.New("immutable state record already exists")
)

// Store owns either the exclusive state lock for a target repository or a
// read-only connection to an existing state baseline.
type Store struct {
	root         string
	stateDir     string
	artifactsDir string
	db           *sql.DB
	lock         *os.File
	readOnly     bool
	mu           sync.Mutex
	closed       bool
}

// Open creates or restores the state database stored beneath root/.scout.
func Open(root string) (*Store, error) {
	if root == "" {
		return nil, errors.New("state root is required")
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve state root: %w", err)
	}
	info, err := os.Stat(absoluteRoot)
	if err != nil {
		return nil, fmt.Errorf("stat state root: %w", err)
	}
	if !info.IsDir() {
		return nil, errors.New("state root must be a directory")
	}

	stateDir := filepath.Join(absoluteRoot, ".scout")
	artifactsDir := filepath.Join(stateDir, "artifacts")
	databasePath := filepath.Join(stateDir, "state.sqlite")
	// An existing store's version is read on a read-only connection before
	// .scout/, artifacts/, the lock or the journal mode are touched, so a store
	// this build cannot migrate is refused unchanged.
	if err := rejectSymlink(databasePath); err != nil {
		return nil, err
	}
	if _, err := os.Lstat(databasePath); err == nil {
		version, err := readUserVersion(databasePath)
		if err != nil {
			return nil, err
		}
		if version != 0 && version != schemaVersion && version != reanchorSourceVersion {
			return nil, unsupportedVersionError(version)
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("inspect state database: %w", err)
	}
	if err := secureDir(stateDir); err != nil {
		return nil, fmt.Errorf("create state directory: %w", err)
	}
	if err := secureDir(artifactsDir); err != nil {
		return nil, fmt.Errorf("create artifact directory: %w", err)
	}

	lock, err := acquireLock(filepath.Join(stateDir, "writer.lock"))
	if err != nil {
		return nil, err
	}
	cleanup := func(openErr error) (*Store, error) {
		_ = releaseLock(lock)
		return nil, openErr
	}

	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return cleanup(fmt.Errorf("open state database: %w", err))
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return cleanup(fmt.Errorf("ping state database: %w", err))
	}
	for _, pragma := range []string{`PRAGMA journal_mode=WAL`, `PRAGMA foreign_keys=ON`, `PRAGMA busy_timeout=5000`} {
		if _, err := db.Exec(pragma); err != nil {
			_ = db.Close()
			return cleanup(fmt.Errorf("configure state database: %w", err))
		}
	}
	if err := os.Chmod(databasePath, 0o600); err != nil && !errors.Is(err, fs.ErrNotExist) {
		_ = db.Close()
		return cleanup(fmt.Errorf("secure state database: %w", err))
	}
	store := &Store{root: absoluteRoot, stateDir: stateDir, artifactsDir: artifactsDir, db: db, lock: lock}
	if err := store.migrate(context.Background()); err != nil {
		_ = db.Close()
		return cleanup(err)
	}
	return store, nil
}

// OpenExistingReadOnly opens an existing Scout baseline without creating state,
// acquiring a writer lock, or applying migrations. It is safe for readiness and
// query-only surfaces that must not initialize a target repository.
func OpenExistingReadOnly(root string) (*Store, error) {
	if root == "" {
		return nil, errors.New("state root is required")
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve state root: %w", err)
	}
	if info, err := os.Stat(absoluteRoot); err != nil {
		return nil, fmt.Errorf("stat state root: %w", err)
	} else if !info.IsDir() {
		return nil, errors.New("state root must be a directory")
	}
	stateDir := filepath.Join(absoluteRoot, ".scout")
	databasePath := filepath.Join(stateDir, "state.sqlite")
	if err := rejectSymlink(stateDir); err != nil {
		return nil, err
	}
	if err := rejectSymlink(databasePath); err != nil {
		return nil, err
	}
	info, err := os.Stat(databasePath)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("stat state database: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("state database must be a regular file")
	}
	version, err := readUserVersion(databasePath)
	if err != nil {
		return nil, err
	}
	switch version {
	case schemaVersion:
	case reanchorSourceVersion:
		return nil, fmt.Errorf("state schema version %d requires writable re-anchor to version %d", version, schemaVersion)
	default:
		return nil, unsupportedVersionError(version)
	}
	db, err := sql.Open("sqlite", (&url.URL{Scheme: "file", Path: databasePath, RawQuery: "mode=ro"}).String())
	if err != nil {
		return nil, fmt.Errorf("open read-only state database: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping read-only state database: %w", err)
	}
	return &Store{root: absoluteRoot, stateDir: stateDir, artifactsDir: filepath.Join(stateDir, "artifacts"), db: db, readOnly: true}, nil
}

// Close releases the database and the exclusive state lock.
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	dbErr := s.db.Close()
	lockErr := releaseLock(s.lock)
	if dbErr != nil {
		return fmt.Errorf("close state database: %w", dbErr)
	}
	if lockErr != nil {
		return fmt.Errorf("release state lock: %w", lockErr)
	}
	return nil
}

// Root returns the absolute assessment target that owns this store.
func (s *Store) Root() string { return s.root }

func (s *Store) withTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errors.New("state store is closed")
	}
	if s.readOnly {
		return errors.New("state store is read-only")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin state transaction: %w", err)
	}
	if err := fn(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("state transaction: %w (rollback: %v)", err, rollbackErr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit state transaction: %w", err)
	}
	return nil
}

func (s *Store) withRead(ctx context.Context, fn func(*sql.DB) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errors.New("state store is closed")
	}
	return fn(s.db)
}

// reanchorSourceVersion is the one older schema a store may carry and still be
// migrated: the Scout v1 implementation's last schema (design §3.3).
const reanchorSourceVersion = 3

func (s *Store) migrate(ctx context.Context) error {
	return s.withTransaction(ctx, func(tx *sql.Tx) error {
		var current int
		if err := tx.QueryRowContext(ctx, "PRAGMA user_version").Scan(&current); err != nil {
			return fmt.Errorf("read state schema version: %w", err)
		}
		switch current {
		case schemaVersion:
			return nil
		case 0:
			if err := execAll(ctx, tx, schemaCurrent); err != nil {
				return fmt.Errorf("create state schema: %w", err)
			}
		case reanchorSourceVersion:
			if err := backfillPlanJobs(ctx, tx); err != nil {
				return fmt.Errorf("backfill plan jobs: %w", err)
			}
			if err := reanchorIssues(ctx, tx); err != nil {
				return fmt.Errorf("re-anchor issues: %w", err)
			}
			if err := execAll(ctx, tx, schemaAddedAfterV1); err != nil {
				return fmt.Errorf("create tables added since the v1 schema: %w", err)
			}
		default:
			return unsupportedVersionError(current)
		}
		if _, err := tx.ExecContext(ctx, fmt.Sprintf("PRAGMA user_version=%d", schemaVersion)); err != nil {
			return fmt.Errorf("record state schema version: %w", err)
		}
		return nil
	})
}

func execAll(ctx context.Context, tx *sql.Tx, statements []string) error {
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

// unsupportedVersionError names the remedy for a store this build cannot open:
// a newer store needs a newer Scout, any other version is re-baselined from the
// state export, which is the durable baseline (design §7.2).
func unsupportedVersionError(version int) error {
	if version > schemaVersion {
		return fmt.Errorf("state schema version %d is newer than supported version %d", version, schemaVersion)
	}
	return fmt.Errorf("state schema version %d is not supported; re-baseline: remove .scout/state.sqlite* and run scout state restore --remote origin (or scout import --from <export directory>)", version)
}

// readUserVersion reads PRAGMA user_version of an existing database on a
// read-only connection: the file is not written and no migration runs.
func readUserVersion(databasePath string) (int, error) {
	databaseURL := (&url.URL{Scheme: "file", Path: databasePath, RawQuery: "mode=ro"}).String()
	db, err := sql.Open("sqlite", databaseURL)
	if err != nil {
		return 0, fmt.Errorf("open read-only state database: %w", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		return 0, fmt.Errorf("ping read-only state database: %w", err)
	}
	var version int
	if err := db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return 0, fmt.Errorf("read state schema version: %w", err)
	}
	return version, nil
}

// schemaBase is the schema of the tables the Scout v1 implementation created,
// in their current shape (runs.finalized is part of the table).
var schemaBase = []string{
	`CREATE TABLE metadata (key TEXT PRIMARY KEY, value TEXT NOT NULL)`,
	`CREATE TABLE plans (plan_id TEXT PRIMARY KEY, run_id TEXT NOT NULL UNIQUE, repository_id TEXT NOT NULL, created_at INTEGER NOT NULL, payload BLOB NOT NULL)`,
	`CREATE TABLE runs (run_id TEXT PRIMARY KEY, plan_id TEXT NOT NULL UNIQUE, repository_id TEXT NOT NULL, state TEXT NOT NULL, started_at INTEGER NOT NULL, finished_at INTEGER, payload BLOB NOT NULL, finalized INTEGER NOT NULL DEFAULT 1, FOREIGN KEY(plan_id) REFERENCES plans(plan_id))`,
	`CREATE TABLE coverage (run_id TEXT NOT NULL, coverage_key TEXT NOT NULL, status TEXT NOT NULL, payload BLOB NOT NULL, PRIMARY KEY(run_id, coverage_key), FOREIGN KEY(run_id) REFERENCES runs(run_id) ON DELETE CASCADE)`,
	`CREATE TABLE observations (observation_id TEXT PRIMARY KEY, run_id TEXT NOT NULL, observed_at INTEGER NOT NULL, payload BLOB NOT NULL, FOREIGN KEY(run_id) REFERENCES runs(run_id) ON DELETE RESTRICT)`,
	`CREATE TABLE artifacts (digest TEXT PRIMARY KEY, artifact_id TEXT NOT NULL UNIQUE, content_type TEXT NOT NULL, size_bytes INTEGER NOT NULL, created_at INTEGER NOT NULL)`,
	`CREATE TABLE issues (issue_id TEXT PRIMARY KEY, state TEXT NOT NULL, updated_at INTEGER NOT NULL, payload BLOB NOT NULL)`,
	`CREATE TABLE issue_observations (issue_id TEXT NOT NULL, observation_id TEXT NOT NULL, PRIMARY KEY(issue_id, observation_id), FOREIGN KEY(issue_id) REFERENCES issues(issue_id) ON DELETE CASCADE, FOREIGN KEY(observation_id) REFERENCES observations(observation_id) ON DELETE RESTRICT)`,
	`CREATE TABLE issue_history (issue_id TEXT NOT NULL, ordinal INTEGER NOT NULL, payload BLOB NOT NULL, PRIMARY KEY(issue_id, ordinal), FOREIGN KEY(issue_id) REFERENCES issues(issue_id) ON DELETE CASCADE)`,
	`CREATE TABLE inventory_snapshots (run_id TEXT PRIMARY KEY, payload BLOB NOT NULL, FOREIGN KEY(run_id) REFERENCES runs(run_id))`,
	`CREATE TABLE run_job_commits (run_id TEXT NOT NULL, job_id TEXT NOT NULL, committed_at INTEGER NOT NULL, payload BLOB NOT NULL, PRIMARY KEY(run_id, job_id), FOREIGN KEY(run_id) REFERENCES runs(run_id) ON DELETE CASCADE)`,
	`CREATE INDEX coverage_success_lookup ON coverage(coverage_key, status, run_id)`,
	`CREATE INDEX runs_success_lookup ON runs(state, finished_at DESC)`,
	`CREATE INDEX observations_by_run ON observations(run_id, observed_at)`,
	`CREATE INDEX issues_by_state ON issues(state, updated_at DESC)`,
	`CREATE INDEX run_job_commits_by_run ON run_job_commits(run_id, committed_at, job_id)`,
}

// schemaAddedAfterV1 holds the tables a v1 store lacks that need no backfill.
// A re-anchored v1 store and a new store both end with them.
var schemaAddedAfterV1 = slices.Concat(schemaWouldBlock, schemaSignals, schemaAnnotations, schemaIngested, schemaPRMetadata, schemaAreas)

// schemaCurrent creates a new store at schemaVersion.
var schemaCurrent = slices.Concat(schemaBase, []string{schemaPlanJobs, schemaMigrationAmbiguities}, schemaAddedAfterV1)

func (s *Store) ensureRepositoryIdentity(ctx context.Context, tx *sql.Tx, repositoryID string) error {
	var stored string
	err := tx.QueryRowContext(ctx, `SELECT value FROM metadata WHERE key = 'repository_id'`).Scan(&stored)
	if errors.Is(err, sql.ErrNoRows) {
		if _, err := tx.ExecContext(ctx, `INSERT INTO metadata(key, value) VALUES ('repository_id', ?)`, repositoryID); err != nil {
			return fmt.Errorf("record repository identity: %w", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("read repository identity: %w", err)
	}
	if stored != repositoryID {
		return fmt.Errorf("%w: stored %q, received %q", ErrRepositoryIdentityMismatch, stored, repositoryID)
	}
	return nil
}

func ensureArtifactTx(ctx context.Context, tx *sql.Tx, ref domain.SafeArtifactRef) error {
	if err := ref.Validate(); err != nil {
		return err
	}
	var artifactID, contentType string
	if err := tx.QueryRowContext(ctx, `SELECT artifact_id, content_type FROM artifacts WHERE digest = ?`, ref.Digest).Scan(&artifactID, &contentType); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("artifact %q: %w", ref.ArtifactID, ErrNotFound)
		}
		return fmt.Errorf("load artifact metadata: %w", err)
	}
	if artifactID != ref.ArtifactID || contentType != ref.ContentType {
		return errors.New("artifact reference does not match recorded metadata")
	}
	return nil
}

func decode(payload []byte, target any) error {
	if err := json.Unmarshal(payload, target); err != nil {
		return fmt.Errorf("decode state record: %w", err)
	}
	return nil
}

func encode(value any) ([]byte, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode state record: %w", err)
	}
	if err := rejectUnsafePersistence(string(payload)); err != nil {
		return nil, err
	}
	return payload, nil
}

func rejectUnsafePersistence(value string) error {
	if redact.Contains(value) {
		return errors.New("refusing to persist credential-bearing content")
	}
	return nil
}

func secureDir(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		if err := os.Mkdir(path, 0o700); err != nil && !errors.Is(err, fs.ErrExist) {
			return err
		}
		info, err = os.Lstat(path)
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return errors.New("state directory must be a real directory, not a symlink")
	}
	return os.Chmod(path, 0o700)
}

func rejectSymlink(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect state path: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return errors.New("state path must not be a symlink")
	}
	return nil
}

func acquireLock(path string) (*os.File, error) {
	if err := rejectSymlink(path); err != nil {
		return nil, err
	}
	lock, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open state lock: %w", err)
	}
	if err := lock.Chmod(0o600); err != nil {
		_ = lock.Close()
		return nil, fmt.Errorf("secure state lock: %w", err)
	}
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = lock.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return nil, ErrStoreLocked
		}
		return nil, fmt.Errorf("acquire state lock: %w", err)
	}
	return lock, nil
}

func releaseLock(lock *os.File) error {
	if lock == nil {
		return nil
	}
	unlockErr := syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	closeErr := lock.Close()
	if unlockErr != nil {
		return unlockErr
	}
	return closeErr
}

func unixMillis(value time.Time) int64 {
	return value.UTC().UnixMilli()
}

func nullUnixMillis(value *time.Time) any {
	if value == nil {
		return nil
	}
	return unixMillis(*value)
}

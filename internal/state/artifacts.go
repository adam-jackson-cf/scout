package state

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"scout/internal/domain"
)

// ArtifactSanitizer must return only publishable content. A sanitizer error prevents
// any artifact from being written; Store never falls back to the caller's raw input.
type ArtifactSanitizer func(raw []byte) ([]byte, error)

// StoreArtifact sanitizes and atomically writes an immutable content-addressed artifact.
func (s *Store) StoreArtifact(ctx context.Context, raw []byte, contentType string, sanitize ArtifactSanitizer) (domain.SafeArtifactRef, error) {
	if s.readOnly {
		return domain.SafeArtifactRef{}, errors.New("state store is read-only")
	}
	if sanitize == nil {
		return domain.SafeArtifactRef{}, errors.New("artifact sanitizer is required")
	}
	if contentType == "" || strings.ContainsAny(contentType, "\x00\r\n") {
		return domain.SafeArtifactRef{}, errors.New("artifact content type is invalid")
	}
	safe, err := sanitize(raw)
	if err != nil {
		return domain.SafeArtifactRef{}, fmt.Errorf("sanitize artifact: %w", err)
	}
	if err := rejectUnsafePersistence(string(safe)); err != nil {
		return domain.SafeArtifactRef{}, fmt.Errorf("sanitized artifact: %w", err)
	}
	digestBytes := sha256.Sum256(safe)
	digest := hex.EncodeToString(digestBytes[:])
	ref := domain.SafeArtifactRef{ArtifactID: "sha256:" + digest, Digest: digest, ContentType: contentType}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return domain.SafeArtifactRef{}, errors.New("state store is closed")
	}
	path, err := s.artifactPath(digest)
	if err != nil {
		return domain.SafeArtifactRef{}, err
	}
	if err := writeContentAddressed(path, safe); err != nil {
		return domain.SafeArtifactRef{}, err
	}
	var recordedContentType string
	err = s.db.QueryRowContext(ctx, `SELECT content_type FROM artifacts WHERE digest = ?`, digest).Scan(&recordedContentType)
	if errors.Is(err, sql.ErrNoRows) {
		if _, err := s.db.ExecContext(ctx, `INSERT INTO artifacts(digest, artifact_id, content_type, size_bytes, created_at) VALUES (?, ?, ?, ?, ?)`, digest, ref.ArtifactID, contentType, len(safe), unixMillis(time.Now())); err != nil {
			return domain.SafeArtifactRef{}, fmt.Errorf("record artifact: %w", err)
		}
		return ref, nil
	}
	if err != nil {
		return domain.SafeArtifactRef{}, fmt.Errorf("load artifact metadata: %w", err)
	}
	if recordedContentType != contentType {
		return domain.SafeArtifactRef{}, errors.New("artifact content address already exists with a different content type")
	}
	return ref, nil
}

// LoadArtifact returns a previously sanitized artifact after checking its digest.
func (s *Store) LoadArtifact(ctx context.Context, ref domain.SafeArtifactRef) ([]byte, error) {
	if err := ref.Validate(); err != nil {
		return nil, fmt.Errorf("artifact reference: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, errors.New("state store is closed")
	}
	var artifactID, contentType string
	if err := s.db.QueryRowContext(ctx, `SELECT artifact_id, content_type FROM artifacts WHERE digest = ?`, ref.Digest).Scan(&artifactID, &contentType); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("artifact %q: %w", ref.ArtifactID, ErrNotFound)
		}
		return nil, fmt.Errorf("load artifact metadata: %w", err)
	}
	if artifactID != ref.ArtifactID || contentType != ref.ContentType {
		return nil, errors.New("artifact reference does not match recorded metadata")
	}
	path, err := s.artifactPath(ref.Digest)
	if err != nil {
		return nil, err
	}
	if err := rejectSymlink(path); err != nil {
		return nil, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("artifact %q file: %w", ref.ArtifactID, ErrNotFound)
		}
		return nil, fmt.Errorf("read artifact: %w", err)
	}
	actual := sha256.Sum256(content)
	if hex.EncodeToString(actual[:]) != ref.Digest {
		return nil, errors.New("artifact digest does not match content")
	}
	return content, nil
}

func (s *Store) artifactPath(digest string) (string, error) {
	if len(digest) != sha256.Size*2 {
		return "", errors.New("artifact digest has invalid length")
	}
	if _, err := hex.DecodeString(digest); err != nil {
		return "", errors.New("artifact digest is not hexadecimal")
	}
	return filepath.Join(s.artifactsDir, digest[:2], digest[2:]), nil
}

func writeContentAddressed(path string, content []byte) error {
	if err := secureDir(filepath.Dir(path)); err != nil {
		return fmt.Errorf("create artifact shard: %w", err)
	}
	if err := rejectSymlink(path); err != nil {
		return err
	}
	if existing, err := os.Open(path); err == nil {
		defer existing.Close()
		hash := sha256.New()
		if _, err := io.Copy(hash, existing); err != nil {
			return fmt.Errorf("verify existing artifact: %w", err)
		}
		if hex.EncodeToString(hash.Sum(nil)) != filepath.Base(filepath.Dir(path))+filepath.Base(path) {
			return errors.New("existing artifact conflicts with its content address")
		}
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect artifact path: %w", err)
	}

	temporary, err := os.CreateTemp(filepath.Dir(path), ".artifact-*")
	if err != nil {
		return fmt.Errorf("create artifact temporary file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("secure artifact temporary file: %w", err)
	}
	if _, err := temporary.Write(content); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write artifact: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync artifact: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close artifact temporary file: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		if errors.Is(err, os.ErrExist) {
			return writeContentAddressed(path, content)
		}
		return fmt.Errorf("publish artifact: %w", err)
	}
	return nil
}

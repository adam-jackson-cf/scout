package state

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"scout/internal/domain"
)

const (
	ExportPRRuns       = ExportDir + "/pr-runs.jsonl"
	ExportShadowTriage = ExportDir + "/would-block-triage.jsonl"
)

var shadowExportPaths = []string{ExportPRRuns, ExportShadowTriage}

// shadowRecord is the v2 tagged union. Data retains exact decimal IDs and the
// public domain field names; Kind is never inferred from a record's contents.
type shadowRecord struct {
	Kind string          `json:"kind"`
	Data json.RawMessage `json:"data"`
}

func (s *Store) exportShadow(ctx context.Context) (map[string][]byte, error) {
	var ledger domain.ShadowLedger
	var inputs domain.ShadowInputs
	err := s.withRead(ctx, func(db *sql.DB) error {
		tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
		if err != nil {
			return err
		}
		defer tx.Rollback()
		if err := readShadowLedgerTx(ctx, tx, &ledger); err != nil {
			return err
		}
		if err := readShadowInputsTx(ctx, tx, &inputs); err != nil {
			return err
		}
		return tx.Commit()
	})
	if err != nil {
		return nil, err
	}
	return EncodeShadow(ledger, inputs)
}

// EncodeShadow renders the shared portion of a ledger. Local records never
// leave the reader store. Legacy scope is explicit and cannot certify CI.
func EncodeShadow(ledger domain.ShadowLedger, inputs domain.ShadowInputs) (map[string][]byte, error) {
	if err := ValidateShadowLedger(ledger); err != nil {
		return nil, err
	}
	if err := validateShadowInputs(ledger, inputs); err != nil {
		return nil, err
	}
	files := map[string][]byte{ExportPRRuns: nil, ExportShadowTriage: nil}
	records, err := shadowRecords(ledger, inputs)
	if err != nil {
		return nil, err
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].Kind != records[j].Kind {
			return records[i].Kind < records[j].Kind
		}
		return bytes.Compare(records[i].Data, records[j].Data) < 0
	})
	for n, record := range records {
		if n > 0 && record.Kind == records[n-1].Kind && bytes.Equal(record.Data, records[n-1].Data) {
			continue
		}
		path := ExportPRRuns
		switch record.Kind {
		case "triage", "refusal", "pending_import":
			path = ExportShadowTriage
		}
		line, err := CanonicalLine(record)
		if err != nil {
			return nil, err
		}
		files[path] = append(files[path], line...)
	}
	return files, nil
}

func appendShadowRecord[T any](out *[]shadowRecord, kind string, rows []T, include func(T) bool) error {
	for _, row := range rows {
		if include != nil && !include(row) {
			continue
		}
		raw, err := json.Marshal(row)
		if err != nil {
			return err
		}
		*out = append(*out, shadowRecord{Kind: kind, Data: raw})
	}
	return nil
}

func shadowRecords(l domain.ShadowLedger, in domain.ShadowInputs) ([]shadowRecord, error) {
	var records []shadowRecord
	steps := []func() error{
		func() error {
			return appendShadowRecord(&records, "fact", l.Facts, func(f domain.PRRunFact) bool { return f.Scope != domain.ShadowScopeLocal })
		},
		func() error {
			return appendShadowRecord(&records, "event", l.Events, func(e domain.ProviderPREvent) bool { return e.Scope != domain.ShadowScopeLocal })
		},
		func() error {
			return appendShadowRecord(&records, "link", l.Links, func(e domain.PRRunEventLink) bool { return e.Scope != domain.ShadowScopeLocal })
		},
		func() error {
			return appendShadowRecord(&records, "triage", l.Triage, func(e domain.ShadowTriage) bool { return e.Scope != domain.ShadowScopeLocal })
		},
		func() error { return appendShadowRecord(&records, "frontier", l.Frontiers, nil) },
		func() error { return appendShadowRecord(&records, "point", l.GapPoints, nil) },
		func() error { return appendShadowRecord(&records, "interval", l.Intervals, nil) },
		func() error { return appendShadowRecord(&records, "publication", l.Publications, nil) },
		func() error {
			return appendShadowRecord(&records, "obligation", in.Obligations, func(e domain.ProviderEventObligation) bool { return e.Scope != domain.ShadowScopeLocal })
		},
		func() error {
			return appendShadowRecord(&records, "accounting", in.Accounting, func(e domain.PRRunObligationLink) bool { return e.Scope != domain.ShadowScopeLocal })
		},
		func() error { return appendShadowRecord(&records, "refusal", in.ImportRefusals, nil) },
		func() error { return appendShadowRecord(&records, "pending_import", in.PendingImports, nil) },
	}
	for _, step := range steps {
		if err := step(); err != nil {
			return nil, err
		}
	}
	return records, nil
}

func decodeShadowRow[T any](data []byte, rows *[]T) error {
	var row T
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&row); err != nil {
		return err
	}
	*rows = append(*rows, row)
	return nil
}

// DecodeShadow decodes and validates the complete immutable graph, including
// publication envelopes and supplemental inputs. Genuine v1 has no shadow rows.
func DecodeShadow(files map[string][]byte) (domain.ShadowLedger, domain.ShadowInputs, error) {
	var l domain.ShadowLedger
	var in domain.ShadowInputs
	for _, path := range shadowExportPaths {
		for n, line := range SplitLines(files[path]) {
			var r shadowRecord
			if err := json.Unmarshal(line, &r); err != nil {
				return l, in, fmt.Errorf("%s line %d: %w", path, n+1, err)
			}
			triageFile := r.Kind == "triage" || r.Kind == "refusal" || r.Kind == "pending_import"
			if triageFile != (path == ExportShadowTriage) {
				return l, in, fmt.Errorf("record %s in wrong shadow file", r.Kind)
			}
			var err error
			switch r.Kind {
			case "fact":
				err = decodeShadowRow(r.Data, &l.Facts)
			case "event":
				err = decodeShadowRow(r.Data, &l.Events)
			case "link":
				err = decodeShadowRow(r.Data, &l.Links)
			case "triage":
				err = decodeShadowRow(r.Data, &l.Triage)
			case "frontier":
				err = decodeShadowRow(r.Data, &l.Frontiers)
			case "point":
				err = decodeShadowRow(r.Data, &l.GapPoints)
			case "interval":
				err = decodeShadowRow(r.Data, &l.Intervals)
			case "publication":
				err = decodeShadowRow(r.Data, &l.Publications)
			case "obligation":
				err = decodeShadowRow(r.Data, &in.Obligations)
			case "accounting":
				err = decodeShadowRow(r.Data, &in.Accounting)
			case "refusal":
				err = decodeShadowRow(r.Data, &in.ImportRefusals)
			case "pending_import":
				err = decodeShadowRow(r.Data, &in.PendingImports)
			default:
				err = fmt.Errorf("unknown shadow record kind %q", r.Kind)
			}
			if err != nil {
				return l, in, fmt.Errorf("%s line %d: %w", path, n+1, err)
			}
		}
	}
	if err := ValidateShadowLedger(l); err != nil {
		return l, in, err
	}
	if err := validateShadowInputs(l, in); err != nil {
		return l, in, err
	}
	return l, in, nil
}

// ValidateState is the read-only compatibility gate used on every fetched CAS
// parent. It neither repairs state nor changes any imported bytes.
func ValidateState(files map[string][]byte) (ManifestMeta, error) {
	if err := verifyManifest(files); err != nil {
		return ManifestMeta{}, err
	}
	if _, _, err := DecodeShadow(files); err != nil {
		return ManifestMeta{}, err
	}
	return ReadManifestMeta(files[ExportManifest])
}

// importShadow runs in the ordinary import transaction, so invalid graph data
// cannot partially import issues or change any baseline metadata.
func importShadow(ctx context.Context, tx *sql.Tx, l domain.ShadowLedger, in domain.ShadowInputs) error {
	// A reader store with history is not a fresh restore merely because it has no
	// issues or frontier. Without an authenticated chain proof, never adopt stamps.
	var destination domain.ShadowLedger
	var destinationInputs domain.ShadowInputs
	if err := readShadowLedgerTx(ctx, tx, &destination); err != nil {
		return err
	}
	if err := readShadowInputsTx(ctx, tx, &destinationInputs); err != nil {
		return err
	}
	if err := ValidateShadowOverlap(destination, l); err != nil {
		return err
	}
	if !shadowLedgerEmpty(destination) || !shadowInputsEmpty(destinationInputs) {
		if ShadowHistoryContains(l, destination) && ShadowInputsContains(in, destinationInputs) {
			return nil
		}
		pending, err := PendingShadowImport(l, in)
		if err != nil {
			return err
		}
		return recordShadowInputs(ctx, tx, domain.ShadowInputs{PendingImports: []domain.ShadowPendingImport{pending}})
	}
	// Restore in chain order; transport order is canonical, not insertion order.
	publications := append([]domain.ShadowPublication(nil), l.Publications...)
	sort.Slice(publications, func(i, j int) bool { return publications[i].RecordedAt.Before(publications[j].RecordedAt) })
	for _, p := range publications {
		if err := applyShadowPublication(ctx, tx, p); err != nil {
			return err
		}
	}
	for _, f := range l.Facts {
		if err := storeShadowFact(ctx, tx, f, true); err != nil {
			return err
		}
	}
	for _, e := range l.Events {
		// Migration placeholders intentionally retain missing legacy provider IDs.
		if e.Scope == domain.ShadowScopeLegacy && e.RepositoryID == "" {
			if err := insertShadowJSON(ctx, tx, "shadow_events", []string{"scope", "repository_id", "event_id"}, []any{e.Scope, e.RepositoryID, e.EventID}, e); err != nil {
				return err
			}
		} else if err := storeShadowEvent(ctx, tx, e, true); err != nil {
			return err
		}
	}
	for _, link := range l.Links {
		if err := storeShadowLink(ctx, tx, link); err != nil {
			return err
		}
	}
	for _, row := range l.Triage {
		if row.Scope != domain.ShadowScopeLegacy {
			continue
		}
		if err := insertShadowJSON(ctx, tx, "would_block_triage", []string{"scope", "repository_id", "action_id", "event_id"}, []any{row.Scope, row.RepositoryID, row.ActionID, row.EventID}, row); err != nil {
			return err
		}
	}
	return recordShadowInputs(ctx, tx, in)
}

func shadowLedgerEmpty(l domain.ShadowLedger) bool {
	return len(l.Facts)+len(l.Events)+len(l.Links)+len(l.Triage)+len(l.Frontiers)+len(l.GapPoints)+len(l.Intervals)+len(l.Publications) == 0
}

func shadowInputsEmpty(in domain.ShadowInputs) bool {
	return len(in.Obligations)+len(in.Accounting)+len(in.ImportRefusals)+len(in.PendingImports) == 0
}

// PendingShadowImport packages all source material without granting it authority.
// Its content identity is only an idempotence key, never a lineage attestation.
func PendingShadowImport(l domain.ShadowLedger, in domain.ShadowInputs) (domain.ShadowPendingImport, error) {
	files, err := EncodeShadow(l, in)
	if err != nil {
		return domain.ShadowPendingImport{}, err
	}
	// Decode canonical transport to remove local rows and stabilize set order.
	history, inputs, err := DecodeShadow(files)
	if err != nil {
		return domain.ShadowPendingImport{}, err
	}
	repository := ""
	for _, f := range history.Facts {
		if f.Scope == domain.ShadowScopeCI {
			if repository != "" && repository != f.ID.RepositoryID {
				return domain.ShadowPendingImport{}, fmt.Errorf("import spans multiple repositories")
			}
			repository = f.ID.RepositoryID
		}
	}
	for _, p := range history.Publications {
		if repository != "" && repository != p.RepositoryID {
			return domain.ShadowPendingImport{}, fmt.Errorf("import spans multiple repositories")
		}
		repository = p.RepositoryID
	}
	for _, r := range inputs.ImportRefusals {
		if repository != "" && repository != r.RepositoryID {
			return domain.ShadowPendingImport{}, fmt.Errorf("import spans multiple repositories")
		}
		repository = r.RepositoryID
	}
	for _, o := range inputs.Obligations {
		if o.Scope == domain.ShadowScopeCI {
			if repository != "" && repository != o.RepositoryID {
				return domain.ShadowPendingImport{}, fmt.Errorf("import spans multiple repositories")
			}
			repository = o.RepositoryID
		}
	}
	for _, p := range inputs.PendingImports {
		if repository != "" && repository != p.RepositoryID {
			return domain.ShadowPendingImport{}, fmt.Errorf("import spans multiple repositories")
		}
		repository = p.RepositoryID
	}
	if repository == "" {
		return domain.ShadowPendingImport{}, fmt.Errorf("pending import has no CI repository identity")
	}
	hash := sha256.New()
	for _, path := range shadowExportPaths {
		hash.Write([]byte(path))
		hash.Write([]byte{0})
		hash.Write(files[path])
	}
	id := hex.EncodeToString(hash.Sum(nil))
	return domain.ShadowPendingImport{ID: id, RepositoryID: repository, Scope: domain.ShadowScopeCI, SourceLineage: "unverified:" + id, History: &history, Inputs: &inputs}, nil
}

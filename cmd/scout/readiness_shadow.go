package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"time"

	"scout/internal/domain"
	"scout/internal/policy"
	"scout/internal/shadow"
	"scout/internal/state"
)

type shadowFlags struct {
	enabled                        bool
	check, version, asOf, encoding string
}

func registerShadowFlags(set *flag.FlagSet) *shadowFlags {
	flags := &shadowFlags{}
	set.BoolVar(&flags.enabled, "shadow", false, "query existing CI shadow readiness without scanning or initializing state")
	set.StringVar(&flags.check, "check", "", "shadow check ID (required with --shadow)")
	set.StringVar(&flags.version, "version", "", "exact shadow rule version (default: latest visible pin)")
	set.StringVar(&flags.asOf, "as-of", "", "historical shadow cutoff (RFC3339); never activation evidence")
	set.StringVar(&flags.encoding, "encoding", shadow.EncodingVersion, "canonical shadow dossier encoding")
	return flags
}

func readinessShadow(ctx context.Context, flags *commonFlags, opts *shadowFlags, set *flag.FlagSet) int {
	query := shadow.Query{CheckID: opts.check, RuleVersion: opts.version, EncodingVersion: opts.encoding, Now: time.Now().UTC()}
	var inputErr error
	set.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "root", "json", "shadow", "check", "version", "as-of", "encoding":
		default:
			inputErr = fmt.Errorf("--%s cannot be combined with --shadow", f.Name)
		}
	})
	if inputErr != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, inputErr)
	}
	if opts.check == "" {
		return emit(flags, "unavailable", exitUnavailable, nil, errors.New("--shadow requires --check"))
	}
	if err := policy.ValidateEligibleID(opts.check); err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	if opts.encoding != shadow.EncodingVersion {
		return emit(flags, "unavailable", exitUnavailable, nil, fmt.Errorf("unsupported shadow encoding %q", opts.encoding))
	}
	mode := "current_window"
	if opts.asOf != "" {
		cutoff, err := time.Parse(time.RFC3339Nano, opts.asOf)
		if err != nil {
			return emit(flags, "unavailable", exitUnavailable, nil, fmt.Errorf("invalid --as-of: %w", err))
		}
		query.AsOf = &cutoff
		mode = "historical_reproduction"
	}
	store, err := state.OpenExistingReadOnly(flags.root)
	if errors.Is(err, state.ErrNotFound) {
		return emit(flags, "unmeasured", exitPartial, nil, err)
	}
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	ledger, readErr := store.ReadShadowLedger(ctx)
	closeErr := store.Close()
	if err := errors.Join(readErr, closeErr); err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	query.RepositoryID, err = shadowRepository(ledger)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	report, err := shadow.Evaluate(ledger, query)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	dossier, digest, err := shadow.CanonicalDossier(report.Dossier)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	result := struct {
		QueryMode     string              `json:"queryMode"`
		Dossier       json.RawMessage     `json:"dossier"`
		DossierDigest string              `json:"dossierDigest"`
		Live          shadow.LiveEnvelope `json:"live"`
	}{mode, dossier, digest, report.Live}
	status := "complete"
	if report.ExitCode == exitPartial {
		status = "unmeasured"
	} else if report.ExitCode == exitUnavailable {
		status = "unavailable"
	}
	return emit(flags, status, report.ExitCode, result, nil)
}

// The persisted provider identity, not a working-tree remote or network lookup,
// selects the repository. Never silently combine independent repositories.
func shadowRepository(ledger domain.ShadowLedger) (string, error) {
	var repository string
	accept := func(id string) error {
		if repository != "" && id != "" && id != repository {
			return errors.New("shadow store contains multiple repository identities")
		}
		if id != "" {
			repository = id
		}
		return nil
	}
	for _, publication := range ledger.Publications {
		if err := accept(publication.RepositoryID); err != nil {
			return "", err
		}
	}
	for _, frontier := range ledger.Frontiers {
		if err := accept(frontier.RepositoryID); err != nil {
			return "", err
		}
	}
	for _, fact := range ledger.Facts {
		if fact.Scope == domain.ShadowScopeCI {
			if err := accept(fact.ID.RepositoryID); err != nil {
				return "", err
			}
		}
	}
	return repository, nil
}

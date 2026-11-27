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

func shadowFlagError(set *flag.FlagSet) error {
	var inputErr error
	set.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "as-of", "version":
			if f.Value.String() == "" {
				inputErr = fmt.Errorf("--%s requires a nonempty value", f.Name)
			}
		case "root", "json", "shadow", "check", "encoding":
		default:
			inputErr = fmt.Errorf("--%s cannot be combined with --shadow", f.Name)
		}
	})
	return inputErr
}

func shadowQuery(opts *shadowFlags) (shadow.Query, string, error) {
	query := shadow.Query{CheckID: opts.check, RuleVersion: opts.version, EncodingVersion: opts.encoding, Now: time.Now().UTC()}
	if opts.check == "" {
		return query, "", errors.New("--shadow requires --check")
	}
	if err := policy.ValidateEligibleID(opts.check); err != nil {
		return query, "", err
	}
	if opts.encoding != shadow.EncodingVersion {
		return query, "", fmt.Errorf("unsupported shadow encoding %q", opts.encoding)
	}
	mode := "current_window"
	if opts.asOf != "" {
		cutoff, err := time.Parse(time.RFC3339Nano, opts.asOf)
		if err != nil {
			return query, "", fmt.Errorf("invalid --as-of: %w", err)
		}
		query.AsOf = &cutoff
		mode = "historical_reproduction"
	}
	return query, mode, nil
}

type shadowResult struct {
	QueryMode     string              `json:"queryMode"`
	Dossier       json.RawMessage     `json:"dossier"`
	DossierDigest string              `json:"dossierDigest"`
	Live          shadow.LiveEnvelope `json:"live"`
}

func evaluateShadow(ledger domain.ShadowLedger, query shadow.Query, mode string) (shadowResult, int, error) {
	var err error
	query.RepositoryID, err = shadowRepository(ledger)
	if err != nil {
		return shadowResult{}, exitUnavailable, err
	}
	report, err := shadow.Evaluate(ledger, query)
	if err != nil {
		return shadowResult{}, exitUnavailable, err
	}
	dossier, digest, err := shadow.CanonicalDossier(report.Dossier)
	if err != nil {
		return shadowResult{}, exitUnavailable, err
	}
	return shadowResult{mode, dossier, digest, report.Live}, report.ExitCode, nil
}

func shadowStatus(code int) string {
	switch code {
	case exitPartial:
		return "unmeasured"
	case exitUnavailable:
		return "unavailable"
	default:
		return "complete"
	}
}

func readinessShadow(ctx context.Context, flags *commonFlags, opts *shadowFlags, set *flag.FlagSet) int {
	if inputErr := shadowFlagError(set); inputErr != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, inputErr)
	}
	query, mode, err := shadowQuery(opts)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
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
	result, code, err := evaluateShadow(ledger, query, mode)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	return emit(flags, shadowStatus(code), code, result, nil)
}

// The persisted provider identity, not a working-tree remote or network lookup,
// selects the repository. Never silently combine independent repositories.
func shadowRepository(ledger domain.ShadowLedger) (string, error) {
	var ids []string
	for _, publication := range ledger.Publications {
		ids = append(ids, publication.RepositoryID)
	}
	for _, frontier := range ledger.Frontiers {
		ids = append(ids, frontier.RepositoryID)
	}
	for _, fact := range ledger.Facts {
		if fact.Scope == domain.ShadowScopeCI {
			ids = append(ids, fact.ID.RepositoryID)
		}
	}
	return singleRepository(ids)
}

func singleRepository(ids []string) (string, error) {
	var repository string
	for _, id := range ids {
		if repository != "" && id != "" && id != repository {
			return "", errors.New("shadow store contains multiple repository identities")
		}
		if id != "" {
			repository = id
		}
	}
	return repository, nil
}

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"

	"scout/internal/areas"
	"scout/internal/decisions"
	"scout/internal/decisions/orderflip"
	"scout/internal/policy"
)

func decisionsCmd(ctx context.Context, args []string) int {
	if len(args) == 0 || args[0] != "orderflip" {
		return emit(nil, "unavailable", exitUnavailable, nil, errors.New("usage: scout decisions orderflip --question merge.which_area --units <file.jsonl> [--json]"))
	}
	return orderflipCmd(ctx, args[1:])
}

type flipMeasurement struct {
	orderflip.Report
	MaxShift float64 `json:"max_shift"`
	// Each unit can carry a different catalogue, hence a different question version.
	QuestionVersions map[string]string `json:"question_versions"`
}

type flipInput struct {
	unit     orderflip.Unit
	question decisions.Question
}

func validateFlipQuestion(question string) error {
	switch question {
	case "merge.which_area":
		return nil
	case "merge.covered", "triage.real_defect":
		return fmt.Errorf("question %q: only a Choice has an option order", question)
	default:
		return fmt.Errorf("unsupported orderflip question %q; use merge.which_area", question)
	}
}

func flipConsent(root string) (policy.Decisions, string, error) {
	pol, err := headPolicy(root)
	if err != nil {
		return policy.Decisions{}, "", err
	}
	consent := pol.Decisions()
	if !consent.EgressAcknowledged {
		return policy.Decisions{}, "", errors.New("decisions.egress_acknowledged must be true for orderflip")
	}
	key := jevKey()
	if reason, allowed := (decisions.Eligibility{Enabled: consent.Enabled, HasKey: key != "", EgressAllowed: consent.EgressAcknowledged}).Skip(); !allowed {
		return policy.Decisions{}, "", errors.New(reason)
	}
	cfg := decisions.Config{Model: consent.Model, MaxPerRun: consent.MaxPerRun, MaxCostUSD: consent.MaxCostUSD}
	if err := cfg.Validate(); err != nil {
		return policy.Decisions{}, "", err
	}
	return consent, key, nil
}

func measureFlips(ctx context.Context, client *decisions.Client, question, model string, inputs []flipInput) (flipMeasurement, error) {
	report := flipMeasurement{Report: orderflip.Report{QuestionID: question, Model: model, Units: len(inputs), Reports: make([]orderflip.UnitReport, 0, len(inputs))}, QuestionVersions: make(map[string]string, len(inputs))}
	for _, input := range inputs {
		one, err := orderflip.Run(ctx, client, input.question, model, []orderflip.Unit{input.unit})
		if err != nil {
			return flipMeasurement{}, err
		}
		report.QuestionVersions[input.unit.ID] = one.Version
		report.Measured += one.Measured
		report.Unmeasured += one.Unmeasured
		report.Flipped += one.Flipped
		report.CostUSD += one.CostUSD
		report.Reports = append(report.Reports, one.Reports...)
		for _, row := range one.Reports {
			report.MaxShift = math.Max(report.MaxShift, row.MaxShift)
		}
	}
	if report.Measured > 0 {
		rate := float64(report.Flipped) / float64(report.Measured)
		report.FlipRate = &rate
	}
	return report, nil
}

func flipSummary(report flipMeasurement, question, model string) string {
	rate := "unmeasured"
	if report.FlipRate != nil {
		rate = fmt.Sprintf("%.6f", *report.FlipRate)
	}
	return fmt.Sprintf("%s (%s)\nflip rate: %s\nmeasured: %d; unmeasured: %d; flipped: %d\nmax probability shift: %g\ncost USD: %g\n", question, model, rate, report.Measured, report.Unmeasured, report.Flipped, report.MaxShift, report.CostUSD)
}

func orderflipCmd(ctx context.Context, args []string) int {
	set := flag.NewFlagSet("decisions orderflip", flag.ContinueOnError)
	flags := common(set)
	var question, path string
	set.StringVar(&question, "question", "", "Choice question id")
	set.StringVar(&path, "units", "", "JSONL units with id and bounded association state including catalogue")
	if !parse(set, args, flags) {
		return exitUnavailable
	}
	refuse := func(err error) int { return emit(flags, "unavailable", exitUnavailable, nil, err) }
	if err := validateFlipQuestion(question); err != nil {
		return refuse(err)
	}
	if path == "" {
		return refuse(errors.New("--units is required"))
	}
	consent, key, err := flipConsent(flags.root)
	if err != nil {
		return refuse(err)
	}
	inputs, err := readFlipUnits(path)
	if err != nil {
		return refuse(err)
	}
	report, err := measureFlips(ctx, &decisions.Client{Key: key}, question, consent.Model, inputs)
	if err != nil {
		return refuse(err)
	}
	if flags.json {
		return emit(flags, "complete", exitComplete, report, nil)
	}
	return printStdout(flipSummary(report, question, consent.Model), exitComplete)
}

type flipCatalogueEntry struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Summary     string   `json:"summary"`
	SampleFiles []string `json:"sample_files"`
}

func flipCatalogue(unitID string, entries []flipCatalogueEntry) ([]areas.Area, error) {
	catalogue := make([]areas.Area, 0, len(entries))
	ids := make(map[string]bool, len(entries))
	for _, area := range entries {
		if area.ID == "" || ids[area.ID] {
			return nil, fmt.Errorf("unit %q: catalogue ids must be nonempty and unique", unitID)
		}
		ids[area.ID] = true
		catalogue = append(catalogue, areas.Area{ScoutAreaID: area.ID, Title: area.Title, Summary: area.Summary, FileSet: area.SampleFiles})
	}
	return catalogue, nil
}

func parseFlipUnit(line int, raw []byte, seen map[string]bool) (flipInput, error) {
	var row struct {
		ID    string          `json:"id"`
		State json.RawMessage `json:"state"`
	}
	if err := json.Unmarshal(raw, &row); err != nil {
		return flipInput{}, fmt.Errorf("units line %d: %w", line, err)
	}
	if row.ID == "" || seen[row.ID] {
		return flipInput{}, fmt.Errorf("units line %d: id must be nonempty and unique", line)
	}
	seen[row.ID] = true
	var state struct {
		Catalogue []flipCatalogueEntry `json:"catalogue"`
	}
	if err := json.Unmarshal(row.State, &state); err != nil {
		return flipInput{}, fmt.Errorf("unit %q state: %w", row.ID, err)
	}
	catalogue, err := flipCatalogue(row.ID, state.Catalogue)
	if err != nil {
		return flipInput{}, err
	}
	q := decisions.MergeWhichArea(catalogue)
	if err := q.Validate(); err != nil {
		return flipInput{}, fmt.Errorf("unit %q: %w", row.ID, err)
	}
	return flipInput{unit: orderflip.Unit{ID: row.ID, State: row.State}, question: q}, nil
}

func readFlipUnits(path string) (inputs []flipInput, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	seen := make(map[string]bool)
	for line := 1; scanner.Scan(); line++ {
		input, err := parseFlipUnit(line, scanner.Bytes(), seen)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, input)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(inputs) == 0 {
		return nil, errors.New("units file is empty")
	}
	return inputs, nil
}

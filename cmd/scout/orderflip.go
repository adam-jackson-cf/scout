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
	switch question {
	case "merge.which_area":
	case "merge.covered", "triage.real_defect":
		return refuse(fmt.Errorf("question %q: only a Choice has an option order", question))
	default:
		return refuse(fmt.Errorf("unsupported orderflip question %q; use merge.which_area", question))
	}
	if path == "" {
		return refuse(errors.New("--units is required"))
	}
	pol, err := headPolicy(flags.root)
	if err != nil {
		return refuse(err)
	}
	consent := pol.Decisions()
	if !consent.EgressAcknowledged {
		return refuse(errors.New("decisions.egress_acknowledged must be true for orderflip"))
	}
	key := jevKey()
	if reason, allowed := (decisions.Eligibility{Enabled: consent.Enabled, HasKey: key != "", EgressAllowed: consent.EgressAcknowledged}).Skip(); !allowed {
		return refuse(errors.New(reason))
	}
	cfg := decisions.Config{Model: consent.Model, MaxPerRun: consent.MaxPerRun, MaxCostUSD: consent.MaxCostUSD}
	if err := cfg.Validate(); err != nil {
		return refuse(err)
	}
	inputs, err := readFlipUnits(path)
	if err != nil {
		return refuse(err)
	}
	report := flipMeasurement{Report: orderflip.Report{QuestionID: question, Model: consent.Model, Units: len(inputs), Reports: make([]orderflip.UnitReport, 0, len(inputs))}, QuestionVersions: make(map[string]string, len(inputs))}
	client := &decisions.Client{Key: key}
	for _, input := range inputs {
		one, err := orderflip.Run(ctx, client, input.question, consent.Model, []orderflip.Unit{input.unit})
		if err != nil {
			return refuse(err)
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
	if flags.json {
		return emit(flags, "complete", exitComplete, report, nil)
	}
	rate := "unmeasured"
	if report.FlipRate != nil {
		rate = fmt.Sprintf("%.6f", *report.FlipRate)
	}
	fmt.Fprintf(os.Stdout, "%s (%s)\nflip rate: %s\nmeasured: %d; unmeasured: %d; flipped: %d\nmax probability shift: %g\ncost USD: %g\n", question, consent.Model, rate, report.Measured, report.Unmeasured, report.Flipped, report.MaxShift, report.CostUSD)
	return exitComplete
}

func readFlipUnits(path string) ([]flipInput, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	var inputs []flipInput
	seen := make(map[string]bool)
	for line := 1; scanner.Scan(); line++ {
		var row struct {
			ID    string          `json:"id"`
			State json.RawMessage `json:"state"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
			return nil, fmt.Errorf("units line %d: %w", line, err)
		}
		if row.ID == "" || seen[row.ID] {
			return nil, fmt.Errorf("units line %d: id must be nonempty and unique", line)
		}
		seen[row.ID] = true
		var state struct {
			Catalogue []struct {
				ID          string   `json:"id"`
				Title       string   `json:"title"`
				Summary     string   `json:"summary"`
				SampleFiles []string `json:"sample_files"`
			} `json:"catalogue"`
		}
		if err := json.Unmarshal(row.State, &state); err != nil {
			return nil, fmt.Errorf("unit %q state: %w", row.ID, err)
		}
		catalogue := make([]areas.Area, 0, len(state.Catalogue))
		ids := make(map[string]bool, len(state.Catalogue))
		for _, area := range state.Catalogue {
			if area.ID == "" || ids[area.ID] {
				return nil, fmt.Errorf("unit %q: catalogue ids must be nonempty and unique", row.ID)
			}
			ids[area.ID] = true
			catalogue = append(catalogue, areas.Area{ScoutAreaID: area.ID, Title: area.Title, Summary: area.Summary, FileSet: area.SampleFiles})
		}
		q := decisions.MergeWhichArea(catalogue)
		if err := q.Validate(); err != nil {
			return nil, fmt.Errorf("unit %q: %w", row.ID, err)
		}
		inputs = append(inputs, flipInput{unit: orderflip.Unit{ID: row.ID, State: row.State}, question: q})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(inputs) == 0 {
		return nil, errors.New("units file is empty")
	}
	return inputs, nil
}

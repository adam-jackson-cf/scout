package main

import (
	"encoding/json"
	"fmt"
	"os"

	"scout/internal/eval"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go run evaluation/replay.go <requests.json>")
		os.Exit(2)
	}
	loaded, err := eval.LoadCorpus("evaluation/reference-v1.json")
	if err != nil || loaded.Status != eval.CorpusAvailable {
		panic(fmt.Sprintf("load reference corpus: %s %v", loaded.Status, err))
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var requests []eval.EvaluationRequest
	if err := json.Unmarshal(data, &requests); err != nil {
		panic(err)
	}
	reports := make([]eval.OutcomeReport, 0, len(requests))
	for _, request := range requests {
		report, err := eval.CompareOutcomes(loaded, request)
		if err != nil {
			panic(fmt.Sprintf("compare %s/%s: %v", request.Language, request.Family, err))
		}
		reports = append(reports, report)
	}
	if err := json.NewEncoder(os.Stdout).Encode(reports); err != nil {
		panic(err)
	}
}

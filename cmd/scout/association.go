package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"scout/internal/association"
	"scout/internal/decisions"
	"scout/internal/domain"
	"scout/internal/policy"
)

// associationGuardReport shares the guard diagnostic across CLI outputs.
type associationGuardReport struct {
	AssociationGuard association.GuardState `json:"associationGuard"`
}

// associationOptions reuses assess's explicit commit range, or the GitHub push
// payload when present. A missing range stays empty: it is never guessed.
func associationOptions(pol policy.Policy, base, head string, fork, noEgress bool) (association.LaneOptions, domain.LaneTrigger, error) {
	trigger := domain.TriggerPush
	eventName := os.Getenv("GITHUB_EVENT_NAME")
	if eventName != "" && eventName != "push" {
		trigger = domain.TriggerPR
		if eventName == "schedule" {
			trigger = domain.TriggerNightly
		}
	}
	if path := os.Getenv("GITHUB_EVENT_PATH"); path != "" {
		payload, err := os.ReadFile(path)
		if err != nil {
			return association.LaneOptions{}, trigger, fmt.Errorf("read GitHub event: %w", err)
		}
		var event struct {
			Before     string `json:"before"`
			After      string `json:"after"`
			Ref        string `json:"ref"`
			Repository struct {
				DefaultBranch string `json:"default_branch"`
			} `json:"repository"`
			PullRequest *struct {
				Head struct {
					Repo struct {
						Fork bool `json:"fork"`
					} `json:"repo"`
				} `json:"head"`
			} `json:"pull_request"`
		}
		if err := json.Unmarshal(payload, &event); err != nil {
			return association.LaneOptions{}, trigger, fmt.Errorf("decode GitHub event: %w", err)
		}
		if event.PullRequest != nil {
			trigger = domain.TriggerPR
			fork = fork || event.PullRequest.Head.Repo.Fork
		}
		if eventName == "push" {
			// Only a default-branch push is eligible for merge association.
			if event.Repository.DefaultBranch == "" || event.Ref != "refs/heads/"+event.Repository.DefaultBranch {
				trigger = domain.TriggerPR
			}
			if base == "" && head == "" {
				base, head = event.Before, event.After
			}
		}
	}
	// GitHub's zero before/after denotes branch creation/deletion, not a range.
	if strings.Trim(base, "0") == "" || strings.Trim(head, "0") == "" {
		base, head = "", ""
	}
	consent := pol.Decisions()
	key := jevKey()
	config := decisions.Config{Enabled: consent.Enabled, Model: consent.Model, MaxPerRun: consent.MaxPerRun, MaxCostUSD: consent.MaxCostUSD, EgressAcknowledged: consent.EgressAcknowledged}
	mapping := pol.Mapping().Association
	return association.LaneOptions{
		Base: base, Head: head, Mapping: mapping, Config: config,
		FallbackConsent: pol.LaneOptedIn(policy.LaneAssociationFallback),
		Eligibility:     decisions.Eligibility{Enabled: consent.Enabled, HasKey: key != "", EgressAllowed: consent.EgressAcknowledged && !noEgress, Fork: fork},
		Client:          &decisions.Client{Key: key},
		Fingerprint:     association.Fingerprint{CoveredScore: mapping.CoveredScore, TChoice: mapping.TChoice, CoveredQuestion: decisions.MergeCovered.Version(consent.Model), ChoiceQuestion: decisions.MergeWhichAreaTextDigest(), JevModel: consent.Model, PromptDigest: association.TerraPromptDigest, FallbackModel: mapping.Fallback.Model},
	}, trigger, nil
}

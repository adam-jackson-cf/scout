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

type gitHubEvent struct {
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

type associationRun struct {
	trigger    domain.LaneTrigger
	fork       bool
	base, head string
}

func eventTrigger(eventName string) domain.LaneTrigger {
	trigger := domain.TriggerPush
	if eventName != "" && eventName != "push" {
		trigger = domain.TriggerPR
		if eventName == "schedule" {
			trigger = domain.TriggerNightly
		}
	}
	return trigger
}

func readGitHubEvent(path string) (gitHubEvent, error) {
	var event gitHubEvent
	payload, err := os.ReadFile(path)
	if err != nil {
		return event, fmt.Errorf("read GitHub event: %w", err)
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		return event, fmt.Errorf("decode GitHub event: %w", err)
	}
	return event, nil
}

func (r associationRun) withEvent(event gitHubEvent, eventName string) associationRun {
	if event.PullRequest != nil {
		r.trigger = domain.TriggerPR
		r.fork = r.fork || event.PullRequest.Head.Repo.Fork
	}
	if eventName == "push" {
		// Only a default-branch push is eligible for merge association.
		if event.Repository.DefaultBranch == "" || event.Ref != "refs/heads/"+event.Repository.DefaultBranch {
			r.trigger = domain.TriggerPR
		}
		if r.base == "" && r.head == "" {
			r.base, r.head = event.Before, event.After
		}
	}
	return r
}

// associationOptions reuses assess's explicit commit range, or the GitHub push
// payload when present. A missing range stays empty: it is never guessed.
func associationOptions(pol policy.Policy, base, head string, fork, noEgress bool) (association.LaneOptions, domain.LaneTrigger, error) {
	eventName := os.Getenv("GITHUB_EVENT_NAME")
	run := associationRun{trigger: eventTrigger(eventName), fork: fork, base: base, head: head}
	if path := os.Getenv("GITHUB_EVENT_PATH"); path != "" {
		event, err := readGitHubEvent(path)
		if err != nil {
			return association.LaneOptions{}, run.trigger, err
		}
		run = run.withEvent(event, eventName)
	}
	// GitHub's zero before/after denotes branch creation/deletion, not a range.
	if strings.Trim(run.base, "0") == "" || strings.Trim(run.head, "0") == "" {
		run.base, run.head = "", ""
	}
	return associationLaneOptions(pol, run, noEgress), run.trigger, nil
}

func associationLaneOptions(pol policy.Policy, run associationRun, noEgress bool) association.LaneOptions {
	consent := pol.Decisions()
	key := jevKey()
	config := decisions.Config{Enabled: consent.Enabled, Model: consent.Model, MaxPerRun: consent.MaxPerRun, MaxCostUSD: consent.MaxCostUSD, EgressAcknowledged: consent.EgressAcknowledged}
	mapping := pol.Mapping().Association
	return association.LaneOptions{
		Base: run.base, Head: run.head, Mapping: mapping, Config: config,
		FallbackConsent: pol.LaneOptedIn(policy.LaneAssociationFallback),
		Eligibility:     decisions.Eligibility{Enabled: consent.Enabled, HasKey: key != "", EgressAllowed: consent.EgressAcknowledged && !noEgress, Fork: run.fork},
		Client:          &decisions.Client{Key: key},
		Fingerprint:     association.Fingerprint{CoveredScore: mapping.CoveredScore, TChoice: mapping.TChoice, CoveredQuestion: decisions.MergeCovered.Version(consent.Model), ChoiceQuestion: decisions.MergeWhichAreaTextDigest(), JevModel: consent.Model, PromptDigest: association.TerraPromptDigest, FallbackModel: mapping.Fallback.Model},
	}
}

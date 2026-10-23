package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type workflowFile struct {
	Permissions map[string]string `yaml:"permissions"`
	Jobs        map[string]struct {
		If          string            `yaml:"if"`
		Uses        string            `yaml:"uses"`
		Permissions map[string]string `yaml:"permissions"`
	} `yaml:"jobs"`
}

func readWorkflow(t *testing.T, name string) (string, workflowFile) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", name))
	if err != nil {
		t.Fatal(err)
	}
	var w workflowFile
	if err := yaml.Unmarshal(raw, &w); err != nil {
		t.Fatalf("%s does not parse: %v", name, err)
	}
	return string(raw), w
}

// jobBlock returns the source text of one top-level job.
func jobBlock(t *testing.T, raw, job string) string {
	t.Helper()
	re := regexp.MustCompile(`(?m)^  ` + regexp.QuoteMeta(job) + `:\n((?:    .*\n|\n)+)`)
	m := re.FindStringSubmatch(raw)
	if m == nil {
		t.Fatalf("job %s not found", job)
	}
	return m[1]
}

func TestGuardianWorkflowTrustBoundary(t *testing.T) {
	raw, w := readWorkflow(t, "scout-guardian.yml")
	code := regexp.MustCompile(`(?m)^\s*#.*$`).ReplaceAllString(raw, "")

	if strings.Contains(code, "go build") || strings.Contains(code, "./cmd/scout") || strings.Contains(code, "bootstrap-tools") {
		t.Fatal("the guardian workflow builds Scout from the checkout")
	}
	for _, want := range []string{"releases/download/${SCOUT_VERSION}", "sha256sum", "clawpatch-manifest.json", "scout-linux-amd64"} {
		if !strings.Contains(code, want) {
			t.Errorf("release download with checksum verification is missing %q", want)
		}
	}
	for _, banned := range []string{"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "CLAWPATCH_OPENAI", "CLAWPATCH_ANTHROPIC", "aws-actions", "AWS_", "SCOUT_AWS", "s3://", "id-token"} {
		if strings.Contains(code, banned) {
			t.Errorf("forbidden reference %q", banned)
		}
	}

	if w.Permissions["contents"] != "read" {
		t.Errorf("workflow default permissions: %v", w.Permissions)
	}
	pr, trusted := w.Jobs["pull-request"], w.Jobs["trusted"]
	if pr.If != "github.event_name == 'pull_request'" || len(pr.Permissions) != 1 || pr.Permissions["contents"] != "read" {
		t.Errorf("pull-request job must be read-only: if=%q perms=%v", pr.If, pr.Permissions)
	}
	if trusted.Permissions["contents"] != "write" || len(trusted.Permissions) != 1 {
		t.Errorf("trusted job permissions: %v", trusted.Permissions)
	}
	if !strings.Contains(trusted.If, "github.event_name != 'pull_request'") || !strings.Contains(trusted.If, "default_branch") {
		t.Errorf("trusted job is not limited to default-branch non-PR runs: %q", trusted.If)
	}

	if strings.Contains(jobBlock(t, raw, "pull-request"), "secrets.") {
		t.Error("the pull-request job references a secret")
	}
	if got := strings.Count(code, "secrets."); got != 1 || !strings.Contains(jobBlock(t, raw, "trusted"), "secrets.OPENROUTER_API_KEY") {
		t.Errorf("only the trusted job may read the single OPENROUTER_API_KEY secret (secret references: %d)", got)
	}
	if strings.Contains(jobBlock(t, raw, "pull-request"), "state publish") || !strings.Contains(jobBlock(t, raw, "trusted"), "state publish") {
		t.Error("only the trusted job publishes scout/state")
	}
	if !strings.Contains(jobBlock(t, raw, "pull-request"), "BASE_SHA") {
		t.Error("pull-request policy must be read from the base commit")
	}
}

func TestNightlyIsACallerOfTheReusableWorkflow(t *testing.T) {
	raw, w := readWorkflow(t, "scout-nightly.yml")
	if len(w.Jobs) == 0 {
		t.Fatal("no jobs")
	}
	for name, job := range w.Jobs {
		if job.Uses != "./.github/workflows/scout-guardian.yml" {
			t.Errorf("job %s does not call the reusable workflow: %q", name, job.Uses)
		}
	}
	for _, banned := range []string{"aws-actions", "AWS", "s3", "id-token", "go build", "bootstrap-tools", "OPENAI_API_KEY", "ANTHROPIC_API_KEY", "SCOUT_PROVIDER"} {
		if strings.Contains(raw, banned) {
			t.Errorf("caller still references %q", banned)
		}
	}
	if w.Jobs["pull-request"].Permissions["contents"] != "read" || w.Jobs["trusted"].Permissions["contents"] != "write" {
		t.Errorf("caller permissions: %v / %v", w.Jobs["pull-request"].Permissions, w.Jobs["trusted"].Permissions)
	}
	if strings.Contains(jobBlock(t, raw, "pull-request"), "secrets") {
		t.Error("the pull-request caller job passes secrets")
	}
	for _, trigger := range []string{"pull_request:", "push:", "schedule:", "workflow_dispatch:", "refresh-map", "authorize-egress"} {
		if !strings.Contains(raw, trigger) {
			t.Errorf("trigger %q missing", trigger)
		}
	}
}

func TestCompositeActionInstallsFromVerifiedRelease(t *testing.T) {
	dir := filepath.Join("..", "..", ".github", "actions", "scout")
	action, err := os.ReadFile(filepath.Join(dir, "action.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Runs struct{ Using string } `yaml:"runs"`
	}
	if err := yaml.Unmarshal(action, &parsed); err != nil || parsed.Runs.Using != "composite" {
		t.Fatalf("not a composite action: %v %q", err, parsed.Runs.Using)
	}
	install, err := os.ReadFile(filepath.Join(dir, "install.sh"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(action) + string(install)
	for _, want := range []string{"releases/download", "digest", "clawpatch-manifest.json"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q", want)
		}
	}
	for _, banned := range []string{"go build", "OPENAI_API_KEY", "ANTHROPIC_API_KEY", "aws"} {
		if strings.Contains(text, banned) {
			t.Errorf("forbidden %q", banned)
		}
	}
}

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
	prRaw, pr := readWorkflow(t, "scout-guardian.yml")
	trRaw, tr := readWorkflow(t, "scout-guardian-trusted.yml")
	for name, raw := range map[string]string{"scout-guardian.yml": prRaw, "scout-guardian-trusted.yml": trRaw} {
		code := regexp.MustCompile(`(?m)^\s*#.*$`).ReplaceAllString(raw, "")
		if strings.Contains(code, "go build") || strings.Contains(code, "./cmd/scout") {
			t.Errorf("%s builds Scout from the checkout", name)
		}
		// The install and the scanner toolchain come from two scripts of a clone of
		// the release tag outside the workspace, never from the assessed checkout.
		for _, want := range []string{
			`git clone --quiet --depth 1 --branch "$SCOUT_VERSION"`,
			`"https://github.com/${RELEASE_REPO}.git" "$src"`,
			`src="$RUNNER_TEMP/scout-source"`,
			`-f "$src/scripts/install-release.sh" && -f "$src/scripts/ci-toolchain.sh"`,
			"use v0.1.6 or later",
			`bash "$src/scripts/install-release.sh"`,
			`bash "$RUNNER_TEMP/scout-source/scripts/ci-toolchain.sh"`,
			"TOOLCHAIN_DIR: ${{ runner.temp }}/scout-toolchain",
		} {
			if !strings.Contains(code, want) {
				t.Errorf("%s: release-tag install is missing %q", name, want)
			}
		}
		if !regexp.MustCompile(`use v0\.1\.6 or later\.'; exit 3`).MatchString(code) {
			t.Errorf("%s: a tag without the install scripts must exit 3 naming the minimum tag", name)
		}
		for _, banned := range []string{"$GITHUB_WORKSPACE/scripts", "bash scripts/", "./scripts/", "source .tools/activate", "GITHUB_WORKSPACE/.tools", "uses: ./.github/actions", "bootstrap-tools.sh", "curl", "sha256sum", "releases/download", "clawpatch-manifest.json"} {
			if strings.Contains(code, banned) {
				t.Errorf("%s: install logic must live in the release scripts, not here (%q)", name, banned)
			}
		}
		install, toolchain, use := strings.Index(code, `bash "$src/scripts/install-release.sh"`), strings.Index(code, `bash "$RUNNER_TEMP/scout-source/scripts/ci-toolchain.sh"`), strings.Index(code, `"$SCOUT_BIN"`)
		if strings.Count(code, `install-release.sh"`) != 2 || strings.Count(code, `ci-toolchain.sh"`) != 2 || strings.Index(code, "git clone") > install || install > toolchain || toolchain > use {
			t.Errorf("%s: the scripts must be checked and invoked once each, after the clone and before Scout", name)
		}
		for _, want := range []string{`[[ "$SCOUT_VERSION" =~`, `[[ "$RELEASE_REPO" =~`} {
			if !strings.Contains(code, want) {
				t.Errorf("%s: input validation (guarding git clone --branch) must stay in the workflow: missing %q", name, want)
			}
		}
		// The release source is the repository hosting the workflow unless the
		// caller names one; no owner is hard-coded.
		if !strings.Contains(code, "default: ''") || strings.Count(code, "RELEASE_REPO: ${{ inputs.release-repo || job.workflow_repository }}") != 2 || strings.Contains(code, "RELEASE_REPO: ${{ inputs.release-repo }}") {
			t.Errorf("%s: release-repo must default to job.workflow_repository in both release steps", name)
		}
	}
	prCode := regexp.MustCompile(`(?m)^\s*#.*$`).ReplaceAllString(prRaw, "")
	trCode := regexp.MustCompile(`(?m)^\s*#.*$`).ReplaceAllString(trRaw, "")

	// Pull-request workflow: one read-only job, no secrets, no token.
	if len(pr.Jobs) != 1 || len(pr.Permissions) != 1 || pr.Permissions["contents"] != "read" {
		t.Errorf("pull-request workflow must hold exactly one job and contents: read: jobs=%d perms=%v", len(pr.Jobs), pr.Permissions)
	}
	job, ok := pr.Jobs["pull-request"]
	if !ok || job.If != "github.event_name == 'pull_request'" || len(job.Permissions) != 1 || job.Permissions["contents"] != "read" {
		t.Errorf("pull-request job must be read-only: %+v", job)
	}
	for _, banned := range []string{"secrets", "GITHUB_TOKEN", "github.token", "pull-requests", "state publish", "map --", "args=(map"} {
		if strings.Contains(prCode, banned) {
			t.Errorf("pull-request workflow must not reference %q", banned)
		}
	}
	if !strings.Contains(prCode, "BASE_SHA") {
		t.Error("pull-request policy must be read from the base commit")
	}

	// Trusted workflow: one job with write + pull-requests read, the single secret.
	if len(tr.Jobs) != 1 || tr.Permissions["contents"] != "write" || tr.Permissions["pull-requests"] != "read" || len(tr.Permissions) != 2 {
		t.Errorf("trusted workflow permissions: jobs=%d perms=%v", len(tr.Jobs), tr.Permissions)
	}
	trusted, ok := tr.Jobs["trusted"]
	if !ok || trusted.Permissions["contents"] != "write" || trusted.Permissions["pull-requests"] != "read" || len(trusted.Permissions) != 2 {
		t.Errorf("trusted job permissions: %+v", trusted)
	}
	if !strings.Contains(trusted.If, "github.event_name != 'pull_request'") || !strings.Contains(trusted.If, "default_branch") {
		t.Errorf("trusted job is not limited to default-branch non-PR runs: %q", trusted.If)
	}
	if strings.Count(trCode, "GITHUB_TOKEN") != 1 || strings.Count(trCode, "github.token") != 1 || !strings.Contains(trCode, "GITHUB_TOKEN: ${{ github.token }}") || !strings.Contains(trCode, "--fetch-pr-metadata --pr-repo") {
		t.Error("GITHUB_TOKEN must be referenced once, with PR metadata fetching")
	}
	mapAt, assessAt := strings.Index(trCode, "args=(map"), strings.Index(trCode, "args=(assess")
	if mapAt < 0 || assessAt < 0 || mapAt > assessAt || strings.Index(trCode, "state restore") > mapAt || strings.Index(trCode, "state publish") < assessAt {
		t.Error("trusted job order must be state restore, map, assess, state publish")
	}
	for _, want := range []string{"--event push --default-branch", "--event nightly --default-branch", "--refresh"} {
		if !strings.Contains(trCode, want) {
			t.Errorf("trusted map step lacks %q", want)
		}
	}
	for _, gone := range []string{"--base", "BEFORE_SHA", "HEAD_SHA"} {
		if strings.Contains(trCode[mapAt:assessAt], gone) {
			t.Errorf("trusted map step must not use %q (refresh-on-merge was removed)", gone)
		}
	}
	if got := strings.Count(trCode, "secrets."); got != 1 || !strings.Contains(trCode, "secrets.OPENROUTER_API_KEY") {
		t.Errorf("only one OPENROUTER_API_KEY secret reference is allowed (references: %d)", got)
	}
	if !strings.Contains(trRaw, "secrets:\n      OPENROUTER_API_KEY:") {
		t.Error("trusted workflow must declare the OPENROUTER_API_KEY secret")
	}
	if strings.Contains(prRaw, "secrets:") {
		t.Error("pull-request workflow must declare no secrets")
	}
}

func TestNightlyIsACallerOfTheReusableWorkflows(t *testing.T) {
	raw, w := readWorkflow(t, "scout-nightly.yml")
	want := map[string]string{
		"pull-request": "./.github/workflows/scout-guardian.yml",
		"trusted":      "./.github/workflows/scout-guardian-trusted.yml",
	}
	if len(w.Jobs) != len(want) {
		t.Fatalf("unexpected jobs: %v", w.Jobs)
	}
	for name, uses := range want {
		if w.Jobs[name].Uses != uses {
			t.Errorf("job %s calls %q, want %q", name, w.Jobs[name].Uses, uses)
		}
	}
	for _, banned := range []string{"aws-actions", "AWS", "s3", "id-token", "go build", "bootstrap-tools", "OPENAI_API_KEY", "ANTHROPIC_API_KEY", "SCOUT_PROVIDER"} {
		if strings.Contains(raw, banned) {
			t.Errorf("caller still references %q", banned)
		}
	}
	pr, tr := w.Jobs["pull-request"].Permissions, w.Jobs["trusted"].Permissions
	if pr["contents"] != "read" || len(pr) != 1 || tr["contents"] != "write" || tr["pull-requests"] != "read" || len(tr) != 2 {
		t.Errorf("caller permissions: %v / %v", pr, tr)
	}
	if strings.Contains(jobBlock(t, raw, "pull-request"), "secrets") {
		t.Error("the pull-request caller job passes secrets")
	}
	if !strings.Contains(jobBlock(t, raw, "trusted"), "OPENROUTER_API_KEY: ${{ secrets.OPENROUTER_API_KEY }}") {
		t.Error("the trusted caller job must pass the OPENROUTER_API_KEY secret")
	}
	for _, trigger := range []string{"pull_request:", "push:", "schedule:", "workflow_dispatch:", "refresh-map", "authorize-egress"} {
		if !strings.Contains(raw, trigger) {
			t.Errorf("trigger %q missing", trigger)
		}
	}
}

// GitHub validates every job of a called workflow against the calling job's
// permissions, regardless of `if:`; a called job asking for more fails the
// whole run at startup.
func TestCalledWorkflowPermissionsFitCallerCeiling(t *testing.T) {
	level := map[string]int{"": 0, "none": 0, "read": 1, "write": 2}
	_, caller := readWorkflow(t, "scout-nightly.yml")
	for name, cj := range caller.Jobs {
		_, called := readWorkflow(t, filepath.Base(cj.Uses))
		ceiling := cj.Permissions
		if ceiling == nil {
			ceiling = caller.Permissions
		}
		for jn, j := range called.Jobs {
			perms := j.Permissions
			if perms == nil {
				perms = called.Permissions
			}
			for scope, v := range perms {
				if level[v] > level[ceiling[scope]] {
					t.Errorf("caller job %s grants %s: %q but %s job %s requests %q", name, scope, ceiling[scope], cj.Uses, jn, v)
				}
			}
		}
	}
}

func TestInstallScriptsInstallFromVerifiedRelease(t *testing.T) {
	if _, err := os.Stat(filepath.Join("..", "..", ".github", "actions", "scout")); !os.IsNotExist(err) {
		t.Errorf("the composite action must be removed: %v", err)
	}
	read := func(name string) string {
		raw, err := os.ReadFile(filepath.Join("..", "..", "scripts", name))
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	install := read("install-release.sh")
	for _, want := range []string{"releases/download", "die()", "scripts/sha256.sh", "sha256 ", "clawpatch-manifest.json", "$RUNNER_TEMP/scout-release", "SCOUT_BIN=", "CLAWPATCH_BIN="} {
		if !strings.Contains(install, want) {
			t.Errorf("install-release.sh missing %q", want)
		}
	}
	for _, banned := range []string{"go build", "OPENAI_API_KEY", "ANTHROPIC_API_KEY", "aws", "sha256sum"} {
		if strings.Contains(install, banned) {
			t.Errorf("install-release.sh forbidden %q", banned)
		}
	}
	tool := read("ci-toolchain.sh")
	for _, want := range []string{`$(dirname "$0")/..`, `bash "$src/scripts/bootstrap-tools.sh"`, "bubblewrap", "SCOUT_TOOLS_DIR=", "GITHUB_PATH"} {
		if !strings.Contains(tool, want) {
			t.Errorf("ci-toolchain.sh missing %q", want)
		}
	}
	for _, banned := range []string{"GITHUB_WORKSPACE", "go build", "git clone"} {
		if strings.Contains(tool, banned) {
			t.Errorf("ci-toolchain.sh must not reference %q", banned)
		}
	}
}

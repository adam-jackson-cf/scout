package eval

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func validCorpus() Corpus {
	return Corpus{SchemaVersion: CorpusSchemaV1, CorpusVersion: "1", Entries: []CorpusEntry{{
		ID: "e1", Target: testTarget, Language: "go", Family: "ci",
		Input: SupportedInput{Supported: true}, AdjudicationStatus: ReferenceAssessmentAdjudicated,
		Issues: []ReferenceIssue{trueIssue},
	}}}
}

func TestCorpusValidate(t *testing.T) {
	if err := validCorpus().Validate(); err != nil {
		t.Fatalf("a well-formed corpus must validate: %v", err)
	}
	cases := map[string]func(*Corpus){
		"wrong schema":      func(c *Corpus) { c.SchemaVersion = "other" },
		"no corpus version": func(c *Corpus) { c.CorpusVersion = " " },
		"entry without id":  func(c *Corpus) { c.Entries[0].ID = "" },
		"duplicate entry id": func(c *Corpus) {
			second := c.Entries[0]
			second.Language = "rust"
			c.Entries = append(c.Entries, second)
			c.Entries[1].ID = c.Entries[0].ID
		},
		"unpinned target":     func(c *Corpus) { c.Entries[0].Target.Revision = "" },
		"no language":         func(c *Corpus) { c.Entries[0].Language = "" },
		"duplicate selection": func(c *Corpus) { c.Entries = append(c.Entries, c.Entries[0]); c.Entries[1].ID = "e2" },
		"unsupported without reason": func(c *Corpus) {
			c.Entries[0].Input = SupportedInput{}
			c.Entries[0].AdjudicationStatus = ReferenceAssessmentUnsupported
		},
		"unknown assessment status":    func(c *Corpus) { c.Entries[0].AdjudicationStatus = "done" },
		"unsupported status, input ok": func(c *Corpus) { c.Entries[0].AdjudicationStatus = ReferenceAssessmentUnsupported },
		"unsupported input, adjudicated": func(c *Corpus) {
			c.Entries[0].Input = SupportedInput{Reason: "none"}
		},
		"issue without id":     func(c *Corpus) { c.Entries[0].Issues[0].ID = "" },
		"unknown adjudication": func(c *Corpus) { c.Entries[0].Issues[0].Adjudication = "guessed" },
		"true issue without path": func(c *Corpus) {
			c.Entries[0].Issues = []ReferenceIssue{{ID: "x", Adjudication: AdjudicationValidatedTrue, Mechanism: "m", Evidence: "e"}}
		},
		"true issue without evidence": func(c *Corpus) {
			c.Entries[0].Issues = []ReferenceIssue{{ID: "x", Adjudication: AdjudicationValidatedTrue, Mechanism: "m", Path: "p"}}
		},
		"duplicate issue id": func(c *Corpus) { c.Entries[0].Issues = []ReferenceIssue{trueIssue, trueIssue} },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			corpus := validCorpus()
			mutate(&corpus)
			if corpus.Validate() == nil {
				t.Fatal("an invalid corpus must be rejected")
			}
		})
	}
}

func TestLoadCorpusDistinguishesMissingUnreadableAndAvailable(t *testing.T) {
	dir := t.TempDir()
	if r, err := LoadCorpus(""); err != nil || r.Status != CorpusMissing {
		t.Fatalf("empty path: %+v %v", r, err)
	}
	if r, err := LoadCorpus(filepath.Join(dir, "absent.json")); err != nil || r.Status != CorpusMissing {
		t.Fatalf("absent file: %+v %v", r, err)
	}
	write := func(name, content string) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	for name, content := range map[string]string{
		"garbage.json":  "{",
		"unknown.json":  `{"schemaVersion":"scout.eval.corpus/v1","corpusVersion":"1","entries":[],"extra":1}`,
		"trailing.json": `{"schemaVersion":"scout.eval.corpus/v1","corpusVersion":"1","entries":[]} {}`,
		"invalid.json":  `{"schemaVersion":"nope","corpusVersion":"1","entries":[]}`,
	} {
		if r, err := LoadCorpus(write(name, content)); err == nil || r.Status != CorpusUnreadable {
			t.Fatalf("%s: want unreadable with an error, got %+v %v", name, r, err)
		}
	}
	r, err := LoadCorpus(write("ok.json", `{"schemaVersion":"scout.eval.corpus/v1","corpusVersion":"1","entries":[]}`))
	if err != nil || r.Status != CorpusAvailable || r.Corpus == nil {
		t.Fatalf("valid empty corpus: %+v %v", r, err)
	}
	if _, err := LoadCorpus(dir); err == nil || !strings.Contains(err.Error(), "corpus") {
		t.Fatalf("a directory is unreadable, not missing: %v", err)
	}
}

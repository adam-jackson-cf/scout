// Command jev scores the residual change clusters of an X2d round (X2d PREREGISTRATION sections 6 and 7) with Jev through Scout's own client, internal/decisions
// (OpenRouter alpha/decisions route, resolved-model pinning, failure kinds, billed usage). It is a thin driver, in the style of X2c's jev driver: the HTTP client, the
// request builder, the pin check and the answer validation are the product code; this file builds the study's questions from the frozen `questions/*@v1.json` files and the
// frozen catalogue, renders the state of every request, appends one row per request to an append-only output and stops on a cumulative billed-cost cap.
//
//	OPENROUTER_API_KEY=... jev -in units.jsonl -catalogue catalogue.json -out scores.jsonl [-arms sem,sib,title,fan] [-flip-all] [-settings settings.json]
//	    [-questions-src DIR] [-questions-dir DIR] [-endpoint URL] [-model M] [-workers 8] [-max-cost-usd X] [-timeout 60s] [-sample-files 5]
//	jev -dump-state -in units.jsonl -catalogue catalogue.json -out states.jsonl [-arms sem,sib,title] [-settings settings.json] [-sample-files 5]
//	    (no key, no request: one row per (unit, enabled variant) with the exact state the driver sends, see below)
//	jev -catalogue catalogue.json -print-versions       (no key, no input: one `<question>/<variant>\t<version key>` line per question and enabled variant)
//
// Input rows are units.jsonl rows (this driver uses `case_id`, `cluster_id`, `state` {unit, pr, diff_summary} and `siblings`). For EVERY unit and every enabled variant
// (-arms: sem = V-sem, sib = V-sib adding `state.siblings`, title = V-title with only `state.pr.title`) it asks merge.covered (Noul, catalogue in the state) and
// merge.which_area (Choice over the whole catalogue: option ids are the area ids, permuted order, unit id = case id; a second permutation, unit id `<case id>#2`, for the
// clusters selected by sha256("x2d-flip" + cluster id) mod 5 == 0 or with -flip-all); with `fan` in -arms also one merge.area_covers (Noul, V-sem only, `state.area`)
// per catalogue area. The Choice is asked regardless of the Noul's answer (composition and tuning decide later). Every state string passes decisions.Redact.
//
// Output: a header row then one `answer` or `failure` row per request (req_id = "<case_id>|<question>|<variant>|<perm>|<area_id or ->"). Re-running with the same input,
// catalogue, arms, settings and model resumes: requests with an answer row are skipped, failed requests are retried and append a new row; a header naming anything else is
// refused. The key is read from the environment only and is never printed or written. Exit codes: 2 usage/refusal, 3 failures or unknown billed cost, 4 cost cap reached.
//
// -settings FILE (JSON {"sample_files_per_area":N (3..8), "v_sib":bool, "include_pr_title":bool (default true; false removes state.pr.title from V-sem and V-sib; the V-title arm
// is always on), "field_order":[...], "order_rule":"permutation:unit_id"|"fixed", ...}) overrides -sample-files and removes the V-sib variant when v_sib is false (unknown keys are
// ignored). Sample files: every area shows sample_files[:n]; the driver refuses n > the longest sample_files of the catalogue (a catalogue whose areas all have an empty
// sample_files list shows none, as the N1 control's catalogue).
//
// -dump-state: no key, no request. Writes (truncating) -out JSONL, one object per unit and enabled variant, in unit then variant order:
// {"case_id","variant","noul_state" (the merge.covered state: unit fields in field order, the catalogue with sample_files[:n]),"choice_state" (the merge.which_area state: no catalogue),
// "choice_options" (the option texts of merge.which_area in catalogue order, as rendered into the question criteria),"choice_criteria" ({area id: option text})}. The states are the
// bytes stateFor produces for the real requests (redacted, field order and include_pr_title applied).
package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"scout/internal/decisions"
	"scout/internal/policy"
)

const (
	qCovered   = "merge.covered"
	qWhichArea = "merge.which_area"
	qAreaCover = "merge.area_covers"
)

// variant of one arm key.
var variantOf = map[string]string{"sem": "V-sem", "sib": "V-sib", "title": "V-title"}

// ---- frozen question files

type variantDef struct {
	InstructionsDelta string   `json:"instructions_delta"`
	StateShapeDelta   []string `json:"state_shape_delta"`
}

type questionFile struct {
	QuestionID       string                    `json:"question_id"`
	Kind             string                    `json:"kind"`
	Instructions     string                    `json:"instructions"`
	Criteria         json.RawMessage           `json:"criteria"`
	CriteriaTemplate string                    `json:"criteria_template"`
	StateShape       []string                  `json:"state_shape"`
	OrderRule        decisions.OptionOrderRule `json:"option_order_rule"`
	Positive         *string                   `json:"positive"`
	Variants         map[string]variantDef     `json:"variants"`
}

func loadQuestionFile(dir, id string) (*questionFile, error) {
	raw, err := os.ReadFile(filepath.Join(dir, id+"@v1.json"))
	if err != nil {
		return nil, err
	}
	var q questionFile
	if err := json.Unmarshal(raw, &q); err != nil {
		return nil, fmt.Errorf("%s@v1.json: %w", id, err)
	}
	if q.QuestionID != id {
		return nil, fmt.Errorf("%s@v1.json names question %q", id, q.QuestionID)
	}
	return &q, nil
}

// applyShape applies a variant's state_shape_delta ("+path" adds, "-path" removes) to a base shape.
func applyShape(base, delta []string) ([]string, error) {
	set := map[string]bool{}
	for _, s := range base {
		set[s] = true
	}
	for _, d := range delta {
		switch {
		case strings.HasPrefix(d, "+"):
			set[d[1:]] = true
		case strings.HasPrefix(d, "-"):
			if !set[d[1:]] {
				return nil, fmt.Errorf("variant removes %q, which is not in the state shape", d[1:])
			}
			delete(set, d[1:])
		default:
			return nil, fmt.Errorf("state shape delta %q has no + or - prefix", d)
		}
	}
	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	sort.Strings(out)
	return out, nil
}

// ---- catalogue

type area struct {
	AreaID      string   `json:"area_id"`
	Scope       string   `json:"scope"`
	Title       string   `json:"title"`
	Summary     string   `json:"summary"`
	SampleFiles []string `json:"sample_files"`
	Files       []string `json:"files"`
}

type catalogueFile struct {
	Schema string `json:"schema"`
	Areas  []area `json:"areas"`
}

func loadCatalogue(path string) (*catalogueFile, string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	var c catalogueFile
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, "", fmt.Errorf("%s: %w", path, err)
	}
	if len(c.Areas) < 2 || len(c.Areas) > 255 {
		return nil, "", fmt.Errorf("%s: %d areas; a Choice needs 2 to 255 options", path, len(c.Areas))
	}
	seen := map[string]bool{}
	for _, a := range c.Areas {
		if a.AreaID == "" || seen[a.AreaID] {
			return nil, "", fmt.Errorf("%s: empty or duplicate area_id %q", path, a.AreaID)
		}
		seen[a.AreaID] = true
	}
	sum := sha256.Sum256(raw)
	return &c, hex.EncodeToString(sum[:]), nil
}

// sampleFiles: the area's sample_files[:n] (the driver refuses n beyond the longest list of a catalogue, see checkSampleFiles).
func (a area) sampleFiles(n int) []string {
	if n > len(a.SampleFiles) {
		n = len(a.SampleFiles)
	}
	return a.SampleFiles[:n]
}

// checkSampleFiles refuses n > the longest sample_files list of the catalogue; a catalogue with no sample files at all (the N1 control) is accepted and shows none.
func checkSampleFiles(c *catalogueFile, n int) error {
	longest := 0
	for _, a := range c.Areas {
		if len(a.SampleFiles) > longest {
			longest = len(a.SampleFiles)
		}
	}
	if longest > 0 && n > longest {
		return fmt.Errorf("sample_files_per_area %d exceeds the catalogue's longest sample_files list (%d): the catalogue cannot supply them", n, longest)
	}
	return nil
}

// ---- redaction (identical in onboard/main.go)

func redactValue(v any) any {
	switch t := v.(type) {
	case string:
		return decisions.Redact(t)
	case []any:
		out := make([]any, len(t))
		for i, e := range t {
			out[i] = redactValue(e)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, e := range t {
			out[k] = redactValue(e)
		}
		return out
	default:
		return v
	}
}

func redactAll(in []string) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = decisions.Redact(s)
	}
	return out
}

// orderedState marshals a JSON object whose keys follow `keys` (the field order setting), then the remaining keys sorted.
type orderedState struct {
	keys []string
	vals map[string]any
}

func (o orderedState) MarshalJSON() ([]byte, error) {
	order := make([]string, 0, len(o.vals))
	used := map[string]bool{}
	for _, k := range o.keys {
		if _, ok := o.vals[k]; ok && !used[k] {
			order = append(order, k)
			used[k] = true
		}
	}
	var rest []string
	for k := range o.vals {
		if !used[k] {
			rest = append(rest, k)
		}
	}
	sort.Strings(rest)
	order = append(order, rest...)
	var b bytes.Buffer
	b.WriteByte('{')
	for i, k := range order {
		if i > 0 {
			b.WriteByte(',')
		}
		kb, _ := json.Marshal(k)
		vb, err := json.Marshal(o.vals[k])
		if err != nil {
			return nil, err
		}
		b.Write(kb)
		b.WriteByte(':')
		b.Write(vb)
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

// ---- units

type unit struct {
	CaseID    string         `json:"case_id"`
	ClusterID string         `json:"cluster_id"`
	State     map[string]any `json:"state"`
	Siblings  []any          `json:"siblings"`
}

func readUnits(path string) ([]unit, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()
	h := sha256.New()
	var units []unit
	sc := bufio.NewScanner(io.TeeReader(f, h))
	sc.Buffer(make([]byte, 1<<20), 64<<20)
	seen := map[string]bool{}
	for line := 1; sc.Scan(); line++ {
		if len(sc.Bytes()) == 0 {
			continue
		}
		dec := json.NewDecoder(bytes.NewReader(sc.Bytes()))
		dec.UseNumber()
		var u unit
		if err := dec.Decode(&u); err != nil || u.CaseID == "" || u.ClusterID == "" {
			return nil, "", fmt.Errorf("%s:%d: not a units row with case_id and cluster_id", path, line)
		}
		for _, k := range []string{"unit", "pr", "diff_summary"} {
			if _, ok := u.State[k].(map[string]any); !ok {
				return nil, "", fmt.Errorf("%s:%d: state.%s is missing", path, line, k)
			}
		}
		if _, ok := u.State["pr"].(map[string]any)["title"].(string); !ok {
			return nil, "", fmt.Errorf("%s:%d: state.pr.title is missing", path, line)
		}
		if seen[u.CaseID] {
			return nil, "", fmt.Errorf("%s:%d: duplicate case_id %s", path, line, u.CaseID)
		}
		seen[u.CaseID] = true
		units = append(units, u)
	}
	return units, hex.EncodeToString(h.Sum(nil)), sc.Err()
}

// flipSelected is x2d.flip_selected: sha256("x2d-flip" + cluster id) read as an integer, mod 5 == 0.
func flipSelected(clusterID string) bool {
	sum := sha256.Sum256([]byte("x2d-flip" + clusterID))
	n := new(big.Int).SetBytes(sum[:])
	return new(big.Int).Mod(n, big.NewInt(5)).Sign() == 0
}

// ---- settings

type settings struct {
	SampleFilesPerArea *int     `json:"sample_files_per_area"`
	VSib               *bool    `json:"v_sib"`
	IncludePRTitle     *bool    `json:"include_pr_title"`
	FieldOrder         []string `json:"field_order"`
	OrderRule          *string  `json:"order_rule"`
}

func loadSettings(path string) (settings, error) {
	var s settings
	raw, err := os.ReadFile(path)
	if err != nil {
		return s, err
	}
	if err := json.Unmarshal(raw, &s); err != nil {
		return s, fmt.Errorf("%s: %w", path, err)
	}
	if s.SampleFilesPerArea != nil && (*s.SampleFilesPerArea < 3 || *s.SampleFilesPerArea > 8) {
		return s, fmt.Errorf("%s: sample_files_per_area %d is outside the menu range 3..8", path, *s.SampleFilesPerArea)
	}
	if s.OrderRule != nil && *s.OrderRule != decisions.OrderPermutation && *s.OrderRule != decisions.OrderFixed {
		return s, fmt.Errorf("%s: order_rule %q is neither %s nor %s", path, *s.OrderRule, decisions.OrderPermutation, decisions.OrderFixed)
	}
	return s, nil
}

// ---- the questions as sent

type built struct {
	Name     string // merge.covered ...
	Variant  string // V-sem ...
	Question decisions.Question
	Version  string
}

func (b built) key() string { return b.Name + "/" + b.Variant }

type questionSet struct {
	byKey map[string]built
	order []string
}

func (qs *questionSet) add(b built) {
	qs.byKey[b.key()] = b
	qs.order = append(qs.order, b.key())
}

func buildQuestions(dir, model string, cat *catalogueFile, nFiles int, orderRule string, variants []string, fan bool) (*questionSet, error) {
	qs := &questionSet{byKey: map[string]built{}}
	criteriaOf := func(f *questionFile) (map[string]string, error) {
		var c map[string]string
		if err := json.Unmarshal(f.Criteria, &c); err != nil {
			return nil, fmt.Errorf("%s: criteria is not a map: %w", f.QuestionID, err)
		}
		return c, nil
	}
	mk := func(f *questionFile, variant string, kind decisions.Kind) (decisions.Question, error) {
		v, ok := f.Variants[variant]
		if !ok {
			return decisions.Question{}, fmt.Errorf("%s has no variant %s", f.QuestionID, variant)
		}
		shape, err := applyShape(f.StateShape, v.StateShapeDelta)
		if err != nil {
			return decisions.Question{}, fmt.Errorf("%s/%s: %w", f.QuestionID, variant, err)
		}
		if string(kind) != f.Kind {
			return decisions.Question{}, fmt.Errorf("%s is a %s question, expected %s", f.QuestionID, f.Kind, kind)
		}
		q := decisions.Question{ID: f.QuestionID, Kind: kind, Instructions: f.Instructions + v.InstructionsDelta, StateShape: shape, Order: f.OrderRule}
		if f.Positive != nil {
			q.Positive = *f.Positive
		}
		return q, nil
	}
	for _, name := range []string{qCovered, qWhichArea, qAreaCover} {
		if name == qAreaCover && !fan {
			continue
		}
		f, err := loadQuestionFile(dir, name)
		if err != nil {
			return nil, err
		}
		vs := variants
		if name == qAreaCover {
			vs = []string{"V-sem"}
		}
		for _, variant := range vs {
			switch name {
			case qCovered, qAreaCover:
				q, err := mk(f, variant, decisions.KindNoul)
				if err != nil {
					return nil, err
				}
				if q.Criteria, err = criteriaOf(f); err != nil {
					return nil, err
				}
				qs.add(built{Name: name, Variant: variant, Question: q})
			case qWhichArea:
				if !strings.HasPrefix(f.CriteriaTemplate, "{title}: {summary} Sample files: ") {
					return nil, fmt.Errorf("%s: criteria_template %q is not the frozen template", name, f.CriteriaTemplate)
				}
				q, err := mk(f, variant, decisions.KindChoice)
				if err != nil {
					return nil, err
				}
				q.Criteria = map[string]string{}
				ids := make([]string, 0, len(cat.Areas))
				for _, a := range cat.Areas {
					q.Criteria[a.AreaID] = decisions.Redact(fmt.Sprintf("%s: %s Sample files: %s", a.Title, a.Summary, strings.Join(a.sampleFiles(nFiles), ", ")))
					ids = append(ids, a.AreaID)
				}
				if orderRule == decisions.OrderFixed {
					q.Order = decisions.OptionOrderRule{Rule: decisions.OrderFixed, Options: ids}
				} else {
					q.Order = decisions.OptionOrderRule{Rule: decisions.OrderPermutation}
				}
				qs.add(built{Name: name, Variant: variant, Question: q})
			}
		}
	}
	for _, k := range qs.order {
		b := qs.byKey[k]
		if err := b.Question.Validate(); err != nil {
			return nil, fmt.Errorf("%s: %w", k, err)
		}
		b.Version = b.Question.Version(model)
		qs.byKey[k] = b
	}
	return qs, nil
}

// ---- states

type stateBuilder struct {
	cat            *catalogueFile
	nFiles         int
	fieldOrder     []string
	includePRTitle bool
}

func (sb stateBuilder) catalogueState() []any {
	out := make([]any, 0, len(sb.cat.Areas))
	for _, a := range sb.cat.Areas {
		out = append(out, map[string]any{"id": a.AreaID, "title": a.Title, "summary": a.Summary, "sample_files": toAny(a.sampleFiles(sb.nFiles))})
	}
	return out
}

func toAny(in []string) []any {
	out := make([]any, len(in))
	for i, s := range in {
		out[i] = s
	}
	return out
}

// stateFor renders the state of one request, every string redacted. variant V-sem/V-sib/V-title; withCatalogue only for the covered Noul; areaID non-empty only for the fan-out.
func (sb stateBuilder) stateFor(u unit, variant string, withCatalogue bool, areaID string) (json.RawMessage, error) {
	vals := map[string]any{}
	switch variant {
	case "V-sem", "V-sib":
		vals["unit"], vals["diff_summary"] = u.State["unit"], u.State["diff_summary"]
		pr := map[string]any{}
		for k, v := range u.State["pr"].(map[string]any) {
			if k == "title" && !sb.includePRTitle {
				continue
			}
			pr[k] = v
		}
		vals["pr"] = pr
		if variant == "V-sib" {
			sib := u.Siblings
			if sib == nil {
				sib = []any{}
			}
			vals["siblings"] = sib
		}
	case "V-title":
		vals["pr"] = map[string]any{"title": u.State["pr"].(map[string]any)["title"]}
	default:
		return nil, fmt.Errorf("unknown variant %s", variant)
	}
	if withCatalogue {
		vals["catalogue"] = sb.catalogueState()
	}
	if areaID != "" {
		found := false
		for _, a := range sb.cat.Areas {
			if a.AreaID == areaID {
				vals["area"] = map[string]any{"id": a.AreaID, "title": a.Title, "summary": a.Summary, "sample_files": toAny(a.sampleFiles(sb.nFiles))}
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf("area %s is not in the catalogue", areaID)
		}
	}
	for k, v := range vals {
		vals[k] = redactValue(v)
	}
	return json.Marshal(orderedState{keys: sb.fieldOrder, vals: vals})
}

// ---- output rows

type header struct {
	Type         string            `json:"type"`
	Model        string            `json:"model"`
	CatalogueSHA string            `json:"catalogue_sha256"`
	InputSHA256  string            `json:"input_sha256"`
	Arms         []string          `json:"arms"`
	Versions     map[string]string `json:"versions"`
	Endpoint     string            `json:"endpoint"`
	Started      string            `json:"started_utc"`
	SampleFiles  int               `json:"sample_files"`
	IncludeTitle bool              `json:"include_pr_title"`
	FlipAll      bool              `json:"flip_all"`
	OrderRule    string            `json:"order_rule"`
	FieldOrder   []string          `json:"field_order"`
	QuestionsSHA map[string]string `json:"questions_src_sha256"`
	InputUnits   int               `json:"input_units"`
}

type answerRow struct {
	Type            string             `json:"type"`
	ReqID           string             `json:"req_id"`
	CaseID          string             `json:"case_id"`
	Question        string             `json:"question"`
	Variant         string             `json:"variant"`
	Perm            int                `json:"perm"`
	AreaID          *string            `json:"area_id"`
	PTrue           *float64           `json:"p_true,omitempty"`
	Choice          string             `json:"choice,omitempty"`
	PChoice         *float64           `json:"p_choice,omitempty"`
	Confidence      *float64           `json:"confidence"`
	Probabilities   map[string]float64 `json:"probabilities,omitempty"`
	CostUSD         float64            `json:"cost_usd"`
	CostKnown       bool               `json:"cost_known"`
	Model           string             `json:"model"`
	QuestionVersion string             `json:"question_version"`
	ResponseID      string             `json:"response_id"`
	RequestSHA256   string             `json:"request_sha256"`
	Usage           json.RawMessage    `json:"usage"`
	LatencyMS       int64              `json:"latency_ms"`
	Attempts        int                `json:"attempts"`
	Timestamp       string             `json:"timestamp"`
}

type failureRow struct {
	Type      string  `json:"type"`
	ReqID     string  `json:"req_id"`
	CaseID    string  `json:"case_id"`
	Question  string  `json:"question"`
	Variant   string  `json:"variant"`
	Perm      int     `json:"perm"`
	AreaID    *string `json:"area_id"`
	Kind      string  `json:"kind"`
	Status    int     `json:"status"`
	Reason    string  `json:"reason"`
	Model     string  `json:"model"`
	CostUSD   float64 `json:"cost_usd"`
	CostKnown bool    `json:"cost_known"`
	LatencyMS int64   `json:"latency_ms"`
	Timestamp string  `json:"timestamp"`
}

// existing reads a previous output file: its header, the request ids that already have an answer row, and the billed cost already spent (the sum of cost_usd over answer and
// failure rows that report a cost), so the cost stop is cumulative over the append-only output and not per invocation.
func existing(path string) (*header, map[string]bool, float64, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, map[string]bool{}, 0, nil
	}
	if err != nil {
		return nil, nil, 0, err
	}
	defer f.Close()
	var hd *header
	done := map[string]bool{}
	spent := 0.0
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 64<<20)
	for line := 1; sc.Scan(); line++ {
		var probe struct {
			Type      string  `json:"type"`
			ReqID     string  `json:"req_id"`
			CostUSD   float64 `json:"cost_usd"`
			CostKnown bool    `json:"cost_known"`
		}
		if err := json.Unmarshal(sc.Bytes(), &probe); err != nil {
			return nil, nil, 0, fmt.Errorf("%s:%d: malformed line (a killed run left a partial line: truncate it)", path, line)
		}
		switch probe.Type {
		case "header":
			var h header
			if err := json.Unmarshal(sc.Bytes(), &h); err != nil {
				return nil, nil, 0, err
			}
			hd = &h
		case "answer":
			done[probe.ReqID] = true
			if probe.CostKnown {
				spent += probe.CostUSD
			}
		case "failure":
			if probe.CostKnown {
				spent += probe.CostUSD
			}
		}
	}
	return hd, done, spent, sc.Err()
}

// choiceComplete: the Choice contract returns a probability for every option and the probabilities sum to 1; the product client checks only the supplied values, so the
// driver requires every option and a sum within 1e-6 of 1 (PREREGISTRATION section 6).
func choiceComplete(probs map[string]float64, q decisions.Question) bool {
	sum := 0.0
	for _, p := range probs {
		sum += p
	}
	return len(probs) == len(q.Criteria) && math.Abs(sum-1) <= 1e-6
}

// confidenceOf reads the route's `confidence` of an answer (nil when it is absent).
func confidenceOf(raw json.RawMessage) *float64 {
	var a struct {
		Confidence *float64 `json:"confidence"`
	}
	if json.Unmarshal(raw, &a) != nil || a.Confidence == nil || math.IsNaN(*a.Confidence) || *a.Confidence < 0 || *a.Confidence > 1 {
		return nil
	}
	return a.Confidence
}

type job struct {
	u      unit
	b      built
	perm   int
	areaID string
	reqID  string
	unitID string
}

func reqID(caseID, question, variant string, perm int, areaID string) string {
	a := "-"
	if areaID != "" {
		a = areaID
	}
	return fmt.Sprintf("%s|%s|%s|%d|%s", caseID, question, variant, perm, a)
}

func writeQuestionFile(dir, model string, b built) error {
	q := b.Question
	raw, _ := json.MarshalIndent(map[string]any{"question_id": q.ID, "variant": b.Variant, "version": b.Version, "model": model, "kind": q.Kind, "instructions": q.Instructions,
		"criteria": q.Criteria, "state_shape": q.StateShape, "option_order_rule": q.Order, "positive": q.Positive}, "", " ")
	return os.WriteFile(filepath.Join(dir, b.Version+".json"), raw, 0o644)
}

func fileSHA(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func die(code int, format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(code)
}

// dumpStates writes the -dump-state rows.
func dumpStates(out string, units []unit, cat *catalogueFile, qs *questionSet, variants []string, sb stateBuilder) error {
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, u := range units {
		for _, variant := range variants {
			noul, err := sb.stateFor(u, variant, true, "")
			if err != nil {
				return err
			}
			choice, err := sb.stateFor(u, variant, false, "")
			if err != nil {
				return err
			}
			crit := qs.byKey[qWhichArea+"/"+variant].Question.Criteria
			opts := make([]string, 0, len(cat.Areas))
			for _, a := range cat.Areas {
				opts = append(opts, crit[a.AreaID])
			}
			noulQ, choiceQ := qs.byKey[qCovered+"/"+variant].Question, qs.byKey[qWhichArea+"/"+variant].Question
			// R2-17: the instructions and criteria exactly as the driver sends them (base + the variant's instructions_delta), so a labeller sees the same question.
			if err := enc.Encode(map[string]any{"case_id": u.CaseID, "variant": variant, "noul_state": noul, "choice_state": choice, "choice_options": opts, "choice_criteria": crit,
				"noul_instructions": noulQ.Instructions, "noul_criteria": noulQ.Criteria, "choice_instructions": choiceQ.Instructions}); err != nil {
				return err
			}
		}
	}
	return nil
}

func main() {
	in := flag.String("in", "", "units.jsonl")
	catPath := flag.String("catalogue", "", "frozen catalogue.json")
	out := flag.String("out", "", "output JSONL (append-only; resumable)")
	arms := flag.String("arms", "sem,sib,title", "comma-separated arms: sem (V-sem), sib (V-sib), title (V-title), fan (merge.area_covers per area)")
	flipAll := flag.Bool("flip-all", false, "ask merge.which_area under the second permutation for every cluster (default: the sha256('x2d-flip' + cluster id) mod 5 == 0 clusters)")
	qsrc := flag.String("questions-src", "evaluation/x2d-merge-association-pilot/questions", "directory of the frozen questions/*@v1.json")
	qdir := flag.String("questions-dir", "", "write each question as sent to <dir>/<id>@<version>.json")
	settingsPath := flag.String("settings", "", "round settings JSON (overrides -sample-files; removes disabled variants)")
	endpoint := flag.String("endpoint", decisions.DefaultEndpoint, "decisions route URL (validation points it at a local stub)")
	model := flag.String("model", policy.DefaultModel, "pinned Jev model id (aliases and undated release lines are refused)")
	workers := flag.Int("workers", 8, "concurrent requests")
	maxCost := flag.Float64("max-cost-usd", 1, "stop when the cumulative billed cost of the output file (answer and failure rows, earlier runs included) exceeds this many USD")
	timeout := flag.Duration("timeout", 60*time.Second, "per-request HTTP timeout")
	sampleFiles := flag.Int("sample-files", 5, "sample files per area shown to Jev")
	dumpState := flag.Bool("dump-state", false, "write the exact request states of every unit and enabled variant to -out (JSONL, truncated) and exit; no key, no request")
	printVersions := flag.Bool("print-versions", false, "print the question version key of every question and enabled variant for -model and -catalogue, then exit (no key, no input)")
	flag.Parse()

	if *catPath == "" {
		die(2, "usage: jev -in units.jsonl -catalogue catalogue.json -out scores.jsonl [-arms sem,sib,title,fan] [-flip-all] [-settings FILE] [-questions-src DIR] [-questions-dir DIR] [-endpoint URL] [-workers N] [-max-cost-usd X]")
	}
	fieldOrder := []string{"unit", "pr", "diff_summary", "catalogue"}
	includePRTitle := true
	orderRule := decisions.OrderPermutation
	armSet := map[string]bool{}
	for _, a := range strings.Split(*arms, ",") {
		a = strings.TrimSpace(a)
		if a == "" {
			continue
		}
		if a != "fan" && variantOf[a] == "" {
			die(2, "unknown arm %q (sem, sib, title, fan)", a)
		}
		armSet[a] = true
	}
	if *settingsPath != "" {
		s, err := loadSettings(*settingsPath)
		if err != nil {
			die(2, "%v", err)
		}
		if s.SampleFilesPerArea != nil {
			*sampleFiles = *s.SampleFilesPerArea
		}
		if s.VSib != nil && !*s.VSib {
			delete(armSet, "sib")
		}
		if s.IncludePRTitle != nil {
			includePRTitle = *s.IncludePRTitle
		}
		if len(s.FieldOrder) > 0 {
			fieldOrder = s.FieldOrder
		}
		if s.OrderRule != nil {
			orderRule = *s.OrderRule
		}
	}
	if *sampleFiles < 1 {
		die(2, "-sample-files must be >= 1")
	}
	var armList, variants []string
	for _, a := range []string{"sem", "sib", "title", "fan"} {
		if armSet[a] {
			armList = append(armList, a)
			if a != "fan" {
				variants = append(variants, variantOf[a])
			}
		}
	}
	if len(armList) == 0 {
		die(2, "no arm enabled")
	}
	if policy.IsModelAlias(*model) {
		die(2, "refusing: %q is an alias or an undated release line, not a pinned build", *model)
	}
	cat, catDigest, err := loadCatalogue(*catPath)
	if err != nil {
		die(2, "%v", err)
	}
	if err := checkSampleFiles(cat, *sampleFiles); err != nil {
		die(2, "%v", err)
	}
	qs, err := buildQuestions(*qsrc, *model, cat, *sampleFiles, orderRule, variants, armSet["fan"])
	if err != nil {
		die(2, "question invalid: %v", err)
	}
	versions := map[string]string{}
	for _, k := range qs.order {
		versions[k] = qs.byKey[k].Version
	}
	if *printVersions {
		keys := append([]string(nil), qs.order...)
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("%s\t%s\n", k, versions[k])
		}
		return
	}
	if *dumpState {
		if *in == "" || *out == "" {
			die(2, "usage: jev -dump-state -in units.jsonl -catalogue catalogue.json -out states.jsonl [-arms ...] [-settings FILE]")
		}
		units, _, err := readUnits(*in)
		if err != nil {
			die(2, "%v", err)
		}
		if err := dumpStates(*out, units, cat, qs, variants, stateBuilder{cat: cat, nFiles: *sampleFiles, fieldOrder: fieldOrder, includePRTitle: includePRTitle}); err != nil {
			die(2, "%v", err)
		}
		return
	}
	if *in == "" || *out == "" {
		die(2, "usage: jev -in units.jsonl -catalogue catalogue.json -out scores.jsonl [...] (see the package comment)")
	}
	key := os.Getenv("OPENROUTER_API_KEY")
	if key == "" {
		die(2, "OPENROUTER_API_KEY is not set (run through `infisical run ... --`; the key is never printed)")
	}
	units, digest, err := readUnits(*in)
	if err != nil {
		die(2, "%v", err)
	}
	qsha := map[string]string{}
	for _, name := range []string{qCovered, qWhichArea, qAreaCover} {
		if name == qAreaCover && !armSet["fan"] {
			continue
		}
		if qsha[name], err = fileSHA(filepath.Join(*qsrc, name+"@v1.json")); err != nil {
			die(2, "%v", err)
		}
	}
	want := header{Type: "header", Model: *model, CatalogueSHA: catDigest, InputSHA256: digest, Arms: armList, Versions: versions, Endpoint: *endpoint,
		Started: time.Now().UTC().Format(time.RFC3339), SampleFiles: *sampleFiles, IncludeTitle: includePRTitle, FlipAll: *flipAll, OrderRule: orderRule, FieldOrder: fieldOrder, QuestionsSHA: qsha, InputUnits: len(units)}
	prev, done, spent, err := existing(*out)
	if err != nil {
		die(2, "%v", err)
	}
	if prev != nil {
		same := prev.Model == want.Model && prev.CatalogueSHA == want.CatalogueSHA && prev.InputSHA256 == want.InputSHA256 && prev.FlipAll == want.FlipAll &&
			prev.SampleFiles == want.SampleFiles && prev.IncludeTitle == want.IncludeTitle && prev.OrderRule == want.OrderRule && strings.Join(prev.Arms, ",") == strings.Join(want.Arms, ",") &&
			strings.Join(prev.FieldOrder, ",") == strings.Join(want.FieldOrder, ",") && len(prev.Versions) == len(want.Versions)
		for k, v := range want.Versions {
			if prev.Versions[k] != v {
				same = false
			}
		}
		if !same {
			die(2, "refusing to resume: the output header names another input, catalogue, model, arms, settings or question version")
		}
	}
	if prev == nil && len(done) > 0 {
		die(2, "refusing: output has answer rows but no header")
	}
	if spent > *maxCost {
		die(2, "refusing: budget already exhausted (billed cost %.6f USD in the output exceeds -max-cost-usd %.4f)", spent, *maxCost)
	}
	sb := stateBuilder{cat: cat, nFiles: *sampleFiles, fieldOrder: fieldOrder, includePRTitle: includePRTitle}

	// the full request list, in a deterministic order
	var jobs []job
	for _, u := range units {
		flip := *flipAll || flipSelected(u.ClusterID)
		for _, variant := range variants {
			jobs = append(jobs, job{u: u, b: qs.byKey[qCovered+"/"+variant], perm: 1, reqID: reqID(u.CaseID, qCovered, variant, 1, ""), unitID: u.CaseID})
			jobs = append(jobs, job{u: u, b: qs.byKey[qWhichArea+"/"+variant], perm: 1, reqID: reqID(u.CaseID, qWhichArea, variant, 1, ""), unitID: u.CaseID})
			if flip {
				jobs = append(jobs, job{u: u, b: qs.byKey[qWhichArea+"/"+variant], perm: 2, reqID: reqID(u.CaseID, qWhichArea, variant, 2, ""), unitID: u.CaseID + "#2"})
			}
		}
		if armSet["fan"] {
			for _, a := range cat.Areas {
				jobs = append(jobs, job{u: u, b: qs.byKey[qAreaCover+"/V-sem"], perm: 1, areaID: a.AreaID, reqID: reqID(u.CaseID, qAreaCover, "V-sem", 1, a.AreaID), unitID: u.CaseID})
			}
		}
	}

	f, err := os.OpenFile(*out, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		die(2, "%v", err)
	}
	defer f.Close()
	var mu sync.Mutex
	enc := json.NewEncoder(f)
	write := func(v any) {
		mu.Lock()
		defer mu.Unlock()
		if err := enc.Encode(v); err != nil {
			fmt.Fprintln(os.Stderr, "write failed:", err)
			os.Exit(2)
		}
	}
	if prev == nil {
		write(want)
	}
	if *qdir != "" {
		if err := os.MkdirAll(*qdir, 0o755); err != nil {
			die(2, "%v", err)
		}
		for _, k := range qs.order {
			if err := writeQuestionFile(*qdir, *model, qs.byKey[k]); err != nil {
				die(2, "%v", err)
			}
		}
	}
	client := &decisions.Client{Endpoint: *endpoint, Key: key, HTTP: &http.Client{Timeout: *timeout}}
	queue := make(chan job)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cost := spent // the cap is cumulative over the append-only output (spent = billed cost of every earlier row)
	var failures, unknownCost, answered, stopped int
	var wg sync.WaitGroup
	for range *workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range queue {
				if ctx.Err() != nil {
					continue
				}
				var aid *string
				if j.areaID != "" {
					a := j.areaID
					aid = &a
				}
				state, serr := sb.stateFor(j.u, j.b.Variant, j.b.Name == qCovered, j.areaID)
				var reply decisions.Reply
				err := serr
				if err == nil {
					reply, err = client.Ask(ctx, []decisions.Question{j.b.Question}, *model, j.unitID, state)
				}
				now := time.Now().UTC().Format(time.RFC3339)
				mu.Lock()
				if reply.CostKnown {
					cost += reply.CostUSD
				} else if reply.Model != "" || err == nil {
					unknownCost++
				}
				over := cost > *maxCost
				mu.Unlock()
				fail := func(kind string, status int, reason string) {
					write(failureRow{Type: "failure", ReqID: j.reqID, CaseID: j.u.CaseID, Question: j.b.Name, Variant: j.b.Variant, Perm: j.perm, AreaID: aid, Kind: kind, Status: status,
						Reason: reason, Model: reply.Model, CostUSD: reply.CostUSD, CostKnown: reply.CostKnown, LatencyMS: reply.Latency.Milliseconds(), Timestamp: now})
					mu.Lock()
					failures++
					mu.Unlock()
				}
				row := answerRow{Type: "answer", ReqID: j.reqID, CaseID: j.u.CaseID, Question: j.b.Name, Variant: j.b.Variant, Perm: j.perm, AreaID: aid, Model: reply.Model,
					ResponseID: reply.ID, QuestionVersion: j.b.Version, CostUSD: reply.CostUSD, CostKnown: reply.CostKnown, Usage: reply.Usage, LatencyMS: reply.Latency.Milliseconds(),
					Attempts: reply.Attempts, Timestamp: now}
				switch {
				case err != nil:
					var fe *decisions.Failure
					if errors.As(err, &fe) {
						fail(string(fe.Kind), fe.Status, fe.Reason)
					} else {
						fail("other", 0, err.Error())
					}
				case len(reply.Answers) != 1:
					fail("schema", 0, "reply does not carry exactly one answer")
				case j.b.Question.Kind == decisions.KindChoice && !choiceComplete(reply.Answers[0].Probabilities, j.b.Question):
					fail("schema", 0, "choice probabilities incomplete or not normalised")
				case j.b.Question.Kind == decisions.KindNoul && (math.IsNaN(reply.Answers[0].Probability) || reply.Answers[0].Kind != decisions.KindNoul):
					fail("schema", 0, "noul answer carries no probability")
				default:
					a := reply.Answers[0]
					sum := sha256.Sum256(reply.RequestBody)
					row.RequestSHA256 = hex.EncodeToString(sum[:])
					row.Confidence = confidenceOf(a.Answer)
					if a.Kind == decisions.KindNoul {
						p := a.Probability
						row.PTrue = &p
					} else {
						row.Choice, row.Probabilities = a.Choice, a.Probabilities
						p := a.Probabilities[a.Choice]
						row.PChoice = &p
					}
					write(row)
					mu.Lock()
					answered++
					mu.Unlock()
				}
				if over {
					mu.Lock()
					stopped = 1
					mu.Unlock()
					cancel()
				}
			}
		}()
	}
	pending := 0
	for _, j := range jobs {
		if done[j.reqID] {
			continue
		}
		pending++
		queue <- j
	}
	close(queue)
	wg.Wait()
	fmt.Printf("jev: %d units, %d requests (%d pending at start), %d answered this run, %d failed, %d with unknown billed cost, cumulative billed cost %.6f USD of which %.6f earlier (model %s)\n",
		len(units), len(jobs), pending, answered, failures, unknownCost, cost, spent, *model)
	if stopped == 1 {
		fmt.Fprintf(os.Stderr, "stopped: billed cost %.4f USD exceeded -max-cost-usd %.4f\n", cost, *maxCost)
		os.Exit(4)
	}
	if failures > 0 || unknownCost > 0 {
		os.Exit(3)
	}
}

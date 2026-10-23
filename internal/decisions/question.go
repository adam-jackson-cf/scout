// Package decisions is the typed-decision adapter (design §6): typed questions
// asked of Jev through OpenRouter's alpha/decisions route, resolved-model
// pinning, failure semantics, the call cap, the ordinal rank formula over the
// stored answers and the interim audit sampler. It only produces advisory
// annotations; it never touches issue state or the gate, and `issues` must
// not import it.
package decisions

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Kind is the type of a question.
type Kind string

// The three question types of the decisions route.
const (
	KindNoul   Kind = "noul"
	KindChoice Kind = "choice"
	KindScore  Kind = "score"
)

// Option-order rules. A fixed rule lists the options; a permutation rule
// orders them by a hash seeded with the unit id, so the order is reproducible
// and the key does not depend on the realised order.
const (
	OrderFixed       = "fixed"
	OrderPermutation = "permutation:unit_id"
)

// OptionOrderRule is the part of a question's version key that describes how
// options are ordered.
type OptionOrderRule struct {
	Rule    string   `json:"rule"`
	Options []string `json:"options,omitempty"`
}

// Question is a typed question as data. ID is for code only and is never sent
// to the model.
type Question struct {
	ID           string
	Kind         Kind
	Instructions string
	// Criteria maps an option (or `true`/`false` for a Noul) to its literal
	// criterion text.
	Criteria   map[string]string
	StateShape []string
	Order      OptionOrderRule
	// Positive names the option whose probability a Choice contributes to a
	// rank. It is a code-side reading of the answer, not part of the version key.
	Positive string
}

// Validate checks that the question is well formed for its kind.
func (q Question) Validate() error {
	if q.ID == "" || strings.TrimSpace(q.Instructions) == "" || len(q.StateShape) == 0 {
		return errors.New("question requires an id, instructions and a state shape")
	}
	switch q.Kind {
	case KindNoul:
		if len(q.Criteria) != 2 || q.Criteria["true"] == "" || q.Criteria["false"] == "" {
			return fmt.Errorf("noul question %q requires exactly the criteria true and false", q.ID)
		}
	case KindChoice:
		if len(q.Criteria) < 2 || len(q.Criteria) > 255 {
			return fmt.Errorf("choice question %q requires 2 to 255 options", q.ID)
		}
	case KindScore:
	default:
		return fmt.Errorf("question %q has unknown kind %q", q.ID, q.Kind)
	}
	switch q.Order.Rule {
	case OrderFixed:
		if q.Kind == KindChoice && !sameSet(q.Order.Options, q.Criteria) {
			return fmt.Errorf("choice question %q: fixed option order must list every option once", q.ID)
		}
	case OrderPermutation:
		if q.Kind != KindChoice {
			return fmt.Errorf("question %q: only a choice may use a permutation order", q.ID)
		}
	default:
		return fmt.Errorf("question %q has unknown option order rule %q", q.ID, q.Order.Rule)
	}
	return nil
}

func sameSet(options []string, criteria map[string]string) bool {
	if len(options) != len(criteria) {
		return false
	}
	seen := map[string]bool{}
	for _, option := range options {
		if _, ok := criteria[option]; !ok || seen[option] {
			return false
		}
		seen[option] = true
	}
	return true
}

// versionPayload is the canonical JSON the version key hashes (design §6).
// Field order is fixed by the struct; maps are marshalled with sorted keys.
type versionPayload struct {
	Kind            Kind              `json:"kind"`
	Instructions    string            `json:"instructions"`
	Criteria        map[string]string `json:"criteria"`
	StateShape      []string          `json:"state_shape"`
	OptionOrderRule OptionOrderRule   `json:"option_order_rule"`
	ResolvedModelID string            `json:"resolved_model_id"`
}

// Version is `question_id@<sha256 of canonical JSON>`. Any change to the
// kind, instructions, criteria, state shape, option-order rule or the pinned
// model changes it; the realised option order does not.
func (q Question) Version(resolvedModelID string) string {
	shape := append([]string(nil), q.StateShape...)
	sort.Strings(shape)
	raw, err := json.Marshal(versionPayload{q.Kind, q.Instructions, q.Criteria, shape, q.Order, resolvedModelID})
	if err != nil {
		panic("decisions: version payload is not encodable: " + err.Error())
	}
	sum := sha256.Sum256(raw)
	return q.ID + "@" + hex.EncodeToString(sum[:])
}

// RealisedOrder returns the option order used for one unit.
func (q Question) RealisedOrder(unitID string) []string {
	switch {
	case q.Kind == KindNoul:
		return []string{"true", "false"}
	case q.Order.Rule == OrderFixed:
		return append([]string(nil), q.Order.Options...)
	}
	options := make([]string, 0, len(q.Criteria))
	for option := range q.Criteria {
		options = append(options, option)
	}
	keys := make(map[string][32]byte, len(options))
	for _, option := range options {
		keys[option] = sha256.Sum256([]byte(unitID + "\x00" + option))
	}
	sort.Slice(options, func(i, j int) bool {
		a, b := keys[options[i]], keys[options[j]]
		if c := bytes.Compare(a[:], b[:]); c != 0 {
			return c < 0
		}
		return options[i] < options[j]
	})
	return options
}

// orderedCriteria marshals as a JSON object whose keys follow the realised
// option order.
type orderedCriteria struct {
	order    []string
	criteria map[string]string
}

func (o orderedCriteria) MarshalJSON() ([]byte, error) {
	var out bytes.Buffer
	out.WriteByte('{')
	for index, option := range o.order {
		if index > 0 {
			out.WriteByte(',')
		}
		key, err := json.Marshal(option)
		if err != nil {
			return nil, err
		}
		value, err := json.Marshal(o.criteria[option])
		if err != nil {
			return nil, err
		}
		out.Write(key)
		out.WriteByte(':')
		out.Write(value)
	}
	out.WriteByte('}')
	return out.Bytes(), nil
}

// QuestionKey is the only key the model ever sees for the i-th question of a
// request (zero-based): `q1`, `q2`, ... Question ids are for code only.
func QuestionKey(i int) string { return "q" + strconv.Itoa(i+1) }

type wireQuestion struct {
	Type         Kind            `json:"type"`
	Instructions string          `json:"instructions"`
	Criteria     orderedCriteria `json:"criteria"`
}

// Request is the exact body sent to the decisions route.
type Request struct {
	Model     string                  `json:"model"`
	State     json.RawMessage         `json:"state"`
	Questions map[string]wireQuestion `json:"questions"`
}

// BuildRequest renders the one request of a unit: every independent question
// over the same state, fanned out (compose-and-integrate: ask independent
// questions over the same state together).
func BuildRequest(questions []Question, model, unitID string, state json.RawMessage) Request {
	wire := make(map[string]wireQuestion, len(questions))
	for index, q := range questions {
		criteria := q.Criteria
		if q.Kind == KindScore {
			criteria = map[string]string{}
		}
		wire[QuestionKey(index)] = wireQuestion{Type: q.Kind, Instructions: q.Instructions, Criteria: orderedCriteria{order: orderFor(q, q.RealisedOrder(unitID)), criteria: criteria}}
	}
	return Request{Model: model, State: state, Questions: wire}
}

func orderFor(q Question, order []string) []string {
	if q.Kind == KindScore {
		return nil
	}
	return order
}

package decisions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// DefaultEndpoint is OpenRouter's alpha decisions route. Its contract, model
// slugs and error codes are undocumented by TypeSafe [INFERENCE on stability,
// design R14]: any non-success is a failed unit.
const DefaultEndpoint = "https://openrouter.ai/api/alpha/decisions"

// FailureKind classifies why a unit has no annotation.
type FailureKind string

// Failure kinds. Every one leaves the unit unannotated; none is a pass or a
// low score.
const (
	FailNetwork       FailureKind = "network"
	FailTimeout       FailureKind = "timeout"
	FailHTTP          FailureKind = "http"
	FailNonRetryable  FailureKind = "non-retryable"
	FailRateLimited   FailureKind = "rate-limited"
	FailSchema        FailureKind = "schema-invalid"
	FailModelMismatch FailureKind = "model-mismatch"
	FailUnsafe        FailureKind = "unsafe-to-persist"
)

// Failure is the error of one unit's call.
type Failure struct {
	Kind   FailureKind
	Status int
	Reason string
}

func (f *Failure) Error() string { return string(f.Kind) + ": " + f.Reason }

// Client calls the decisions route. Sleep and Now exist so tests run with a
// fake clock; zero values use the real one.
type Client struct {
	Endpoint string
	// Key is the OpenRouter key. It is never logged, stored or put in an error.
	Key  string
	HTTP *http.Client
	// Sleep waits between retries; nil uses a real timer.
	Sleep func(ctx context.Context, d time.Duration) error
	// Now reads the clock for HTTP-date Retry-After values; nil uses time.Now.
	Now func() time.Time
	// MaxAttempts bounds attempts for 429/529 (default 5).
	MaxAttempts int
	// BaseBackoff doubles per attempt (default 1s); MaxWait bounds every wait,
	// including a Retry-After (default 60s).
	BaseBackoff time.Duration
	MaxWait     time.Duration
}

const maxResponseBytes = 1 << 20

// Response is one validated answer of a reply.
type Response struct {
	Kind        Kind
	Probability float64
	// Choice is the option a Choice answer selected; empty for Noul and Score.
	Choice string
	// Probabilities is a Choice's probability per option as the route
	// reported it; nil for Noul and Score.
	Probabilities map[string]float64
	Answer        json.RawMessage
	// OptionOrder is the option order realised in the request.
	OptionOrder []string
}

// Reply is one request's result: the model that answered, the billed usage,
// the exact request and one validated Response per question, in question order.
type Reply struct {
	Model string
	ID    string
	// Usage is the response's `usage` object; CostUSD is its billed `cost`
	// and CostKnown is false when it is missing or not a finite non-negative
	// number. A reply that failed validation still carries both, because the
	// call was billed.
	Usage     json.RawMessage
	CostUSD   float64
	CostKnown bool
	// Latency is the end-to-end time of the request, including any backoff
	// waits and retries.
	Latency time.Duration
	// RequestBody is the exact request sent, already redacted by construction.
	RequestBody json.RawMessage
	Attempts    int
	Answers     []Response
}

func (c *Client) sleep(ctx context.Context, d time.Duration) error {
	if c.Sleep != nil {
		return c.Sleep(ctx, d)
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (c *Client) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

// Ask sends one unit's questions as one request and validates the reply. A
// reply from a model other than the pin is a failure of kind model-mismatch.
// The returned Reply carries billed usage even when the error is a failure of
// a reply that was received.
func (c *Client) Ask(ctx context.Context, questions []Question, model, unitID string, state json.RawMessage) (Reply, error) {
	if len(questions) == 0 {
		return Reply{}, errors.New("a request needs at least one question")
	}
	seen := map[string]bool{}
	for _, q := range questions {
		if err := q.Validate(); err != nil {
			return Reply{}, err
		}
		if seen[q.ID] {
			return Reply{}, fmt.Errorf("question %q is asked twice in one request", q.ID)
		}
		seen[q.ID] = true
	}
	if c.Key == "" {
		return Reply{}, &Failure{Kind: FailNetwork, Reason: "no key"}
	}
	request := BuildRequest(questions, model, unitID, state)
	body, err := json.Marshal(request)
	if err != nil {
		return Reply{}, &Failure{Kind: FailSchema, Reason: "request is not encodable: " + err.Error()}
	}
	endpoint := c.Endpoint
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 60 * time.Second}
	}
	attempts := c.MaxAttempts
	if attempts <= 0 {
		attempts = 5
	}
	base := c.BaseBackoff
	if base <= 0 {
		base = time.Second
	}
	maxWait := c.MaxWait
	if maxWait <= 0 {
		maxWait = time.Minute
	}
	started := c.now()
	for attempt := 1; ; attempt++ {
		status, header, payload, err := c.post(ctx, httpClient, endpoint, body)
		if err != nil {
			return Reply{}, err
		}
		switch {
		case status == http.StatusOK:
			reply, err := parseResponse(questions, model, unitID, payload)
			reply.RequestBody = body
			reply.Attempts = attempt
			reply.Latency = c.now().Sub(started)
			return reply, err
		case status == http.StatusBadRequest || status == http.StatusUnprocessableEntity:
			reason := fmt.Sprintf("HTTP %d", status)
			if bytes.Contains(payload, []byte("max_tokens_exceeded")) {
				reason += " max_tokens_exceeded"
			}
			return Reply{}, &Failure{Kind: FailNonRetryable, Status: status, Reason: reason}
		case status == http.StatusTooManyRequests || status == 529:
			if attempt >= attempts {
				return Reply{}, &Failure{Kind: FailRateLimited, Status: status, Reason: fmt.Sprintf("HTTP %d after %d attempts", status, attempt)}
			}
			wait := base << (attempt - 1)
			if retry, ok := retryAfter(header.Get("Retry-After"), c.now()); ok && retry > wait {
				wait = retry
			}
			if wait > maxWait {
				wait = maxWait
			}
			if err := c.sleep(ctx, wait); err != nil {
				return Reply{}, &Failure{Kind: FailTimeout, Reason: "cancelled while backing off: " + err.Error()}
			}
		default:
			return Reply{}, &Failure{Kind: FailHTTP, Status: status, Reason: fmt.Sprintf("HTTP %d", status)}
		}
	}
}

func (c *Client) post(ctx context.Context, client *http.Client, endpoint string, body []byte) (int, http.Header, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return 0, nil, nil, &Failure{Kind: FailNetwork, Reason: "invalid endpoint"}
	}
	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		var netErr net.Error
		if errors.Is(err, context.DeadlineExceeded) || errors.As(err, &netErr) && netErr.Timeout() {
			return 0, nil, nil, &Failure{Kind: FailTimeout, Reason: "request timed out"}
		}
		// The key is in a header, never in the URL; the error text is the
		// transport's and is reduced to its kind.
		return 0, nil, nil, &Failure{Kind: FailNetwork, Reason: "request failed"}
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return 0, nil, nil, &Failure{Kind: FailNetwork, Reason: "response unreadable"}
	}
	return resp.StatusCode, resp.Header, payload, nil
}

// retryAfter parses delta-seconds or an HTTP date.
func retryAfter(value string, now time.Time) (time.Duration, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	if seconds, err := strconv.ParseFloat(value, 64); err == nil && seconds >= 0 && !math.IsInf(seconds, 0) && seconds < 1e6 {
		return time.Duration(seconds * float64(time.Second)), true
	}
	if at, err := http.ParseTime(value); err == nil {
		if d := at.Sub(now); d > 0 {
			return d, true
		}
		return 0, true
	}
	return 0, false
}

type wireResponse struct {
	ID      string                     `json:"id"`
	Model   string                     `json:"model"`
	Answers map[string]json.RawMessage `json:"answers"`
	Usage   json.RawMessage            `json:"usage"`
}

type wireAnswer struct {
	Type          string             `json:"type"`
	Noul          *float64           `json:"noul"`
	Score         *float64           `json:"score"`
	Choice        string             `json:"choice"`
	Probabilities map[string]float64 `json:"probabilities"`
	Confidence    *float64           `json:"confidence"`
}

func schemaFailure(format string, args ...any) *Failure {
	return &Failure{Kind: FailSchema, Reason: fmt.Sprintf(format, args...)}
}

func probability(value float64) bool { return !math.IsNaN(value) && value >= 0 && value <= 1 }

// billedCost reads the billed `cost` of a usage object.
func billedCost(usage json.RawMessage) (float64, bool) {
	var parsed struct {
		Cost *float64 `json:"cost"`
	}
	if len(usage) == 0 || json.Unmarshal(usage, &parsed) != nil || parsed.Cost == nil {
		return 0, false
	}
	cost := *parsed.Cost
	if math.IsNaN(cost) || math.IsInf(cost, 0) || cost < 0 {
		return 0, false
	}
	return cost, true
}

// parseResponse enforces the model pin first, then the answer schema of every
// question. Usage is read before either check so a discarded reply still
// reports what it cost.
func parseResponse(questions []Question, pinned, unitID string, payload []byte) (Reply, error) {
	var wire wireResponse
	if err := json.Unmarshal(payload, &wire); err != nil {
		return Reply{}, schemaFailure("response is not JSON")
	}
	reply := Reply{Model: wire.Model, ID: wire.ID, Usage: wire.Usage}
	reply.CostUSD, reply.CostKnown = billedCost(wire.Usage)
	if wire.Model != pinned {
		return reply, &Failure{Kind: FailModelMismatch, Reason: fmt.Sprintf("resolved model %q differs from pin %q", wire.Model, pinned)}
	}
	if len(wire.Answers) != len(questions) {
		return reply, schemaFailure("response has %d answers for %d questions", len(wire.Answers), len(questions))
	}
	for index, q := range questions {
		raw, ok := wire.Answers[QuestionKey(index)]
		if !ok {
			return reply, schemaFailure("response lacks the answer to question %d", index+1)
		}
		response, err := parseAnswer(q, unitID, raw)
		if err != nil {
			return reply, err
		}
		reply.Answers = append(reply.Answers, response)
	}
	return reply, nil
}

func parseAnswer(q Question, unitID string, raw json.RawMessage) (Response, error) {
	var answer wireAnswer
	if err := json.Unmarshal(raw, &answer); err != nil {
		return Response{}, schemaFailure("answer is not an object")
	}
	if answer.Type != string(q.Kind) {
		return Response{}, schemaFailure("answer type %q is not %q", answer.Type, q.Kind)
	}
	response := Response{Kind: q.Kind, Answer: raw, OptionOrder: orderFor(q, q.RealisedOrder(unitID))}
	switch q.Kind {
	case KindNoul:
		if answer.Noul == nil || !probability(*answer.Noul) {
			return Response{}, schemaFailure("noul answer is missing or outside [0, 1]")
		}
		response.Probability = *answer.Noul
	case KindScore:
		if answer.Score == nil || math.IsNaN(*answer.Score) || math.IsInf(*answer.Score, 0) {
			return Response{}, schemaFailure("score answer is missing or not finite")
		}
		response.Probability = *answer.Score
	case KindChoice:
		if _, ok := q.Criteria[answer.Choice]; !ok {
			return Response{}, schemaFailure("choice %q is not an option", answer.Choice)
		}
		response.Choice = answer.Choice
		response.Probabilities = answer.Probabilities
		for option, p := range answer.Probabilities {
			if _, ok := q.Criteria[option]; !ok || !probability(p) {
				return Response{}, schemaFailure("choice probabilities are invalid")
			}
		}
		if q.Positive != "" {
			p, ok := answer.Probabilities[q.Positive]
			if !ok {
				return Response{}, schemaFailure("choice answer lacks a probability for %q", q.Positive)
			}
			response.Probability = p
		}
	}
	return response, nil
}

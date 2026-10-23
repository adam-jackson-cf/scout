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

// Response is a validated answer.
type Response struct {
	Model       string
	ID          string
	Kind        Kind
	Probability float64
	Answer      json.RawMessage
	Usage       json.RawMessage
	// RequestBody is the exact request sent, already redacted by construction.
	RequestBody json.RawMessage
	OptionOrder []string
	Attempts    int
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

// Ask sends one unit's question and validates the reply. A reply from a model
// other than the pin is a failure of kind model-mismatch.
func (c *Client) Ask(ctx context.Context, q Question, model, unitID string, state json.RawMessage) (Response, error) {
	if err := q.Validate(); err != nil {
		return Response{}, err
	}
	if c.Key == "" {
		return Response{}, &Failure{Kind: FailNetwork, Reason: "no key"}
	}
	request := BuildRequest(q, model, unitID, state)
	body, err := json.Marshal(request)
	if err != nil {
		return Response{}, &Failure{Kind: FailSchema, Reason: "request is not encodable: " + err.Error()}
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
	for attempt := 1; ; attempt++ {
		status, header, payload, err := c.post(ctx, httpClient, endpoint, body)
		if err != nil {
			return Response{}, err
		}
		switch {
		case status == http.StatusOK:
			response, err := parseResponse(q, model, payload)
			if err != nil {
				return Response{}, err
			}
			response.RequestBody = body
			response.OptionOrder = orderFor(q, q.RealisedOrder(unitID))
			response.Attempts = attempt
			return response, nil
		case status == http.StatusBadRequest || status == http.StatusUnprocessableEntity:
			reason := fmt.Sprintf("HTTP %d", status)
			if bytes.Contains(payload, []byte("max_tokens_exceeded")) {
				reason += " max_tokens_exceeded"
			}
			return Response{}, &Failure{Kind: FailNonRetryable, Status: status, Reason: reason}
		case status == http.StatusTooManyRequests || status == 529:
			if attempt >= attempts {
				return Response{}, &Failure{Kind: FailRateLimited, Status: status, Reason: fmt.Sprintf("HTTP %d after %d attempts", status, attempt)}
			}
			wait := base << (attempt - 1)
			if retry, ok := retryAfter(header.Get("Retry-After"), c.now()); ok && retry > wait {
				wait = retry
			}
			if wait > maxWait {
				wait = maxWait
			}
			if err := c.sleep(ctx, wait); err != nil {
				return Response{}, &Failure{Kind: FailTimeout, Reason: "cancelled while backing off: " + err.Error()}
			}
		default:
			return Response{}, &Failure{Kind: FailHTTP, Status: status, Reason: fmt.Sprintf("HTTP %d", status)}
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

// parseResponse enforces the model pin first, then the answer schema.
func parseResponse(q Question, pinned string, payload []byte) (Response, error) {
	var wire wireResponse
	if err := json.Unmarshal(payload, &wire); err != nil {
		return Response{}, schemaFailure("response is not JSON")
	}
	if wire.Model != pinned {
		return Response{}, &Failure{Kind: FailModelMismatch, Reason: fmt.Sprintf("resolved model %q differs from pin %q", wire.Model, pinned)}
	}
	raw, ok := wire.Answers[questionName]
	if !ok || len(wire.Answers) != 1 {
		return Response{}, schemaFailure("response has no single answer")
	}
	var answer wireAnswer
	if err := json.Unmarshal(raw, &answer); err != nil {
		return Response{}, schemaFailure("answer is not an object")
	}
	if answer.Type != string(q.Kind) {
		return Response{}, schemaFailure("answer type %q is not %q", answer.Type, q.Kind)
	}
	response := Response{Model: wire.Model, ID: wire.ID, Kind: q.Kind, Answer: raw, Usage: wire.Usage}
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

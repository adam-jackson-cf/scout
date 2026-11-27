package fetch

import (
	"net/http"
	"testing"

	"canary/client"
)

type refuse struct{}

func (refuse) RoundTrip(*http.Request) (*http.Response, error) { return nil, http.ErrHandlerTimeout }

func TestStatusIsZeroWhenTheRequestFails(t *testing.T) {
	client.Default.Transport = refuse{}
	if Status("http://example.invalid") != 0 {
		t.Fatal("status")
	}
}

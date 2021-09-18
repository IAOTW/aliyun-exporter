package ratelimit

import (
	"net/http"

	"go.uber.org/ratelimit"
)

// Transport implements http.RoundTripper.
type Transport struct {
	Transport http.RoundTripper // Used to make actual requests.
	Limiter   ratelimit.Limiter
}

// RoundTrip ensures requests are performed within the rate limiting constraints.
func (t *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	_ = t.Limiter.Take()
	if t.Transport == nil {
		t.Transport = http.DefaultTransport
	}
	return t.Transport.RoundTrip(r)
}

// NewWithLimiter returns a RoundTripper capable of rate limiting http requests.
func NewWithLimiter(limiter ratelimit.Limiter) http.RoundTripper {
	return &Transport{Transport: http.DefaultTransport, Limiter: limiter}
}

// New ..
func New(limit int) http.RoundTripper {
	return NewWithLimiter(ratelimit.New(limit))
}

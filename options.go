package shopline

import (
	"net/http"
	"time"
)

// Option configures a Client.
type Option func(*Client)

// WithVersion sets the API version for the client.
// Example: shopline.WithVersion("v20251201")
func WithVersion(version string) Option {
	return func(c *Client) {
		c.apiVersion = version
	}
}

// WithRetry sets the maximum number of retries for failed requests.
// Retries are performed on HTTP 429 (rate limited) and HTTP 503 responses.
func WithRetry(retries int) Option {
	return func(c *Client) {
		c.maxRetries = retries
	}
}

// WithHTTPClient sets a custom HTTP client for API requests.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithLogger sets a logger for the client.
func WithLogger(logger Logger) Option {
	return func(c *Client) {
		c.log = logger
	}
}

// WithBaseURL sets a custom base URL for the client (useful for testing).
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURLOverride = baseURL
	}
}

// WithTokenManager enables automatic token management with persistence
// and concurrency-safe refresh. The TokenStore is used to persist tokens
// across process restarts.
//
// When this option is set, the static token string passed to NewClient
// is ignored. Instead, tokens are managed automatically:
//   - Loaded from the store on first use
//   - Refreshed proactively before expiry
//   - Only one goroutine refreshes at a time (singleflight pattern)
//
// Example:
//
//	store := shopline.NewFileTokenStore("./tokens")
//	client, _ := shopline.NewClient(app, "myshop", "",
//	    shopline.WithTokenManager(store),
//	)
func WithTokenManager(store TokenStore, opts ...TokenManagerOption) Option {
	return func(c *Client) {
		tm := NewTokenManager(c.app, c.handle, store, opts...)
		tm.log = c.log // share logger
		c.tokenManager = tm
	}
}

// WithCircuitBreaker enables a circuit breaker that stops sending requests when
// the upstream service is consistently failing.
//
// Parameters:
//   - threshold: consecutive failures before the circuit opens (recommended: 5)
//   - cooldown: how long to stay in Open state before probing again (recommended: 30s)
//
// When the circuit is Open, requests fail immediately with an error rather than
// waiting for a timeout, protecting both the client and the upstream service.
func WithCircuitBreaker(threshold int, cooldown time.Duration) Option {
	return func(c *Client) {
		c.cb = newCircuitBreaker(threshold, cooldown)
	}
}

// WithTimeout overrides the HTTP client's request timeout.
// The default timeout is 30 seconds.
//
// Use a longer timeout for bulk operations:
//
//	client, _ := shopline.NewClient(app, handle, token,
//	    shopline.WithTimeout(5 * time.Minute), // for /bulk/ endpoints
//	)
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

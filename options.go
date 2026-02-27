package shopline

import "net/http"

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

package shopline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	// contentType is the required content type for Shopline API.
	contentType = "application/json; charset=utf-8"

	// maxResponseBodySize limits response body reads to 10MB to prevent OOM
	// from malicious or abnormally large responses.
	maxResponseBodySize = 10 * 1024 * 1024

	// maxBackoff caps the exponential backoff duration.
	maxBackoff = 30 * time.Second
)

// timeNow is a function variable for testing.
var timeNow = time.Now

// CreatePath builds the API URL path for a given resource.
// e.g. /admin/openapi/v20251201/products.json
func (c *Client) CreatePath(resource string) string {
	return fmt.Sprintf("/admin/openapi/%s/%s", c.apiVersion, resource)
}

// NewRequest creates an HTTP request with proper headers for the Shopline API.
func (c *Client) NewRequest(ctx context.Context, method, relPath string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(relPath)
	if err != nil {
		return nil, fmt.Errorf("shopline: invalid path %q: %w", relPath, err)
	}

	reqURL := c.baseURL.ResolveReference(rel)

	var buf io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("shopline: failed to marshal request body: %w", err)
		}
		buf = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("shopline: failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", UserAgent)

	// Set authorization header
	// If TokenManager is set, dynamically fetch a valid token (may trigger refresh).
	// Otherwise, use the static token string for backward compatibility.
	if c.tokenManager != nil {
		token, err := c.tokenManager.GetToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("shopline: failed to get access token: %w", err)
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	} else if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return req, nil
}

// Do sends an HTTP request and decodes the JSON response.
// It handles retries for rate limiting (429) and server errors (503)
// with exponential backoff and jitter. It respects context cancellation
// during retry waits.
func (c *Client) Do(req *http.Request, result interface{}) (*http.Response, error) {
	var resp *http.Response
	var err error

	// P0-1: Pre-save request body before the retry loop.
	// The body is a one-time-use stream — if we don't save it before the first
	// attempt, retries will send empty bodies silently.
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("shopline: failed to read request body: %w", err)
		}
		req.Body.Close()
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			c.logDebugf("Retry attempt %d/%d for %s %s", attempt, c.maxRetries, req.Method, req.URL)
			// Restore body for retry
			if bodyBytes != nil {
				req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
		}

		resp, err = c.httpClient.Do(req)
		if err != nil {
			if attempt < c.maxRetries {
				// P1-4: Exponential backoff with jitter for network errors
				backoff := backoffDuration(attempt, time.Second)
				c.logDebugf("Request error: %v, backing off %s", err, backoff)
				// P0-2: Respect context cancellation during sleep
				if sleepErr := sleepWithContext(req.Context(), backoff); sleepErr != nil {
					return nil, fmt.Errorf("shopline: request cancelled during retry: %w", sleepErr)
				}
				continue
			}
			return nil, fmt.Errorf("shopline: request failed after %d retries: %w", c.maxRetries, err)
		}

		// Check for retryable status codes
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
			if attempt < c.maxRetries {
				// P1-5: Correctly parse Retry-After header
				retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
				if retryAfter <= 0 {
					// Fall back to exponential backoff
					retryAfter = backoffDuration(attempt, 2*time.Second)
				}
				// Read and discard body before closing to allow connection reuse
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				c.logDebugf("Rate limited or service unavailable (HTTP %d), retrying after %s", resp.StatusCode, retryAfter)
				// P0-2: Respect context cancellation during sleep
				if sleepErr := sleepWithContext(req.Context(), retryAfter); sleepErr != nil {
					return nil, fmt.Errorf("shopline: request cancelled during retry: %w", sleepErr)
				}
				continue
			}
		}

		break
	}

	if resp == nil {
		return nil, fmt.Errorf("shopline: no response received")
	}

	// P1-6: Limit response body size to prevent OOM
	// P0-3: Read body fully, then close — do NOT defer close and return resp
	//       with an open body, which creates a data race for callers.
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize))
	resp.Body.Close()

	if readErr != nil {
		return resp, fmt.Errorf("shopline: failed to read response body: %w", readErr)
	}

	// Check for errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp, parseResponseErrorFromBytes(resp, body)
	}

	// Decode response body
	if result != nil && len(body) > 0 {
		if err := json.Unmarshal(body, result); err != nil {
			return resp, fmt.Errorf("shopline: failed to decode response: %w (body: %s)", err, string(body))
		}
	}

	return resp, nil
}

// Get performs a GET request to the given path and decodes the response.
func (c *Client) Get(ctx context.Context, path string, result interface{}, opts interface{}) error {
	if opts != nil {
		queryString := buildQueryString(opts)
		if queryString != "" {
			if strings.Contains(path, "?") {
				path += "&" + queryString
			} else {
				path += "?" + queryString
			}
		}
	}

	req, err := c.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}

	_, err = c.Do(req, result)
	return err
}

// Post performs a POST request to the given path with the given body.
func (c *Client) Post(ctx context.Context, path string, body, result interface{}) error {
	req, err := c.NewRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}

	_, err = c.Do(req, result)
	return err
}

// Put performs a PUT request to the given path with the given body.
func (c *Client) Put(ctx context.Context, path string, body, result interface{}) error {
	req, err := c.NewRequest(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}

	_, err = c.Do(req, result)
	return err
}

// Delete performs a DELETE request to the given path.
func (c *Client) Delete(ctx context.Context, path string) error {
	req, err := c.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	_, err = c.Do(req, nil)
	return err
}

// sleepWithContext sleeps for the specified duration or until the context is
// cancelled, whichever comes first. Returns ctx.Err() if cancelled.
func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// backoffDuration returns an exponential backoff duration with jitter.
// The formula is: base * 2^attempt, capped at maxBackoff, with ±25% jitter.
func backoffDuration(attempt int, base time.Duration) time.Duration {
	backoff := base * time.Duration(1<<uint(attempt))
	if backoff > maxBackoff {
		backoff = maxBackoff
	}
	// Add jitter: 75%-125% of backoff to prevent thundering herd
	jitter := time.Duration(rand.Int63n(int64(backoff/2))) - backoff/4
	return backoff + jitter
}

// parseRetryAfter parses the Retry-After HTTP header value.
// It supports two formats per RFC 7231:
//   - Delay in seconds (integer or float): "120", "2.5"
//   - HTTP-date: "Fri, 31 Dec 2025 23:59:59 GMT"
//
// Returns 0 if the header is empty or unparseable.
func parseRetryAfter(header string) time.Duration {
	if header == "" {
		return 0
	}
	// Try as seconds (integer or float)
	if seconds, err := strconv.ParseFloat(header, 64); err == nil {
		return time.Duration(seconds * float64(time.Second))
	}
	// Try as HTTP-date
	if t, err := http.ParseTime(header); err == nil {
		d := time.Until(t)
		if d > 0 {
			return d
		}
	}
	return 0
}

// buildQueryString converts a struct with `url` tags to a query string.
// This is a simplified version — supports basic types.
func buildQueryString(opts interface{}) string {
	if opts == nil {
		return ""
	}

	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return ""
	}

	params := url.Values{}
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		tag := fieldType.Tag.Get("url")
		if tag == "" || tag == "-" {
			continue
		}

		parts := strings.Split(tag, ",")
		name := parts[0]
		omitempty := len(parts) > 1 && parts[1] == "omitempty"

		if omitempty && field.IsZero() {
			continue
		}

		params.Set(name, fmt.Sprintf("%v", field.Interface()))
	}

	return params.Encode()
}

package shopline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

const (
	// contentType is the required content type for Shopline API.
	contentType = "application/json; charset=utf-8"
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
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return req, nil
}

// Do sends an HTTP request and decodes the JSON response.
// It handles retries for rate limiting (429) and server errors (503).
func (c *Client) Do(req *http.Request, result interface{}) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			c.logDebugf("Retry attempt %d/%d for %s %s", attempt, c.maxRetries, req.Method, req.URL)
		}

		// Clone the request body if we need to retry
		var bodyBytes []byte
		if req.Body != nil && attempt > 0 {
			bodyBytes, _ = io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		resp, err = c.httpClient.Do(req)
		if err != nil {
			if attempt < c.maxRetries {
				time.Sleep(time.Duration(attempt+1) * time.Second)
				if bodyBytes != nil {
					req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
				continue
			}
			return nil, fmt.Errorf("shopline: request failed: %w", err)
		}

		// Check if we should retry
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
			if attempt < c.maxRetries {
				retryAfter := 2 * time.Second
				if ra := resp.Header.Get("Retry-After"); ra != "" {
					if d, parseErr := time.ParseDuration(ra + "s"); parseErr == nil {
						retryAfter = d
					}
				}
				resp.Body.Close()
				c.logDebugf("Rate limited or service unavailable, retrying after %s", retryAfter)
				time.Sleep(retryAfter)
				if bodyBytes != nil {
					req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
				continue
			}
		}

		break
	}

	if resp == nil {
		return nil, fmt.Errorf("shopline: no response received")
	}

	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp, parseResponseError(resp)
	}

	// Decode response body
	if result != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp, fmt.Errorf("shopline: failed to read response body: %w", err)
		}

		if len(body) > 0 {
			if err := json.Unmarshal(body, result); err != nil {
				return resp, fmt.Errorf("shopline: failed to decode response: %w (body: %s)", err, string(body))
			}
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

package shopline

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ResponseError represents an error response from the Shopline API.
type ResponseError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	TraceID string `json:"traceId"`
	// Errors can be a string, []string, or map[string][]string depending on the endpoint.
	Errors  interface{} `json:"errors"`
	RawBody []byte      `json:"-"`
}

// Error implements the error interface.
func (e *ResponseError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("shopline: %d %s (traceId: %s)", e.Status, e.Message, e.TraceID)
	}
	if e.Errors != nil {
		return fmt.Sprintf("shopline: %d %v (traceId: %s)", e.Status, e.Errors, e.TraceID)
	}
	return fmt.Sprintf("shopline: %d (traceId: %s)", e.Status, e.TraceID)
}

// GetErrors returns the error details as a formatted string.
func (e *ResponseError) GetErrors() string {
	switch v := e.Errors.(type) {
	case string:
		return v
	case []interface{}:
		msgs := make([]string, 0, len(v))
		for _, item := range v {
			msgs = append(msgs, fmt.Sprintf("%v", item))
		}
		return strings.Join(msgs, "; ")
	case map[string]interface{}:
		msgs := make([]string, 0, len(v))
		for key, val := range v {
			msgs = append(msgs, fmt.Sprintf("%s: %v", key, val))
		}
		return strings.Join(msgs, "; ")
	default:
		return fmt.Sprintf("%v", e.Errors)
	}
}

// RateLimitError represents a rate limit error (HTTP 429).
type RateLimitError struct {
	ResponseError
	RetryAfter time.Duration
}

// Error implements the error interface.
func (e *RateLimitError) Error() string {
	return fmt.Sprintf("shopline: rate limited (429), retry after %s (traceId: %s)", e.RetryAfter, e.TraceID)
}

// parseResponseError creates a ResponseError from an HTTP response.
// This is a convenience wrapper that reads the body first.
func parseResponseError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ResponseError{
			Status:  resp.StatusCode,
			Message: "failed to read error response body",
		}
	}
	return parseResponseErrorFromBytes(resp, body)
}

// parseResponseErrorFromBytes creates a ResponseError from an HTTP response
// and pre-read body bytes. This avoids double-reading the response body.
func parseResponseErrorFromBytes(resp *http.Response, body []byte) error {
	respErr := &ResponseError{
		Status:  resp.StatusCode,
		RawBody: body,
	}

	// Try to parse JSON body
	if len(body) > 0 {
		var parsed map[string]interface{}
		if jsonErr := json.Unmarshal(body, &parsed); jsonErr == nil {
			if msg, ok := parsed["message"].(string); ok {
				respErr.Message = msg
			}
			if traceID, ok := parsed["traceId"].(string); ok {
				respErr.TraceID = traceID
			}
			if errors, ok := parsed["errors"]; ok {
				respErr.Errors = errors
			}
			// Some Shopline errors use "error" instead of "errors"
			if errMsg, ok := parsed["error"].(string); ok && respErr.Message == "" {
				respErr.Message = errMsg
			}
		} else {
			// If not valid JSON, use body as message
			respErr.Message = string(body)
		}
	}

	// Handle rate limiting
	if resp.StatusCode == http.StatusTooManyRequests {
		rlErr := &RateLimitError{
			ResponseError: *respErr,
		}
		if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
			if d := parseRetryAfter(retryAfter); d > 0 {
				rlErr.RetryAfter = d
			}
		}
		if rlErr.RetryAfter == 0 {
			rlErr.RetryAfter = 2 * time.Second // default
		}
		return rlErr
	}

	return respErr
}

package shopline

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// authHTTPClient is a dedicated HTTP client for auth endpoints with
// proper timeout and connection pool settings. Using http.DefaultClient
// in production is dangerous — it has no timeout, so a single slow
// response can block a goroutine forever.
var authHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        20,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     90 * time.Second,
	},
}

// TokenResponse represents the response from token creation/refresh API.
type TokenResponse struct {
	Code     int    `json:"code"`
	I18nCode string `json:"i18nCode"`
	Message  string `json:"message"`
	Data     struct {
		AccessToken string `json:"accessToken"`
		ExpireTime  string `json:"expireTime"`
		Scope       string `json:"scope"`
	} `json:"data"`
	TraceID string `json:"traceId"`
}

// AuthorizeURL generates the OAuth authorization URL for a merchant.
//
// Parameters:
//   - handle: Store handle (e.g. "open001")
//   - state: A random nonce for CSRF protection (passed as customField)
//
// The merchant should be redirected to this URL to authorize the app.
func (app App) AuthorizeURL(handle, state string) string {
	params := url.Values{
		"appKey":       {app.AppKey},
		"responseType": {"code"},
		"scope":        {app.Scope},
		"redirectUri":  {app.RedirectURL},
	}
	if state != "" {
		params.Set("customField", state)
	}
	return fmt.Sprintf(
		"https://%s.myshopline.com/admin/oauth-web/#/oauth/authorize?%s",
		handle,
		params.Encode(),
	)
}

// GenerateSignature generates an HMAC-SHA256 signature for API requests.
//
// The signature is computed by:
// 1. Sorting the parameter keys alphabetically
// 2. Concatenating the key-value pairs as "key=value"
// 3. Joining them with "&"
// 4. Computing HMAC-SHA256 with the AppSecret as key
func (app App) GenerateSignature(params map[string]string) string {
	// Sort keys
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" { // exclude sign itself
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// Build string to sign
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	message := strings.Join(parts, "&")

	// HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(app.AppSecret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature verifies the HMAC-SHA256 signature from a Shopline callback request.
func (app App) VerifySignature(query url.Values) bool {
	sign := query.Get("sign")
	if sign == "" {
		return false
	}

	params := make(map[string]string)
	for k, v := range query {
		if k != "sign" && len(v) > 0 {
			params[k] = v[0]
		}
	}

	expected := app.GenerateSignature(params)
	return hmac.Equal([]byte(sign), []byte(expected))
}

// GetAccessToken exchanges an authorization code for an access token.
//
// This corresponds to Step 4 of the Shopline OAuth flow.
// POST https://{handle}.myshopline.com/admin/oauth/token/create
func (app App) GetAccessToken(ctx context.Context, handle, code string) (*TokenResponse, error) {
	bodyJSON, err := json.Marshal(map[string]string{"code": code})
	if err != nil {
		return nil, fmt.Errorf("shopline: failed to marshal body: %w", err)
	}
	return app.doAuthRequest(ctx, handle, "create", bytes.NewReader(bodyJSON))
}

// RefreshAccessToken refreshes the access token before it expires (10-hour validity).
//
// This corresponds to Step 6 of the Shopline OAuth flow.
// POST https://{handle}.myshopline.com/admin/oauth/token/refresh
func (app App) RefreshAccessToken(ctx context.Context, handle string) (*TokenResponse, error) {
	return app.doAuthRequest(ctx, handle, "refresh", nil)
}

// doAuthRequest is the shared implementation for token create/refresh requests.
// It handles signature generation, header setting, request execution, and
// response parsing in a single place to eliminate code duplication.
func (app App) doAuthRequest(ctx context.Context, handle, endpoint string, body io.Reader) (*TokenResponse, error) {
	// P1-5: Validate handle to prevent empty or malicious URL construction
	if handle == "" {
		return nil, fmt.Errorf("shopline: handle must not be empty")
	}

	timestamp := fmt.Sprintf("%d", currentTimeMillis())
	sign := app.GenerateSignature(map[string]string{
		"appkey":    app.AppKey,
		"timestamp": timestamp,
	})

	apiURL := fmt.Sprintf("https://%s.myshopline.com/admin/oauth/token/%s", handle, endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, body)
	if err != nil {
		return nil, fmt.Errorf("shopline: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("appkey", app.AppKey)
	req.Header.Set("timestamp", timestamp)
	req.Header.Set("sign", sign)

	// P0-1: Use dedicated client with timeout instead of http.DefaultClient
	resp, err := authHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("shopline: %s token request failed: %w", endpoint, err)
	}
	defer resp.Body.Close()

	// P1-3: Limit response body size to prevent OOM
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize))
	if err != nil {
		return nil, fmt.Errorf("shopline: failed to read %s response: %w", endpoint, err)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return nil, fmt.Errorf("shopline: failed to parse %s response: %w (body: %s)", endpoint, err, string(respBody))
	}

	if tokenResp.Code != 200 {
		return &tokenResp, fmt.Errorf("shopline: %s token request failed: %s (code: %d, traceId: %s)",
			endpoint, tokenResp.Message, tokenResp.Code, tokenResp.TraceID)
	}

	return &tokenResp, nil
}

// VerifyWebhookRequest verifies the HMAC signature of a Shopline webhook request.
//
// Shopline sends a signature in the X-Shopline-Hmac-SHA256 header.
// The signature is computed over the raw request body using AppSecret.
//
// After verification, the request body is restored so downstream handlers
// can still read it.
func (app App) VerifyWebhookRequest(r *http.Request) bool {
	signature := r.Header.Get("X-Shopline-Hmac-SHA256")
	if signature == "" {
		return false
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxResponseBodySize))
	if err != nil {
		return false
	}
	// P0-2: Restore the body so downstream handlers can read it.
	// Without this, any handler after verification gets an empty body.
	r.Body = io.NopCloser(bytes.NewReader(body))

	mac := hmac.New(sha256.New, []byte(app.AppSecret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expected))
}

// currentTimeMillis returns the current time in milliseconds.
func currentTimeMillis() int64 {
	return timeNow().UnixMilli()
}

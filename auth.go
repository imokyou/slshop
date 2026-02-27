package shopline

import (
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
)

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
	timestamp := fmt.Sprintf("%d", currentTimeMillis())
	signParams := map[string]string{
		"appkey":    app.AppKey,
		"timestamp": timestamp,
	}
	sign := app.GenerateSignature(signParams)

	apiURL := fmt.Sprintf("https://%s.myshopline.com/admin/oauth/token/create", handle)

	body := map[string]string{"code": code}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("shopline: failed to marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, fmt.Errorf("shopline: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("appkey", app.AppKey)
	req.Header.Set("timestamp", timestamp)
	req.Header.Set("sign", sign)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("shopline: token request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("shopline: failed to read token response: %w", err)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return nil, fmt.Errorf("shopline: failed to parse token response: %w", err)
	}

	if tokenResp.Code != 200 {
		return &tokenResp, fmt.Errorf("shopline: token request failed: %s (code: %d)", tokenResp.Message, tokenResp.Code)
	}

	return &tokenResp, nil
}

// RefreshAccessToken refreshes the access token before it expires (10-hour validity).
//
// This corresponds to Step 6 of the Shopline OAuth flow.
// POST https://{handle}.myshopline.com/admin/oauth/token/refresh
func (app App) RefreshAccessToken(ctx context.Context, handle string) (*TokenResponse, error) {
	timestamp := fmt.Sprintf("%d", currentTimeMillis())
	signParams := map[string]string{
		"appkey":    app.AppKey,
		"timestamp": timestamp,
	}
	sign := app.GenerateSignature(signParams)

	apiURL := fmt.Sprintf("https://%s.myshopline.com/admin/oauth/token/refresh", handle)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("shopline: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("appkey", app.AppKey)
	req.Header.Set("timestamp", timestamp)
	req.Header.Set("sign", sign)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("shopline: refresh token request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("shopline: failed to read refresh response: %w", err)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return nil, fmt.Errorf("shopline: failed to parse refresh response: %w", err)
	}

	if tokenResp.Code != 200 {
		return &tokenResp, fmt.Errorf("shopline: refresh token failed: %s (code: %d)", tokenResp.Message, tokenResp.Code)
	}

	return &tokenResp, nil
}

// VerifyWebhookRequest verifies the HMAC signature of a Shopline webhook request.
//
// Shopline sends a signature in the X-Shopline-Hmac-SHA256 header.
// The signature is computed over the raw request body using AppSecret.
func (app App) VerifyWebhookRequest(r *http.Request) bool {
	signature := r.Header.Get("X-Shopline-Hmac-SHA256")
	if signature == "" {
		return false
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(app.AppSecret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expected))
}

// currentTimeMillis returns the current time in milliseconds.
func currentTimeMillis() int64 {
	return timeNow().UnixMilli()
}

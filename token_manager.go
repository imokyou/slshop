package shopline

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	// defaultRefreshBuffer is the default time before expiry to trigger a refresh.
	// Refreshing 5 minutes early avoids edge cases where requests fail because
	// the token expires mid-flight.
	defaultRefreshBuffer = 5 * time.Minute
)

// TokenManager handles automatic token lifecycle management with:
//   - Concurrency-safe refresh: only ONE goroutine refreshes at a time (singleflight via mutex + channel)
//   - Pre-emptive refresh: triggers refresh before the token actually expires
//   - Persistence: saves refreshed tokens to a TokenStore
//
// When multiple goroutines discover the token is expiring simultaneously:
//  1. The first goroutine acquires the "refresher" role by creating a broadcast channel
//  2. Subsequent goroutines see the channel and wait on it via select
//  3. The refresher calls RefreshAccessToken, updates the token, and closes the channel
//  4. All waiting goroutines wake up and use the new token
//
// This eliminates thundering herd problems without external dependencies.
type TokenManager struct {
	app    App
	handle string
	store  TokenStore
	log    Logger

	mu            sync.Mutex
	token         *ManagedToken
	refreshCh     chan struct{} // non-nil while a refresh is in progress; closed when done
	refreshBuffer time.Duration
	initialized   bool // true after first load from store
}

// NewTokenManager creates a TokenManager for the given app and store handle.
//
// Parameters:
//   - app: Application credentials (AppKey + AppSecret needed for refresh)
//   - handle: Store handle (e.g. "open001")
//   - store: TokenStore implementation for persistence
//   - opts: Optional configuration
func NewTokenManager(app App, handle string, store TokenStore, opts ...TokenManagerOption) *TokenManager {
	tm := &TokenManager{
		app:           app,
		handle:        handle,
		store:         store,
		refreshBuffer: defaultRefreshBuffer,
	}
	for _, opt := range opts {
		opt(tm)
	}
	return tm
}

// TokenManagerOption configures a TokenManager.
type TokenManagerOption func(*TokenManager)

// WithRefreshBuffer sets the time before expiry to trigger a proactive refresh.
// Default is 5 minutes.
func WithRefreshBuffer(d time.Duration) TokenManagerOption {
	return func(tm *TokenManager) {
		tm.refreshBuffer = d
	}
}

// WithTokenManagerLogger sets a logger for the TokenManager.
func WithTokenManagerLogger(log Logger) TokenManagerOption {
	return func(tm *TokenManager) {
		tm.log = log
	}
}

// storeKey returns the persistence key for this manager's token.
func (tm *TokenManager) storeKey() string {
	return fmt.Sprintf("%s:%s", tm.handle, tm.app.AppKey)
}

// GetToken returns a valid access token, refreshing automatically if needed.
//
// This method is safe to call from multiple goroutines concurrently.
// Only one goroutine will perform the actual refresh; others wait for it.
func (tm *TokenManager) GetToken(ctx context.Context) (string, error) {
	tm.mu.Lock()

	// First call: try to load from persistent store
	if !tm.initialized {
		tm.initialized = true
		if tm.store != nil {
			tm.mu.Unlock()
			if err := tm.loadFromStore(ctx); err != nil {
				tm.logDebugf("Failed to load token from store: %v", err)
			}
			tm.mu.Lock()
		}
	}

	// Fast path: token is valid and not near expiry
	if tm.token != nil && !tm.token.IsExpiring(tm.refreshBuffer) {
		token := tm.token.AccessToken
		tm.mu.Unlock()
		return token, nil
	}

	// Slow path: need to refresh
	if tm.refreshCh != nil {
		// Another goroutine is already refreshing — wait for it
		ch := tm.refreshCh
		tm.mu.Unlock()
		select {
		case <-ch:
			// Refresh completed (successfully or not), retry
			return tm.GetToken(ctx)
		case <-ctx.Done():
			return "", fmt.Errorf("shopline: context cancelled while waiting for token refresh: %w", ctx.Err())
		}
	}

	// We are the refresher — create the broadcast channel
	tm.refreshCh = make(chan struct{})
	tm.mu.Unlock()

	// Perform the refresh outside the lock
	tm.logDebugf("Refreshing access token for %s", tm.handle)
	newToken, err := tm.doRefresh(ctx)

	tm.mu.Lock()
	if err == nil {
		tm.token = newToken
	}
	ch := tm.refreshCh
	tm.refreshCh = nil
	tm.mu.Unlock()

	// Wake up all waiting goroutines
	close(ch)

	if err != nil {
		return "", fmt.Errorf("shopline: token refresh failed: %w", err)
	}
	return newToken.AccessToken, nil
}

// SetInitialToken sets a token obtained via GetAccessToken (OAuth code exchange).
// This should be called after the initial OAuth flow completes.
func (tm *TokenManager) SetInitialToken(ctx context.Context, accessToken string, expireAt time.Time, scope string) error {
	token := &ManagedToken{
		AccessToken: accessToken,
		ExpireAt:    expireAt,
		Scope:       scope,
	}

	tm.mu.Lock()
	tm.token = token
	tm.initialized = true
	tm.mu.Unlock()

	// Persist to store
	if tm.store != nil {
		if err := tm.store.Set(ctx, tm.storeKey(), token); err != nil {
			return fmt.Errorf("shopline: failed to persist initial token: %w", err)
		}
	}
	return nil
}

// InvalidateToken clears the cached token and removes it from the store.
// Call this when you know the token is revoked or invalid.
func (tm *TokenManager) InvalidateToken(ctx context.Context) error {
	tm.mu.Lock()
	tm.token = nil
	tm.mu.Unlock()

	if tm.store != nil {
		return tm.store.Delete(ctx, tm.storeKey())
	}
	return nil
}

// doRefresh calls the Shopline refresh API and persists the new token.
func (tm *TokenManager) doRefresh(ctx context.Context) (*ManagedToken, error) {
	resp, err := tm.app.RefreshAccessToken(ctx, tm.handle)
	if err != nil {
		return nil, err
	}

	// Parse expiry time from API response
	expireAt, err := time.Parse(time.RFC3339, resp.Data.ExpireTime)
	if err != nil {
		// Fall back to a reasonable default if the format is unexpected
		// Shopline tokens typically expire in 10 hours
		tm.logDebugf("Failed to parse expire time %q, using 10h default: %v", resp.Data.ExpireTime, err)
		expireAt = time.Now().Add(10 * time.Hour)
	}

	token := &ManagedToken{
		AccessToken: resp.Data.AccessToken,
		ExpireAt:    expireAt,
		Scope:       resp.Data.Scope,
	}

	// Persist to store
	if tm.store != nil {
		if err := tm.store.Set(ctx, tm.storeKey(), token); err != nil {
			tm.logDebugf("Failed to persist refreshed token: %v", err)
			// Don't fail the refresh — the token is still valid in memory
		}
	}

	tm.logDebugf("Token refreshed successfully, expires at %s", expireAt.Format(time.RFC3339))
	return token, nil
}

// loadFromStore loads a token from the persistent store into memory.
func (tm *TokenManager) loadFromStore(ctx context.Context) error {
	token, err := tm.store.Get(ctx, tm.storeKey())
	if err != nil {
		return err
	}
	if token == nil {
		return nil // no persisted token
	}
	if token.IsExpired() {
		tm.logDebugf("Persisted token is expired, will refresh on next GetToken call")
		return nil
	}

	tm.mu.Lock()
	tm.token = token
	tm.mu.Unlock()

	tm.logDebugf("Loaded token from store, expires at %s", token.ExpireAt.Format(time.RFC3339))
	return nil
}

// logDebugf logs a debug message if a logger is set.
func (tm *TokenManager) logDebugf(format string, args ...interface{}) {
	if tm.log != nil {
		tm.log.Debugf(format, args...)
	}
}

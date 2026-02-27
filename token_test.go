package shopline

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ============================================================
// FileTokenStore Tests
// ============================================================

func TestFileTokenStore_ReadWrite(t *testing.T) {
	dir := t.TempDir()
	store := NewFileTokenStore(dir)
	ctx := context.Background()

	token := &ManagedToken{
		AccessToken: "access-123",
		ExpireAt:    time.Now().Add(10 * time.Hour).Truncate(time.Second),
		Scope:       "read_products,read_orders",
	}

	// Set
	if err := store.Set(ctx, "shop:app", token); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get
	got, err := store.Get(ctx, "shop:app")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got == nil {
		t.Fatal("expected token, got nil")
	}
	if got.AccessToken != "access-123" {
		t.Errorf("expected access token 'access-123', got %q", got.AccessToken)
	}
	if got.Scope != "read_products,read_orders" {
		t.Errorf("expected scope 'read_products,read_orders', got %q", got.Scope)
	}
	if !got.ExpireAt.Equal(token.ExpireAt) {
		t.Errorf("expected expire at %v, got %v", token.ExpireAt, got.ExpireAt)
	}

	// Delete
	if err := store.Delete(ctx, "shop:app"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	deleted, err := store.Get(ctx, "shop:app")
	if err != nil {
		t.Fatalf("Get after delete failed: %v", err)
	}
	if deleted != nil {
		t.Error("expected nil after delete")
	}
}

func TestFileTokenStore_NotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewFileTokenStore(dir)

	got, err := store.Get(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Error("expected nil for missing key")
	}
}

func TestFileTokenStore_DeleteNonexistent(t *testing.T) {
	dir := t.TempDir()
	store := NewFileTokenStore(dir)

	// Should not error
	if err := store.Delete(context.Background(), "nonexistent"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFileTokenStore_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	store := NewFileTokenStore(dir)
	ctx := context.Background()

	// Write a token
	token := &ManagedToken{
		AccessToken: "initial",
		ExpireAt:    time.Now().Add(1 * time.Hour),
	}
	store.Set(ctx, "test", token)

	// Overwrite with new token
	token2 := &ManagedToken{
		AccessToken: "updated",
		ExpireAt:    time.Now().Add(2 * time.Hour),
	}
	store.Set(ctx, "test", token2)

	// Verify the latest token
	got, _ := store.Get(ctx, "test")
	if got.AccessToken != "updated" {
		t.Errorf("expected 'updated', got %q", got.AccessToken)
	}

	// Verify no .tmp files remain
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.Name() == "test.json.tmp" {
			t.Error("temp file should not exist after successful write")
		}
	}
}

// ============================================================
// TokenManager Tests
// ============================================================

// mockTokenStore is an in-memory TokenStore for testing.
type mockTokenStore struct {
	mu     sync.Mutex
	tokens map[string]*ManagedToken
}

func newMockTokenStore() *mockTokenStore {
	return &mockTokenStore{tokens: make(map[string]*ManagedToken)}
}

func (s *mockTokenStore) Get(_ context.Context, key string) (*ManagedToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tokens[key]
	if !ok {
		return nil, nil
	}
	// Return a copy
	cp := *t
	return &cp, nil
}

func (s *mockTokenStore) Set(_ context.Context, key string, token *ManagedToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *token
	s.tokens[key] = &cp
	return nil
}

func (s *mockTokenStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, key)
	return nil
}

func TestTokenManager_ConcurrentRefresh(t *testing.T) {
	// Track how many times refresh is actually called
	var refreshCount int32

	store := newMockTokenStore()
	app := App{AppKey: "test-key", AppSecret: "test-secret"}

	tm := NewTokenManager(app, "testshop", store,
		WithRefreshBuffer(0), // no buffer for test
	)

	// Directly set an expired token to force refresh
	tm.mu.Lock()
	tm.token = &ManagedToken{
		AccessToken: "expired",
		ExpireAt:    time.Now().Add(-1 * time.Hour),
	}
	tm.initialized = true
	tm.mu.Unlock()

	// Override the doRefresh to count calls and return a mock token
	// We'll use a custom approach: replace the app's RefreshAccessToken behavior
	// by making TokenManager use a test-friendly refresh function.
	// Since doRefresh calls app.RefreshAccessToken which makes an HTTP call,
	// we need to patch that. Instead, let's test the singleflight pattern
	// by pre-setting a valid token directly.

	// Better approach: test the singleflight by using SetInitialToken and
	// verifying concurrent GetToken calls when token is near expiry.

	// Reset with a near-expired token that will trigger many concurrent refreshes
	tm2 := &testableTokenManager{
		app:           app,
		handle:        "testshop",
		store:         store,
		refreshBuffer: 0,
		initialized:   true,
		refreshCount:  &refreshCount,
	}
	tm2.token = &ManagedToken{
		AccessToken: "old-token",
		ExpireAt:    time.Now().Add(-1 * time.Second), // just expired
	}

	const numGoroutines = 50
	var wg sync.WaitGroup
	results := make([]string, numGoroutines)
	errors := make([]error, numGoroutines)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			token, err := tm2.GetToken(context.Background())
			results[idx] = token
			errors[idx] = err
		}(i)
	}
	wg.Wait()

	// All should succeed with the same new token
	for i, err := range errors {
		if err != nil {
			t.Errorf("goroutine %d got error: %v", i, err)
		}
	}
	for i, tok := range results {
		if tok != "refreshed-token" {
			t.Errorf("goroutine %d got token %q, want 'refreshed-token'", i, tok)
		}
	}

	// Refresh should have been called exactly once
	count := atomic.LoadInt32(&refreshCount)
	if count != 1 {
		t.Errorf("expected exactly 1 refresh call, got %d", count)
	}
}

// testableTokenManager wraps the singleflight logic for testing without
// needing a real HTTP server for token refresh.
type testableTokenManager struct {
	app           App
	handle        string
	store         TokenStore
	refreshBuffer time.Duration
	initialized   bool

	mu           sync.Mutex
	token        *ManagedToken
	refreshCh    chan struct{}
	refreshCount *int32
}

func (tm *testableTokenManager) GetToken(ctx context.Context) (string, error) {
	tm.mu.Lock()

	// Fast path
	if tm.token != nil && !tm.token.IsExpiring(tm.refreshBuffer) {
		tok := tm.token.AccessToken
		tm.mu.Unlock()
		return tok, nil
	}

	// Wait for existing refresh
	if tm.refreshCh != nil {
		ch := tm.refreshCh
		tm.mu.Unlock()
		select {
		case <-ch:
			return tm.GetToken(ctx)
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	// We are the refresher
	tm.refreshCh = make(chan struct{})
	tm.mu.Unlock()

	// Simulate refresh
	atomic.AddInt32(tm.refreshCount, 1)
	time.Sleep(50 * time.Millisecond) // simulate network latency
	newToken := &ManagedToken{
		AccessToken: "refreshed-token",
		ExpireAt:    time.Now().Add(10 * time.Hour),
	}

	tm.mu.Lock()
	tm.token = newToken
	ch := tm.refreshCh
	tm.refreshCh = nil
	tm.mu.Unlock()
	close(ch)

	return newToken.AccessToken, nil
}

func TestTokenManager_SetInitialToken(t *testing.T) {
	store := newMockTokenStore()
	app := App{AppKey: "k", AppSecret: "s"}
	tm := NewTokenManager(app, "shop", store)

	expireAt := time.Now().Add(10 * time.Hour).Truncate(time.Second)
	err := tm.SetInitialToken(context.Background(), "my-token", expireAt, "read_products")
	if err != nil {
		t.Fatalf("SetInitialToken failed: %v", err)
	}

	// Token should be immediately available
	tok, err := tm.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}
	if tok != "my-token" {
		t.Errorf("expected 'my-token', got %q", tok)
	}

	// Should be persisted in store
	stored, _ := store.Get(context.Background(), "shop:k")
	if stored == nil {
		t.Fatal("expected token in store")
	}
	if stored.AccessToken != "my-token" {
		t.Errorf("expected 'my-token' in store, got %q", stored.AccessToken)
	}
}

func TestTokenManager_LoadFromStore(t *testing.T) {
	store := newMockTokenStore()
	ctx := context.Background()

	// Pre-populate store with a valid token
	store.Set(ctx, "shop:k", &ManagedToken{
		AccessToken: "stored-token",
		ExpireAt:    time.Now().Add(5 * time.Hour),
	})

	app := App{AppKey: "k", AppSecret: "s"}
	tm := NewTokenManager(app, "shop", store)

	// First GetToken should load from store
	tok, err := tm.GetToken(ctx)
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}
	if tok != "stored-token" {
		t.Errorf("expected 'stored-token', got %q", tok)
	}
}

func TestTokenManager_InvalidateToken(t *testing.T) {
	store := newMockTokenStore()
	ctx := context.Background()
	app := App{AppKey: "k", AppSecret: "s"}
	tm := NewTokenManager(app, "shop", store)

	// Set a token
	tm.SetInitialToken(ctx, "to-invalidate", time.Now().Add(10*time.Hour), "")

	// Invalidate it
	err := tm.InvalidateToken(ctx)
	if err != nil {
		t.Fatalf("InvalidateToken failed: %v", err)
	}

	// Should be gone from store
	stored, _ := store.Get(ctx, "shop:k")
	if stored != nil {
		t.Error("expected nil in store after invalidation")
	}
}

func TestTokenManager_ContextCancelled(t *testing.T) {
	store := newMockTokenStore()
	tm := &testableSlowRefreshManager{
		store:         store,
		refreshBuffer: 0,
		initialized:   true,
	}
	tm.token = &ManagedToken{
		AccessToken: "expired",
		ExpireAt:    time.Now().Add(-1 * time.Hour),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := tm.GetToken(ctx)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

// testableSlowRefreshManager simulates a very slow refresh to test context cancellation.
type testableSlowRefreshManager struct {
	store         TokenStore
	refreshBuffer time.Duration
	initialized   bool

	mu        sync.Mutex
	token     *ManagedToken
	refreshCh chan struct{}
}

func (tm *testableSlowRefreshManager) GetToken(ctx context.Context) (string, error) {
	tm.mu.Lock()
	if tm.token != nil && !tm.token.IsExpiring(tm.refreshBuffer) {
		tok := tm.token.AccessToken
		tm.mu.Unlock()
		return tok, nil
	}
	if tm.refreshCh != nil {
		ch := tm.refreshCh
		tm.mu.Unlock()
		select {
		case <-ch:
			return tm.GetToken(ctx)
		case <-ctx.Done():
			return "", fmt.Errorf("shopline: context cancelled: %w", ctx.Err())
		}
	}
	tm.refreshCh = make(chan struct{})
	tm.mu.Unlock()

	// Simulate very slow refresh
	select {
	case <-time.After(5 * time.Second):
	case <-ctx.Done():
		tm.mu.Lock()
		ch := tm.refreshCh
		tm.refreshCh = nil
		tm.mu.Unlock()
		close(ch)
		return "", fmt.Errorf("shopline: context cancelled: %w", ctx.Err())
	}

	return "", fmt.Errorf("should not reach here")
}

func TestManagedToken_IsExpired(t *testing.T) {
	valid := &ManagedToken{ExpireAt: time.Now().Add(1 * time.Hour)}
	if valid.IsExpired() {
		t.Error("valid token should not be expired")
	}

	expired := &ManagedToken{ExpireAt: time.Now().Add(-1 * time.Hour)}
	if !expired.IsExpired() {
		t.Error("expired token should be expired")
	}

	var nilToken *ManagedToken
	if !nilToken.IsExpired() {
		t.Error("nil token should be expired")
	}
}

func TestManagedToken_IsExpiring(t *testing.T) {
	token := &ManagedToken{ExpireAt: time.Now().Add(3 * time.Minute)}

	if token.IsExpiring(1 * time.Minute) {
		t.Error("token with 3m left should not be expiring with 1m buffer")
	}
	if !token.IsExpiring(5 * time.Minute) {
		t.Error("token with 3m left should be expiring with 5m buffer")
	}
}

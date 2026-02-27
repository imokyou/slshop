package shopline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ManagedToken represents a token managed by TokenManager with expiry tracking.
type ManagedToken struct {
	AccessToken string    `json:"access_token"`
	ExpireAt    time.Time `json:"expire_at"`
	Scope       string    `json:"scope,omitempty"`
}

// IsExpired returns true if the token has expired.
func (t *ManagedToken) IsExpired() bool {
	return t == nil || time.Now().After(t.ExpireAt)
}

// IsExpiring returns true if the token will expire within the given buffer duration.
func (t *ManagedToken) IsExpiring(buffer time.Duration) bool {
	return t == nil || time.Now().Add(buffer).After(t.ExpireAt)
}

// TokenStore defines the interface for token persistence.
// Users can implement this for any backend (Redis, MySQL, etc.).
//
// The key is typically "handle:appkey" to support multi-store scenarios.
//
// Example Redis implementation:
//
//	type RedisTokenStore struct { client *redis.Client }
//	func (s *RedisTokenStore) Get(ctx context.Context, key string) (*ManagedToken, error) { ... }
//	func (s *RedisTokenStore) Set(ctx context.Context, key string, token *ManagedToken) error { ... }
//	func (s *RedisTokenStore) Delete(ctx context.Context, key string) error { ... }
type TokenStore interface {
	// Get retrieves a token by key. Returns (nil, nil) if not found.
	Get(ctx context.Context, key string) (*ManagedToken, error)

	// Set persists a token with the given key.
	Set(ctx context.Context, key string, token *ManagedToken) error

	// Delete removes a token by key.
	Delete(ctx context.Context, key string) error
}

// ============================================================
// FileTokenStore — built-in file-based implementation
// ============================================================

// FileTokenStore persists tokens as JSON files in a directory.
// Each key maps to a file named {sanitized_key}.json.
//
// This is suitable for local development and single-process deployments.
// For multi-process or distributed environments, implement TokenStore
// with a shared backend like Redis.
type FileTokenStore struct {
	dir string
	mu  sync.Mutex // serialize file writes to prevent corruption
}

// NewFileTokenStore creates a FileTokenStore that stores tokens in the given directory.
// The directory will be created if it does not exist.
func NewFileTokenStore(dir string) *FileTokenStore {
	return &FileTokenStore{dir: dir}
}

// Get reads a token from a JSON file.
func (s *FileTokenStore) Get(_ context.Context, key string) (*ManagedToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.filePath(key)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // not found is not an error
		}
		return nil, fmt.Errorf("shopline: failed to read token file %s: %w", path, err)
	}

	var token ManagedToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("shopline: failed to parse token file %s: %w", path, err)
	}

	return &token, nil
}

// Set writes a token to a JSON file atomically (write-to-temp + rename).
func (s *FileTokenStore) Set(_ context.Context, key string, token *ManagedToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure directory exists
	if err := os.MkdirAll(s.dir, 0700); err != nil {
		return fmt.Errorf("shopline: failed to create token directory: %w", err)
	}

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("shopline: failed to marshal token: %w", err)
	}

	path := s.filePath(key)

	// Atomic write: write to temp file, then rename
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("shopline: failed to write token file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath) // clean up temp file on failure
		return fmt.Errorf("shopline: failed to rename token file: %w", err)
	}

	return nil
}

// Delete removes a token file.
func (s *FileTokenStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.filePath(key)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("shopline: failed to delete token file: %w", err)
	}
	return nil
}

// filePath returns the file path for a given key, sanitizing special characters.
func (s *FileTokenStore) filePath(key string) string {
	// Replace characters that are problematic in filenames
	safe := strings.NewReplacer(":", "_", "/", "_", "\\", "_").Replace(key)
	return filepath.Join(s.dir, safe+".json")
}

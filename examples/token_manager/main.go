// Example: Automatic Token Management with Persistence
//
// This example demonstrates how to use TokenManager for:
//   - Persisting tokens to local files (survives process restarts)
//   - Automatic token refresh before expiry
//   - Concurrency-safe refresh (singleflight pattern)
//
// TokenManager makes token lifecycle completely transparent —
// your business code never needs to worry about token expiry.
//
// Usage:
//
//	export SHOPLINE_APP_KEY="your-app-key"
//	export SHOPLINE_APP_SECRET="your-app-secret"
//	export SHOPLINE_HANDLE="your-store-handle"
//	export SHOPLINE_TOKEN="your-initial-access-token"
//	go run examples/token_manager/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	shopline "github.com/imokyou/slshop"
)

// simpleLogger implements shopline.Logger for demo purposes.
type simpleLogger struct{}

func (l *simpleLogger) Debugf(format string, args ...interface{}) {
	log.Printf("[DEBUG] "+format, args...)
}
func (l *simpleLogger) Infof(format string, args ...interface{}) {
	log.Printf("[INFO]  "+format, args...)
}
func (l *simpleLogger) Errorf(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}

func main() {
	app := shopline.App{
		AppKey:    os.Getenv("SHOPLINE_APP_KEY"),
		AppSecret: os.Getenv("SHOPLINE_APP_SECRET"),
	}

	handle := os.Getenv("SHOPLINE_HANDLE")
	initialToken := os.Getenv("SHOPLINE_TOKEN")

	if handle == "" {
		log.Fatal("Please set SHOPLINE_HANDLE")
	}

	// ============================
	// Method 1: FileTokenStore (Local Development)
	// ============================
	fmt.Println("=== Using FileTokenStore ===")

	// Tokens will be saved as JSON files in ./tokens/ directory
	store := shopline.NewFileTokenStore("./tokens")

	client, err := shopline.NewClient(app, handle, "",
		shopline.WithTokenManager(store),
		shopline.WithLogger(&simpleLogger{}),
		shopline.WithRetry(3),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// If you have an initial token from OAuth, seed it:
	if initialToken != "" {
		fmt.Println("Seeding initial token...")
		// In production, parse the actual expireTime from the OAuth response.
		// Here we use 10 hours as Shopline's default token lifetime.
		expireAt := time.Now().Add(10 * time.Hour)
		if err := client.TokenManager().SetInitialToken(ctx, initialToken, expireAt, ""); err != nil {
			log.Fatalf("Failed to set initial token: %v", err)
		}
		fmt.Println("Token seeded and persisted to ./tokens/")
	}

	// Now make API calls — token management is fully transparent!
	shop, err := client.Store.GetShop(ctx)
	if err != nil {
		log.Printf("Failed to get shop: %v", err)
	} else {
		fmt.Printf("Shop: %s (%s)\n", shop.Name, shop.Domain)
	}

	// ============================
	// Method 2: Concurrent Access Pattern
	// ============================
	fmt.Println("\n=== Concurrent Access Demo ===")
	fmt.Println("Launching 10 concurrent API calls...")

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			products, err := client.Product.List(ctx, nil)
			if err != nil {
				log.Printf("  Worker %d: error: %v", id, err)
			} else {
				fmt.Printf("  Worker %d: fetched %d products\n", id, len(products))
			}
		}(i)
	}
	wg.Wait()

	fmt.Println("\nAll workers completed!")

	// ============================
	// Notes on Custom TokenStore (Redis, MySQL, etc.)
	// ============================
	fmt.Print(`
=== Custom TokenStore ===
For production environments, implement the shopline.TokenStore interface
with your preferred backend:

    type RedisTokenStore struct {
        client *redis.Client
    }

    func (s *RedisTokenStore) Get(ctx context.Context, key string) (*shopline.ManagedToken, error) {
        data, err := s.client.Get(ctx, "shopline:token:"+key).Bytes()
        if err == redis.Nil { return nil, nil }
        if err != nil { return nil, err }
        var token shopline.ManagedToken
        json.Unmarshal(data, &token)
        return &token, nil
    }

    func (s *RedisTokenStore) Set(ctx context.Context, key string, token *shopline.ManagedToken) error {
        data, _ := json.Marshal(token)
        ttl := time.Until(token.ExpireAt)
        return s.client.Set(ctx, "shopline:token:"+key, data, ttl).Err()
    }

    func (s *RedisTokenStore) Delete(ctx context.Context, key string) error {
        return s.client.Del(ctx, "shopline:token:"+key).Err()
    }
`)
}

// Example: Complete OAuth2 Authorization Flow
//
// This example demonstrates the full Shopline OAuth2 flow:
//  1. Generate the authorization URL
//  2. Start a local HTTP server to receive the callback
//  3. Exchange the authorization code for an access token
//  4. Make an API call with the new token
//
// Usage:
//
//	export SHOPLINE_APP_KEY="your-app-key"
//	export SHOPLINE_APP_SECRET="your-app-secret"
//	export SHOPLINE_HANDLE="your-store-handle"
//	go run examples/oauth/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	shopline "github.com/imokyou/slshop"
)

func main() {
	app := shopline.App{
		AppKey:      os.Getenv("SHOPLINE_APP_KEY"),
		AppSecret:   os.Getenv("SHOPLINE_APP_SECRET"),
		RedirectURL: "http://localhost:9090/callback",
		Scope:       "read_products,read_orders,read_customers",
	}

	handle := os.Getenv("SHOPLINE_HANDLE")
	if handle == "" {
		log.Fatal("Please set SHOPLINE_HANDLE environment variable")
	}

	// ============================
	// Step 1: Generate Auth URL
	// ============================
	nonce := fmt.Sprintf("state_%d", time.Now().UnixNano())
	authURL := app.AuthorizeURL(handle, nonce)
	fmt.Println("================================")
	fmt.Println("Please open the following URL in your browser to authorize the app:")
	fmt.Println()
	fmt.Println(authURL)
	fmt.Println()
	fmt.Println("Waiting for callback on http://localhost:9090/callback ...")
	fmt.Println("================================")

	// ============================
	// Step 2: Wait for Callback
	// ============================
	codeCh := make(chan string, 1)

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Verify the signature from Shopline
		if !app.VerifySignature(r.URL.Query()) {
			http.Error(w, "Invalid signature", http.StatusForbidden)
			log.Println("WARNING: Received callback with invalid signature!")
			return
		}

		code := r.URL.Query().Get("code")
		customField := r.URL.Query().Get("customField")

		// Verify state to prevent CSRF
		if customField != nonce {
			http.Error(w, "Invalid state", http.StatusForbidden)
			log.Printf("WARNING: State mismatch: expected %q, got %q\n", nonce, customField)
			return
		}

		fmt.Fprintf(w, "Authorization successful! You can close this tab.")
		codeCh <- code
	})

	server := &http.Server{Addr: ":9090"}
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	code := <-codeCh
	server.Shutdown(context.Background())
	fmt.Printf("\nReceived authorization code: %s\n\n", code)

	// ============================
	// Step 3: Exchange Code for Token
	// ============================
	ctx := context.Background()
	tokenResp, err := app.GetAccessToken(ctx, handle, code)
	if err != nil {
		log.Fatalf("Failed to get access token: %v", err)
	}

	fmt.Println("=== Token Response ===")
	fmt.Printf("Access Token: %s...\n", tokenResp.Data.AccessToken[:20])
	fmt.Printf("Expires At:   %s\n", tokenResp.Data.ExpireTime)
	fmt.Printf("Scope:        %s\n", tokenResp.Data.Scope)

	// ============================
	// Step 4: Use the Token
	// ============================
	client, err := shopline.NewClient(app, handle, tokenResp.Data.AccessToken,
		shopline.WithRetry(3),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	shop, err := client.Store.GetShop(ctx)
	if err != nil {
		log.Fatalf("Failed to get shop info: %v", err)
	}

	fmt.Printf("\n=== Shop Info ===\n")
	fmt.Printf("Name:   %s\n", shop.Name)
	fmt.Printf("Domain: %s\n", shop.Domain)

	fmt.Println("\nOAuth flow completed successfully! 🎉")
}

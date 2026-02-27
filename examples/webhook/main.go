// Example: Webhook Verification and Processing
//
// This example demonstrates how to:
//  1. Verify webhook signatures from Shopline
//  2. Process different webhook topics
//  3. Handle the request body safely (body is preserved after verification)
//
// Usage:
//
//	export SHOPLINE_APP_SECRET="your-app-secret"
//	go run examples/webhook/main.go
//
// Test with curl:
//
//	# Generate HMAC for testing (replace with real signature in production)
//	echo -n '{"id":123,"topic":"orders/create"}' | openssl dgst -sha256 -hmac "your-app-secret"
//	curl -X POST http://localhost:8080/webhook \
//	  -H "Content-Type: application/json" \
//	  -H "X-Shopline-Hmac-SHA256: <hmac-from-above>" \
//	  -d '{"id":123,"topic":"orders/create"}'
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	shopline "github.com/imokyou/slshop"
)

func main() {
	appSecret := os.Getenv("SHOPLINE_APP_SECRET")
	if appSecret == "" {
		log.Fatal("Please set SHOPLINE_APP_SECRET environment variable")
	}

	app := shopline.App{
		AppSecret: appSecret,
	}

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// ============================
		// Step 1: Verify Signature
		// ============================
		// VerifyWebhookRequest reads the body, computes HMAC-SHA256,
		// and RESTORES the body so you can still read it afterward.
		if !app.VerifyWebhookRequest(r) {
			log.Println("⚠️  Webhook signature verification FAILED")
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
		log.Println("✅ Webhook signature verified")

		// ============================
		// Step 2: Read & Parse Body
		// ============================
		// Body is still available because VerifyWebhookRequest restores it!
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// ============================
		// Step 3: Route by Topic
		// ============================
		topic := r.Header.Get("X-Shopline-Topic")
		log.Printf("📨 Received webhook: topic=%s", topic)

		switch topic {
		case "orders/create":
			handleOrderCreate(payload)
		case "orders/updated":
			handleOrderUpdate(payload)
		case "orders/cancelled":
			handleOrderCancel(payload)
		case "products/create":
			handleProductCreate(payload)
		case "products/update":
			handleProductUpdate(payload)
		case "app/uninstalled":
			handleAppUninstalled(payload)
		default:
			log.Printf("   Unhandled topic: %s", topic)
		}

		// Always respond 200 OK quickly — Shopline will retry on non-2xx
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	addr := ":8080"
	fmt.Printf("🚀 Webhook server listening on %s\n", addr)
	fmt.Println("Endpoints:")
	fmt.Println("  POST /webhook  — Shopline webhook receiver")
	fmt.Println("  GET  /health   — Health check")
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleOrderCreate(payload map[string]interface{}) {
	log.Printf("   📦 New order created: %v", payload["id"])
	// TODO: Process new order (sync to ERP, send notification, etc.)
}

func handleOrderUpdate(payload map[string]interface{}) {
	log.Printf("   📦 Order updated: %v", payload["id"])
}

func handleOrderCancel(payload map[string]interface{}) {
	log.Printf("   ❌ Order cancelled: %v", payload["id"])
}

func handleProductCreate(payload map[string]interface{}) {
	log.Printf("   🛍️  New product created: %v", payload["id"])
}

func handleProductUpdate(payload map[string]interface{}) {
	log.Printf("   🛍️  Product updated: %v", payload["id"])
}

func handleAppUninstalled(payload map[string]interface{}) {
	log.Printf("   🗑️  App uninstalled by merchant")
	// TODO: Clean up merchant data, revoke tokens, etc.
}

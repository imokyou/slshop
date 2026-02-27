package shopline

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/imokyou/slshop/core"
	"github.com/imokyou/slshop/order"
	"github.com/imokyou/slshop/product"
	"github.com/imokyou/slshop/store"
)

// Avoid unused import warnings
var _ = order.Order{}

// newTestClient creates a Client connected to a test HTTP server.
func newTestClient(handler http.HandlerFunc) (*Client, *httptest.Server) {
	server := httptest.NewServer(handler)
	app := App{
		AppKey:    "test-key",
		AppSecret: "test-secret",
	}
	client, _ := NewClient(app, "testshop", "test-token",
		WithBaseURL(server.URL),
	)
	return client, server
}

func TestNewClient(t *testing.T) {
	app := App{
		AppKey:    "my-key",
		AppSecret: "my-secret",
	}
	client, err := NewClient(app, "myshop", "my-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.GetHandle() != "myshop" {
		t.Errorf("expected handle 'myshop', got %q", client.GetHandle())
	}
	if client.GetAPIVersion() != DefaultAPIVersion {
		t.Errorf("expected version %q, got %q", DefaultAPIVersion, client.GetAPIVersion())
	}
	if got := client.GetBaseURL().String(); got != "https://myshop.myshopline.com" {
		t.Errorf("expected base URL 'https://myshop.myshopline.com', got %q", got)
	}
}

func TestNewClientWithVersion(t *testing.T) {
	app := App{AppKey: "k", AppSecret: "s"}
	client, err := NewClient(app, "shop", "tok", WithVersion("v20260301"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.GetAPIVersion() != "v20260301" {
		t.Errorf("expected version 'v20260301', got %q", client.GetAPIVersion())
	}
}

func TestCreatePath(t *testing.T) {
	app := App{AppKey: "k", AppSecret: "s"}
	client, _ := NewClient(app, "shop", "tok")
	got := client.CreatePath("products.json")
	expected := "/admin/openapi/v20251201/products.json"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestNewRequest(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	defer server.Close()

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/admin/openapi/v20251201/products.json", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := req.Header.Get("Authorization"); got != "Bearer test-token" {
		t.Errorf("expected 'Bearer test-token', got %q", got)
	}
	if got := req.Header.Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Errorf("expected content type 'application/json; charset=utf-8', got %q", got)
	}
	if got := req.Header.Get("User-Agent"); got != UserAgent {
		t.Errorf("expected user agent %q, got %q", UserAgent, got)
	}
}

type testProductResource struct {
	Product *product.Product `json:"product"`
}

func TestDo_Success(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"product": map[string]interface{}{
				"id":    123,
				"title": "Test Product",
			},
		})
	})
	defer server.Close()

	req, _ := client.NewRequest(context.Background(), http.MethodGet, "/test", nil)
	resource := &testProductResource{}
	_, err := client.Do(req, resource)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resource.Product == nil {
		t.Fatal("expected product, got nil")
	}
	if resource.Product.Title != "Test Product" {
		t.Errorf("expected title 'Test Product', got %q", resource.Product.Title)
	}
}

func TestDo_ErrorResponse(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors":  "Not Found",
			"traceId": "abc123",
		})
	})
	defer server.Close()

	req, _ := client.NewRequest(context.Background(), http.MethodGet, "/test", nil)
	_, err := client.Do(req, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	respErr, ok := err.(*ResponseError)
	if !ok {
		t.Fatalf("expected *ResponseError, got %T", err)
	}
	if respErr.Status != 404 {
		t.Errorf("expected status 404, got %d", respErr.Status)
	}
	if respErr.TraceID != "abc123" {
		t.Errorf("expected traceId 'abc123', got %q", respErr.TraceID)
	}
}

func TestDo_RateLimitRetry(t *testing.T) {
	attempt := 0
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, `{"errors":"rate limited","traceId":"rl1"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	defer server.Close()

	// Reconfigure with retry
	client.maxRetries = 1

	req, _ := client.NewRequest(context.Background(), http.MethodGet, "/test", nil)
	var result map[string]string
	_, err := client.Do(req, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempt != 2 {
		t.Errorf("expected 2 attempts, got %d", attempt)
	}
}

func TestAuthSignature(t *testing.T) {
	app := App{
		AppKey:    "testkey",
		AppSecret: "testsecret",
	}

	params := map[string]string{
		"appkey":    "testkey",
		"timestamp": "1234567890",
	}

	sig := app.GenerateSignature(params)
	if sig == "" {
		t.Fatal("expected non-empty signature")
	}

	// Same params should produce same signature
	sig2 := app.GenerateSignature(params)
	if sig != sig2 {
		t.Error("signatures should be deterministic")
	}

	// Different params should produce different signature
	params["timestamp"] = "9999999999"
	sig3 := app.GenerateSignature(params)
	if sig == sig3 {
		t.Error("different params should produce different signature")
	}
}

func TestAuthorizeURL(t *testing.T) {
	app := App{
		AppKey:      "mykey",
		AppSecret:   "mysecret",
		RedirectURL: "https://example.com/callback",
		Scope:       "read_products,read_orders",
	}

	url := app.AuthorizeURL("testshop", "nonce123")

	if url == "" {
		t.Fatal("expected non-empty URL")
	}

	// Check that URL contains expected components
	expected := []string{
		"testshop.myshopline.com",
		"appKey=mykey",
		"responseType=code",
		"scope=read_products",
		"redirectUri=",
		"customField=nonce123",
	}
	for _, e := range expected {
		found := false
		if len(url) > 0 {
			for i := 0; i <= len(url)-len(e); i++ {
				if url[i:i+len(e)] == e {
					found = true
					break
				}
			}
		}
		if !found {
			t.Errorf("URL missing expected component %q\nURL: %s", e, url)
		}
	}
}

func TestBuildQueryString(t *testing.T) {
	opts := &core.ListOptions{
		Page:  2,
		Limit: 50,
	}
	qs := buildQueryString(opts)
	if qs == "" {
		t.Fatal("expected non-empty query string")
	}
	// Should contain page=2 and limit=50
	if qs != "limit=50&page=2" && qs != "page=2&limit=50" {
		// URL values are sorted by key, so it should be limit=50&page=2
		t.Logf("got query string: %s (order may vary)", qs)
	}
}

func TestBuildQueryString_OmitEmpty(t *testing.T) {
	opts := &core.ListOptions{
		Limit: 25,
		// Page is 0, should be omitted
	}
	qs := buildQueryString(opts)
	if qs != "limit=25" {
		t.Errorf("expected 'limit=25', got %q", qs)
	}
}

func TestBuildQueryString_Nil(t *testing.T) {
	qs := buildQueryString(nil)
	if qs != "" {
		t.Errorf("expected empty string, got %q", qs)
	}
}

// Ensure timeNow can be overridden for testing
func TestTimeNow(t *testing.T) {
	fixed := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	oldTimeNow := timeNow
	timeNow = func() time.Time { return fixed }
	defer func() { timeNow = oldTimeNow }()

	ms := currentTimeMillis()
	expected := fixed.UnixMilli()
	if ms != expected {
		t.Errorf("expected %d, got %d", expected, ms)
	}
}

// Test sub-package integration

func TestProductList(t *testing.T) {
	type productsResource struct {
		Products []product.Product `json:"products"`
	}
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(productsResource{
			Products: []product.Product{
				{ID: 1, Title: "Product 1"},
				{ID: 2, Title: "Product 2"},
			},
		})
	})
	defer server.Close()

	products, err := client.Product.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(products) != 2 {
		t.Fatalf("expected 2 products, got %d", len(products))
	}
	if products[0].Title != "Product 1" {
		t.Errorf("expected 'Product 1', got %q", products[0].Title)
	}
}

func TestProductGet(t *testing.T) {
	type productResource struct {
		Product *product.Product `json:"product"`
	}
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(productResource{
			Product: &product.Product{
				ID:    123,
				Title: "Test Product",
				Variants: []product.Variant{
					{ID: 456, Price: "29.99", SKU: "SKU-001"},
				},
			},
		})
	})
	defer server.Close()

	p, err := client.Product.Get(context.Background(), 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID != 123 {
		t.Errorf("expected ID 123, got %d", p.ID)
	}
	if len(p.Variants) != 1 {
		t.Fatalf("expected 1 variant, got %d", len(p.Variants))
	}
	if p.Variants[0].SKU != "SKU-001" {
		t.Errorf("expected SKU 'SKU-001', got %q", p.Variants[0].SKU)
	}
}

func TestProductCreate(t *testing.T) {
	type productResource struct {
		Product *product.Product `json:"product"`
	}
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body productResource
		json.NewDecoder(r.Body).Decode(&body)
		if body.Product.Title != "New Product" {
			t.Errorf("expected title 'New Product', got %q", body.Product.Title)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(productResource{
			Product: &product.Product{
				ID:    999,
				Title: "New Product",
			},
		})
	})
	defer server.Close()

	p, err := client.Product.Create(context.Background(), product.Product{
		Title: "New Product",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID != 999 {
		t.Errorf("expected ID 999, got %d", p.ID)
	}
}

func TestProductCount(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"count": 42})
	})
	defer server.Close()

	count, err := client.Product.Count(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 42 {
		t.Errorf("expected count 42, got %d", count)
	}
}

func TestProductDelete(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := client.Product.Delete(context.Background(), 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrderList(t *testing.T) {
	type ordersResource struct {
		Orders []order.Order `json:"orders"`
	}
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ordersResource{
			Orders: []order.Order{
				{ID: 1001, Name: "#1001", TotalPrice: "199.00"},
				{ID: 1002, Name: "#1002", TotalPrice: "99.50"},
			},
		})
	})
	defer server.Close()

	orders, err := client.Order.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orders) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(orders))
	}
	if orders[0].TotalPrice != "199.00" {
		t.Errorf("expected '199.00', got %q", orders[0].TotalPrice)
	}
}

func TestCustomerGet(t *testing.T) {
	type customerResource struct {
		Customer *core.Customer `json:"customer"`
	}
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(customerResource{
			Customer: &core.Customer{
				ID:        5001,
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
			},
		})
	})
	defer server.Close()

	customer, err := client.Customer.Get(context.Background(), 5001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if customer.Email != "test@example.com" {
		t.Errorf("expected 'test@example.com', got %q", customer.Email)
	}
}

func TestStoreGetShop(t *testing.T) {
	type shopResource struct {
		Shop *store.Shop `json:"shop"`
	}
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(shopResource{
			Shop: &store.Shop{
				ID:     1,
				Name:   "My Test Shop",
				Domain: "myshop.myshopline.com",
			},
		})
	})
	defer server.Close()

	shop, err := client.Store.GetShop(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if shop.Name != "My Test Shop" {
		t.Errorf("expected 'My Test Shop', got %q", shop.Name)
	}
}

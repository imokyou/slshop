package webhook

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/imokyou/slshop/core"
)

// mockRequester implements core.Requester for webhook tests.
type mockRequester struct {
	server     *httptest.Server
	apiVersion string
}

func newMockRequester(handler http.HandlerFunc) (*mockRequester, func()) {
	srv := httptest.NewServer(handler)
	return &mockRequester{server: srv, apiVersion: "v20251201"}, srv.Close
}

func (m *mockRequester) CreatePath(resource string) string {
	return "/admin/openapi/" + m.apiVersion + "/" + resource
}
func (m *mockRequester) Get(ctx context.Context, path string, result interface{}, opts interface{}) error {
	return m.do(ctx, http.MethodGet, path, nil, result)
}
func (m *mockRequester) Post(ctx context.Context, path string, body, result interface{}) error {
	return m.do(ctx, http.MethodPost, path, body, result)
}
func (m *mockRequester) Put(ctx context.Context, path string, body, result interface{}) error {
	return m.do(ctx, http.MethodPut, path, body, result)
}
func (m *mockRequester) Delete(ctx context.Context, path string) error {
	return m.do(ctx, http.MethodDelete, path, nil, nil)
}
func (m *mockRequester) do(_ context.Context, method, path string, body, result interface{}) error {
	var b []byte
	if body != nil {
		b, _ = json.Marshal(body)
	}
	req, _ := http.NewRequest(method, m.server.URL+path, strings.NewReader(string(b)))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

// Ensure mockRequester satisfies core.Requester
var _ core.Requester = (*mockRequester)(nil)

func TestWebhookList(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "webhooks.json") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(webhooksResource{Webhooks: []Subscription{
			{ID: 1, Topic: "orders/create", Address: "https://example.com/hook1"},
			{ID: 2, Topic: "products/update", Address: "https://example.com/hook2"},
		}})
	})
	defer close()

	svc := NewService(mock)
	hooks, err := svc.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hooks) != 2 {
		t.Fatalf("expected 2 webhooks, got %d", len(hooks))
	}
	if hooks[0].Topic != "orders/create" {
		t.Errorf("expected 'orders/create', got %q", hooks[0].Topic)
	}
}

func TestWebhookGet(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/42.json") {
			t.Errorf("expected /42.json in path, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(webhookResource{Webhook: &Subscription{ID: 42, Topic: "orders/paid"}})
	})
	defer close()

	svc := NewService(mock)
	hook, err := svc.Get(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hook.ID != 42 {
		t.Errorf("expected ID 42, got %d", hook.ID)
	}
	if hook.Topic != "orders/paid" {
		t.Errorf("expected 'orders/paid', got %q", hook.Topic)
	}
}

func TestWebhookCreate(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "webhooks.json") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body webhookResource
		json.NewDecoder(r.Body).Decode(&body)
		if body.Webhook == nil {
			t.Fatal("expected non-nil webhook in body")
		}
		json.NewEncoder(w).Encode(webhookResource{Webhook: &Subscription{
			ID:      99,
			Topic:   body.Webhook.Topic,
			Address: body.Webhook.Address,
		}})
	})
	defer close()

	svc := NewService(mock)
	hook, err := svc.Create(context.Background(), Subscription{
		Topic:   "customers/create",
		Address: "https://example.com/webhook",
		Format:  "json",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hook.ID != 99 {
		t.Errorf("expected ID 99, got %d", hook.ID)
	}
	if hook.Topic != "customers/create" {
		t.Errorf("expected 'customers/create', got %q", hook.Topic)
	}
}

func TestWebhookUpdate(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/42.json") {
			t.Errorf("expected /42.json in path, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(webhookResource{Webhook: &Subscription{
			ID:      42,
			Address: "https://example.com/new-hook",
		}})
	})
	defer close()

	svc := NewService(mock)
	hook, err := svc.Update(context.Background(), Subscription{
		ID:      42,
		Address: "https://example.com/new-hook",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hook.Address != "https://example.com/new-hook" {
		t.Errorf("expected updated address, got %q", hook.Address)
	}
}

func TestWebhookDelete(t *testing.T) {
	called := false
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/42.json") {
			t.Errorf("expected /42.json in path, got %s", r.URL.Path)
		}
		called = true
		w.WriteHeader(http.StatusOK)
	})
	defer close()

	svc := NewService(mock)
	err := svc.Delete(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("DELETE handler was not called")
	}
}

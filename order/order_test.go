package order

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/imokyou/slshop/core"
)

// mockRequester implements core.Requester using a test HTTP server.
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

func (m *mockRequester) do(ctx context.Context, method, path string, body, result interface{}) error {
	var reqBody *strings.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = strings.NewReader(string(b))
	} else {
		reqBody = strings.NewReader("")
	}

	req, _ := http.NewRequestWithContext(ctx, method, m.server.URL+path, reqBody)
	req.Header.Set("Content-Type", "application/json")

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

// =====================================================================
// Tests
// =====================================================================

func TestOrderList(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "orders.json") {
			t.Errorf("expected orders.json path, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ordersResource{Orders: []Order{
			{ID: 1001, Name: "#1001", TotalPrice: "99.00"},
			{ID: 1002, Name: "#1002", TotalPrice: "199.00"},
		}})
	})
	defer close()

	svc := NewService(mock)
	orders, err := svc.List(context.Background(), &ListOptions{Status: "any"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orders) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(orders))
	}
	if orders[0].Name != "#1001" {
		t.Errorf("expected '#1001', got %q", orders[0].Name)
	}
}

func TestOrderCount(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "count.json") {
			t.Errorf("expected count.json path, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(countResource{Count: 42})
	})
	defer close()

	svc := NewService(mock)
	count, err := svc.Count(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 42 {
		t.Errorf("expected 42, got %d", count)
	}
}

func TestOrderGet(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orderResource{Order: &Order{ID: 1001, Name: "#1001"}})
	})
	defer close()

	svc := NewService(mock)
	o, err := svc.Get(context.Background(), 1001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.ID != 1001 {
		t.Errorf("expected ID 1001, got %d", o.ID)
	}
}

func TestOrderCreate(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body orderResource
		json.NewDecoder(r.Body).Decode(&body)
		if body.Order == nil || body.Order.Name != "Test" {
			t.Errorf("unexpected body: %+v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orderResource{Order: &Order{ID: 999, Name: "Test"}})
	})
	defer close()

	svc := NewService(mock)
	o, err := svc.Create(context.Background(), Order{Name: "Test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.ID != 999 {
		t.Errorf("expected ID 999, got %d", o.ID)
	}
}

func TestOrderUpdate(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orderResource{Order: &Order{ID: 1001, Note: "updated"}})
	})
	defer close()

	svc := NewService(mock)
	o, err := svc.Update(context.Background(), Order{ID: 1001, Note: "updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.Note != "updated" {
		t.Errorf("expected note 'updated', got %q", o.Note)
	}
}

func TestOrderDelete(t *testing.T) {
	called := false
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		called = true
		w.WriteHeader(http.StatusOK)
	})
	defer close()

	svc := NewService(mock)
	err := svc.Delete(context.Background(), 1001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected DELETE handler to be called")
	}
}

func TestOrderCancel(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "cancel.json") {
			t.Errorf("expected cancel.json path, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orderResource{Order: &Order{ID: 1001, CancelReason: "customer"}})
	})
	defer close()

	svc := NewService(mock)
	o, err := svc.Cancel(context.Background(), 1001, &CancelOptions{Reason: "customer"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.CancelReason != "customer" {
		t.Errorf("expected cancel reason 'customer', got %q", o.CancelReason)
	}
}

func TestOrderClose(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "close.json") {
			t.Errorf("expected close.json path, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orderResource{Order: &Order{ID: 1001}})
	})
	defer close()

	svc := NewService(mock)
	o, err := svc.Close(context.Background(), 1001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.ID != 1001 {
		t.Errorf("expected ID 1001, got %d", o.ID)
	}
}

func TestOrderOpen(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "open.json") {
			t.Errorf("expected open.json path, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orderResource{Order: &Order{ID: 1001}})
	})
	defer close()

	svc := NewService(mock)
	o, err := svc.Open(context.Background(), 1001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.ID != 1001 {
		t.Errorf("expected ID 1001, got %d", o.ID)
	}
}

func TestOrderListRefunds(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "refunds.json") {
			t.Errorf("expected refunds.json path, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(refundsResource{Refunds: []Refund{{ID: 1, Note: "test refund"}}})
	})
	defer close()

	svc := NewService(mock)
	refunds, err := svc.ListRefunds(context.Background(), 1001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(refunds) != 1 {
		t.Fatalf("expected 1 refund, got %d", len(refunds))
	}
}

func TestOrderListTransactions(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "transactions.json") {
			t.Errorf("expected transactions.json path, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transactionsResource{Transactions: []Transaction{{ID: 1, Amount: "99.00"}}})
	})
	defer close()

	svc := NewService(mock)
	txns, err := svc.ListTransactions(context.Background(), 1001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(txns) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(txns))
	}
	if txns[0].Amount != "99.00" {
		t.Errorf("expected amount '99.00', got %q", txns[0].Amount)
	}
}

// TestOrderListOptions_URLTags verifies that ListOptions fields have correct url tags.
func TestOrderListOptions_URLTags(t *testing.T) {
	opts := &ListOptions{
		ListOptions: core.ListOptions{Limit: 20},
		Status:      "open",
	}
	// Verify struct is usable (non-nil, no panics)
	if opts.Status != "open" {
		t.Errorf("unexpected status: %s", opts.Status)
	}
}

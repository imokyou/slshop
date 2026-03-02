package marketing

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/imokyou/slshop/core"
)

// mockRequester implements core.Requester for marketing tests.
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

var _ core.Requester = (*mockRequester)(nil)

// =====================================================================
// PriceRule Tests
// =====================================================================

func TestListPriceRules(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "price_rules.json") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(priceRulesResource{PriceRules: []PriceRule{
			{ID: 1, Title: "SUMMER10", ValueType: "percentage", Value: "-10.0"},
		}})
	})
	defer close()

	svc := NewDiscountService(mock)
	rules, err := svc.ListPriceRules(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Title != "SUMMER10" {
		t.Errorf("expected 'SUMMER10', got %q", rules[0].Title)
	}
}

func TestGetPriceRule(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "price_rules/1.json") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(priceRuleResource{PriceRule: &PriceRule{ID: 1, Title: "SUMMER10"}})
	})
	defer close()

	svc := NewDiscountService(mock)
	rule, err := svc.GetPriceRule(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.ID != 1 {
		t.Errorf("expected ID 1, got %d", rule.ID)
	}
}

func TestCreatePriceRule(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body priceRuleResource
		json.NewDecoder(r.Body).Decode(&body)
		if body.PriceRule == nil {
			t.Fatal("expected non-nil price rule in body")
		}
		json.NewEncoder(w).Encode(priceRuleResource{PriceRule: &PriceRule{ID: 100, Title: body.PriceRule.Title}})
	})
	defer close()

	svc := NewDiscountService(mock)
	rule, err := svc.CreatePriceRule(context.Background(), PriceRule{
		Title:            "NEWRULE",
		ValueType:        "fixed_amount",
		Value:            "-20.0",
		TargetType:       "line_item",
		AllocationMethod: "across",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.ID != 100 {
		t.Errorf("expected ID 100, got %d", rule.ID)
	}
	if rule.Title != "NEWRULE" {
		t.Errorf("expected 'NEWRULE', got %q", rule.Title)
	}
}

func TestUpdatePriceRule(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "price_rules/1.json") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(priceRuleResource{PriceRule: &PriceRule{ID: 1, UsageLimit: 50}})
	})
	defer close()

	svc := NewDiscountService(mock)
	rule, err := svc.UpdatePriceRule(context.Background(), PriceRule{ID: 1, UsageLimit: 50})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.UsageLimit != 50 {
		t.Errorf("expected usage limit 50, got %d", rule.UsageLimit)
	}
}

func TestDeletePriceRule(t *testing.T) {
	called := false
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		called = true
		w.WriteHeader(http.StatusOK)
	})
	defer close()

	svc := NewDiscountService(mock)
	err := svc.DeletePriceRule(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("DELETE was not called")
	}
}

// =====================================================================
// DiscountCode Tests
// =====================================================================

func TestListDiscountCodes(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "price_rules/1/discount_codes.json") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(discountCodesResource{DiscountCodes: []DiscountCode{
			{ID: 10, Code: "SAVE10", PriceRuleID: 1},
		}})
	})
	defer close()

	svc := NewDiscountService(mock)
	codes, err := svc.ListDiscountCodes(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(codes) != 1 || codes[0].Code != "SAVE10" {
		t.Errorf("unexpected codes: %+v", codes)
	}
}

func TestCreateDiscountCode(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(discountCodeResource{DiscountCode: &DiscountCode{ID: 200, Code: "WELCOME"}})
	})
	defer close()

	svc := NewDiscountService(mock)
	code, err := svc.CreateDiscountCode(context.Background(), 1, DiscountCode{Code: "WELCOME"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code.Code != "WELCOME" {
		t.Errorf("expected 'WELCOME', got %q", code.Code)
	}
}

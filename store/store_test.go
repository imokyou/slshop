package store

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/imokyou/slshop/core"
)

// mockRequester implements core.Requester for store tests.
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

func TestGetShop(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "shop.json") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(shopResource{Shop: &Shop{
			ID:     1,
			Name:   "My Test Shop",
			Domain: "testshop.myshopline.com",
		}})
	})
	defer close()

	svc := NewService(mock)
	shop, err := svc.GetShop(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if shop.Name != "My Test Shop" {
		t.Errorf("expected 'My Test Shop', got %q", shop.Name)
	}
	if shop.Domain != "testshop.myshopline.com" {
		t.Errorf("expected 'testshop.myshopline.com', got %q", shop.Domain)
	}
}

func TestGetSettlementCurrency(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "currency/currencies.json") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(currenciesResource{Currencies: []Currency{
			{Code: "USD", Name: "US Dollar", Symbol: "$", Primary: true, Enabled: true},
			{Code: "EUR", Name: "Euro", Symbol: "€", Enabled: true},
		}})
	})
	defer close()

	svc := NewService(mock)
	currencies, err := svc.GetSettlementCurrency(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(currencies) != 2 {
		t.Fatalf("expected 2 currencies, got %d", len(currencies))
	}
	if currencies[0].Code != "USD" {
		t.Errorf("expected 'USD', got %q", currencies[0].Code)
	}
	if !currencies[0].Primary {
		t.Error("expected USD to be primary currency")
	}
}

func TestListStaffMembers(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "store/list/staff.json") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(staffListResource{Staff: []StaffMember{
			{UID: "uid-001", Email: "admin@shop.com", AccountOwner: true},
			{UID: "uid-002", Email: "staff@shop.com"},
		}})
	})
	defer close()

	svc := NewService(mock)
	staff, err := svc.ListStaffMembers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(staff) != 2 {
		t.Fatalf("expected 2 staff members, got %d", len(staff))
	}
	if !staff[0].AccountOwner {
		t.Error("expected first staff member to be account owner")
	}
}

func TestGetStaffMember(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "store/staff/uid-001.json") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(staffResource{Staff: &StaffMember{
			UID:   "uid-001",
			Email: "admin@shop.com",
		}})
	})
	defer close()

	svc := NewService(mock)
	member, err := svc.GetStaffMember(context.Background(), "uid-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if member.UID != "uid-001" {
		t.Errorf("expected UID 'uid-001', got %q", member.UID)
	}
}

func TestListOperationLogs(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "store/operation_logs.json") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(opLogsResource{OperationLogs: []OperationLog{
			{ID: 1, Action: "created", Subject: "product"},
			{ID: 2, Action: "updated", Subject: "order"},
		}})
	})
	defer close()

	svc := NewService(mock)
	logs, err := svc.ListOperationLogs(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("expected 2 logs, got %d", len(logs))
	}
	if logs[0].Action != "created" {
		t.Errorf("expected 'created', got %q", logs[0].Action)
	}
}

func TestGetInfo(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "merchants/shop.json") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(infoResource{Data: &Info{
			ID:       1,
			Name:     "Test Store",
			Currency: "USD",
		}})
	})
	defer close()

	svc := NewService(mock)
	info, err := svc.GetInfo(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Currency != "USD" {
		t.Errorf("expected 'USD', got %q", info.Currency)
	}
}

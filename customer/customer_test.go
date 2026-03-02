package customer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/imokyou/slshop/core"
)

// mockRequester implements core.Requester for customer tests.
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

func TestCustomerList(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "v2/customers.json") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(customersResource{Customers: []core.Customer{
			{ID: 1, Email: "a@test.com"},
			{ID: 2, Email: "b@test.com"},
		}})
	})
	defer close()

	svc := NewService(mock)
	customers, err := svc.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(customers) != 2 {
		t.Fatalf("expected 2 customers, got %d", len(customers))
	}
	if customers[0].Email != "a@test.com" {
		t.Errorf("expected 'a@test.com', got %q", customers[0].Email)
	}
}

func TestCustomerGet(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/5001.json") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(customerResource{Customer: &core.Customer{
			ID: 5001, Email: "john@test.com", FirstName: "John",
		}})
	})
	defer close()

	svc := NewService(mock)
	c, err := svc.Get(context.Background(), 5001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Email != "john@test.com" {
		t.Errorf("expected 'john@test.com', got %q", c.Email)
	}
}

func TestCustomerCreate(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body customerResource
		json.NewDecoder(r.Body).Decode(&body)
		if body.Customer == nil {
			t.Error("expected non-nil customer in body")
		}
		json.NewEncoder(w).Encode(customerResource{Customer: &core.Customer{ID: 999, Email: body.Customer.Email}})
	})
	defer close()

	svc := NewService(mock)
	c, err := svc.Create(context.Background(), core.Customer{Email: "new@test.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.ID != 999 {
		t.Errorf("expected ID 999, got %d", c.ID)
	}
}

func TestCustomerUpdate(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(customerResource{Customer: &core.Customer{ID: 5001, FirstName: "Jane"}})
	})
	defer close()

	svc := NewService(mock)
	c, err := svc.Update(context.Background(), core.Customer{ID: 5001, FirstName: "Jane"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.FirstName != "Jane" {
		t.Errorf("expected 'Jane', got %q", c.FirstName)
	}
}

func TestCustomerDelete(t *testing.T) {
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
	err := svc.Delete(context.Background(), 5001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("DELETE handler was not called")
	}
}

func TestCustomerCount(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "count.json") {
			t.Errorf("expected count path, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(countResource{Count: 77})
	})
	defer close()

	svc := NewService(mock)
	count, err := svc.Count(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 77 {
		t.Errorf("expected 77, got %d", count)
	}
}

func TestCustomerSearch(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "search.json") {
			t.Errorf("expected search path, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(customersResource{Customers: []core.Customer{{ID: 1, Email: "found@test.com"}}})
	})
	defer close()

	svc := NewService(mock)
	results, err := svc.Search(context.Background(), "found", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestCustomerListGroups(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "groups.json") {
			t.Errorf("expected groups.json path, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(groupsResource{CustomerGroups: []Group{{ID: 1, Name: "VIP"}}})
	})
	defer close()

	svc := NewService(mock)
	groups, err := svc.ListGroups(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 1 || groups[0].Name != "VIP" {
		t.Errorf("unexpected groups: %+v", groups)
	}
}

func TestCustomerCreateGroup(t *testing.T) {
	mock, close := newMockRequester(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(groupResource{CustomerGroup: &Group{ID: 10, Name: "Wholesale"}})
	})
	defer close()

	svc := NewService(mock)
	g, err := svc.CreateGroup(context.Background(), Group{Name: "Wholesale"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Name != "Wholesale" {
		t.Errorf("expected 'Wholesale', got %q", g.Name)
	}
}

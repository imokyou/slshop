package bulk

import (
	"context"
	"time"

	"github.com/imokyou/slshop/core"
)

// =====================================================================
// Bulk Operations Service
// =====================================================================

type Service interface {
	GetCurrent(ctx context.Context, opType string) (*BulkOperation, error)
	CreateQuery(ctx context.Context, query BulkQueryRequest) (*BulkOperation, error)
	CreateMutation(ctx context.Context, mutation BulkMutationRequest) (*BulkOperation, error)
	Cancel(ctx context.Context, id string) (*BulkOperation, error)
}

func NewService(client core.Requester) Service {
	return &serviceOp{client: client}
}

type serviceOp struct{ client core.Requester }

// =====================================================================
// Models
// =====================================================================

type BulkOperation struct {
	ID              string     `json:"id,omitempty"`
	Status          string     `json:"status,omitempty"`
	Type            string     `json:"type,omitempty"`
	Query           string     `json:"query,omitempty"`
	ErrorCode       string     `json:"error_code,omitempty"`
	URL             string     `json:"url,omitempty"`
	RootObjectCount int        `json:"root_object_count,omitempty"`
	ObjectCount     int        `json:"object_count,omitempty"`
	FileSize        int64      `json:"file_size,omitempty"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

type BulkQueryRequest struct {
	Query string `json:"query,omitempty"`
}

type BulkMutationRequest struct {
	Query            string `json:"query,omitempty"`
	StagedUploadPath string `json:"staged_upload_path,omitempty"`
}

// JSON wrappers
type bulkOpResource struct {
	Data *BulkOperation `json:"data"`
}

// =====================================================================
// Implementation
// =====================================================================

// GET current_bulk_operation.json?type={type}
func (s *serviceOp) GetCurrent(ctx context.Context, opType string) (*BulkOperation, error) {
	r := &bulkOpResource{}
	opts := struct {
		Type string `url:"type,omitempty"`
	}{Type: opType}
	err := s.client.Get(ctx, s.client.CreatePath("current_bulk_operation.json"), r, &opts)
	return r.Data, err
}

// POST bulk_operations.json (query)
func (s *serviceOp) CreateQuery(ctx context.Context, query BulkQueryRequest) (*BulkOperation, error) {
	r := &bulkOpResource{}
	err := s.client.Post(ctx, s.client.CreatePath("bulk_operations.json"), query, r)
	return r.Data, err
}

// POST bulk_mutations.json
func (s *serviceOp) CreateMutation(ctx context.Context, mutation BulkMutationRequest) (*BulkOperation, error) {
	r := &bulkOpResource{}
	err := s.client.Post(ctx, s.client.CreatePath("bulk_mutations.json"), mutation, r)
	return r.Data, err
}

// POST current_bulk_operation/cancel.json
func (s *serviceOp) Cancel(ctx context.Context, id string) (*BulkOperation, error) {
	r := &bulkOpResource{}
	body := map[string]string{"id": id}
	err := s.client.Post(ctx, s.client.CreatePath("current_bulk_operation/cancel.json"), body, r)
	return r.Data, err
}

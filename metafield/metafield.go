package metafield

import (
	"context"
	"fmt"
	"time"

	"github.com/imokyou/slshop/core"
)

// =====================================================================
// Metafield Definition Service
// =====================================================================

type DefinitionService interface {
	Create(ctx context.Context, def MetafieldDefinition) (*MetafieldDefinition, error)
	Update(ctx context.Context, def MetafieldDefinition) (*MetafieldDefinition, error)
	List(ctx context.Context, opts *DefinitionListOptions) ([]MetafieldDefinition, error)
	Get(ctx context.Context, id int64) (*MetafieldDefinition, error)
	Delete(ctx context.Context, id int64) error
	Count(ctx context.Context, opts *DefinitionCountOptions) (int, error)
}

func NewDefinitionService(client core.Requester) DefinitionService {
	return &defOp{client: client}
}

type defOp struct{ client core.Requester }

// =====================================================================
// Resource Metafield Service
// =====================================================================

type ResourceService interface {
	Create(ctx context.Context, ownerResource string, ownerID int64, m Metafield) (*Metafield, error)
	Update(ctx context.Context, ownerResource string, ownerID int64, m Metafield) (*Metafield, error)
	List(ctx context.Context, ownerResource string, ownerID int64, opts *core.ListOptions) ([]Metafield, error)
	Get(ctx context.Context, ownerResource string, ownerID, metafieldID int64) (*Metafield, error)
	Delete(ctx context.Context, ownerResource string, ownerID, metafieldID int64) error
	Count(ctx context.Context, ownerResource string, ownerID int64) (int, error)
}

func NewResourceService(client core.Requester) ResourceService {
	return &resOp{client: client}
}

type resOp struct{ client core.Requester }

// =====================================================================
// Store Metafield Service
// =====================================================================

type StoreService interface {
	Create(ctx context.Context, m Metafield) (*Metafield, error)
	Update(ctx context.Context, m Metafield) (*Metafield, error)
	List(ctx context.Context, opts *core.ListOptions) ([]Metafield, error)
	Get(ctx context.Context, metafieldID int64) (*Metafield, error)
	Delete(ctx context.Context, metafieldID int64) error
	Count(ctx context.Context) (int, error)
}

func NewStoreService(client core.Requester) StoreService {
	return &storeOp{client: client}
}

type storeOp struct{ client core.Requester }

// =====================================================================
// Models
// =====================================================================

type MetafieldDefinition struct {
	ID             int64                 `json:"id,omitempty"`
	Name           string                `json:"name,omitempty"`
	Namespace      string                `json:"namespace,omitempty"`
	Key            string                `json:"key,omitempty"`
	Description    string                `json:"description,omitempty"`
	Type           string                `json:"type,omitempty"`
	OwnerType      string                `json:"owner_type,omitempty"`
	PinnedPosition int                   `json:"pinned_position,omitempty"`
	Validations    []MetafieldValidation `json:"validations,omitempty"`
	CreatedAt      *time.Time            `json:"created_at,omitempty"`
	UpdatedAt      *time.Time            `json:"updated_at,omitempty"`
}

type MetafieldValidation struct {
	Name  string `json:"name,omitempty"`
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

type Metafield struct {
	ID            int64      `json:"id,omitempty"`
	Namespace     string     `json:"namespace,omitempty"`
	Key           string     `json:"key,omitempty"`
	Value         string     `json:"value,omitempty"`
	Type          string     `json:"type,omitempty"`
	Description   string     `json:"description,omitempty"`
	OwnerID       int64      `json:"owner_id,omitempty"`
	OwnerResource string     `json:"owner_resource,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}

type DefinitionListOptions struct {
	core.ListOptions
	OwnerType string `url:"owner_type,omitempty"`
	Namespace string `url:"namespace,omitempty"`
}

type DefinitionCountOptions struct {
	OwnerType string `url:"owner_type,omitempty"`
	Namespace string `url:"namespace,omitempty"`
}

// JSON wrappers
type defResource struct {
	MetafieldDefinition *MetafieldDefinition `json:"metafield_definition"`
}
type defsResource struct {
	MetafieldDefinitions []MetafieldDefinition `json:"metafield_definitions"`
}
type mfResource struct {
	Metafield *Metafield `json:"metafield"`
}
type mfsResource struct {
	Metafields []Metafield `json:"metafields"`
}
type countResource struct {
	Count int `json:"count"`
}

// =====================================================================
// MetafieldDefinition Implementation
// =====================================================================

func (s *defOp) Create(ctx context.Context, def MetafieldDefinition) (*MetafieldDefinition, error) {
	r := &defResource{}
	err := s.client.Post(ctx, s.client.CreatePath("metafield_definitions.json"), defResource{MetafieldDefinition: &def}, r)
	return r.MetafieldDefinition, err
}
func (s *defOp) Update(ctx context.Context, def MetafieldDefinition) (*MetafieldDefinition, error) {
	r := &defResource{}
	err := s.client.Put(ctx, s.client.CreatePath(fmt.Sprintf("metafield_definitions/%d.json", def.ID)), defResource{MetafieldDefinition: &def}, r)
	return r.MetafieldDefinition, err
}
func (s *defOp) List(ctx context.Context, opts *DefinitionListOptions) ([]MetafieldDefinition, error) {
	r := &defsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("metafield_definitions.json"), r, opts)
	return r.MetafieldDefinitions, err
}
func (s *defOp) Get(ctx context.Context, id int64) (*MetafieldDefinition, error) {
	r := &defResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("metafield_definitions/%d.json", id)), r, nil)
	return r.MetafieldDefinition, err
}
func (s *defOp) Delete(ctx context.Context, id int64) error {
	return s.client.Delete(ctx, s.client.CreatePath(fmt.Sprintf("metafield_definitions/%d.json", id)))
}
func (s *defOp) Count(ctx context.Context, opts *DefinitionCountOptions) (int, error) {
	r := &countResource{}
	err := s.client.Get(ctx, s.client.CreatePath("metafield_definitions/count.json"), r, opts)
	return r.Count, err
}

// =====================================================================
// Resource Metafield Implementation
// =====================================================================

func (s *resOp) Create(ctx context.Context, ownerResource string, ownerID int64, m Metafield) (*Metafield, error) {
	r := &mfResource{}
	path := fmt.Sprintf("%s/%d/metafields.json", ownerResource, ownerID)
	err := s.client.Post(ctx, s.client.CreatePath(path), mfResource{Metafield: &m}, r)
	return r.Metafield, err
}
func (s *resOp) Update(ctx context.Context, ownerResource string, ownerID int64, m Metafield) (*Metafield, error) {
	r := &mfResource{}
	path := fmt.Sprintf("%s/%d/metafields/%d.json", ownerResource, ownerID, m.ID)
	err := s.client.Put(ctx, s.client.CreatePath(path), mfResource{Metafield: &m}, r)
	return r.Metafield, err
}
func (s *resOp) List(ctx context.Context, ownerResource string, ownerID int64, opts *core.ListOptions) ([]Metafield, error) {
	r := &mfsResource{}
	path := fmt.Sprintf("%s/%d/metafields.json", ownerResource, ownerID)
	err := s.client.Get(ctx, s.client.CreatePath(path), r, opts)
	return r.Metafields, err
}
func (s *resOp) Get(ctx context.Context, ownerResource string, ownerID, metafieldID int64) (*Metafield, error) {
	r := &mfResource{}
	path := fmt.Sprintf("%s/%d/metafields/%d.json", ownerResource, ownerID, metafieldID)
	err := s.client.Get(ctx, s.client.CreatePath(path), r, nil)
	return r.Metafield, err
}
func (s *resOp) Delete(ctx context.Context, ownerResource string, ownerID, metafieldID int64) error {
	path := fmt.Sprintf("%s/%d/metafields/%d.json", ownerResource, ownerID, metafieldID)
	return s.client.Delete(ctx, s.client.CreatePath(path))
}
func (s *resOp) Count(ctx context.Context, ownerResource string, ownerID int64) (int, error) {
	r := &countResource{}
	path := fmt.Sprintf("%s/%d/metafields/count.json", ownerResource, ownerID)
	err := s.client.Get(ctx, s.client.CreatePath(path), r, nil)
	return r.Count, err
}

// =====================================================================
// Store Metafield Implementation
// =====================================================================

func (s *storeOp) Create(ctx context.Context, m Metafield) (*Metafield, error) {
	r := &mfResource{}
	err := s.client.Post(ctx, s.client.CreatePath("metafields.json"), mfResource{Metafield: &m}, r)
	return r.Metafield, err
}
func (s *storeOp) Update(ctx context.Context, m Metafield) (*Metafield, error) {
	r := &mfResource{}
	err := s.client.Put(ctx, s.client.CreatePath(fmt.Sprintf("metafields/%d.json", m.ID)), mfResource{Metafield: &m}, r)
	return r.Metafield, err
}
func (s *storeOp) List(ctx context.Context, opts *core.ListOptions) ([]Metafield, error) {
	r := &mfsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("metafields.json"), r, opts)
	return r.Metafields, err
}
func (s *storeOp) Get(ctx context.Context, metafieldID int64) (*Metafield, error) {
	r := &mfResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("metafields/%d.json", metafieldID)), r, nil)
	return r.Metafield, err
}
func (s *storeOp) Delete(ctx context.Context, metafieldID int64) error {
	return s.client.Delete(ctx, s.client.CreatePath(fmt.Sprintf("metafields/%d.json", metafieldID)))
}
func (s *storeOp) Count(ctx context.Context) (int, error) {
	r := &countResource{}
	err := s.client.Get(ctx, s.client.CreatePath("metafields/count.json"), r, nil)
	return r.Count, err
}

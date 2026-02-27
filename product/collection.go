package product

import (
	"context"
	"fmt"
	"time"

	"github.com/imokyou/slshop/core"
)

// =====================================================================
// Collection (Smart + Manual)
// =====================================================================

type CollectionService interface {
	List(ctx context.Context, opts *core.ListOptions) ([]Collection, error)
	Get(ctx context.Context, id int64) (*Collection, error)
	Create(ctx context.Context, c Collection) (*Collection, error)
	Update(ctx context.Context, c Collection) (*Collection, error)
	Delete(ctx context.Context, id int64) error
	Count(ctx context.Context) (int, error)
}

func NewCollectionService(client core.Requester) CollectionService {
	return &collectionOp{client: client}
}

type collectionOp struct{ client core.Requester }

type SmartCollectionService interface {
	List(ctx context.Context, opts *core.ListOptions) ([]SmartCollection, error)
	Get(ctx context.Context, id int64) (*SmartCollection, error)
	Create(ctx context.Context, c SmartCollection) (*SmartCollection, error)
	Update(ctx context.Context, c SmartCollection) (*SmartCollection, error)
	Delete(ctx context.Context, id int64) error
}

func NewSmartCollectionService(client core.Requester) SmartCollectionService {
	return &smartCollectionOp{client: client}
}

type smartCollectionOp struct{ client core.Requester }

type ManualCollectionService interface {
	List(ctx context.Context, opts *core.ListOptions) ([]ManualCollection, error)
	Get(ctx context.Context, id int64) (*ManualCollection, error)
	Create(ctx context.Context, c ManualCollection) (*ManualCollection, error)
	Update(ctx context.Context, c ManualCollection) (*ManualCollection, error)
	Delete(ctx context.Context, id int64) error
}

func NewManualCollectionService(client core.Requester) ManualCollectionService {
	return &manualCollectionOp{client: client}
}

type manualCollectionOp struct{ client core.Requester }

type Collection struct {
	ID             int64      `json:"id,omitempty"`
	Title          string     `json:"title,omitempty"`
	Handle         string     `json:"handle,omitempty"`
	BodyHTML       string     `json:"body_html,omitempty"`
	SortOrder      string     `json:"sort_order,omitempty"`
	TemplateSuffix string     `json:"template_suffix,omitempty"`
	Published      bool       `json:"published,omitempty"`
	PublishedAt    *time.Time `json:"published_at,omitempty"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty"`
}

type SmartCollection struct {
	ID             int64            `json:"id,omitempty"`
	Title          string           `json:"title,omitempty"`
	Handle         string           `json:"handle,omitempty"`
	BodyHTML       string           `json:"body_html,omitempty"`
	SortOrder      string           `json:"sort_order,omitempty"`
	TemplateSuffix string           `json:"template_suffix,omitempty"`
	Published      bool             `json:"published,omitempty"`
	Disjunctive    bool             `json:"disjunctive,omitempty"`
	Rules          []CollectionRule `json:"rules,omitempty"`
	PublishedAt    *time.Time       `json:"published_at,omitempty"`
	UpdatedAt      *time.Time       `json:"updated_at,omitempty"`
}

type ManualCollection struct {
	ID             int64      `json:"id,omitempty"`
	Title          string     `json:"title,omitempty"`
	Handle         string     `json:"handle,omitempty"`
	BodyHTML       string     `json:"body_html,omitempty"`
	SortOrder      string     `json:"sort_order,omitempty"`
	TemplateSuffix string     `json:"template_suffix,omitempty"`
	Published      bool       `json:"published,omitempty"`
	PublishedAt    *time.Time `json:"published_at,omitempty"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty"`
}

type CollectionRule struct {
	Column    string `json:"column,omitempty"`
	Relation  string `json:"relation,omitempty"`
	Condition string `json:"condition,omitempty"`
}

// JSON wrappers
type collectionResource struct {
	Collection *Collection `json:"collection"`
}
type collectionsResource struct {
	Collections []Collection `json:"collections"`
}
type smartCollectionResource struct {
	SmartCollection *SmartCollection `json:"smart_collection"`
}
type smartCollectionsResource struct {
	SmartCollections []SmartCollection `json:"smart_collections"`
}
type manualCollectionResource struct {
	CustomCollection *ManualCollection `json:"custom_collection"`
}
type manualCollectionsResource struct {
	CustomCollections []ManualCollection `json:"custom_collections"`
}

// === Collection ===
func (s *collectionOp) List(ctx context.Context, opts *core.ListOptions) ([]Collection, error) {
	r := &collectionsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("collections.json"), r, opts)
	return r.Collections, err
}
func (s *collectionOp) Get(ctx context.Context, id int64) (*Collection, error) {
	r := &collectionResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("collections/%d.json", id)), r, nil)
	return r.Collection, err
}
func (s *collectionOp) Create(ctx context.Context, c Collection) (*Collection, error) {
	r := &collectionResource{}
	err := s.client.Post(ctx, s.client.CreatePath("collections.json"), collectionResource{Collection: &c}, r)
	return r.Collection, err
}
func (s *collectionOp) Update(ctx context.Context, c Collection) (*Collection, error) {
	r := &collectionResource{}
	err := s.client.Put(ctx, s.client.CreatePath(fmt.Sprintf("collections/%d.json", c.ID)), collectionResource{Collection: &c}, r)
	return r.Collection, err
}
func (s *collectionOp) Delete(ctx context.Context, id int64) error {
	return s.client.Delete(ctx, s.client.CreatePath(fmt.Sprintf("collections/%d.json", id)))
}
func (s *collectionOp) Count(ctx context.Context) (int, error) {
	r := &countResource{}
	err := s.client.Get(ctx, s.client.CreatePath("collections/count.json"), r, nil)
	return r.Count, err
}

// === Smart Collection ===
func (s *smartCollectionOp) List(ctx context.Context, opts *core.ListOptions) ([]SmartCollection, error) {
	r := &smartCollectionsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("smart_collections.json"), r, opts)
	return r.SmartCollections, err
}
func (s *smartCollectionOp) Get(ctx context.Context, id int64) (*SmartCollection, error) {
	r := &smartCollectionResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("smart_collections/%d.json", id)), r, nil)
	return r.SmartCollection, err
}
func (s *smartCollectionOp) Create(ctx context.Context, c SmartCollection) (*SmartCollection, error) {
	r := &smartCollectionResource{}
	err := s.client.Post(ctx, s.client.CreatePath("smart_collections.json"), smartCollectionResource{SmartCollection: &c}, r)
	return r.SmartCollection, err
}
func (s *smartCollectionOp) Update(ctx context.Context, c SmartCollection) (*SmartCollection, error) {
	r := &smartCollectionResource{}
	err := s.client.Put(ctx, s.client.CreatePath(fmt.Sprintf("smart_collections/%d.json", c.ID)), smartCollectionResource{SmartCollection: &c}, r)
	return r.SmartCollection, err
}
func (s *smartCollectionOp) Delete(ctx context.Context, id int64) error {
	return s.client.Delete(ctx, s.client.CreatePath(fmt.Sprintf("smart_collections/%d.json", id)))
}

// === Manual Collection ===
func (s *manualCollectionOp) List(ctx context.Context, opts *core.ListOptions) ([]ManualCollection, error) {
	r := &manualCollectionsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("custom_collections.json"), r, opts)
	return r.CustomCollections, err
}
func (s *manualCollectionOp) Get(ctx context.Context, id int64) (*ManualCollection, error) {
	r := &manualCollectionResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("custom_collections/%d.json", id)), r, nil)
	return r.CustomCollection, err
}
func (s *manualCollectionOp) Create(ctx context.Context, c ManualCollection) (*ManualCollection, error) {
	r := &manualCollectionResource{}
	err := s.client.Post(ctx, s.client.CreatePath("custom_collections.json"), manualCollectionResource{CustomCollection: &c}, r)
	return r.CustomCollection, err
}
func (s *manualCollectionOp) Update(ctx context.Context, c ManualCollection) (*ManualCollection, error) {
	r := &manualCollectionResource{}
	err := s.client.Put(ctx, s.client.CreatePath(fmt.Sprintf("custom_collections/%d.json", c.ID)), manualCollectionResource{CustomCollection: &c}, r)
	return r.CustomCollection, err
}
func (s *manualCollectionOp) Delete(ctx context.Context, id int64) error {
	return s.client.Delete(ctx, s.client.CreatePath(fmt.Sprintf("custom_collections/%d.json", id)))
}

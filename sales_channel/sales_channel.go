package saleschannel

import (
	"context"
	"fmt"

	"github.com/imokyou/slshop/core"
)

// =====================================================================
// Sales Channel Service
// =====================================================================

type Service interface {
	// Product listings
	ListProducts(ctx context.Context, opts *core.ListOptions) ([]ProductListing, error)
	GetProduct(ctx context.Context, productID int64) (*ProductListing, error)
	AddProduct(ctx context.Context, productID int64) (*ProductListing, error)
	RemoveProduct(ctx context.Context, productID int64) error
	CountProducts(ctx context.Context) (int, error)
	ListProductIDs(ctx context.Context, opts *core.ListOptions) ([]int64, error)

	// Collection listings
	ListCollections(ctx context.Context, opts *core.ListOptions) ([]CollectionListing, error)
	GetCollection(ctx context.Context, collectionID int64) (*CollectionListing, error)
	AddCollection(ctx context.Context, collectionID int64) (*CollectionListing, error)
	RemoveCollection(ctx context.Context, collectionID int64) error
	ListCollectionProductIDs(ctx context.Context, collectionID int64, opts *core.ListOptions) ([]int64, error)
}

func NewService(client core.Requester) Service {
	return &serviceOp{client: client}
}

type serviceOp struct{ client core.Requester }

// =====================================================================
// Models
// =====================================================================

type ProductListing struct {
	ProductID   int64         `json:"product_id,omitempty"`
	Title       string        `json:"title,omitempty"`
	Handle      string        `json:"handle,omitempty"`
	BodyHTML    string        `json:"body_html,omitempty"`
	ProductType string        `json:"product_type,omitempty"`
	Vendor      string        `json:"vendor,omitempty"`
	Tags        string        `json:"tags,omitempty"`
	Variants    []interface{} `json:"variants,omitempty"`
	Images      []interface{} `json:"images,omitempty"`
}

type CollectionListing struct {
	CollectionID int64  `json:"collection_id,omitempty"`
	Title        string `json:"title,omitempty"`
	Handle       string `json:"handle,omitempty"`
	BodyHTML     string `json:"body_html,omitempty"`
	SortOrder    string `json:"sort_order,omitempty"`
}

// JSON wrappers
type productListingResource struct {
	ProductListing *ProductListing `json:"product_listing"`
}
type productListingsResource struct {
	ProductListings []ProductListing `json:"product_listings"`
}
type collectionListingResource struct {
	CollectionListing *CollectionListing `json:"collection_listing"`
}
type collectionListingsResource struct {
	CollectionListings []CollectionListing `json:"collection_listings"`
}
type countResource struct {
	Count int `json:"count"`
}
type productIDsResource struct {
	ProductIDs []int64 `json:"product_ids"`
}

// =====================================================================
// Product Listings
// =====================================================================

// GET product_listings.json
func (s *serviceOp) ListProducts(ctx context.Context, opts *core.ListOptions) ([]ProductListing, error) {
	r := &productListingsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("product_listings.json"), r, opts)
	return r.ProductListings, err
}

// GET product_listings/{product_id}.json
func (s *serviceOp) GetProduct(ctx context.Context, productID int64) (*ProductListing, error) {
	r := &productListingResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("product_listings/%d.json", productID)), r, nil)
	return r.ProductListing, err
}

// PUT product_listings/{product_id}.json
func (s *serviceOp) AddProduct(ctx context.Context, productID int64) (*ProductListing, error) {
	r := &productListingResource{}
	err := s.client.Put(ctx, s.client.CreatePath(fmt.Sprintf("product_listings/%d.json", productID)), nil, r)
	return r.ProductListing, err
}

// DELETE product_listings/{product_id}.json
func (s *serviceOp) RemoveProduct(ctx context.Context, productID int64) error {
	return s.client.Delete(ctx, s.client.CreatePath(fmt.Sprintf("product_listings/%d.json", productID)))
}

// GET product_listings/count.json
func (s *serviceOp) CountProducts(ctx context.Context) (int, error) {
	r := &countResource{}
	err := s.client.Get(ctx, s.client.CreatePath("product_listings/count.json"), r, nil)
	return r.Count, err
}

// GET product_listings/product_ids.json
func (s *serviceOp) ListProductIDs(ctx context.Context, opts *core.ListOptions) ([]int64, error) {
	r := &productIDsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("product_listings/product_ids.json"), r, opts)
	return r.ProductIDs, err
}

// =====================================================================
// Collection Listings
// =====================================================================

// GET collection_listings.json
func (s *serviceOp) ListCollections(ctx context.Context, opts *core.ListOptions) ([]CollectionListing, error) {
	r := &collectionListingsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("collection_listings.json"), r, opts)
	return r.CollectionListings, err
}

// GET collection_listings/{collection_id}.json
func (s *serviceOp) GetCollection(ctx context.Context, collectionID int64) (*CollectionListing, error) {
	r := &collectionListingResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("collection_listings/%d.json", collectionID)), r, nil)
	return r.CollectionListing, err
}

// PUT collection_listings/{collection_id}.json
func (s *serviceOp) AddCollection(ctx context.Context, collectionID int64) (*CollectionListing, error) {
	r := &collectionListingResource{}
	err := s.client.Put(ctx, s.client.CreatePath(fmt.Sprintf("collection_listings/%d.json", collectionID)), nil, r)
	return r.CollectionListing, err
}

// DELETE collection_listings/{collection_id}.json
func (s *serviceOp) RemoveCollection(ctx context.Context, collectionID int64) error {
	return s.client.Delete(ctx, s.client.CreatePath(fmt.Sprintf("collection_listings/%d.json", collectionID)))
}

// GET collection_listings/{collection_id}/product_ids.json
func (s *serviceOp) ListCollectionProductIDs(ctx context.Context, collectionID int64, opts *core.ListOptions) ([]int64, error) {
	r := &productIDsResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("collection_listings/%d/product_ids.json", collectionID)), r, opts)
	return r.ProductIDs, err
}

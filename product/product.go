package product

import (
	"context"
	"fmt"
	"time"

	"github.com/imokyou/slshop/core"
)

const productsBasePath = "products"

// =====================================================================
// Product Service
// =====================================================================

type Service interface {
	List(ctx context.Context, opts *core.ListOptions) ([]Product, error)
	Count(ctx context.Context, opts *core.CountOptions) (int, error)
	Get(ctx context.Context, id int64) (*Product, error)
	Create(ctx context.Context, p Product) (*Product, error)
	Update(ctx context.Context, p Product) (*Product, error)
	Delete(ctx context.Context, id int64) error
}

func NewService(client core.Requester) Service {
	return &serviceOp{client: client}
}

type serviceOp struct{ client core.Requester }

type Product struct {
	ID          int64      `json:"id,omitempty"`
	Title       string     `json:"title,omitempty"`
	BodyHTML    string     `json:"body_html,omitempty"`
	Vendor      string     `json:"vendor,omitempty"`
	ProductType string     `json:"product_type,omitempty"`
	Handle      string     `json:"handle,omitempty"`
	Status      string     `json:"status,omitempty"`
	Tags        string     `json:"tags,omitempty"`
	Variants    []Variant  `json:"variants,omitempty"`
	Options     []Option   `json:"options,omitempty"`
	Images      []Image    `json:"images,omitempty"`
	Image       *Image     `json:"image,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type Variant struct {
	ID                  int64      `json:"id,omitempty"`
	ProductID           int64      `json:"product_id,omitempty"`
	Title               string     `json:"title,omitempty"`
	Price               string     `json:"price,omitempty"`
	CompareAtPrice      string     `json:"compare_at_price,omitempty"`
	SKU                 string     `json:"sku,omitempty"`
	Barcode             string     `json:"barcode,omitempty"`
	Position            int        `json:"position,omitempty"`
	Grams               int        `json:"grams,omitempty"`
	Weight              float64    `json:"weight,omitempty"`
	WeightUnit          string     `json:"weight_unit,omitempty"`
	InventoryItemID     int64      `json:"inventory_item_id,omitempty"`
	InventoryQuantity   int        `json:"inventory_quantity,omitempty"`
	InventoryManagement string     `json:"inventory_management,omitempty"`
	InventoryPolicy     string     `json:"inventory_policy,omitempty"`
	FulfillmentService  string     `json:"fulfillment_service,omitempty"`
	Option1             string     `json:"option1,omitempty"`
	Option2             string     `json:"option2,omitempty"`
	Option3             string     `json:"option3,omitempty"`
	RequiresShipping    bool       `json:"requires_shipping,omitempty"`
	Taxable             bool       `json:"taxable,omitempty"`
	ImageID             int64      `json:"image_id,omitempty"`
	CreatedAt           *time.Time `json:"created_at,omitempty"`
	UpdatedAt           *time.Time `json:"updated_at,omitempty"`
}

type Option struct {
	ID        int64    `json:"id,omitempty"`
	ProductID int64    `json:"product_id,omitempty"`
	Name      string   `json:"name,omitempty"`
	Position  int      `json:"position,omitempty"`
	Values    []string `json:"values,omitempty"`
}

type Image struct {
	ID         int64      `json:"id,omitempty"`
	ProductID  int64      `json:"product_id,omitempty"`
	Position   int        `json:"position,omitempty"`
	Src        string     `json:"src,omitempty"`
	Width      int        `json:"width,omitempty"`
	Height     int        `json:"height,omitempty"`
	VariantIDs []int64    `json:"variant_ids,omitempty"`
	Alt        string     `json:"alt,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
}

type productResource struct {
	Product *Product `json:"product"`
}
type productsResource struct {
	Products []Product `json:"products"`
}
type countResource struct {
	Count int `json:"count"`
}

func (s *serviceOp) List(ctx context.Context, opts *core.ListOptions) ([]Product, error) {
	r := &productsResource{}
	err := s.client.Get(ctx, s.client.CreatePath(productsBasePath+".json"), r, opts)
	return r.Products, err
}
func (s *serviceOp) Count(ctx context.Context, opts *core.CountOptions) (int, error) {
	r := &countResource{}
	err := s.client.Get(ctx, s.client.CreatePath(productsBasePath+"/count.json"), r, opts)
	return r.Count, err
}
func (s *serviceOp) Get(ctx context.Context, id int64) (*Product, error) {
	r := &productResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("%s/%d.json", productsBasePath, id)), r, nil)
	return r.Product, err
}
func (s *serviceOp) Create(ctx context.Context, p Product) (*Product, error) {
	r := &productResource{}
	err := s.client.Post(ctx, s.client.CreatePath(productsBasePath+".json"), productResource{Product: &p}, r)
	return r.Product, err
}
func (s *serviceOp) Update(ctx context.Context, p Product) (*Product, error) {
	r := &productResource{}
	err := s.client.Put(ctx, s.client.CreatePath(fmt.Sprintf("%s/%d.json", productsBasePath, p.ID)), productResource{Product: &p}, r)
	return r.Product, err
}
func (s *serviceOp) Delete(ctx context.Context, id int64) error {
	return s.client.Delete(ctx, s.client.CreatePath(fmt.Sprintf("%s/%d.json", productsBasePath, id)))
}

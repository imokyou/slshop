package appopenapi

import (
	"context"
	"fmt"

	"github.com/imokyou/slshop/core"
)

// =====================================================================
// Size Chart Service
// =====================================================================

type SizeChartService interface {
	BatchQueryProductSizes(ctx context.Context, productIDs []int64) ([]ProductSizeData, error)
	BatchQueryCategoryTemplates(ctx context.Context, categoryIDs []int64) ([]SizeTemplate, error)
	BatchQueryStoreTemplates(ctx context.Context, opts *core.ListOptions) ([]SizeTemplate, error)
	BatchDeleteProductSizes(ctx context.Context, productIDs []int64) error
	BatchCreateOrUpdateProductSizes(ctx context.Context, data []ProductSizeData) ([]ProductSizeData, error)
}

func NewSizeChartService(client core.Requester) SizeChartService {
	return &sizeChartOp{client: client}
}

type sizeChartOp struct{ client core.Requester }

// =====================================================================
// Customer Data Platform Service
// =====================================================================

type CDPService interface {
	ReportBehaviorEvents(ctx context.Context, events []BehaviorEvent) error
	ReportIdentity(ctx context.Context, identity IdentityReport) error
}

func NewCDPService(client core.Requester) CDPService {
	return &cdpOp{client: client}
}

type cdpOp struct{ client core.Requester }

// =====================================================================
// Variant Images Service
// =====================================================================

type VariantImageService interface {
	QueryVariantImages(ctx context.Context, variantID int64) ([]VariantImage, error)
	BatchUpdateVariantImages(ctx context.Context, updates []VariantImageUpdate) error
}

func NewVariantImageService(client core.Requester) VariantImageService {
	return &variantImgOp{client: client}
}

type variantImgOp struct{ client core.Requester }

// =====================================================================
// Models
// =====================================================================

type ProductSizeData struct {
	ProductID  int64       `json:"product_id,omitempty"`
	SizeData   interface{} `json:"size_data,omitempty"`
	TemplateID int64       `json:"template_id,omitempty"`
}

type SizeTemplate struct {
	ID         int64       `json:"id,omitempty"`
	Name       string      `json:"name,omitempty"`
	CategoryID int64       `json:"category_id,omitempty"`
	Data       interface{} `json:"data,omitempty"`
}

type BehaviorEvent struct {
	EventName  string                 `json:"event_name,omitempty"`
	CustomerID string                 `json:"customer_id,omitempty"`
	Timestamp  int64                  `json:"timestamp,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

type IdentityReport struct {
	CustomerID string                 `json:"customer_id,omitempty"`
	Email      string                 `json:"email,omitempty"`
	Phone      string                 `json:"phone,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

type VariantImage struct {
	VariantID int64  `json:"variant_id,omitempty"`
	ImageID   int64  `json:"image_id,omitempty"`
	Src       string `json:"src,omitempty"`
	Position  int    `json:"position,omitempty"`
}

type VariantImageUpdate struct {
	VariantID int64   `json:"variant_id,omitempty"`
	ImageIDs  []int64 `json:"image_ids,omitempty"`
}

// JSON wrappers
type productSizesResource struct {
	Data []ProductSizeData `json:"data"`
}
type sizeTemplatesResource struct {
	Data []SizeTemplate `json:"data"`
}
type variantImagesResource struct {
	VariantImages []VariantImage `json:"variant_images"`
}

// =====================================================================
// Size Chart Implementation
// =====================================================================

// POST app_open_api/size_chart/products/batch_query.json
func (s *sizeChartOp) BatchQueryProductSizes(ctx context.Context, productIDs []int64) ([]ProductSizeData, error) {
	r := &productSizesResource{}
	body := map[string][]int64{"product_ids": productIDs}
	err := s.client.Post(ctx, s.client.CreatePath("app_open_api/size_chart/products/batch_query.json"), body, r)
	return r.Data, err
}

// POST app_open_api/size_chart/categories/batch_query.json
func (s *sizeChartOp) BatchQueryCategoryTemplates(ctx context.Context, categoryIDs []int64) ([]SizeTemplate, error) {
	r := &sizeTemplatesResource{}
	body := map[string][]int64{"category_ids": categoryIDs}
	err := s.client.Post(ctx, s.client.CreatePath("app_open_api/size_chart/categories/batch_query.json"), body, r)
	return r.Data, err
}

// GET app_open_api/size_chart/templates.json
func (s *sizeChartOp) BatchQueryStoreTemplates(ctx context.Context, opts *core.ListOptions) ([]SizeTemplate, error) {
	r := &sizeTemplatesResource{}
	err := s.client.Get(ctx, s.client.CreatePath("app_open_api/size_chart/templates.json"), r, opts)
	return r.Data, err
}

// POST app_open_api/size_chart/products/batch_delete.json
func (s *sizeChartOp) BatchDeleteProductSizes(ctx context.Context, productIDs []int64) error {
	body := map[string][]int64{"product_ids": productIDs}
	return s.client.Post(ctx, s.client.CreatePath("app_open_api/size_chart/products/batch_delete.json"), body, nil)
}

// POST app_open_api/size_chart/products/batch_upsert.json
func (s *sizeChartOp) BatchCreateOrUpdateProductSizes(ctx context.Context, data []ProductSizeData) ([]ProductSizeData, error) {
	r := &productSizesResource{}
	body := map[string][]ProductSizeData{"data": data}
	err := s.client.Post(ctx, s.client.CreatePath("app_open_api/size_chart/products/batch_upsert.json"), body, r)
	return r.Data, err
}

// =====================================================================
// CDP Implementation
// =====================================================================

// POST app_open_api/cdp/events.json
func (s *cdpOp) ReportBehaviorEvents(ctx context.Context, events []BehaviorEvent) error {
	body := map[string][]BehaviorEvent{"events": events}
	return s.client.Post(ctx, s.client.CreatePath("app_open_api/cdp/events.json"), body, nil)
}

// POST app_open_api/cdp/identity.json
func (s *cdpOp) ReportIdentity(ctx context.Context, identity IdentityReport) error {
	return s.client.Post(ctx, s.client.CreatePath("app_open_api/cdp/identity.json"), identity, nil)
}

// =====================================================================
// Variant Images Implementation
// =====================================================================

// GET app_open_api/variants/{variant_id}/images.json
func (s *variantImgOp) QueryVariantImages(ctx context.Context, variantID int64) ([]VariantImage, error) {
	r := &variantImagesResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("app_open_api/variants/%d/images.json", variantID)), r, nil)
	return r.VariantImages, err
}

// PUT app_open_api/variants/images/batch_update.json
func (s *variantImgOp) BatchUpdateVariantImages(ctx context.Context, updates []VariantImageUpdate) error {
	body := map[string][]VariantImageUpdate{"updates": updates}
	return s.client.Put(ctx, s.client.CreatePath("app_open_api/variants/images/batch_update.json"), body, nil)
}

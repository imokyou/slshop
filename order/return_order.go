package order

import (
	"context"
	"fmt"
	"time"

	"github.com/imokyou/slshop/core"
)

// === Return Order ===

type ReturnService interface {
	List(ctx context.Context, opts *core.ListOptions) ([]Return, error)
	Create(ctx context.Context, orderID int64, ret Return) (*Return, error)
	Close(ctx context.Context, returnID int64) (*Return, error)
	ListFulfillments(ctx context.Context, opts *core.ListOptions) ([]ReturnFulfillment, error)
	CreateFulfillment(ctx context.Context, returnID int64, f ReturnFulfillment) (*ReturnFulfillment, error)
	UpdateFulfillmentTracking(ctx context.Context, returnID, fID int64, t FulfillmentTracking) (*ReturnFulfillment, error)
	ListFulfillmentOrders(ctx context.Context, opts *core.ListOptions) ([]ReturnFulfillmentOrder, error)
}

func NewReturnService(client core.Requester) ReturnService {
	return &returnOp{client: client}
}

type returnOp struct{ client core.Requester }

type Return struct {
	ID              int64            `json:"id,omitempty"`
	OrderID         int64            `json:"order_id,omitempty"`
	Status          string           `json:"status,omitempty"`
	Note            string           `json:"note,omitempty"`
	ReturnLineItems []ReturnLineItem `json:"return_line_items,omitempty"`
	CreatedAt       *time.Time       `json:"created_at,omitempty"`
	UpdatedAt       *time.Time       `json:"updated_at,omitempty"`
	ClosedAt        *time.Time       `json:"closed_at,omitempty"`
}

type ReturnLineItem struct {
	ID               int64  `json:"id,omitempty"`
	LineItemID       int64  `json:"line_item_id,omitempty"`
	Quantity         int    `json:"quantity,omitempty"`
	RestockType      string `json:"restock_type,omitempty"`
	ReturnReason     string `json:"return_reason,omitempty"`
	ReturnReasonNote string `json:"return_reason_note,omitempty"`
}

type ReturnFulfillment struct {
	ID              int64      `json:"id,omitempty"`
	ReturnID        int64      `json:"return_id,omitempty"`
	Status          string     `json:"status,omitempty"`
	TrackingCompany string     `json:"tracking_company,omitempty"`
	TrackingNumber  string     `json:"tracking_number,omitempty"`
	TrackingURL     string     `json:"tracking_url,omitempty"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty"`
}

type ReturnFulfillmentOrder struct {
	ID       int64  `json:"id,omitempty"`
	ReturnID int64  `json:"return_id,omitempty"`
	Status   string `json:"status,omitempty"`
	OrderID  int64  `json:"order_id,omitempty"`
}

type returnResource struct {
	Return *Return `json:"return"`
}
type returnsResource struct {
	Returns []Return `json:"returns"`
}
type returnFulfillmentResource struct {
	ReturnFulfillment *ReturnFulfillment `json:"return_fulfillment"`
}
type returnFulfillmentsResource struct {
	ReturnFulfillments []ReturnFulfillment `json:"return_fulfillments"`
}
type returnFulfillmentOrdersResource struct {
	ReturnFulfillmentOrders []ReturnFulfillmentOrder `json:"return_fulfillment_orders"`
}

func (s *returnOp) List(ctx context.Context, opts *core.ListOptions) ([]Return, error) {
	r := &returnsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("returns.json"), r, opts)
	return r.Returns, err
}
func (s *returnOp) Create(ctx context.Context, orderID int64, ret Return) (*Return, error) {
	r := &returnResource{}
	err := s.client.Post(ctx, s.client.CreatePath(fmt.Sprintf("orders/%d/returns.json", orderID)), returnResource{Return: &ret}, r)
	return r.Return, err
}
func (s *returnOp) Close(ctx context.Context, returnID int64) (*Return, error) {
	r := &returnResource{}
	err := s.client.Post(ctx, s.client.CreatePath(fmt.Sprintf("returns/%d/close.json", returnID)), nil, r)
	return r.Return, err
}
func (s *returnOp) ListFulfillments(ctx context.Context, opts *core.ListOptions) ([]ReturnFulfillment, error) {
	r := &returnFulfillmentsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("return_fulfillments.json"), r, opts)
	return r.ReturnFulfillments, err
}
func (s *returnOp) CreateFulfillment(ctx context.Context, returnID int64, f ReturnFulfillment) (*ReturnFulfillment, error) {
	r := &returnFulfillmentResource{}
	err := s.client.Post(ctx, s.client.CreatePath(fmt.Sprintf("returns/%d/return_fulfillments.json", returnID)), returnFulfillmentResource{ReturnFulfillment: &f}, r)
	return r.ReturnFulfillment, err
}
func (s *returnOp) UpdateFulfillmentTracking(ctx context.Context, returnID, fID int64, t FulfillmentTracking) (*ReturnFulfillment, error) {
	r := &returnFulfillmentResource{}
	err := s.client.Post(ctx, s.client.CreatePath(fmt.Sprintf("returns/%d/return_fulfillments/%d/update_tracking.json", returnID, fID)), t, r)
	return r.ReturnFulfillment, err
}
func (s *returnOp) ListFulfillmentOrders(ctx context.Context, opts *core.ListOptions) ([]ReturnFulfillmentOrder, error) {
	r := &returnFulfillmentOrdersResource{}
	err := s.client.Get(ctx, s.client.CreatePath("return_fulfillment_orders.json"), r, opts)
	return r.ReturnFulfillmentOrders, err
}

// === Order Archive ===

type ArchiveService interface {
	Archive(ctx context.Context, orderID int64) error
	Unarchive(ctx context.Context, orderID int64) error
}

func NewArchiveService(client core.Requester) ArchiveService {
	return &archiveOp{client: client}
}

type archiveOp struct{ client core.Requester }

func (s *archiveOp) Archive(ctx context.Context, orderID int64) error {
	return s.client.Put(ctx, s.client.CreatePath(fmt.Sprintf("orders/%d/archive.json", orderID)), nil, nil)
}
func (s *archiveOp) Unarchive(ctx context.Context, orderID int64) error {
	return s.client.Put(ctx, s.client.CreatePath(fmt.Sprintf("orders/%d/unarchive.json", orderID)), nil, nil)
}

// === Order Edit ===

type EditService interface {
	Start(ctx context.Context, orderID int64) (*EditSession, error)
	SetQuantity(ctx context.Context, orderID int64, e EditSetQuantity) (*EditSession, error)
	AddLineItem(ctx context.Context, orderID int64, e EditAddLineItem) (*EditSession, error)
	AddCustomItem(ctx context.Context, orderID int64, e EditAddCustomItem) (*EditSession, error)
	AddDiscount(ctx context.Context, orderID int64, e EditAddDiscount) (*EditSession, error)
	RemoveDiscount(ctx context.Context, orderID int64, e EditRemoveDiscount) (*EditSession, error)
	Commit(ctx context.Context, orderID int64) (*Order, error)
}

func NewEditService(client core.Requester) EditService {
	return &editOp{client: client}
}

type editOp struct{ client core.Requester }

type EditSession struct {
	ID      int64  `json:"id,omitempty"`
	OrderID int64  `json:"order_id,omitempty"`
	Status  string `json:"status,omitempty"`
}

type EditSetQuantity struct {
	LineItemID int64 `json:"line_item_id"`
	Quantity   int   `json:"quantity"`
}
type EditAddLineItem struct {
	VariantID  int64 `json:"variant_id"`
	Quantity   int   `json:"quantity"`
	LocationID int64 `json:"location_id,omitempty"`
}
type EditAddCustomItem struct {
	Title            string `json:"title"`
	Price            string `json:"price"`
	Quantity         int    `json:"quantity"`
	RequiresShipping bool   `json:"requires_shipping,omitempty"`
	Taxable          bool   `json:"taxable,omitempty"`
}
type EditAddDiscount struct {
	LineItemID   int64  `json:"line_item_id"`
	DiscountType string `json:"discount_type,omitempty"`
	Value        string `json:"value"`
	Description  string `json:"description,omitempty"`
}
type EditRemoveDiscount struct {
	DiscountID int64 `json:"discount_id"`
}

type editSessionResource struct {
	EditSession *EditSession `json:"edit_session"`
}

func (s *editOp) Start(ctx context.Context, orderID int64) (*EditSession, error) {
	r := &editSessionResource{}
	err := s.client.Post(ctx, s.client.CreatePath(fmt.Sprintf("orders/%d/edit/start.json", orderID)), nil, r)
	return r.EditSession, err
}
func (s *editOp) SetQuantity(ctx context.Context, orderID int64, e EditSetQuantity) (*EditSession, error) {
	r := &editSessionResource{}
	err := s.client.Post(ctx, s.client.CreatePath(fmt.Sprintf("orders/%d/edit/set_quantity.json", orderID)), e, r)
	return r.EditSession, err
}
func (s *editOp) AddLineItem(ctx context.Context, orderID int64, e EditAddLineItem) (*EditSession, error) {
	r := &editSessionResource{}
	err := s.client.Post(ctx, s.client.CreatePath(fmt.Sprintf("orders/%d/edit/add_line_item.json", orderID)), e, r)
	return r.EditSession, err
}
func (s *editOp) AddCustomItem(ctx context.Context, orderID int64, e EditAddCustomItem) (*EditSession, error) {
	r := &editSessionResource{}
	err := s.client.Post(ctx, s.client.CreatePath(fmt.Sprintf("orders/%d/edit/add_custom_item.json", orderID)), e, r)
	return r.EditSession, err
}
func (s *editOp) AddDiscount(ctx context.Context, orderID int64, e EditAddDiscount) (*EditSession, error) {
	r := &editSessionResource{}
	err := s.client.Post(ctx, s.client.CreatePath(fmt.Sprintf("orders/%d/edit/add_discount.json", orderID)), e, r)
	return r.EditSession, err
}
func (s *editOp) RemoveDiscount(ctx context.Context, orderID int64, e EditRemoveDiscount) (*EditSession, error) {
	r := &editSessionResource{}
	err := s.client.Post(ctx, s.client.CreatePath(fmt.Sprintf("orders/%d/edit/remove_discount.json", orderID)), e, r)
	return r.EditSession, err
}
func (s *editOp) Commit(ctx context.Context, orderID int64) (*Order, error) {
	r := &orderResource{}
	err := s.client.Post(ctx, s.client.CreatePath(fmt.Sprintf("orders/%d/edit/commit.json", orderID)), nil, r)
	return r.Order, err
}

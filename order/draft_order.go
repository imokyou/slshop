package order

import (
	"context"
	"fmt"
	"time"

	"github.com/imokyou/slshop/core"
)

const draftOrdersBasePath = "orders/draft_orders"

type DraftOrderService interface {
	Create(ctx context.Context, order DraftOrder) (*DraftOrder, error)
	Update(ctx context.Context, order DraftOrder) (*DraftOrder, error)
	Get(ctx context.Context, id int64) (*DraftOrder, error)
	Delete(ctx context.Context, id int64) error
	Complete(ctx context.Context, id int64) (*DraftOrder, error)
	Count(ctx context.Context) (int, error)
	SendInvoice(ctx context.Context, id int64, invoice DraftOrderInvoice) (*DraftOrderInvoice, error)
}

func NewDraftOrderService(client core.Requester) DraftOrderService {
	return &draftOrderOp{client: client}
}

type draftOrderOp struct{ client core.Requester }

type DraftOrder struct {
	ID              int64                    `json:"id,omitempty"`
	Name            string                   `json:"name,omitempty"`
	Email           string                   `json:"email,omitempty"`
	Currency        string                   `json:"currency,omitempty"`
	Status          string                   `json:"status,omitempty"`
	Note            string                   `json:"note,omitempty"`
	Tags            string                   `json:"tags,omitempty"`
	TotalPrice      string                   `json:"total_price,omitempty"`
	SubtotalPrice   string                   `json:"subtotal_price,omitempty"`
	TotalTax        string                   `json:"total_tax,omitempty"`
	TaxesIncluded   bool                     `json:"taxes_included,omitempty"`
	Customer        *core.Customer       `json:"customer,omitempty"`
	BillingAddress  *core.Address        `json:"billing_address,omitempty"`
	ShippingAddress *core.Address        `json:"shipping_address,omitempty"`
	ShippingLine    *core.ShippingLine   `json:"shipping_line,omitempty"`
	LineItems       []core.LineItem      `json:"line_items,omitempty"`
	TaxLines        []core.TaxLine       `json:"tax_lines,omitempty"`
	NoteAttributes  []core.NoteAttribute `json:"note_attributes,omitempty"`
	OrderID         int64                    `json:"order_id,omitempty"`
	InvoiceURL      string                   `json:"invoice_url,omitempty"`
	CreatedAt       *time.Time               `json:"created_at,omitempty"`
	UpdatedAt       *time.Time               `json:"updated_at,omitempty"`
	CompletedAt     *time.Time               `json:"completed_at,omitempty"`
}

type DraftOrderInvoice struct {
	To            string   `json:"to,omitempty"`
	From          string   `json:"from,omitempty"`
	Subject       string   `json:"subject,omitempty"`
	CustomMessage string   `json:"custom_message,omitempty"`
	Bcc           []string `json:"bcc,omitempty"`
}

type draftOrderResource struct {
	DraftOrder *DraftOrder `json:"draft_order"`
}
type draftOrdersCountResource struct {
	Count int `json:"count"`
}
type draftOrderInvoiceResource struct {
	DraftOrderInvoice *DraftOrderInvoice `json:"draft_order_invoice"`
}

func (s *draftOrderOp) Create(ctx context.Context, order DraftOrder) (*DraftOrder, error) {
	path := s.client.CreatePath(draftOrdersBasePath + ".json")
	body := draftOrderResource{DraftOrder: &order}
	resource := &draftOrderResource{}
	err := s.client.Post(ctx, path, body, resource)
	return resource.DraftOrder, err
}

func (s *draftOrderOp) Update(ctx context.Context, order DraftOrder) (*DraftOrder, error) {
	path := s.client.CreatePath(fmt.Sprintf("%s/%d.json", draftOrdersBasePath, order.ID))
	body := draftOrderResource{DraftOrder: &order}
	resource := &draftOrderResource{}
	err := s.client.Put(ctx, path, body, resource)
	return resource.DraftOrder, err
}

func (s *draftOrderOp) Get(ctx context.Context, id int64) (*DraftOrder, error) {
	path := s.client.CreatePath(fmt.Sprintf("%s/%d.json", draftOrdersBasePath, id))
	resource := &draftOrderResource{}
	err := s.client.Get(ctx, path, resource, nil)
	return resource.DraftOrder, err
}

func (s *draftOrderOp) Delete(ctx context.Context, id int64) error {
	return s.client.Delete(ctx, s.client.CreatePath(fmt.Sprintf("%s/%d.json", draftOrdersBasePath, id)))
}

func (s *draftOrderOp) Complete(ctx context.Context, id int64) (*DraftOrder, error) {
	path := s.client.CreatePath(fmt.Sprintf("%s/%d/complete.json", draftOrdersBasePath, id))
	resource := &draftOrderResource{}
	err := s.client.Put(ctx, path, nil, resource)
	return resource.DraftOrder, err
}

func (s *draftOrderOp) Count(ctx context.Context) (int, error) {
	path := s.client.CreatePath(draftOrdersBasePath + "/count.json")
	resource := &draftOrdersCountResource{}
	err := s.client.Get(ctx, path, resource, nil)
	return resource.Count, err
}

func (s *draftOrderOp) SendInvoice(ctx context.Context, id int64, invoice DraftOrderInvoice) (*DraftOrderInvoice, error) {
	path := s.client.CreatePath(fmt.Sprintf("%s/%d/send_invoice.json", draftOrdersBasePath, id))
	body := draftOrderInvoiceResource{DraftOrderInvoice: &invoice}
	resource := &draftOrderInvoiceResource{}
	err := s.client.Post(ctx, path, body, resource)
	return resource.DraftOrderInvoice, err
}

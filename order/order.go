package order

import (
	"context"
	"fmt"
	"time"

	"github.com/imokyou/slshop/core"
)

const ordersBasePath = "orders"

// =====================================================================
// Service Interface
// =====================================================================

type Service interface {
	List(ctx context.Context, opts *ListOptions) ([]Order, error)
	Count(ctx context.Context, opts *CountOptions) (int, error)
	Get(ctx context.Context, id int64) (*Order, error)
	Create(ctx context.Context, order Order) (*Order, error)
	Update(ctx context.Context, order Order) (*Order, error)
	Delete(ctx context.Context, id int64) error

	Cancel(ctx context.Context, id int64, opts *CancelOptions) (*Order, error)
	Close(ctx context.Context, id int64) (*Order, error)
	Open(ctx context.Context, id int64) (*Order, error)

	ListRefunds(ctx context.Context, orderID int64) ([]Refund, error)
	GetRefund(ctx context.Context, orderID, refundID int64) (*Refund, error)
	CreateRefund(ctx context.Context, orderID int64, refund Refund) (*Refund, error)
	CalculateRefund(ctx context.Context, orderID int64, refund Refund) (*Refund, error)

	ListRisks(ctx context.Context, orderID int64) ([]Risk, error)
	GetRisk(ctx context.Context, orderID, riskID int64) (*Risk, error)
	CreateRisk(ctx context.Context, orderID int64, risk Risk) (*Risk, error)
	UpdateRisk(ctx context.Context, orderID int64, risk Risk) (*Risk, error)
	DeleteRisk(ctx context.Context, orderID, riskID int64) error
	DeleteAllRisks(ctx context.Context, orderID int64) error

	ListTransactions(ctx context.Context, orderID int64) ([]Transaction, error)
	GetTransaction(ctx context.Context, orderID, transactionID int64) (*Transaction, error)
}

// NewService creates a new order Service.
func NewService(client core.Requester) Service {
	return &serviceOp{client: client}
}

type serviceOp struct{ client core.Requester }

// =====================================================================
// Query Options
// =====================================================================

type ListOptions struct {
	core.ListOptions
	Status            string `url:"status,omitempty"`
	FinancialStatus   string `url:"financial_status,omitempty"`
	FulfillmentStatus string `url:"fulfillment_status,omitempty"`
	IDs               string `url:"ids,omitempty"`
	Name              string `url:"name,omitempty"`
	ProcessedAtMin    string `url:"processed_at_min,omitempty"`
	ProcessedAtMax    string `url:"processed_at_max,omitempty"`
}

type CountOptions struct {
	core.CountOptions
	Status            string `url:"status,omitempty"`
	FinancialStatus   string `url:"financial_status,omitempty"`
	FulfillmentStatus string `url:"fulfillment_status,omitempty"`
}

type CancelOptions struct {
	Reason  string `json:"reason,omitempty"`
	Restock bool   `json:"restock,omitempty"`
	Email   bool   `json:"email,omitempty"`
}

// =====================================================================
// Models
// =====================================================================

type Order struct {
	ID                      int64                    `json:"id,omitempty"`
	Name                    string                   `json:"name,omitempty"`
	OrderNumber             int                      `json:"order_number,omitempty"`
	Email                   string                   `json:"email,omitempty"`
	Phone                   string                   `json:"phone,omitempty"`
	Token                   string                   `json:"token,omitempty"`
	Note                    string                   `json:"note,omitempty"`
	OrderNote               string                   `json:"order_note,omitempty"`
	BuyerNote               string                   `json:"buyer_note,omitempty"`
	Tags                    string                   `json:"tags,omitempty"`
	Currency                string                   `json:"currency,omitempty"`
	ExchangeRate            string                   `json:"exchange_rate,omitempty"`
	CustomerLocale          string                   `json:"customer_locale,omitempty"`
	MarketRegionCountryCode string                   `json:"market_region_country_code,omitempty"`
	CompanyLocationID       string                   `json:"company_location_id,omitempty"`
	TotalPrice              string                   `json:"total_price,omitempty"`
	SubtotalPrice           string                   `json:"subtotal_price,omitempty"`
	TotalTax                string                   `json:"total_tax,omitempty"`
	TotalDiscounts          string                   `json:"total_discounts,omitempty"`
	TotalShippingPrice      string                   `json:"total_shipping_price,omitempty"`
	TotalWeight             float64                  `json:"total_weight,omitempty"`
	TotalLineItemsPrice     string                   `json:"total_line_items_price,omitempty"`
	PriceInfo               *PriceInfo               `json:"price_info,omitempty"`
	FinancialStatus         string                   `json:"financial_status,omitempty"`
	FulfillmentStatus       string                   `json:"fulfillment_status,omitempty"`
	CancelReason            string                   `json:"cancel_reason,omitempty"`
	InventoryBehaviour      string                   `json:"inventory_behaviour,omitempty"`
	SendReceipt             *bool                    `json:"send_receipt,omitempty"`
	SendFulfillmentReceipt  *bool                    `json:"send_fulfillment_receipt,omitempty"`
	Gateway                 string                   `json:"gateway,omitempty"`
	Test                    bool                     `json:"test,omitempty"`
	Confirmed               bool                     `json:"confirmed,omitempty"`
	BuyerAcceptsMarketing   bool                     `json:"buyer_accepts_marketing,omitempty"`
	TaxesIncluded           bool                     `json:"taxes_included,omitempty"`
	Customer                *core.Customer       `json:"customer,omitempty"`
	BillingAddress          *core.Address        `json:"billing_address,omitempty"`
	ShippingAddress         *core.Address        `json:"shipping_address,omitempty"`
	ShippingLine            *core.ShippingLine   `json:"shipping_line,omitempty"`
	LineItems               []core.LineItem      `json:"line_items,omitempty"`
	ShippingLines           []core.ShippingLine  `json:"shipping_lines,omitempty"`
	TaxLines                []core.TaxLine       `json:"tax_lines,omitempty"`
	DiscountCodes           []core.DiscountCode  `json:"discount_codes,omitempty"`
	Refunds                 []Refund                 `json:"refunds,omitempty"`
	NoteAttributes          []core.NoteAttribute `json:"note_attributes,omitempty"`
	TransactionList         []Transaction            `json:"transaction_list,omitempty"`
	Transactions            *TransactionRef          `json:"transactions,omitempty"`
	CreatedAt               *time.Time               `json:"created_at,omitempty"`
	UpdatedAt               *time.Time               `json:"updated_at,omitempty"`
	ClosedAt                *time.Time               `json:"closed_at,omitempty"`
	CancelledAt             *time.Time               `json:"cancelled_at,omitempty"`
	ProcessedAt             *time.Time               `json:"processed_at,omitempty"`
}

type PriceInfo struct {
	CurrentExtraTotalDiscounts string `json:"current_extra_total_discounts,omitempty"`
	TaxesIncluded              bool   `json:"taxes_included,omitempty"`
	TotalShippingPrice         string `json:"total_shipping_price,omitempty"`
}

type TransactionRef struct {
	ID string `json:"id,omitempty"`
}

type Refund struct {
	ID              int64            `json:"id,omitempty"`
	OrderID         int64            `json:"order_id,omitempty"`
	Note            string           `json:"note,omitempty"`
	Restock         bool             `json:"restock,omitempty"`
	Shipping        *RefundShipping  `json:"shipping,omitempty"`
	RefundLineItems []RefundLineItem `json:"refund_line_items,omitempty"`
	Transactions    []Transaction    `json:"transactions,omitempty"`
	Currency        string           `json:"currency,omitempty"`
	CreatedAt       *time.Time       `json:"created_at,omitempty"`
	ProcessedAt     *time.Time       `json:"processed_at,omitempty"`
}

type RefundShipping struct {
	Amount     string `json:"amount,omitempty"`
	Tax        string `json:"tax,omitempty"`
	FullRefund bool   `json:"full_refund,omitempty"`
}

type RefundLineItem struct {
	ID          int64              `json:"id,omitempty"`
	LineItemID  int64              `json:"line_item_id,omitempty"`
	LineItem    *core.LineItem `json:"line_item,omitempty"`
	Quantity    int                `json:"quantity,omitempty"`
	RestockType string             `json:"restock_type,omitempty"`
	LocationID  int64              `json:"location_id,omitempty"`
	Subtotal    string             `json:"subtotal,omitempty"`
	TotalTax    string             `json:"total_tax,omitempty"`
}

type Risk struct {
	ID              int64  `json:"id,omitempty"`
	OrderID         int64  `json:"order_id,omitempty"`
	CauseCancel     bool   `json:"cause_cancel,omitempty"`
	Display         bool   `json:"display,omitempty"`
	MerchantMessage string `json:"merchant_message,omitempty"`
	Message         string `json:"message,omitempty"`
	Recommendation  string `json:"recommendation,omitempty"`
	Score           string `json:"score,omitempty"`
	Source          string `json:"source,omitempty"`
}

type Transaction struct {
	ID            int64      `json:"id,omitempty"`
	OrderID       int64      `json:"order_id,omitempty"`
	Amount        string     `json:"amount,omitempty"`
	Currency      string     `json:"currency,omitempty"`
	Kind          string     `json:"kind,omitempty"`
	Status        string     `json:"status,omitempty"`
	Gateway       string     `json:"gateway,omitempty"`
	Gateways      string     `json:"gateways,omitempty"`
	Message       string     `json:"message,omitempty"`
	ErrorCode     string     `json:"error_code,omitempty"`
	Test          bool       `json:"test,omitempty"`
	Authorization string     `json:"authorization,omitempty"`
	ParentID      int64      `json:"parent_id,omitempty"`
	ProcessedAt   *time.Time `json:"processed_at,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
}

// =====================================================================
// JSON Wrappers
// =====================================================================

type orderResource struct {
	Order *Order `json:"order"`
}
type ordersResource struct {
	Orders []Order `json:"orders"`
}
type countResource struct {
	Count int `json:"count"`
}
type refundResource struct {
	Refund *Refund `json:"refund"`
}
type refundsResource struct {
	Refunds []Refund `json:"refunds"`
}
type riskResource struct {
	Risk *Risk `json:"risk"`
}
type risksResource struct {
	Risks []Risk `json:"risks"`
}
type transactionResource struct {
	Transaction *Transaction `json:"transaction"`
}
type transactionsResource struct {
	Transactions []Transaction `json:"transactions"`
}

// =====================================================================
// Implementation
// =====================================================================

func (s *serviceOp) List(ctx context.Context, opts *ListOptions) ([]Order, error) {
	path := s.client.CreatePath(ordersBasePath + ".json")
	resource := &ordersResource{}
	err := s.client.Get(ctx, path, resource, opts)
	return resource.Orders, err
}

func (s *serviceOp) Count(ctx context.Context, opts *CountOptions) (int, error) {
	path := s.client.CreatePath(ordersBasePath + "/count.json")
	resource := &countResource{}
	err := s.client.Get(ctx, path, resource, opts)
	return resource.Count, err
}

func (s *serviceOp) Get(ctx context.Context, id int64) (*Order, error) {
	path := s.client.CreatePath(fmt.Sprintf("%s/%d.json", ordersBasePath, id))
	resource := &orderResource{}
	err := s.client.Get(ctx, path, resource, nil)
	return resource.Order, err
}

func (s *serviceOp) Create(ctx context.Context, order Order) (*Order, error) {
	path := s.client.CreatePath(ordersBasePath + ".json")
	body := orderResource{Order: &order}
	resource := &orderResource{}
	err := s.client.Post(ctx, path, body, resource)
	return resource.Order, err
}

func (s *serviceOp) Update(ctx context.Context, order Order) (*Order, error) {
	path := s.client.CreatePath(fmt.Sprintf("%s/%d.json", ordersBasePath, order.ID))
	body := orderResource{Order: &order}
	resource := &orderResource{}
	err := s.client.Put(ctx, path, body, resource)
	return resource.Order, err
}

func (s *serviceOp) Delete(ctx context.Context, id int64) error {
	return s.client.Delete(ctx, s.client.CreatePath(fmt.Sprintf("%s/%d.json", ordersBasePath, id)))
}

func (s *serviceOp) Cancel(ctx context.Context, id int64, opts *CancelOptions) (*Order, error) {
	path := s.client.CreatePath(fmt.Sprintf("%s/%d/cancel.json", ordersBasePath, id))
	resource := &orderResource{}
	err := s.client.Post(ctx, path, opts, resource)
	return resource.Order, err
}

func (s *serviceOp) Close(ctx context.Context, id int64) (*Order, error) {
	path := s.client.CreatePath(fmt.Sprintf("%s/%d/close.json", ordersBasePath, id))
	resource := &orderResource{}
	err := s.client.Post(ctx, path, nil, resource)
	return resource.Order, err
}

func (s *serviceOp) Open(ctx context.Context, id int64) (*Order, error) {
	path := s.client.CreatePath(fmt.Sprintf("%s/%d/open.json", ordersBasePath, id))
	resource := &orderResource{}
	err := s.client.Post(ctx, path, nil, resource)
	return resource.Order, err
}

func (s *serviceOp) ListRefunds(ctx context.Context, orderID int64) ([]Refund, error) {
	path := s.client.CreatePath(fmt.Sprintf("%s/%d/refunds.json", ordersBasePath, orderID))
	resource := &refundsResource{}
	err := s.client.Get(ctx, path, resource, nil)
	return resource.Refunds, err
}

func (s *serviceOp) GetRefund(ctx context.Context, orderID, refundID int64) (*Refund, error) {
	path := s.client.CreatePath(fmt.Sprintf("%s/%d/refunds/%d.json", ordersBasePath, orderID, refundID))
	resource := &refundResource{}
	err := s.client.Get(ctx, path, resource, nil)
	return resource.Refund, err
}

func (s *serviceOp) CreateRefund(ctx context.Context, orderID int64, refund Refund) (*Refund, error) {
	path := s.client.CreatePath(fmt.Sprintf("%s/%d/refunds.json", ordersBasePath, orderID))
	body := refundResource{Refund: &refund}
	resource := &refundResource{}
	err := s.client.Post(ctx, path, body, resource)
	return resource.Refund, err
}

func (s *serviceOp) CalculateRefund(ctx context.Context, orderID int64, refund Refund) (*Refund, error) {
	path := s.client.CreatePath(fmt.Sprintf("%s/%d/refunds/calculate.json", ordersBasePath, orderID))
	body := refundResource{Refund: &refund}
	resource := &refundResource{}
	err := s.client.Post(ctx, path, body, resource)
	return resource.Refund, err
}

func (s *serviceOp) ListRisks(ctx context.Context, orderID int64) ([]Risk, error) {
	path := s.client.CreatePath(fmt.Sprintf("%s/%d/risks.json", ordersBasePath, orderID))
	resource := &risksResource{}
	err := s.client.Get(ctx, path, resource, nil)
	return resource.Risks, err
}

func (s *serviceOp) GetRisk(ctx context.Context, orderID, riskID int64) (*Risk, error) {
	path := s.client.CreatePath(fmt.Sprintf("%s/%d/risks/%d.json", ordersBasePath, orderID, riskID))
	resource := &riskResource{}
	err := s.client.Get(ctx, path, resource, nil)
	return resource.Risk, err
}

func (s *serviceOp) CreateRisk(ctx context.Context, orderID int64, risk Risk) (*Risk, error) {
	path := s.client.CreatePath(fmt.Sprintf("%s/%d/risks.json", ordersBasePath, orderID))
	body := riskResource{Risk: &risk}
	resource := &riskResource{}
	err := s.client.Post(ctx, path, body, resource)
	return resource.Risk, err
}

func (s *serviceOp) UpdateRisk(ctx context.Context, orderID int64, risk Risk) (*Risk, error) {
	path := s.client.CreatePath(fmt.Sprintf("%s/%d/risks/%d.json", ordersBasePath, orderID, risk.ID))
	body := riskResource{Risk: &risk}
	resource := &riskResource{}
	err := s.client.Put(ctx, path, body, resource)
	return resource.Risk, err
}

func (s *serviceOp) DeleteRisk(ctx context.Context, orderID, riskID int64) error {
	return s.client.Delete(ctx, s.client.CreatePath(fmt.Sprintf("%s/%d/risks/%d.json", ordersBasePath, orderID, riskID)))
}

func (s *serviceOp) DeleteAllRisks(ctx context.Context, orderID int64) error {
	return s.client.Delete(ctx, s.client.CreatePath(fmt.Sprintf("%s/%d/risks.json", ordersBasePath, orderID)))
}

func (s *serviceOp) ListTransactions(ctx context.Context, orderID int64) ([]Transaction, error) {
	path := s.client.CreatePath(fmt.Sprintf("%s/%d/transactions.json", ordersBasePath, orderID))
	resource := &transactionsResource{}
	err := s.client.Get(ctx, path, resource, nil)
	return resource.Transactions, err
}

func (s *serviceOp) GetTransaction(ctx context.Context, orderID, transactionID int64) (*Transaction, error) {
	path := s.client.CreatePath(fmt.Sprintf("%s/%d/transactions/%d.json", ordersBasePath, orderID, transactionID))
	resource := &transactionResource{}
	err := s.client.Get(ctx, path, resource, nil)
	return resource.Transaction, err
}

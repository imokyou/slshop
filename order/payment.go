package order

import (
	"context"
	"fmt"
	"time"

	"github.com/imokyou/slshop/core"
)

// Payment + TransactionList + AbandonedCheckout + Subscription + Tax
// + Return + OrderArchive + OrderEdit

// === Payment ===

type PaymentService interface {
	CreatePaymentSlip(ctx context.Context, slip PaymentSlip) (*PaymentSlip, error)
	UpdatePaymentSlip(ctx context.Context, slip PaymentSlip) (*PaymentSlip, error)
	GetSettings(ctx context.Context) (*PaymentSettings, error)
	ListChannels(ctx context.Context) ([]PaymentChannel, error)
	ListPayments(ctx context.Context, orderID int64) ([]OrderPayment, error)
}

func NewPaymentService(client core.Requester) PaymentService {
	return &paymentOp{client: client}
}

type paymentOp struct{ client core.Requester }

type PaymentSlip struct {
	ID          int64      `json:"id,omitempty"`
	OrderID     int64      `json:"order_id,omitempty"`
	OrderSeq    string     `json:"order_seq,omitempty"`
	Amount      string     `json:"amount,omitempty"`
	Currency    string     `json:"currency,omitempty"`
	Status      string     `json:"status,omitempty"`
	Gateway     string     `json:"gateway,omitempty"`
	Kind        string     `json:"kind,omitempty"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
}

type PaymentSettings struct {
	Provider            string   `json:"provider,omitempty"`
	TestMode            bool     `json:"test_mode,omitempty"`
	SupportedCurrencies []string `json:"supported_currencies,omitempty"`
}

type PaymentChannel struct {
	ID       int64  `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Gateway  string `json:"gateway,omitempty"`
	Enabled  bool   `json:"enabled,omitempty"`
	Provider string `json:"provider,omitempty"`
}

type OrderPayment struct {
	ID          int64      `json:"id,omitempty"`
	OrderID     int64      `json:"order_id,omitempty"`
	Amount      string     `json:"amount,omitempty"`
	Currency    string     `json:"currency,omitempty"`
	Gateway     string     `json:"gateway,omitempty"`
	Status      string     `json:"status,omitempty"`
	Kind        string     `json:"kind,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}

type paymentSlipResource struct {
	Transaction *PaymentSlip `json:"transaction"`
}
type paymentSettingsResource struct {
	Settings *PaymentSettings `json:"settings"`
}
type paymentChannelsResource struct {
	Channels []PaymentChannel `json:"channels"`
}
type orderPaymentsResource struct {
	Payments []OrderPayment `json:"payments"`
}

func (s *paymentOp) CreatePaymentSlip(ctx context.Context, slip PaymentSlip) (*PaymentSlip, error) {
	r := &paymentSlipResource{}
	err := s.client.Post(ctx, s.client.CreatePath("orders/transactions.json"), paymentSlipResource{Transaction: &slip}, r)
	return r.Transaction, err
}
func (s *paymentOp) UpdatePaymentSlip(ctx context.Context, slip PaymentSlip) (*PaymentSlip, error) {
	r := &paymentSlipResource{}
	err := s.client.Post(ctx, s.client.CreatePath("orders/payment_slip/update.json"), paymentSlipResource{Transaction: &slip}, r)
	return r.Transaction, err
}
func (s *paymentOp) GetSettings(ctx context.Context) (*PaymentSettings, error) {
	r := &paymentSettingsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("payment/settings.json"), r, nil)
	return r.Settings, err
}
func (s *paymentOp) ListChannels(ctx context.Context) ([]PaymentChannel, error) {
	r := &paymentChannelsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("payment/channels.json"), r, nil)
	return r.Channels, err
}
func (s *paymentOp) ListPayments(ctx context.Context, orderID int64) ([]OrderPayment, error) {
	r := &orderPaymentsResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("orders/%d/payments.json", orderID)), r, nil)
	return r.Payments, err
}

// === AbandonedCheckout ===

type AbandonedCheckoutService interface {
	List(ctx context.Context, opts *core.ListOptions) ([]AbandonedCheckout, error)
	Count(ctx context.Context) (int, error)
	Archive(ctx context.Context, ids []int64) error
}

func NewAbandonedCheckoutService(client core.Requester) AbandonedCheckoutService {
	return &checkoutOp{client: client}
}

type checkoutOp struct{ client core.Requester }

type AbandonedCheckout struct {
	ID                   int64               `json:"id,omitempty"`
	Token                string              `json:"token,omitempty"`
	Email                string              `json:"email,omitempty"`
	Phone                string              `json:"phone,omitempty"`
	Currency             string              `json:"currency,omitempty"`
	TotalPrice           string              `json:"total_price,omitempty"`
	SubtotalPrice        string              `json:"subtotal_price,omitempty"`
	TotalTax             string              `json:"total_tax,omitempty"`
	Customer             *core.Customer  `json:"customer,omitempty"`
	BillingAddress       *core.Address   `json:"billing_address,omitempty"`
	ShippingAddress      *core.Address   `json:"shipping_address,omitempty"`
	LineItems            []core.LineItem `json:"line_items,omitempty"`
	AbandonedCheckoutURL string              `json:"abandoned_checkout_url,omitempty"`
	CreatedAt            *time.Time          `json:"created_at,omitempty"`
	UpdatedAt            *time.Time          `json:"updated_at,omitempty"`
}

type checkoutsResource struct {
	Checkouts []AbandonedCheckout `json:"checkouts"`
}

func (s *checkoutOp) List(ctx context.Context, opts *core.ListOptions) ([]AbandonedCheckout, error) {
	r := &checkoutsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("checkouts.json"), r, opts)
	return r.Checkouts, err
}
func (s *checkoutOp) Count(ctx context.Context) (int, error) {
	r := &countResource{}
	err := s.client.Get(ctx, s.client.CreatePath("checkouts/count.json"), r, nil)
	return r.Count, err
}
func (s *checkoutOp) Archive(ctx context.Context, ids []int64) error {
	return s.client.Post(ctx, s.client.CreatePath("checkouts/checkouts_archive.json"), map[string][]int64{"ids": ids}, nil)
}

// === Subscription ===

type SubscriptionService interface {
	Get(ctx context.Context, id int64) (*SubscriptionContract, error)
	List(ctx context.Context, opts *core.ListOptions) ([]SubscriptionContract, error)
	Update(ctx context.Context, c SubscriptionContract) (*SubscriptionContract, error)
	Cancel(ctx context.Context, id int64) (*SubscriptionContract, error)
	ReviseNextBillTime(ctx context.Context, id int64, t time.Time) (*SubscriptionContract, error)
	SkipNextBill(ctx context.Context, id int64) (*SubscriptionContract, error)
	CreateOrder(ctx context.Context, id int64) (*Order, error)
}

func NewSubscriptionService(client core.Requester) SubscriptionService {
	return &subscriptionOp{client: client}
}

type subscriptionOp struct{ client core.Requester }

type SubscriptionContract struct {
	ID              int64                  `json:"id,omitempty"`
	Status          string                 `json:"status,omitempty"`
	CustomerID      int64                  `json:"customer_id,omitempty"`
	Customer        *core.Customer     `json:"customer,omitempty"`
	BillingPolicy   *SubscriptionPolicy    `json:"billing_policy,omitempty"`
	DeliveryPolicy  *SubscriptionPolicy    `json:"delivery_policy,omitempty"`
	NextBillingDate *time.Time             `json:"next_billing_date,omitempty"`
	Currency        string                 `json:"currency,omitempty"`
	LineItems       []SubscriptionLineItem `json:"line_items,omitempty"`
	ShippingAddress *core.Address      `json:"shipping_address,omitempty"`
	BillingAddress  *core.Address      `json:"billing_address,omitempty"`
	Note            string                 `json:"note,omitempty"`
	CreatedAt       *time.Time             `json:"created_at,omitempty"`
	UpdatedAt       *time.Time             `json:"updated_at,omitempty"`
	CancelledAt     *time.Time             `json:"cancelled_at,omitempty"`
}

type SubscriptionPolicy struct {
	Interval      string `json:"interval,omitempty"`
	IntervalCount int    `json:"interval_count,omitempty"`
	MaxCycles     int    `json:"max_cycles,omitempty"`
	MinCycles     int    `json:"min_cycles,omitempty"`
}

type SubscriptionLineItem struct {
	ID        int64  `json:"id,omitempty"`
	VariantID int64  `json:"variant_id,omitempty"`
	ProductID int64  `json:"product_id,omitempty"`
	Title     string `json:"title,omitempty"`
	Quantity  int    `json:"quantity,omitempty"`
	Price     string `json:"price,omitempty"`
}

type subscriptionResource struct {
	SubscriptionContract *SubscriptionContract `json:"subscription_contract"`
}
type subscriptionsResource struct {
	SubscriptionContracts []SubscriptionContract `json:"subscription_contracts"`
}

func (s *subscriptionOp) Get(ctx context.Context, id int64) (*SubscriptionContract, error) {
	r := &subscriptionResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("subscription_contracts/%d.json", id)), r, nil)
	return r.SubscriptionContract, err
}
func (s *subscriptionOp) List(ctx context.Context, opts *core.ListOptions) ([]SubscriptionContract, error) {
	r := &subscriptionsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("subscription_contracts.json"), r, opts)
	return r.SubscriptionContracts, err
}
func (s *subscriptionOp) Update(ctx context.Context, c SubscriptionContract) (*SubscriptionContract, error) {
	r := &subscriptionResource{}
	err := s.client.Put(ctx, s.client.CreatePath(fmt.Sprintf("subscription_contracts/%d.json", c.ID)), subscriptionResource{SubscriptionContract: &c}, r)
	return r.SubscriptionContract, err
}
func (s *subscriptionOp) Cancel(ctx context.Context, id int64) (*SubscriptionContract, error) {
	r := &subscriptionResource{}
	err := s.client.Post(ctx, s.client.CreatePath(fmt.Sprintf("subscription_contracts/%d/cancel.json", id)), nil, r)
	return r.SubscriptionContract, err
}
func (s *subscriptionOp) ReviseNextBillTime(ctx context.Context, id int64, t time.Time) (*SubscriptionContract, error) {
	r := &subscriptionResource{}
	err := s.client.Post(ctx, s.client.CreatePath(fmt.Sprintf("subscription_contracts/%d/revise_next_bill_time.json", id)), map[string]string{"next_billing_date": t.Format(time.RFC3339)}, r)
	return r.SubscriptionContract, err
}
func (s *subscriptionOp) SkipNextBill(ctx context.Context, id int64) (*SubscriptionContract, error) {
	r := &subscriptionResource{}
	err := s.client.Post(ctx, s.client.CreatePath(fmt.Sprintf("subscription_contracts/%d/skip_next_bill.json", id)), nil, r)
	return r.SubscriptionContract, err
}
func (s *subscriptionOp) CreateOrder(ctx context.Context, id int64) (*Order, error) {
	r := &orderResource{}
	err := s.client.Post(ctx, s.client.CreatePath(fmt.Sprintf("subscription_contracts/%d/create_order.json", id)), nil, r)
	return r.Order, err
}

// === Tax ===

type TaxService interface {
	ListCountries(ctx context.Context) ([]TaxCountry, error)
	GetCountry(ctx context.Context, id int64) (*TaxCountry, error)
	CountCountries(ctx context.Context) (int, error)
	ListProvinces(ctx context.Context, countryID int64) ([]TaxProvince, error)
	GetProvince(ctx context.Context, id int64) (*TaxProvince, error)
	CountProvinces(ctx context.Context, countryID int64) (int, error)
	ListTaxChannels(ctx context.Context) ([]TaxChannel, error)
	UpdateTaxChannel(ctx context.Context, c TaxChannel) (*TaxChannel, error)
	DeleteTaxChannel(ctx context.Context, id int64) error
}

func NewTaxService(client core.Requester) TaxService {
	return &taxOp{client: client}
}

type taxOp struct{ client core.Requester }

type TaxCountry struct {
	ID        int64         `json:"id,omitempty"`
	Name      string        `json:"name,omitempty"`
	Code      string        `json:"code,omitempty"`
	Tax       float64       `json:"tax,omitempty"`
	TaxName   string        `json:"tax_name,omitempty"`
	Provinces []TaxProvince `json:"provinces,omitempty"`
}

type TaxProvince struct {
	ID        int64   `json:"id,omitempty"`
	CountryID int64   `json:"country_id,omitempty"`
	Name      string  `json:"name,omitempty"`
	Code      string  `json:"code,omitempty"`
	Tax       float64 `json:"tax,omitempty"`
	TaxName   string  `json:"tax_name,omitempty"`
	TaxType   string  `json:"tax_type,omitempty"`
}

type TaxChannel struct {
	ID      int64  `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Enabled bool   `json:"enabled,omitempty"`
}

type taxCountryResource struct {
	Country *TaxCountry `json:"country"`
}
type taxCountriesResource struct {
	Countries []TaxCountry `json:"countries"`
}
type taxProvinceResource struct {
	Province *TaxProvince `json:"province"`
}
type taxProvincesResource struct {
	Provinces []TaxProvince `json:"provinces"`
}
type taxChannelResource struct {
	TaxChannel *TaxChannel `json:"tax_channel"`
}
type taxChannelsResource struct {
	TaxChannels []TaxChannel `json:"tax_channels"`
}

func (s *taxOp) ListCountries(ctx context.Context) ([]TaxCountry, error) {
	r := &taxCountriesResource{}
	err := s.client.Get(ctx, s.client.CreatePath("countries.json"), r, nil)
	return r.Countries, err
}
func (s *taxOp) GetCountry(ctx context.Context, id int64) (*TaxCountry, error) {
	r := &taxCountryResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("countries/%d.json", id)), r, nil)
	return r.Country, err
}
func (s *taxOp) CountCountries(ctx context.Context) (int, error) {
	r := &countResource{}
	err := s.client.Get(ctx, s.client.CreatePath("countries/count.json"), r, nil)
	return r.Count, err
}
func (s *taxOp) ListProvinces(ctx context.Context, countryID int64) ([]TaxProvince, error) {
	r := &taxProvincesResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("countries/%d/provinces.json", countryID)), r, nil)
	return r.Provinces, err
}
func (s *taxOp) GetProvince(ctx context.Context, id int64) (*TaxProvince, error) {
	r := &taxProvinceResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("provinces/%d.json", id)), r, nil)
	return r.Province, err
}
func (s *taxOp) CountProvinces(ctx context.Context, countryID int64) (int, error) {
	r := &countResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("countries/%d/provinces/count.json", countryID)), r, nil)
	return r.Count, err
}
func (s *taxOp) ListTaxChannels(ctx context.Context) ([]TaxChannel, error) {
	r := &taxChannelsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("tax_channels.json"), r, nil)
	return r.TaxChannels, err
}
func (s *taxOp) UpdateTaxChannel(ctx context.Context, c TaxChannel) (*TaxChannel, error) {
	r := &taxChannelResource{}
	err := s.client.Post(ctx, s.client.CreatePath("tax_channels.json"), taxChannelResource{TaxChannel: &c}, r)
	return r.TaxChannel, err
}
func (s *taxOp) DeleteTaxChannel(ctx context.Context, id int64) error {
	return s.client.Delete(ctx, s.client.CreatePath(fmt.Sprintf("tax_channels/%d.json", id)))
}

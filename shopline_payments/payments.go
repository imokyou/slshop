package shoplinepay

import (
	"context"
	"time"

	"github.com/imokyou/slshop/core"
)

// =====================================================================
// SHOPLINE Payments Service
// =====================================================================

type Service interface {
	GetBalance(ctx context.Context) (*Balance, error)
	ListPayouts(ctx context.Context, opts *PayoutListOptions) ([]Payout, error)
	ListBillingRecords(ctx context.Context, opts *BillingListOptions) ([]BillingRecord, error)
	CreatePayout(ctx context.Context, payout PayoutRequest) (*Payout, error)
	ListTransactions(ctx context.Context, opts *TransactionListOptions) ([]Transaction, error)
}

func NewService(client core.Requester) Service {
	return &serviceOp{client: client}
}

type serviceOp struct{ client core.Requester }

// =====================================================================
// Models
// =====================================================================

type Balance struct {
	Currency  string `json:"currency,omitempty"`
	Amount    string `json:"amount,omitempty"`
	Available string `json:"available,omitempty"`
	Pending   string `json:"pending,omitempty"`
}

type Payout struct {
	ID         int64      `json:"id,omitempty"`
	Amount     string     `json:"amount,omitempty"`
	Currency   string     `json:"currency,omitempty"`
	Status     string     `json:"status,omitempty"`
	PayoutDate string     `json:"payout_date,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
}

type PayoutRequest struct {
	Amount   string `json:"amount,omitempty"`
	Currency string `json:"currency,omitempty"`
}

type BillingRecord struct {
	ID         int64      `json:"id,omitempty"`
	Type       string     `json:"type,omitempty"`
	Amount     string     `json:"amount,omitempty"`
	Fee        string     `json:"fee,omitempty"`
	Net        string     `json:"net,omitempty"`
	Currency   string     `json:"currency,omitempty"`
	OrderID    int64      `json:"order_id,omitempty"`
	SourceType string     `json:"source_type,omitempty"`
	SourceID   int64      `json:"source_id,omitempty"`
	Status     string     `json:"status,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
}

type Transaction struct {
	ID        int64      `json:"id,omitempty"`
	Type      string     `json:"type,omitempty"`
	Amount    string     `json:"amount,omitempty"`
	Currency  string     `json:"currency,omitempty"`
	OrderID   int64      `json:"order_id,omitempty"`
	Status    string     `json:"status,omitempty"`
	Gateway   string     `json:"gateway,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

type PayoutListOptions struct {
	core.ListOptions
	Status string `url:"status,omitempty"`
}

type BillingListOptions struct {
	core.ListOptions
	Type string `url:"type,omitempty"`
}

type TransactionListOptions struct {
	core.ListOptions
	OrderID int64  `url:"order_id,omitempty"`
	Type    string `url:"type,omitempty"`
}

// JSON wrappers
type balanceResource struct {
	Balance *Balance `json:"balance"`
}
type payoutsResource struct {
	Payouts []Payout `json:"payouts"`
}
type payoutResource struct {
	Payout *Payout `json:"payout"`
}
type billingRecordsResource struct {
	BillingRecords []BillingRecord `json:"billing_records"`
}
type transactionsResource struct {
	Transactions []Transaction `json:"transactions"`
}

// =====================================================================
// Implementation
// =====================================================================

// GET payments/store/balance.json
func (s *serviceOp) GetBalance(ctx context.Context) (*Balance, error) {
	r := &balanceResource{}
	err := s.client.Get(ctx, s.client.CreatePath("payments/store/balance.json"), r, nil)
	return r.Balance, err
}

// GET payments/store/payouts.json
func (s *serviceOp) ListPayouts(ctx context.Context, opts *PayoutListOptions) ([]Payout, error) {
	r := &payoutsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("payments/store/payouts.json"), r, opts)
	return r.Payouts, err
}

// GET payments/store/billing_records.json
func (s *serviceOp) ListBillingRecords(ctx context.Context, opts *BillingListOptions) ([]BillingRecord, error) {
	r := &billingRecordsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("payments/store/billing_records.json"), r, opts)
	return r.BillingRecords, err
}

// POST payments/store/payouts.json
func (s *serviceOp) CreatePayout(ctx context.Context, payout PayoutRequest) (*Payout, error) {
	r := &payoutResource{}
	err := s.client.Post(ctx, s.client.CreatePath("payments/store/payouts.json"), payoutResource{Payout: &Payout{Amount: payout.Amount, Currency: payout.Currency}}, r)
	return r.Payout, err
}

// GET payments/store/transactions.json
func (s *serviceOp) ListTransactions(ctx context.Context, opts *TransactionListOptions) ([]Transaction, error) {
	r := &transactionsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("payments/store/transactions.json"), r, opts)
	return r.Transactions, err
}

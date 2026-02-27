package marketing

import (
	"context"
	"fmt"
	"time"

	"github.com/imokyou/slshop/core"
)

// =====================================================================
// Discount (PriceRule + DiscountCode)
// =====================================================================

type DiscountService interface {
	ListPriceRules(ctx context.Context, opts *core.ListOptions) ([]PriceRule, error)
	GetPriceRule(ctx context.Context, id int64) (*PriceRule, error)
	CreatePriceRule(ctx context.Context, r PriceRule) (*PriceRule, error)
	UpdatePriceRule(ctx context.Context, r PriceRule) (*PriceRule, error)
	DeletePriceRule(ctx context.Context, id int64) error

	ListDiscountCodes(ctx context.Context, priceRuleID int64) ([]DiscountCode, error)
	GetDiscountCode(ctx context.Context, priceRuleID, codeID int64) (*DiscountCode, error)
	CreateDiscountCode(ctx context.Context, priceRuleID int64, c DiscountCode) (*DiscountCode, error)
	UpdateDiscountCode(ctx context.Context, priceRuleID int64, c DiscountCode) (*DiscountCode, error)
	DeleteDiscountCode(ctx context.Context, priceRuleID, codeID int64) error
}

func NewDiscountService(client core.Requester) DiscountService {
	return &discountOp{client: client}
}

type discountOp struct{ client core.Requester }

type PriceRule struct {
	ID                int64      `json:"id,omitempty"`
	Title             string     `json:"title,omitempty"`
	TargetType        string     `json:"target_type,omitempty"`
	TargetSelection   string     `json:"target_selection,omitempty"`
	AllocationMethod  string     `json:"allocation_method,omitempty"`
	ValueType         string     `json:"value_type,omitempty"`
	Value             string     `json:"value,omitempty"`
	OncePerCustomer   bool       `json:"once_per_customer,omitempty"`
	UsageLimit        int        `json:"usage_limit,omitempty"`
	CustomerSelection string     `json:"customer_selection,omitempty"`
	StartsAt          *time.Time `json:"starts_at,omitempty"`
	EndsAt            *time.Time `json:"ends_at,omitempty"`
	CreatedAt         *time.Time `json:"created_at,omitempty"`
	UpdatedAt         *time.Time `json:"updated_at,omitempty"`
}

type DiscountCode struct {
	ID          int64      `json:"id,omitempty"`
	PriceRuleID int64      `json:"price_rule_id,omitempty"`
	Code        string     `json:"code,omitempty"`
	UsageCount  int        `json:"usage_count,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
}

type priceRuleResource struct {
	PriceRule *PriceRule `json:"price_rule"`
}
type priceRulesResource struct {
	PriceRules []PriceRule `json:"price_rules"`
}
type discountCodeResource struct {
	DiscountCode *DiscountCode `json:"discount_code"`
}
type discountCodesResource struct {
	DiscountCodes []DiscountCode `json:"discount_codes"`
}

func (s *discountOp) ListPriceRules(ctx context.Context, opts *core.ListOptions) ([]PriceRule, error) {
	r := &priceRulesResource{}
	err := s.client.Get(ctx, s.client.CreatePath("price_rules.json"), r, opts)
	return r.PriceRules, err
}
func (s *discountOp) GetPriceRule(ctx context.Context, id int64) (*PriceRule, error) {
	r := &priceRuleResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("price_rules/%d.json", id)), r, nil)
	return r.PriceRule, err
}
func (s *discountOp) CreatePriceRule(ctx context.Context, rule PriceRule) (*PriceRule, error) {
	r := &priceRuleResource{}
	err := s.client.Post(ctx, s.client.CreatePath("price_rules.json"), priceRuleResource{PriceRule: &rule}, r)
	return r.PriceRule, err
}
func (s *discountOp) UpdatePriceRule(ctx context.Context, rule PriceRule) (*PriceRule, error) {
	r := &priceRuleResource{}
	err := s.client.Put(ctx, s.client.CreatePath(fmt.Sprintf("price_rules/%d.json", rule.ID)), priceRuleResource{PriceRule: &rule}, r)
	return r.PriceRule, err
}
func (s *discountOp) DeletePriceRule(ctx context.Context, id int64) error {
	return s.client.Delete(ctx, s.client.CreatePath(fmt.Sprintf("price_rules/%d.json", id)))
}
func (s *discountOp) ListDiscountCodes(ctx context.Context, priceRuleID int64) ([]DiscountCode, error) {
	r := &discountCodesResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("price_rules/%d/discount_codes.json", priceRuleID)), r, nil)
	return r.DiscountCodes, err
}
func (s *discountOp) GetDiscountCode(ctx context.Context, priceRuleID, codeID int64) (*DiscountCode, error) {
	r := &discountCodeResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("price_rules/%d/discount_codes/%d.json", priceRuleID, codeID)), r, nil)
	return r.DiscountCode, err
}
func (s *discountOp) CreateDiscountCode(ctx context.Context, priceRuleID int64, c DiscountCode) (*DiscountCode, error) {
	r := &discountCodeResource{}
	err := s.client.Post(ctx, s.client.CreatePath(fmt.Sprintf("price_rules/%d/discount_codes.json", priceRuleID)), discountCodeResource{DiscountCode: &c}, r)
	return r.DiscountCode, err
}
func (s *discountOp) UpdateDiscountCode(ctx context.Context, priceRuleID int64, c DiscountCode) (*DiscountCode, error) {
	r := &discountCodeResource{}
	err := s.client.Put(ctx, s.client.CreatePath(fmt.Sprintf("price_rules/%d/discount_codes/%d.json", priceRuleID, c.ID)), discountCodeResource{DiscountCode: &c}, r)
	return r.DiscountCode, err
}
func (s *discountOp) DeleteDiscountCode(ctx context.Context, priceRuleID, codeID int64) error {
	return s.client.Delete(ctx, s.client.CreatePath(fmt.Sprintf("price_rules/%d/discount_codes/%d.json", priceRuleID, codeID)))
}

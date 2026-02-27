package store

import (
	"context"
	"fmt"
	"time"

	"github.com/imokyou/slshop/core"
)

// =====================================================================
// Store Service
// =====================================================================

type Service interface {
	GetInfo(ctx context.Context) (*Info, error)
	GetSettlementCurrency(ctx context.Context) ([]Currency, error)
	GetStaffMember(ctx context.Context, uid string) (*StaffMember, error)
	ListStaffMembers(ctx context.Context) ([]StaffMember, error)
	ListOperationLogs(ctx context.Context, opts *core.ListOptions) ([]OperationLog, error)
	GetOperationLog(ctx context.Context, id int64) (*OperationLog, error)
	CountOperationLogs(ctx context.Context) (int, error)
	GetActiveSubscription(ctx context.Context) (*Subscription, error)

	// Shop (legacy)
	GetShop(ctx context.Context) (*Shop, error)
}

func NewService(client core.Requester) Service {
	return &serviceOp{client: client}
}

type serviceOp struct{ client core.Requester }

// =====================================================================
// Models
// =====================================================================

type Info struct {
	ID                  int64          `json:"id,omitempty"`
	Name                string         `json:"name,omitempty"`
	Email               string         `json:"email,omitempty"`
	CustomerEmail       string         `json:"customer_email,omitempty"`
	Domain              string         `json:"domain,omitempty"`
	Currency            string         `json:"currency,omitempty"`
	Language            string         `json:"language,omitempty"`
	IanaTimezone        string         `json:"iana_timezone,omitempty"`
	MerchantID          string         `json:"merchant_id,omitempty"`
	LocationCountryCode string         `json:"location_country_code,omitempty"`
	StandardLogo        string         `json:"standard_logo,omitempty"`
	BizStoreStatus      int            `json:"biz_store_status,omitempty"`
	SalesChannels       []SalesChannel `json:"sales_channels,omitempty"`
	CreatedAt           string         `json:"created_at,omitempty"`
	UpdatedAt           string         `json:"updated_at,omitempty"`
}

type SalesChannel struct {
	ChannelHandle string `json:"channel_handle,omitempty"`
}

type Currency struct {
	Code          string `json:"code,omitempty"`
	Name          string `json:"name,omitempty"`
	Symbol        string `json:"symbol,omitempty"`
	Primary       bool   `json:"primary,omitempty"`
	Enabled       bool   `json:"enabled,omitempty"`
	RateToDefault string `json:"rate_to_default,omitempty"`
}

type StaffMember struct {
	UID          string     `json:"uid,omitempty"`
	Email        string     `json:"email,omitempty"`
	FirstName    string     `json:"first_name,omitempty"`
	LastName     string     `json:"last_name,omitempty"`
	Phone        string     `json:"phone,omitempty"`
	Locale       string     `json:"locale,omitempty"`
	AccountOwner bool       `json:"account_owner,omitempty"`
	Permissions  []string   `json:"permissions,omitempty"`
	Avatar       string     `json:"avatar,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

type OperationLog struct {
	ID          int64      `json:"id,omitempty"`
	Action      string     `json:"action,omitempty"`
	Subject     string     `json:"subject,omitempty"`
	SubjectType string     `json:"subject_type,omitempty"`
	SubjectID   int64      `json:"subject_id,omitempty"`
	Author      string     `json:"author,omitempty"`
	AuthorID    string     `json:"author_id,omitempty"`
	Body        string     `json:"body,omitempty"`
	Message     string     `json:"message,omitempty"`
	Path        string     `json:"path,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
}

type Subscription struct {
	ID              int64      `json:"id,omitempty"`
	PlanName        string     `json:"plan_name,omitempty"`
	PlanDisplayName string     `json:"plan_display_name,omitempty"`
	Status          string     `json:"status,omitempty"`
	TrialDays       int        `json:"trial_days,omitempty"`
	TrialEndsAt     *time.Time `json:"trial_ends_at,omitempty"`
	ActivatedAt     *time.Time `json:"activated_at,omitempty"`
	BillingOn       *time.Time `json:"billing_on,omitempty"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty"`
}

type Shop struct {
	ID                      int64      `json:"id,omitempty"`
	Name                    string     `json:"name,omitempty"`
	Email                   string     `json:"email,omitempty"`
	Domain                  string     `json:"domain,omitempty"`
	MyshoplineDomain        string     `json:"myshopline_domain,omitempty"`
	Phone                   string     `json:"phone,omitempty"`
	Address1                string     `json:"address1,omitempty"`
	Address2                string     `json:"address2,omitempty"`
	City                    string     `json:"city,omitempty"`
	Province                string     `json:"province,omitempty"`
	ProvinceCode            string     `json:"province_code,omitempty"`
	Country                 string     `json:"country,omitempty"`
	CountryCode             string     `json:"country_code,omitempty"`
	Zip                     string     `json:"zip,omitempty"`
	Currency                string     `json:"currency,omitempty"`
	MoneyFormat             string     `json:"money_format,omitempty"`
	MoneyWithCurrencyFormat string     `json:"money_with_currency_format,omitempty"`
	Timezone                string     `json:"timezone,omitempty"`
	IanaTimezone            string     `json:"iana_timezone,omitempty"`
	WeightUnit              string     `json:"weight_unit,omitempty"`
	PlanName                string     `json:"plan_name,omitempty"`
	PlanDisplayName         string     `json:"plan_display_name,omitempty"`
	CreatedAt               *time.Time `json:"created_at,omitempty"`
	UpdatedAt               *time.Time `json:"updated_at,omitempty"`
}

// JSON wrappers
type infoResource struct {
	Data *Info `json:"data"`
}
type currenciesResource struct {
	Currencies []Currency `json:"currencies"`
}
type staffResource struct {
	Staff *StaffMember `json:"staff"`
}
type staffListResource struct {
	Staff []StaffMember `json:"staff"`
}
type opLogResource struct {
	OperationLog *OperationLog `json:"operation_log"`
}
type opLogsResource struct {
	OperationLogs []OperationLog `json:"operation_logs"`
}
type countResource struct {
	Count int `json:"count"`
}
type subscriptionResource struct {
	Subscription *Subscription `json:"subscription"`
}
type shopResource struct {
	Shop *Shop `json:"shop"`
}

// =====================================================================
// Implementation
// =====================================================================

func (s *serviceOp) GetInfo(ctx context.Context) (*Info, error) {
	r := &infoResource{}
	err := s.client.Get(ctx, s.client.CreatePath("merchants/shop.json"), r, nil)
	return r.Data, err
}
func (s *serviceOp) GetSettlementCurrency(ctx context.Context) ([]Currency, error) {
	r := &currenciesResource{}
	err := s.client.Get(ctx, s.client.CreatePath("currency/currencies.json"), r, nil)
	return r.Currencies, err
}
func (s *serviceOp) GetStaffMember(ctx context.Context, uid string) (*StaffMember, error) {
	r := &staffResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("store/staff/%s.json", uid)), r, nil)
	return r.Staff, err
}
func (s *serviceOp) ListStaffMembers(ctx context.Context) ([]StaffMember, error) {
	r := &staffListResource{}
	err := s.client.Get(ctx, s.client.CreatePath("store/list/staff.json"), r, nil)
	return r.Staff, err
}
func (s *serviceOp) ListOperationLogs(ctx context.Context, opts *core.ListOptions) ([]OperationLog, error) {
	r := &opLogsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("store/operation_logs.json"), r, opts)
	return r.OperationLogs, err
}
func (s *serviceOp) GetOperationLog(ctx context.Context, id int64) (*OperationLog, error) {
	r := &opLogResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("store/operation_logs/%d.json", id)), r, nil)
	return r.OperationLog, err
}
func (s *serviceOp) CountOperationLogs(ctx context.Context) (int, error) {
	r := &countResource{}
	err := s.client.Get(ctx, s.client.CreatePath("store/operation_logs/count.json"), r, nil)
	return r.Count, err
}
func (s *serviceOp) GetActiveSubscription(ctx context.Context) (*Subscription, error) {
	r := &subscriptionResource{}
	err := s.client.Get(ctx, s.client.CreatePath("store/subscription"), r, nil)
	return r.Subscription, err
}
func (s *serviceOp) GetShop(ctx context.Context) (*Shop, error) {
	r := &shopResource{}
	err := s.client.Get(ctx, s.client.CreatePath("shop.json"), r, nil)
	return r.Shop, err
}

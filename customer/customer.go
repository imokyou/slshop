package customer

import (
	"context"
	"fmt"
	"time"

	"github.com/imokyou/slshop/core"
)

const basePath = "v2/customers"

// =====================================================================
// Service Interface
// =====================================================================

type Service interface {
	List(ctx context.Context, opts *ListOptions) ([]core.Customer, error)
	Get(ctx context.Context, id int64) (*core.Customer, error)
	Count(ctx context.Context, opts *core.CountOptions) (int, error)
	Search(ctx context.Context, query string, opts *core.ListOptions) ([]core.Customer, error)
	Create(ctx context.Context, c core.Customer) (*core.Customer, error)
	Update(ctx context.Context, c core.Customer) (*core.Customer, error)
	Delete(ctx context.Context, id int64) error

	SendInvite(ctx context.Context, id int64) error
	ActivationURL(ctx context.Context, id int64) (string, error)
	CheckEmail(ctx context.Context, email string) (*core.Customer, error)
	ListOrders(ctx context.Context, id int64, opts *core.ListOptions) ([]Order, error)
	BatchMarketingStates(ctx context.Context, opts *MarketingOptions) ([]MarketingState, error)

	DeleteTag(ctx context.Context, customerID int64, tag string) error
	AddToBlacklist(ctx context.Context, id int64) error
	RemoveFromBlacklist(ctx context.Context, id int64) error

	ListGroups(ctx context.Context, opts *core.ListOptions) ([]Group, error)
	GetGroup(ctx context.Context, groupID int64) (*Group, error)
	CreateGroup(ctx context.Context, g Group) (*Group, error)
	UpdateGroup(ctx context.Context, g Group) (*Group, error)
	DeleteGroup(ctx context.Context, groupID int64) error
	ListGroupCustomers(ctx context.Context, groupID int64, opts *core.ListOptions) ([]core.Customer, error)
	ListStoreGroups(ctx context.Context) ([]Group, error)

	CreateAddress(ctx context.Context, customerID int64, addr core.Address) (*core.Address, error)
	UpdateAddress(ctx context.Context, customerID int64, addr core.Address) (*core.Address, error)
	DeleteAddress(ctx context.Context, customerID, addressID int64) error
	GetAddress(ctx context.Context, customerID, addressID int64) (*core.Address, error)
	SetDefaultAddress(ctx context.Context, customerID, addressID int64) (*core.Address, error)
	BatchSetAddress(ctx context.Context, customerID int64, addrs []core.Address) ([]core.Address, error)
	BatchQueryAddress(ctx context.Context, customerIDs []int64) ([]AddressResult, error)

	ListSocialLogin(ctx context.Context) ([]SocialLoginConfig, error)
	UpdateSocialLogin(ctx context.Context, cfg SocialLoginConfig) (*SocialLoginConfig, error)
	DeleteSocialLogin(ctx context.Context) error
}

func NewService(client core.Requester) Service {
	return &serviceOp{client: client}
}

type serviceOp struct{ client core.Requester }

// =====================================================================
// Models
// =====================================================================

type ListOptions struct {
	core.ListOptions
	IDs          string `url:"ids,omitempty"`
	SinceID      int64  `url:"since_id,omitempty"`
	CreatedAtMin string `url:"created_at_min,omitempty"`
	CreatedAtMax string `url:"created_at_max,omitempty"`
	UpdatedAtMin string `url:"updated_at_min,omitempty"`
	UpdatedAtMax string `url:"updated_at_max,omitempty"`
}

type MarketingOptions struct {
	core.ListOptions
	CustomerIDs string `url:"customer_ids,omitempty"`
}

type MarketingState struct {
	CustomerID       int64  `json:"customer_id,omitempty"`
	Email            string `json:"email,omitempty"`
	AcceptsMarketing bool   `json:"accepts_marketing,omitempty"`
}

type Group struct {
	ID        int64      `json:"id,omitempty"`
	Name      string     `json:"name,omitempty"`
	Query     string     `json:"query,omitempty"`
	SortOrder string     `json:"sort_order,omitempty"`
	Count     int        `json:"count,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type AddressResult struct {
	CustomerID int64              `json:"customer_id,omitempty"`
	Addresses  []core.Address `json:"addresses,omitempty"`
}

type SocialLoginConfig struct {
	ID       int64  `json:"id,omitempty"`
	Provider string `json:"provider,omitempty"`
	AppID    string `json:"app_id,omitempty"`
	AppKey   string `json:"app_key,omitempty"`
	Secret   string `json:"secret,omitempty"`
	Enabled  bool   `json:"enabled,omitempty"`
}

// Order is a minimal order representation for customer order listing.
type Order struct {
	ID              int64      `json:"id,omitempty"`
	Name            string     `json:"name,omitempty"`
	TotalPrice      string     `json:"total_price,omitempty"`
	Currency        string     `json:"currency,omitempty"`
	FinancialStatus string     `json:"financial_status,omitempty"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
}

// =====================================================================
// JSON Wrappers
// =====================================================================

type customerResource struct {
	Customer *core.Customer `json:"customer"`
}
type customersResource struct {
	Customers []core.Customer `json:"customers"`
}
type countResource struct {
	Count int `json:"count"`
}
type activationURLResource struct {
	ActivationURL string `json:"activation_url"`
}
type marketingStatesResource struct {
	MarketingStates []MarketingState `json:"marketing_states"`
}
type groupResource struct {
	CustomerGroup *Group `json:"customer_group"`
}
type groupsResource struct {
	CustomerGroups []Group `json:"customer_groups"`
}
type addressResource struct {
	Address *core.Address `json:"address"`
}
type addressesResource struct {
	Addresses []core.Address `json:"addresses"`
}
type addressResultsResource struct {
	Results []AddressResult `json:"results"`
}
type socialLoginResource struct {
	SocialLoginConfig *SocialLoginConfig `json:"social_login"`
}
type socialLoginsResource struct {
	SocialLoginConfigs []SocialLoginConfig `json:"social_logins"`
}
type ordersResource struct {
	Orders []Order `json:"orders"`
}

// =====================================================================
// Customer CRUD
// =====================================================================

func (s *serviceOp) List(ctx context.Context, opts *ListOptions) ([]core.Customer, error) {
	r := &customersResource{}
	err := s.client.Get(ctx, s.client.CreatePath(basePath+".json"), r, opts)
	return r.Customers, err
}
func (s *serviceOp) Get(ctx context.Context, id int64) (*core.Customer, error) {
	r := &customerResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("%s/%d.json", basePath, id)), r, nil)
	return r.Customer, err
}
func (s *serviceOp) Count(ctx context.Context, opts *core.CountOptions) (int, error) {
	r := &countResource{}
	err := s.client.Get(ctx, s.client.CreatePath(basePath+"/count.json"), r, opts)
	return r.Count, err
}
func (s *serviceOp) Search(ctx context.Context, query string, opts *core.ListOptions) ([]core.Customer, error) {
	so := struct {
		*core.ListOptions
		Query string `url:"query,omitempty"`
	}{ListOptions: opts, Query: query}
	r := &customersResource{}
	err := s.client.Get(ctx, s.client.CreatePath(basePath+"/search.json"), r, &so)
	return r.Customers, err
}
func (s *serviceOp) Create(ctx context.Context, c core.Customer) (*core.Customer, error) {
	r := &customerResource{}
	err := s.client.Post(ctx, s.client.CreatePath(basePath+".json"), customerResource{Customer: &c}, r)
	return r.Customer, err
}
func (s *serviceOp) Update(ctx context.Context, c core.Customer) (*core.Customer, error) {
	r := &customerResource{}
	err := s.client.Put(ctx, s.client.CreatePath(fmt.Sprintf("%s/%d.json", basePath, c.ID)), customerResource{Customer: &c}, r)
	return r.Customer, err
}
func (s *serviceOp) Delete(ctx context.Context, id int64) error {
	return s.client.Delete(ctx, s.client.CreatePath(fmt.Sprintf("%s/%d.json", basePath, id)))
}

// =====================================================================
// Customer Actions
// =====================================================================

func (s *serviceOp) SendInvite(ctx context.Context, id int64) error {
	return s.client.Post(ctx, s.client.CreatePath(fmt.Sprintf("%s/%d/send_invite.json", basePath, id)), nil, nil)
}
func (s *serviceOp) ActivationURL(ctx context.Context, id int64) (string, error) {
	r := &activationURLResource{}
	err := s.client.Post(ctx, s.client.CreatePath(fmt.Sprintf("%s/%d/activation_url.json", basePath, id)), nil, r)
	return r.ActivationURL, err
}
func (s *serviceOp) CheckEmail(ctx context.Context, email string) (*core.Customer, error) {
	r := &customerResource{}
	err := s.client.Post(ctx, s.client.CreatePath(basePath+"/check_email.json"), map[string]string{"email": email}, r)
	return r.Customer, err
}
func (s *serviceOp) ListOrders(ctx context.Context, id int64, opts *core.ListOptions) ([]Order, error) {
	r := &ordersResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("%s/%d/orders.json", basePath, id)), r, opts)
	return r.Orders, err
}
func (s *serviceOp) BatchMarketingStates(ctx context.Context, opts *MarketingOptions) ([]MarketingState, error) {
	r := &marketingStatesResource{}
	err := s.client.Get(ctx, s.client.CreatePath(basePath+"/marketing_states.json"), r, opts)
	return r.MarketingStates, err
}
func (s *serviceOp) DeleteTag(ctx context.Context, customerID int64, tag string) error {
	return s.client.Post(ctx, s.client.CreatePath(fmt.Sprintf("%s/%d/tags/%s.json", basePath, customerID, tag)), nil, nil)
}
func (s *serviceOp) AddToBlacklist(ctx context.Context, id int64) error {
	return s.client.Post(ctx, s.client.CreatePath(fmt.Sprintf("%s/%d/blacklist.json", basePath, id)), nil, nil)
}
func (s *serviceOp) RemoveFromBlacklist(ctx context.Context, id int64) error {
	return s.client.Post(ctx, s.client.CreatePath(fmt.Sprintf("%s/%d/unblacklist.json", basePath, id)), nil, nil)
}

// =====================================================================
// Customer Groups
// =====================================================================

func (s *serviceOp) ListGroups(ctx context.Context, opts *core.ListOptions) ([]Group, error) {
	r := &groupsResource{}
	err := s.client.Get(ctx, s.client.CreatePath(basePath+"/groups.json"), r, opts)
	return r.CustomerGroups, err
}
func (s *serviceOp) GetGroup(ctx context.Context, groupID int64) (*Group, error) {
	r := &groupResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("%s/groups/%d.json", basePath, groupID)), r, nil)
	return r.CustomerGroup, err
}
func (s *serviceOp) CreateGroup(ctx context.Context, g Group) (*Group, error) {
	r := &groupResource{}
	err := s.client.Post(ctx, s.client.CreatePath(basePath+"/groups.json"), groupResource{CustomerGroup: &g}, r)
	return r.CustomerGroup, err
}
func (s *serviceOp) UpdateGroup(ctx context.Context, g Group) (*Group, error) {
	r := &groupResource{}
	err := s.client.Put(ctx, s.client.CreatePath(fmt.Sprintf("%s/groups/%d.json", basePath, g.ID)), groupResource{CustomerGroup: &g}, r)
	return r.CustomerGroup, err
}
func (s *serviceOp) DeleteGroup(ctx context.Context, groupID int64) error {
	return s.client.Delete(ctx, s.client.CreatePath(fmt.Sprintf("%s/groups/%d.json", basePath, groupID)))
}
func (s *serviceOp) ListGroupCustomers(ctx context.Context, groupID int64, opts *core.ListOptions) ([]core.Customer, error) {
	r := &customersResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("%s/groups/%d/customers.json", basePath, groupID)), r, opts)
	return r.Customers, err
}
func (s *serviceOp) ListStoreGroups(ctx context.Context) ([]Group, error) {
	r := &groupsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("v2/customer_groups.json"), r, nil)
	return r.CustomerGroups, err
}

// =====================================================================
// Customer Address
// =====================================================================

func (s *serviceOp) CreateAddress(ctx context.Context, customerID int64, addr core.Address) (*core.Address, error) {
	r := &addressResource{}
	err := s.client.Post(ctx, s.client.CreatePath(fmt.Sprintf("%s/%d/addresses.json", basePath, customerID)), addressResource{Address: &addr}, r)
	return r.Address, err
}
func (s *serviceOp) UpdateAddress(ctx context.Context, customerID int64, addr core.Address) (*core.Address, error) {
	r := &addressResource{}
	err := s.client.Put(ctx, s.client.CreatePath(fmt.Sprintf("%s/%d/addresses/%d.json", basePath, customerID, addr.ID)), addressResource{Address: &addr}, r)
	return r.Address, err
}
func (s *serviceOp) DeleteAddress(ctx context.Context, customerID, addressID int64) error {
	return s.client.Delete(ctx, s.client.CreatePath(fmt.Sprintf("%s/%d/addresses/%d.json", basePath, customerID, addressID)))
}
func (s *serviceOp) GetAddress(ctx context.Context, customerID, addressID int64) (*core.Address, error) {
	r := &addressResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("%s/%d/addresses/%d.json", basePath, customerID, addressID)), r, nil)
	return r.Address, err
}
func (s *serviceOp) SetDefaultAddress(ctx context.Context, customerID, addressID int64) (*core.Address, error) {
	r := &addressResource{}
	err := s.client.Put(ctx, s.client.CreatePath(fmt.Sprintf("%s/%d/addresses/%d/default.json", basePath, customerID, addressID)), nil, r)
	return r.Address, err
}
func (s *serviceOp) BatchSetAddress(ctx context.Context, customerID int64, addrs []core.Address) ([]core.Address, error) {
	r := &addressesResource{}
	err := s.client.Put(ctx, s.client.CreatePath(fmt.Sprintf("%s/%d/addresses/set.json", basePath, customerID)), addressesResource{Addresses: addrs}, r)
	return r.Addresses, err
}
func (s *serviceOp) BatchQueryAddress(ctx context.Context, customerIDs []int64) ([]AddressResult, error) {
	r := &addressResultsResource{}
	err := s.client.Post(ctx, s.client.CreatePath(basePath+"/addresses/list.json"), map[string][]int64{"customer_ids": customerIDs}, r)
	return r.Results, err
}

// =====================================================================
// Third-party Login
// =====================================================================

func (s *serviceOp) ListSocialLogin(ctx context.Context) ([]SocialLoginConfig, error) {
	r := &socialLoginsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("customers_social_login.json"), r, nil)
	return r.SocialLoginConfigs, err
}
func (s *serviceOp) UpdateSocialLogin(ctx context.Context, cfg SocialLoginConfig) (*SocialLoginConfig, error) {
	r := &socialLoginResource{}
	err := s.client.Post(ctx, s.client.CreatePath("customers_social_login.json"), socialLoginResource{SocialLoginConfig: &cfg}, r)
	return r.SocialLoginConfig, err
}
func (s *serviceOp) DeleteSocialLogin(ctx context.Context) error {
	return s.client.Delete(ctx, s.client.CreatePath("customers_social_login.json"))
}

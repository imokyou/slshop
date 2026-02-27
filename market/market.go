package market

import (
	"context"
	"fmt"
	"time"

	"github.com/imokyou/slshop/core"
)

// =====================================================================
// Market
// =====================================================================

type MarketService interface {
	List(ctx context.Context, opts *core.ListOptions) ([]Market, error)
	Get(ctx context.Context, id int64) (*Market, error)
}

func NewMarketService(client core.Requester) MarketService {
	return &marketOp{client: client}
}

type marketOp struct{ client core.Requester }

type Market struct {
	ID      int64  `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Handle  string `json:"handle,omitempty"`
	Enabled bool   `json:"enabled,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

type marketResource struct {
	Market *Market `json:"market"`
}
type marketsResource struct {
	Markets []Market `json:"markets"`
}

func (s *marketOp) List(ctx context.Context, opts *core.ListOptions) ([]Market, error) {
	r := &marketsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("markets.json"), r, opts)
	return r.Markets, err
}
func (s *marketOp) Get(ctx context.Context, id int64) (*Market, error) {
	r := &marketResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("markets/%d.json", id)), r, nil)
	return r.Market, err
}

// =====================================================================
// Location
// =====================================================================

type LocationService interface {
	List(ctx context.Context) ([]Location, error)
	Get(ctx context.Context, id int64) (*Location, error)
}

func NewLocationService(client core.Requester) LocationService {
	return &locationOp{client: client}
}

type locationOp struct{ client core.Requester }

type Location struct {
	ID           int64      `json:"id,omitempty"`
	Name         string     `json:"name,omitempty"`
	Address1     string     `json:"address1,omitempty"`
	Address2     string     `json:"address2,omitempty"`
	City         string     `json:"city,omitempty"`
	Province     string     `json:"province,omitempty"`
	ProvinceCode string     `json:"province_code,omitempty"`
	Country      string     `json:"country,omitempty"`
	CountryCode  string     `json:"country_code,omitempty"`
	Zip          string     `json:"zip,omitempty"`
	Phone        string     `json:"phone,omitempty"`
	Active       bool       `json:"active,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

type locationResource struct {
	Location *Location `json:"location"`
}
type locationsResource struct {
	Locations []Location `json:"locations"`
}

func (s *locationOp) List(ctx context.Context) ([]Location, error) {
	r := &locationsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("locations.json"), r, nil)
	return r.Locations, err
}
func (s *locationOp) Get(ctx context.Context, id int64) (*Location, error) {
	r := &locationResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("locations/%d.json", id)), r, nil)
	return r.Location, err
}

// =====================================================================
// Publication
// =====================================================================

type PublicationService interface {
	List(ctx context.Context, opts *core.ListOptions) ([]Publication, error)
}

func NewPublicationService(client core.Requester) PublicationService {
	return &publicationOp{client: client}
}

type publicationOp struct{ client core.Requester }

type Publication struct {
	ID        int64      `json:"id,omitempty"`
	Name      string     `json:"name,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

type publicationsResource struct {
	Publications []Publication `json:"publications"`
}

func (s *publicationOp) List(ctx context.Context, opts *core.ListOptions) ([]Publication, error) {
	r := &publicationsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("publications.json"), r, opts)
	return r.Publications, err
}

// =====================================================================
// GiftCard
// =====================================================================

type GiftCardService interface {
	List(ctx context.Context, opts *core.ListOptions) ([]GiftCard, error)
	Get(ctx context.Context, id int64) (*GiftCard, error)
	Create(ctx context.Context, c GiftCard) (*GiftCard, error)
}

func NewGiftCardService(client core.Requester) GiftCardService {
	return &giftCardOp{client: client}
}

type giftCardOp struct{ client core.Requester }

type GiftCard struct {
	ID           int64      `json:"id,omitempty"`
	Code         string     `json:"code,omitempty"`
	Balance      string     `json:"balance,omitempty"`
	Currency     string     `json:"currency,omitempty"`
	InitialValue string     `json:"initial_value,omitempty"`
	Note         string     `json:"note,omitempty"`
	DisabledAt   *time.Time `json:"disabled_at,omitempty"`
	ExpiresOn    string     `json:"expires_on,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

type giftCardResource struct {
	GiftCard *GiftCard `json:"gift_card"`
}
type giftCardsResource struct {
	GiftCards []GiftCard `json:"gift_cards"`
}

func (s *giftCardOp) List(ctx context.Context, opts *core.ListOptions) ([]GiftCard, error) {
	r := &giftCardsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("gift_cards.json"), r, opts)
	return r.GiftCards, err
}
func (s *giftCardOp) Get(ctx context.Context, id int64) (*GiftCard, error) {
	r := &giftCardResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("gift_cards/%d.json", id)), r, nil)
	return r.GiftCard, err
}
func (s *giftCardOp) Create(ctx context.Context, c GiftCard) (*GiftCard, error) {
	r := &giftCardResource{}
	err := s.client.Post(ctx, s.client.CreatePath("gift_cards.json"), giftCardResource{GiftCard: &c}, r)
	return r.GiftCard, err
}

package access

import (
	"context"

	"github.com/imokyou/slshop/core"
)

type StorefrontAccessTokenService interface {
	Create(ctx context.Context, title string) (*StorefrontAccessToken, error)
	List(ctx context.Context) ([]StorefrontAccessToken, error)
	Delete(ctx context.Context, id int64) error
}

func NewStorefrontAccessTokenService(client core.Requester) StorefrontAccessTokenService {
	return &storefrontOp{client: client}
}

type storefrontOp struct{ client core.Requester }

type StorefrontAccessToken struct {
	ID          int64  `json:"id,omitempty"`
	Title       string `json:"title,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
}

type tokenResource struct {
	StorefrontAccessToken *StorefrontAccessToken `json:"storefront_access_token"`
}
type tokensResource struct {
	StorefrontAccessTokens []StorefrontAccessToken `json:"storefront_access_tokens"`
}

func (s *storefrontOp) Create(ctx context.Context, title string) (*StorefrontAccessToken, error) {
	r := &tokenResource{}
	body := tokenResource{StorefrontAccessToken: &StorefrontAccessToken{Title: title}}
	err := s.client.Post(ctx, s.client.CreatePath("storefront_access_tokens.json"), body, r)
	return r.StorefrontAccessToken, err
}
func (s *storefrontOp) List(ctx context.Context) ([]StorefrontAccessToken, error) {
	r := &tokensResource{}
	err := s.client.Get(ctx, s.client.CreatePath("storefront_access_tokens.json"), r, nil)
	return r.StorefrontAccessTokens, err
}
func (s *storefrontOp) Delete(ctx context.Context, id int64) error {
	return s.client.Delete(ctx, s.client.CreatePath("storefront_access_tokens.json"))
}

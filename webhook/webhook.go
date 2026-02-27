package webhook

import (
	"context"
	"fmt"
	"time"

	"github.com/imokyou/slshop/core"
)

type Service interface {
	List(ctx context.Context, opts *core.ListOptions) ([]Subscription, error)
	Get(ctx context.Context, id int64) (*Subscription, error)
	Create(ctx context.Context, w Subscription) (*Subscription, error)
	Update(ctx context.Context, w Subscription) (*Subscription, error)
	Delete(ctx context.Context, id int64) error
}

func NewService(client core.Requester) Service {
	return &serviceOp{client: client}
}

type serviceOp struct{ client core.Requester }

type Subscription struct {
	ID        int64      `json:"id,omitempty"`
	Address   string     `json:"address,omitempty"`
	Topic     string     `json:"topic,omitempty"`
	Format    string     `json:"format,omitempty"`
	Fields    []string   `json:"fields,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type webhookResource struct {
	Webhook *Subscription `json:"webhook"`
}
type webhooksResource struct {
	Webhooks []Subscription `json:"webhooks"`
}

func (s *serviceOp) List(ctx context.Context, opts *core.ListOptions) ([]Subscription, error) {
	r := &webhooksResource{}
	err := s.client.Get(ctx, s.client.CreatePath("webhooks.json"), r, opts)
	return r.Webhooks, err
}
func (s *serviceOp) Get(ctx context.Context, id int64) (*Subscription, error) {
	r := &webhookResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("webhooks/%d.json", id)), r, nil)
	return r.Webhook, err
}
func (s *serviceOp) Create(ctx context.Context, w Subscription) (*Subscription, error) {
	r := &webhookResource{}
	err := s.client.Post(ctx, s.client.CreatePath("webhooks.json"), webhookResource{Webhook: &w}, r)
	return r.Webhook, err
}
func (s *serviceOp) Update(ctx context.Context, w Subscription) (*Subscription, error) {
	r := &webhookResource{}
	err := s.client.Put(ctx, s.client.CreatePath(fmt.Sprintf("webhooks/%d.json", w.ID)), webhookResource{Webhook: &w}, r)
	return r.Webhook, err
}
func (s *serviceOp) Delete(ctx context.Context, id int64) error {
	return s.client.Delete(ctx, s.client.CreatePath(fmt.Sprintf("webhooks/%d.json", id)))
}

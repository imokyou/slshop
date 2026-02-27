package onlinestore

import (
	"context"
	"fmt"
	"time"

	"github.com/imokyou/slshop/core"
)

// =====================================================================
// Theme
// =====================================================================

type ThemeService interface {
	List(ctx context.Context) ([]Theme, error)
	Get(ctx context.Context, id int64) (*Theme, error)
}

func NewThemeService(client core.Requester) ThemeService {
	return &themeOp{client: client}
}

type themeOp struct{ client core.Requester }

type Theme struct {
	ID           int64      `json:"id,omitempty"`
	Name         string     `json:"name,omitempty"`
	Role         string     `json:"role,omitempty"`
	ThemeStoreID int64      `json:"theme_store_id,omitempty"`
	Previewable  bool       `json:"previewable,omitempty"`
	Processing   bool       `json:"processing,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

type themeResource struct {
	Theme *Theme `json:"theme"`
}
type themesResource struct {
	Themes []Theme `json:"themes"`
}

func (s *themeOp) List(ctx context.Context) ([]Theme, error) {
	r := &themesResource{}
	err := s.client.Get(ctx, s.client.CreatePath("themes.json"), r, nil)
	return r.Themes, err
}
func (s *themeOp) Get(ctx context.Context, id int64) (*Theme, error) {
	r := &themeResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("themes/%d.json", id)), r, nil)
	return r.Theme, err
}

// =====================================================================
// Page
// =====================================================================

type PageService interface {
	List(ctx context.Context, opts *core.ListOptions) ([]Page, error)
	Get(ctx context.Context, id int64) (*Page, error)
	Create(ctx context.Context, p Page) (*Page, error)
	Update(ctx context.Context, p Page) (*Page, error)
	Delete(ctx context.Context, id int64) error
}

func NewPageService(client core.Requester) PageService {
	return &pageOp{client: client}
}

type pageOp struct{ client core.Requester }

type Page struct {
	ID             int64      `json:"id,omitempty"`
	Title          string     `json:"title,omitempty"`
	Handle         string     `json:"handle,omitempty"`
	BodyHTML       string     `json:"body_html,omitempty"`
	Author         string     `json:"author,omitempty"`
	TemplateSuffix string     `json:"template_suffix,omitempty"`
	Published      bool       `json:"published,omitempty"`
	CreatedAt      *time.Time `json:"created_at,omitempty"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty"`
	PublishedAt    *time.Time `json:"published_at,omitempty"`
}

type pageResource struct {
	Page *Page `json:"page"`
}
type pagesResource struct {
	Pages []Page `json:"pages"`
}

func (s *pageOp) List(ctx context.Context, opts *core.ListOptions) ([]Page, error) {
	r := &pagesResource{}
	err := s.client.Get(ctx, s.client.CreatePath("pages.json"), r, opts)
	return r.Pages, err
}
func (s *pageOp) Get(ctx context.Context, id int64) (*Page, error) {
	r := &pageResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("pages/%d.json", id)), r, nil)
	return r.Page, err
}
func (s *pageOp) Create(ctx context.Context, p Page) (*Page, error) {
	r := &pageResource{}
	err := s.client.Post(ctx, s.client.CreatePath("pages.json"), pageResource{Page: &p}, r)
	return r.Page, err
}
func (s *pageOp) Update(ctx context.Context, p Page) (*Page, error) {
	r := &pageResource{}
	err := s.client.Put(ctx, s.client.CreatePath(fmt.Sprintf("pages/%d.json", p.ID)), pageResource{Page: &p}, r)
	return r.Page, err
}
func (s *pageOp) Delete(ctx context.Context, id int64) error {
	return s.client.Delete(ctx, s.client.CreatePath(fmt.Sprintf("pages/%d.json", id)))
}

// =====================================================================
// ScriptTag
// =====================================================================

type ScriptTagService interface {
	List(ctx context.Context, opts *core.ListOptions) ([]ScriptTag, error)
	Get(ctx context.Context, id int64) (*ScriptTag, error)
	Create(ctx context.Context, t ScriptTag) (*ScriptTag, error)
	Delete(ctx context.Context, id int64) error
}

func NewScriptTagService(client core.Requester) ScriptTagService {
	return &scriptTagOp{client: client}
}

type scriptTagOp struct{ client core.Requester }

type ScriptTag struct {
	ID           int64      `json:"id,omitempty"`
	Event        string     `json:"event,omitempty"`
	Src          string     `json:"src,omitempty"`
	DisplayScope string     `json:"display_scope,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

type scriptTagResource struct {
	ScriptTag *ScriptTag `json:"script_tag"`
}
type scriptTagsResource struct {
	ScriptTags []ScriptTag `json:"script_tags"`
}

func (s *scriptTagOp) List(ctx context.Context, opts *core.ListOptions) ([]ScriptTag, error) {
	r := &scriptTagsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("script_tags.json"), r, opts)
	return r.ScriptTags, err
}
func (s *scriptTagOp) Get(ctx context.Context, id int64) (*ScriptTag, error) {
	r := &scriptTagResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("script_tags/%d.json", id)), r, nil)
	return r.ScriptTag, err
}
func (s *scriptTagOp) Create(ctx context.Context, t ScriptTag) (*ScriptTag, error) {
	r := &scriptTagResource{}
	err := s.client.Post(ctx, s.client.CreatePath("script_tags.json"), scriptTagResource{ScriptTag: &t}, r)
	return r.ScriptTag, err
}
func (s *scriptTagOp) Delete(ctx context.Context, id int64) error {
	return s.client.Delete(ctx, s.client.CreatePath(fmt.Sprintf("script_tags/%d.json", id)))
}

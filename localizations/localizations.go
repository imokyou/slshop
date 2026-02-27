package localizations

import (
	"context"

	"github.com/imokyou/slshop/core"
)

// =====================================================================
// Localizations Service
// =====================================================================

type Service interface {
	// Language management
	GetLanguages(ctx context.Context) (*LanguageData, error)
	AddLanguages(ctx context.Context, languages []string) (*LanguageData, error)
	DeleteLanguages(ctx context.Context, languages []string) (*LanguageData, error)
	GetAvailableLanguages(ctx context.Context) ([]AvailableLanguage, error)

	// Translation management
	GetTranslation(ctx context.Context, opts *TranslationQuery) (*TranslationData, error)
	UpdateTranslation(ctx context.Context, data TranslationUpdateRequest) error
	DeleteTranslation(ctx context.Context, data TranslationDeleteRequest) error
	BatchQueryTranslation(ctx context.Context, opts *TranslationBatchQuery) ([]TranslationData, error)
}

func NewService(client core.Requester) Service {
	return &serviceOp{client: client}
}

type serviceOp struct{ client core.Requester }

// =====================================================================
// Models
// =====================================================================

type LanguageData struct {
	DefaultLanguage    string   `json:"default_language,omitempty"`
	SupportedLanguages []string `json:"supported_languages,omitempty"`
}

type AvailableLanguage struct {
	Locale string `json:"locale,omitempty"`
	Name   string `json:"name,omitempty"`
}

type TranslationQuery struct {
	ResourceID   string `url:"resource_id,omitempty"`
	ResourceType string `url:"resource_type,omitempty"`
	Locale       string `url:"locale,omitempty"`
	Outdated     string `url:"outdated,omitempty"`
	Key          string `url:"key,omitempty"`
}

type TranslationBatchQuery struct {
	core.ListOptions
	ResourceType string `url:"resource_type,omitempty"`
	Locale       string `url:"locale,omitempty"`
}

type TranslationData struct {
	ResourceID   string             `json:"resource_id,omitempty"`
	ResourceType string             `json:"resource_type,omitempty"`
	Translations []TranslationEntry `json:"translations,omitempty"`
}

type TranslationEntry struct {
	Key      string `json:"key,omitempty"`
	Value    string `json:"value,omitempty"`
	Locale   string `json:"locale,omitempty"`
	Outdated bool   `json:"outdated,omitempty"`
}

type TranslationUpdateRequest struct {
	ResourceID   string             `json:"resource_id,omitempty"`
	ResourceType string             `json:"resource_type,omitempty"`
	Translations []TranslationEntry `json:"translations,omitempty"`
}

type TranslationDeleteRequest struct {
	ResourceID   string   `json:"resource_id,omitempty"`
	ResourceType string   `json:"resource_type,omitempty"`
	Keys         []string `json:"keys,omitempty"`
	Locales      []string `json:"locales,omitempty"`
}

// JSON wrappers
type languageDataResource struct {
	Data *LanguageData `json:"data"`
}
type availableLanguagesResource struct {
	Languages []AvailableLanguage `json:"languages"`
}
type translationDataResource struct {
	Data *TranslationData `json:"data"`
}
type translationBatchResource struct {
	Data []TranslationData `json:"data"`
}

// =====================================================================
// Implementation
// =====================================================================

// GET store/languages.json
func (s *serviceOp) GetLanguages(ctx context.Context) (*LanguageData, error) {
	r := &languageDataResource{}
	err := s.client.Get(ctx, s.client.CreatePath("store/languages.json"), r, nil)
	return r.Data, err
}

// POST store/languages.json
func (s *serviceOp) AddLanguages(ctx context.Context, languages []string) (*LanguageData, error) {
	r := &languageDataResource{}
	body := map[string][]string{"languages": languages}
	err := s.client.Post(ctx, s.client.CreatePath("store/languages.json"), body, r)
	return r.Data, err
}

// DELETE store/languages.json (body: {"languages": [...]})
func (s *serviceOp) DeleteLanguages(ctx context.Context, languages []string) (*LanguageData, error) {
	// Note: Shopline uses DELETE with body for this endpoint
	r := &languageDataResource{}
	body := map[string][]string{"languages": languages}
	err := s.client.Post(ctx, s.client.CreatePath("store/languages/delete.json"), body, r)
	return r.Data, err
}

// GET store/available_languages.json
func (s *serviceOp) GetAvailableLanguages(ctx context.Context) ([]AvailableLanguage, error) {
	r := &availableLanguagesResource{}
	err := s.client.Get(ctx, s.client.CreatePath("store/available_languages.json"), r, nil)
	return r.Languages, err
}

// GET ugc/resource.json
func (s *serviceOp) GetTranslation(ctx context.Context, opts *TranslationQuery) (*TranslationData, error) {
	r := &translationDataResource{}
	err := s.client.Get(ctx, s.client.CreatePath("ugc/resource.json"), r, opts)
	return r.Data, err
}

// PUT ugc/resource.json
func (s *serviceOp) UpdateTranslation(ctx context.Context, data TranslationUpdateRequest) error {
	return s.client.Put(ctx, s.client.CreatePath("ugc/resource.json"), data, nil)
}

// DELETE ugc/resource.json (delete translations)
func (s *serviceOp) DeleteTranslation(ctx context.Context, data TranslationDeleteRequest) error {
	return s.client.Post(ctx, s.client.CreatePath("ugc/resource/delete.json"), data, nil)
}

// GET ugc/resources.json
func (s *serviceOp) BatchQueryTranslation(ctx context.Context, opts *TranslationBatchQuery) ([]TranslationData, error) {
	r := &translationBatchResource{}
	err := s.client.Get(ctx, s.client.CreatePath("ugc/resources.json"), r, opts)
	return r.Data, err
}

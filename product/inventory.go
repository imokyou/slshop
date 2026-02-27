package product

import (
	"context"
	"fmt"
	"time"

	"github.com/imokyou/slshop/core"
)

// =====================================================================
// Inventory
// =====================================================================

type InventoryService interface {
	ListItems(ctx context.Context, opts *core.ListOptions) ([]InventoryItem, error)
	GetItem(ctx context.Context, id int64) (*InventoryItem, error)
	UpdateItem(ctx context.Context, item InventoryItem) (*InventoryItem, error)

	ListLevels(ctx context.Context, opts *InventoryLevelListOptions) ([]InventoryLevel, error)
	SetLevel(ctx context.Context, level InventoryLevel) (*InventoryLevel, error)
	AdjustLevel(ctx context.Context, inventoryItemID, locationID int64, adjustment int) (*InventoryLevel, error)
}

func NewInventoryService(client core.Requester) InventoryService {
	return &inventoryOp{client: client}
}

type inventoryOp struct{ client core.Requester }

type InventoryItem struct {
	ID                   int64      `json:"id,omitempty"`
	SKU                  string     `json:"sku,omitempty"`
	Cost                 string     `json:"cost,omitempty"`
	Tracked              bool       `json:"tracked,omitempty"`
	CountryCodeOfOrigin  string     `json:"country_code_of_origin,omitempty"`
	ProvinceCodeOfOrigin string     `json:"province_code_of_origin,omitempty"`
	HarmonizedSystemCode string     `json:"harmonized_system_code,omitempty"`
	RequiresShipping     bool       `json:"requires_shipping,omitempty"`
	CreatedAt            *time.Time `json:"created_at,omitempty"`
	UpdatedAt            *time.Time `json:"updated_at,omitempty"`
}

type InventoryLevel struct {
	InventoryItemID int64      `json:"inventory_item_id,omitempty"`
	LocationID      int64      `json:"location_id,omitempty"`
	Available       int        `json:"available,omitempty"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty"`
}

type InventoryLevelListOptions struct {
	core.ListOptions
	InventoryItemIDs string `url:"inventory_item_ids,omitempty"`
	LocationIDs      string `url:"location_ids,omitempty"`
}

type inventoryItemResource struct {
	InventoryItem *InventoryItem `json:"inventory_item"`
}
type inventoryItemsResource struct {
	InventoryItems []InventoryItem `json:"inventory_items"`
}
type inventoryLevelResource struct {
	InventoryLevel *InventoryLevel `json:"inventory_level"`
}
type inventoryLevelsResource struct {
	InventoryLevels []InventoryLevel `json:"inventory_levels"`
}

func (s *inventoryOp) ListItems(ctx context.Context, opts *core.ListOptions) ([]InventoryItem, error) {
	r := &inventoryItemsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("inventory_items.json"), r, opts)
	return r.InventoryItems, err
}
func (s *inventoryOp) GetItem(ctx context.Context, id int64) (*InventoryItem, error) {
	r := &inventoryItemResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("inventory_items/%d.json", id)), r, nil)
	return r.InventoryItem, err
}
func (s *inventoryOp) UpdateItem(ctx context.Context, item InventoryItem) (*InventoryItem, error) {
	r := &inventoryItemResource{}
	err := s.client.Put(ctx, s.client.CreatePath(fmt.Sprintf("inventory_items/%d.json", item.ID)), inventoryItemResource{InventoryItem: &item}, r)
	return r.InventoryItem, err
}
func (s *inventoryOp) ListLevels(ctx context.Context, opts *InventoryLevelListOptions) ([]InventoryLevel, error) {
	r := &inventoryLevelsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("inventory_levels.json"), r, opts)
	return r.InventoryLevels, err
}
func (s *inventoryOp) SetLevel(ctx context.Context, level InventoryLevel) (*InventoryLevel, error) {
	r := &inventoryLevelResource{}
	err := s.client.Post(ctx, s.client.CreatePath("inventory_levels/set.json"), inventoryLevelResource{InventoryLevel: &level}, r)
	return r.InventoryLevel, err
}
func (s *inventoryOp) AdjustLevel(ctx context.Context, inventoryItemID, locationID int64, adjustment int) (*InventoryLevel, error) {
	body := map[string]interface{}{
		"inventory_item_id":    inventoryItemID,
		"location_id":          locationID,
		"available_adjustment": adjustment,
	}
	r := &inventoryLevelResource{}
	err := s.client.Post(ctx, s.client.CreatePath("inventory_levels/adjust.json"), body, r)
	return r.InventoryLevel, err
}

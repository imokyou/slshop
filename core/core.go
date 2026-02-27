package core

import (
	"context"
	"time"
)

// Requester defines the interface for making HTTP requests to the Shopline API.
// Sub-packages depend on this interface instead of the root Client directly,
// which avoids circular dependencies.
type Requester interface {
	Get(ctx context.Context, path string, result interface{}, opts interface{}) error
	Post(ctx context.Context, path string, body, result interface{}) error
	Put(ctx context.Context, path string, body, result interface{}) error
	Delete(ctx context.Context, path string) error
	CreatePath(resource string) string
}

// =====================================================================
// 共享类型 — 多个子包引用的通用模型
// =====================================================================

// ListOptions specifies the optional parameters to various List methods.
type ListOptions struct {
	Page         int    `url:"page,omitempty"`
	Limit        int    `url:"limit,omitempty"`
	SinceID      int64  `url:"since_id,omitempty"`
	CreatedAtMin string `url:"created_at_min,omitempty"`
	CreatedAtMax string `url:"created_at_max,omitempty"`
	UpdatedAtMin string `url:"updated_at_min,omitempty"`
	UpdatedAtMax string `url:"updated_at_max,omitempty"`
	Fields       string `url:"fields,omitempty"`
}

// CountOptions specifies the optional parameters for Count methods.
type CountOptions struct {
	CreatedAtMin string `url:"created_at_min,omitempty"`
	CreatedAtMax string `url:"created_at_max,omitempty"`
	UpdatedAtMin string `url:"updated_at_min,omitempty"`
	UpdatedAtMax string `url:"updated_at_max,omitempty"`
}

// Address represents a mailing address (used by Order, Customer, etc.).
type Address struct {
	ID             int64   `json:"id,omitempty"`
	FirstName      string  `json:"first_name,omitempty"`
	LastName       string  `json:"last_name,omitempty"`
	Company        string  `json:"company,omitempty"`
	Address1       string  `json:"address1,omitempty"`
	Address2       string  `json:"address2,omitempty"`
	City           string  `json:"city,omitempty"`
	CityCode       string  `json:"city_code,omitempty"`
	Province       string  `json:"province,omitempty"`
	ProvinceCode   string  `json:"province_code,omitempty"`
	Country        string  `json:"country,omitempty"`
	CountryCode    string  `json:"country_code,omitempty"`
	Area           string  `json:"area,omitempty"`
	AreaCode       string  `json:"area_code,omitempty"`
	Zip            string  `json:"zip,omitempty"`
	Phone          string  `json:"phone,omitempty"`
	Email          string  `json:"email,omitempty"`
	Name           string  `json:"name,omitempty"`
	Latitude       float64 `json:"latitude,omitempty"`
	Longitude      float64 `json:"longitude,omitempty"`
	Default        bool    `json:"default,omitempty"`
	SameAsReceiver *bool   `json:"same_as_receiver,omitempty"`
}

// Customer represents a Shopline customer (shared, used by Order and others).
type Customer struct {
	ID                        int64      `json:"id,omitempty"`
	Email                     string     `json:"email,omitempty"`
	Phone                     string     `json:"phone,omitempty"`
	FirstName                 string     `json:"first_name,omitempty"`
	LastName                  string     `json:"last_name,omitempty"`
	State                     string     `json:"state,omitempty"`
	Note                      string     `json:"note,omitempty"`
	Tags                      string     `json:"tags,omitempty"`
	Currency                  string     `json:"currency,omitempty"`
	TotalSpent                string     `json:"total_spent,omitempty"`
	OrdersCount               int        `json:"orders_count,omitempty"`
	TaxExempt                 bool       `json:"tax_exempt,omitempty"`
	VerifiedEmail             bool       `json:"verified_email,omitempty"`
	AcceptsMarketing          bool       `json:"accepts_marketing,omitempty"`
	Addresses                 []Address  `json:"addresses,omitempty"`
	DefaultAddress            *Address   `json:"default_address,omitempty"`
	LastOrderID               int64      `json:"last_order_id,omitempty"`
	LastOrderName             string     `json:"last_order_name,omitempty"`
	Password                  string     `json:"password,omitempty"`
	PasswordConfirmation      string     `json:"password_confirmation,omitempty"`
	SendEmailWelcome          *bool      `json:"send_email_welcome,omitempty"`
	SendEmailInvite           *bool      `json:"send_email_invite,omitempty"`
	AcceptsMarketingUpdatedAt *time.Time `json:"accepts_marketing_updated_at,omitempty"`
	CreatedAt                 *time.Time `json:"created_at,omitempty"`
	UpdatedAt                 *time.Time `json:"updated_at,omitempty"`
}

// LineItem represents a line item in an order.
type LineItem struct {
	ID                  int64              `json:"id,omitempty"`
	VariantID           interface{}        `json:"variant_id,omitempty"`
	ProductID           interface{}        `json:"product_id,omitempty"`
	Title               string             `json:"title,omitempty"`
	VariantTitle        string             `json:"variant_title,omitempty"`
	Name                string             `json:"name,omitempty"`
	SKU                 string             `json:"sku,omitempty"`
	Price               string             `json:"price,omitempty"`
	Quantity            int                `json:"quantity,omitempty"`
	TotalDiscount       string             `json:"total_discount,omitempty"`
	Grams               int                `json:"grams,omitempty"`
	Vendor              string             `json:"vendor,omitempty"`
	FulfillmentStatus   string             `json:"fulfillment_status,omitempty"`
	FulfillableQuantity int                `json:"fulfillable_quantity,omitempty"`
	Taxable             *bool              `json:"taxable,omitempty"`
	GiftCard            bool               `json:"gift_card,omitempty"`
	RequiresShipping    *bool              `json:"requires_shipping,omitempty"`
	LocationID          string             `json:"location_id,omitempty"`
	ShippingLineTitle   string             `json:"shipping_line_title,omitempty"`
	Properties          []LineItemProperty `json:"properties,omitempty"`
	TaxLines            []TaxLine          `json:"tax_lines,omitempty"`
	TaxLine             *TaxLine           `json:"tax_line,omitempty"`
	DiscountPrice       *LineItemDiscount  `json:"discount_price,omitempty"`
}

// LineItemProperty represents a custom property on a line item.
type LineItemProperty struct {
	Name  string   `json:"name,omitempty"`
	Value string   `json:"value,omitempty"`
	Show  *bool    `json:"show,omitempty"`
	Type  string   `json:"type,omitempty"`
	URLs  []string `json:"urls,omitempty"`
}

// LineItemDiscount represents a custom discount on a line item.
type LineItemDiscount struct {
	Amount string `json:"amount,omitempty"`
	Title  string `json:"title,omitempty"`
}

// ShippingLine represents a shipping line in an order.
type ShippingLine struct {
	ID       int64     `json:"id,omitempty"`
	Title    string    `json:"title,omitempty"`
	Price    string    `json:"price,omitempty"`
	Code     string    `json:"code,omitempty"`
	Source   string    `json:"source,omitempty"`
	TaxLines []TaxLine `json:"tax_lines,omitempty"`
	TaxLine  *TaxLine  `json:"tax_line,omitempty"`
}

// TaxLine represents a tax line item.
type TaxLine struct {
	Title string  `json:"title,omitempty"`
	Price string  `json:"price,omitempty"`
	Rate  float64 `json:"rate,omitempty"`
}

// DiscountCode represents a discount code applied to an order.
type DiscountCode struct {
	Code   string `json:"code,omitempty"`
	Amount string `json:"amount,omitempty"`
	Type   string `json:"type,omitempty"`
}

// NoteAttribute represents a key-value note attribute.
type NoteAttribute struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

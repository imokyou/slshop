package order

import (
	"context"
	"fmt"
	"time"

	"github.com/imokyou/slshop/core"
)

// =====================================================================
// Fulfillment
// =====================================================================

type FulfillmentService interface {
	List(ctx context.Context, orderID int64, opts *core.ListOptions) ([]Fulfillment, error)
	Create(ctx context.Context, orderID int64, f Fulfillment) (*Fulfillment, error)
	Cancel(ctx context.Context, orderID, fulfillmentID int64) (*Fulfillment, error)
	UpdateTracking(ctx context.Context, orderID, fulfillmentID int64, t FulfillmentTracking) (*Fulfillment, error)

	ListByFulfillmentOrder(ctx context.Context, foID int64) ([]Fulfillment, error)
	GetByFulfillmentOrder(ctx context.Context, foID, fID int64) (*Fulfillment, error)
	CreateByFulfillmentOrder(ctx context.Context, foID int64, f Fulfillment) (*Fulfillment, error)

	UpdateTrackingGlobal(ctx context.Context, fID int64, t FulfillmentTracking) (*Fulfillment, error)
	CancelGlobal(ctx context.Context, fID int64) (*Fulfillment, error)
	Count(ctx context.Context) (int, error)

	MoveFulfillmentOrder(ctx context.Context, foID, locationID int64) error
	HoldFulfillmentOrder(ctx context.Context, foID int64, hold FulfillmentHold) error

	ListInventoryLocations(ctx context.Context) ([]InventoryLocation, error)
	ListShippingMethods(ctx context.Context) ([]ShippingMethod, error)
	ListPickupMethods(ctx context.Context) ([]PickupMethod, error)
}

func NewFulfillmentService(client core.Requester) FulfillmentService {
	return &fulfillmentOp{client: client}
}

type fulfillmentOp struct{ client core.Requester }

type Fulfillment struct {
	ID              int64               `json:"id,omitempty"`
	OrderID         int64               `json:"order_id,omitempty"`
	Status          string              `json:"status,omitempty"`
	TrackingCompany string              `json:"tracking_company,omitempty"`
	TrackingNumber  string              `json:"tracking_number,omitempty"`
	TrackingNumbers []string            `json:"tracking_numbers,omitempty"`
	TrackingURL     string              `json:"tracking_url,omitempty"`
	TrackingURLs    []string            `json:"tracking_urls,omitempty"`
	LineItems       []core.LineItem `json:"line_items,omitempty"`
	NotifyCustomer  bool                `json:"notify_customer,omitempty"`
	LocationID      int64               `json:"location_id,omitempty"`
	CreatedAt       *time.Time          `json:"created_at,omitempty"`
	UpdatedAt       *time.Time          `json:"updated_at,omitempty"`
}

type FulfillmentTracking struct {
	TrackingNumber  string   `json:"tracking_number,omitempty"`
	TrackingNumbers []string `json:"tracking_numbers,omitempty"`
	TrackingCompany string   `json:"tracking_company,omitempty"`
	TrackingURL     string   `json:"tracking_url,omitempty"`
	TrackingURLs    []string `json:"tracking_urls,omitempty"`
	NotifyCustomer  bool     `json:"notify_customer,omitempty"`
}

type FulfillmentHold struct {
	Reason      string `json:"reason,omitempty"`
	ReasonNotes string `json:"reason_notes,omitempty"`
}

type InventoryLocation struct {
	ID       int64  `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Address1 string `json:"address1,omitempty"`
	Address2 string `json:"address2,omitempty"`
	City     string `json:"city,omitempty"`
	Country  string `json:"country,omitempty"`
	Active   bool   `json:"active,omitempty"`
}

type ShippingMethod struct {
	ID    int64  `json:"id,omitempty"`
	Title string `json:"title,omitempty"`
	Price string `json:"price,omitempty"`
}

type PickupMethod struct {
	ID    int64  `json:"id,omitempty"`
	Title string `json:"title,omitempty"`
}

// Carrier Service
type CarrierServiceService interface {
	List(ctx context.Context) ([]CarrierService, error)
	Get(ctx context.Context, id int64) (*CarrierService, error)
	Create(ctx context.Context, c CarrierService) (*CarrierService, error)
	Update(ctx context.Context, c CarrierService) (*CarrierService, error)
	Delete(ctx context.Context, id int64) error
}

func NewCarrierServiceService(client core.Requester) CarrierServiceService {
	return &carrierOp{client: client}
}

type carrierOp struct{ client core.Requester }

type CarrierService struct {
	ID                 int64  `json:"id,omitempty"`
	Name               string `json:"name,omitempty"`
	Active             bool   `json:"active,omitempty"`
	ServiceDiscovery   bool   `json:"service_discovery,omitempty"`
	CarrierServiceType string `json:"carrier_service_type,omitempty"`
	Format             string `json:"format,omitempty"`
	CallbackURL        string `json:"callback_url,omitempty"`
}

// Fulfillment Service Definition
type FulfillmentServiceDefService interface {
	List(ctx context.Context) ([]FulfillmentServiceDef, error)
	Get(ctx context.Context, id int64) (*FulfillmentServiceDef, error)
	Create(ctx context.Context, svc FulfillmentServiceDef) (*FulfillmentServiceDef, error)
	Update(ctx context.Context, svc FulfillmentServiceDef) (*FulfillmentServiceDef, error)
	Delete(ctx context.Context, id int64) error
	CreateLocation(ctx context.Context, loc FulfillmentServiceLocation) (*FulfillmentServiceLocation, error)
}

func NewFulfillmentServiceDefService(client core.Requester) FulfillmentServiceDefService {
	return &fulfillmentSvcOp{client: client}
}

type fulfillmentSvcOp struct{ client core.Requester }

type FulfillmentServiceDef struct {
	ID                     int64  `json:"id,omitempty"`
	Name                   string `json:"name,omitempty"`
	Email                  string `json:"email,omitempty"`
	Handle                 string `json:"handle,omitempty"`
	CallbackURL            string `json:"callback_url,omitempty"`
	FulfillmentOrdersOptIn bool   `json:"fulfillment_orders_opt_in,omitempty"`
	IncludePendingStock    bool   `json:"include_pending_stock,omitempty"`
	TrackingSupport        bool   `json:"tracking_support,omitempty"`
	InventoryManagement    bool   `json:"inventory_management,omitempty"`
}

type FulfillmentServiceLocation struct {
	ID       int64  `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Address1 string `json:"address1,omitempty"`
	City     string `json:"city,omitempty"`
	Country  string `json:"country,omitempty"`
}

// JSON wrappers
type fulfillmentResource struct {
	Fulfillment *Fulfillment `json:"fulfillment"`
}
type fulfillmentsResource struct {
	Fulfillments []Fulfillment `json:"fulfillments"`
}
type inventoryLocationsResource struct {
	InventoryLocations []InventoryLocation `json:"inventory_locations"`
}
type shippingMethodsResource struct {
	ShippingMethods []ShippingMethod `json:"shipping_methods"`
}
type pickupMethodsResource struct {
	PickupMethods []PickupMethod `json:"pickup_methods"`
}
type carrierServiceResource struct {
	CarrierService *CarrierService `json:"carrier_service"`
}
type carrierServicesResource struct {
	CarrierServices []CarrierService `json:"carrier_services"`
}
type fulfillmentSvcDefResource struct {
	FulfillmentService *FulfillmentServiceDef `json:"fulfillment_service"`
}
type fulfillmentSvcDefsResource struct {
	FulfillmentServices []FulfillmentServiceDef `json:"fulfillment_services"`
}
type fulfillmentSvcLocResource struct {
	FulfillmentServiceLocation *FulfillmentServiceLocation `json:"fulfillment_service_location"`
}

// === Fulfillment implementation ===

func (s *fulfillmentOp) List(ctx context.Context, orderID int64, opts *core.ListOptions) ([]Fulfillment, error) {
	path := s.client.CreatePath(fmt.Sprintf("orders/%d/fulfillments.json", orderID))
	r := &fulfillmentsResource{}
	err := s.client.Get(ctx, path, r, opts)
	return r.Fulfillments, err
}
func (s *fulfillmentOp) Create(ctx context.Context, orderID int64, f Fulfillment) (*Fulfillment, error) {
	path := s.client.CreatePath(fmt.Sprintf("orders/%d/fulfillments.json", orderID))
	r := &fulfillmentResource{}
	err := s.client.Post(ctx, path, fulfillmentResource{Fulfillment: &f}, r)
	return r.Fulfillment, err
}
func (s *fulfillmentOp) Cancel(ctx context.Context, orderID, fID int64) (*Fulfillment, error) {
	path := s.client.CreatePath(fmt.Sprintf("orders/%d/fulfillments/%d/cancel.json", orderID, fID))
	r := &fulfillmentResource{}
	err := s.client.Post(ctx, path, nil, r)
	return r.Fulfillment, err
}
func (s *fulfillmentOp) UpdateTracking(ctx context.Context, orderID, fID int64, t FulfillmentTracking) (*Fulfillment, error) {
	path := s.client.CreatePath(fmt.Sprintf("orders/%d/fulfillments/%d/update_tracking.json", orderID, fID))
	r := &fulfillmentResource{}
	err := s.client.Post(ctx, path, t, r)
	return r.Fulfillment, err
}
func (s *fulfillmentOp) ListByFulfillmentOrder(ctx context.Context, foID int64) ([]Fulfillment, error) {
	path := s.client.CreatePath(fmt.Sprintf("fulfillment_orders/%d/fulfillments.json", foID))
	r := &fulfillmentsResource{}
	err := s.client.Get(ctx, path, r, nil)
	return r.Fulfillments, err
}
func (s *fulfillmentOp) GetByFulfillmentOrder(ctx context.Context, foID, fID int64) (*Fulfillment, error) {
	path := s.client.CreatePath(fmt.Sprintf("fulfillment_orders/%d/fulfillments/%d.json", foID, fID))
	r := &fulfillmentResource{}
	err := s.client.Get(ctx, path, r, nil)
	return r.Fulfillment, err
}
func (s *fulfillmentOp) CreateByFulfillmentOrder(ctx context.Context, foID int64, f Fulfillment) (*Fulfillment, error) {
	path := s.client.CreatePath(fmt.Sprintf("fulfillment_orders/%d/fulfillments.json", foID))
	r := &fulfillmentResource{}
	err := s.client.Post(ctx, path, fulfillmentResource{Fulfillment: &f}, r)
	return r.Fulfillment, err
}
func (s *fulfillmentOp) UpdateTrackingGlobal(ctx context.Context, fID int64, t FulfillmentTracking) (*Fulfillment, error) {
	path := s.client.CreatePath(fmt.Sprintf("fulfillments/%d/update_tracking.json", fID))
	r := &fulfillmentResource{}
	err := s.client.Post(ctx, path, t, r)
	return r.Fulfillment, err
}
func (s *fulfillmentOp) CancelGlobal(ctx context.Context, fID int64) (*Fulfillment, error) {
	path := s.client.CreatePath(fmt.Sprintf("fulfillments/%d/cancel.json", fID))
	r := &fulfillmentResource{}
	err := s.client.Post(ctx, path, nil, r)
	return r.Fulfillment, err
}
func (s *fulfillmentOp) Count(ctx context.Context) (int, error) {
	path := s.client.CreatePath("fulfillments/count.json")
	r := &countResource{}
	err := s.client.Get(ctx, path, r, nil)
	return r.Count, err
}
func (s *fulfillmentOp) MoveFulfillmentOrder(ctx context.Context, foID, locationID int64) error {
	path := s.client.CreatePath(fmt.Sprintf("fulfillment_orders/%d/move.json", foID))
	return s.client.Post(ctx, path, map[string]int64{"location_id": locationID}, nil)
}
func (s *fulfillmentOp) HoldFulfillmentOrder(ctx context.Context, foID int64, hold FulfillmentHold) error {
	path := s.client.CreatePath(fmt.Sprintf("fulfillment_orders/%d/hold.json", foID))
	return s.client.Post(ctx, path, hold, nil)
}
func (s *fulfillmentOp) ListInventoryLocations(ctx context.Context) ([]InventoryLocation, error) {
	r := &inventoryLocationsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("inventory_locations.json"), r, nil)
	return r.InventoryLocations, err
}
func (s *fulfillmentOp) ListShippingMethods(ctx context.Context) ([]ShippingMethod, error) {
	r := &shippingMethodsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("shipping_methods.json"), r, nil)
	return r.ShippingMethods, err
}
func (s *fulfillmentOp) ListPickupMethods(ctx context.Context) ([]PickupMethod, error) {
	r := &pickupMethodsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("pickup_methods.json"), r, nil)
	return r.PickupMethods, err
}

// === Carrier Service implementation ===
func (s *carrierOp) List(ctx context.Context) ([]CarrierService, error) {
	r := &carrierServicesResource{}
	err := s.client.Get(ctx, s.client.CreatePath("carrier_services.json"), r, nil)
	return r.CarrierServices, err
}
func (s *carrierOp) Get(ctx context.Context, id int64) (*CarrierService, error) {
	r := &carrierServiceResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("carrier_services/%d.json", id)), r, nil)
	return r.CarrierService, err
}
func (s *carrierOp) Create(ctx context.Context, c CarrierService) (*CarrierService, error) {
	r := &carrierServiceResource{}
	err := s.client.Post(ctx, s.client.CreatePath("carrier_services.json"), carrierServiceResource{CarrierService: &c}, r)
	return r.CarrierService, err
}
func (s *carrierOp) Update(ctx context.Context, c CarrierService) (*CarrierService, error) {
	r := &carrierServiceResource{}
	err := s.client.Put(ctx, s.client.CreatePath(fmt.Sprintf("carrier_services/%d.json", c.ID)), carrierServiceResource{CarrierService: &c}, r)
	return r.CarrierService, err
}
func (s *carrierOp) Delete(ctx context.Context, id int64) error {
	return s.client.Delete(ctx, s.client.CreatePath(fmt.Sprintf("carrier_services/%d.json", id)))
}

// === Fulfillment Service Def implementation ===
func (s *fulfillmentSvcOp) List(ctx context.Context) ([]FulfillmentServiceDef, error) {
	r := &fulfillmentSvcDefsResource{}
	err := s.client.Get(ctx, s.client.CreatePath("fulfillment_services.json"), r, nil)
	return r.FulfillmentServices, err
}
func (s *fulfillmentSvcOp) Get(ctx context.Context, id int64) (*FulfillmentServiceDef, error) {
	r := &fulfillmentSvcDefResource{}
	err := s.client.Get(ctx, s.client.CreatePath(fmt.Sprintf("fulfillment_services/%d.json", id)), r, nil)
	return r.FulfillmentService, err
}
func (s *fulfillmentSvcOp) Create(ctx context.Context, svc FulfillmentServiceDef) (*FulfillmentServiceDef, error) {
	r := &fulfillmentSvcDefResource{}
	err := s.client.Post(ctx, s.client.CreatePath("fulfillment_services.json"), fulfillmentSvcDefResource{FulfillmentService: &svc}, r)
	return r.FulfillmentService, err
}
func (s *fulfillmentSvcOp) Update(ctx context.Context, svc FulfillmentServiceDef) (*FulfillmentServiceDef, error) {
	r := &fulfillmentSvcDefResource{}
	err := s.client.Put(ctx, s.client.CreatePath(fmt.Sprintf("fulfillment_services/%d.json", svc.ID)), fulfillmentSvcDefResource{FulfillmentService: &svc}, r)
	return r.FulfillmentService, err
}
func (s *fulfillmentSvcOp) Delete(ctx context.Context, id int64) error {
	return s.client.Delete(ctx, s.client.CreatePath(fmt.Sprintf("fulfillment_services/%d.json", id)))
}
func (s *fulfillmentSvcOp) CreateLocation(ctx context.Context, loc FulfillmentServiceLocation) (*FulfillmentServiceLocation, error) {
	r := &fulfillmentSvcLocResource{}
	err := s.client.Post(ctx, s.client.CreatePath("fulfillment_services/fulfillment_service_location.json"), fulfillmentSvcLocResource{FulfillmentServiceLocation: &loc}, r)
	return r.FulfillmentServiceLocation, err
}

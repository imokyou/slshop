package shopline

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/imokyou/slshop/access"
	appopenapi "github.com/imokyou/slshop/app_openapi"
	"github.com/imokyou/slshop/bulk"
	"github.com/imokyou/slshop/customer"
	"github.com/imokyou/slshop/localizations"
	"github.com/imokyou/slshop/market"
	"github.com/imokyou/slshop/marketing"
	"github.com/imokyou/slshop/metafield"
	onlinestore "github.com/imokyou/slshop/online_store"
	"github.com/imokyou/slshop/order"
	paymentsapp "github.com/imokyou/slshop/payments_app"
	"github.com/imokyou/slshop/product"
	saleschannel "github.com/imokyou/slshop/sales_channel"
	shoplinepay "github.com/imokyou/slshop/shopline_payments"
	"github.com/imokyou/slshop/store"
	"github.com/imokyou/slshop/webhook"
)

// Logger is the interface for logging.
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// App represents a Shopline application with its credentials.
type App struct {
	// AppKey is the application key from Shopline Developer Center.
	AppKey string

	// AppSecret is the application secret used for generating/verifying signatures.
	AppSecret string

	// RedirectURL is the OAuth callback URL.
	RedirectURL string

	// Scope defines the access permissions (e.g. "read_products,read_orders").
	Scope string
}

// Client is the Shopline Admin API client.
type Client struct {
	app             App
	handle          string        // Store handle (e.g. "open001" from open001.myshopline.com)
	token           string        // Bearer access token (static, used when tokenManager is nil)
	tokenManager    *TokenManager // automatic token management (overrides token field)
	apiVersion      string
	httpClient      *http.Client
	baseURL         *url.URL
	baseURLOverride string
	maxRetries      int
	log             Logger

	// ========================
	// Sub-package Services
	// ========================

	// Order 大类
	Order             order.Service
	DraftOrder        order.DraftOrderService
	Fulfillment       order.FulfillmentService
	CarrierService    order.CarrierServiceService
	FulfillmentSvcDef order.FulfillmentServiceDefService
	Payment           order.PaymentService
	AbandonedCheckout order.AbandonedCheckoutService
	Subscription      order.SubscriptionService
	Tax               order.TaxService
	Return            order.ReturnService
	OrderArchive      order.ArchiveService
	OrderEdit         order.EditService

	// Customer 大类
	Customer customer.Service

	// Product 大类
	Product          product.Service
	Collection       product.CollectionService
	SmartCollection  product.SmartCollectionService
	ManualCollection product.ManualCollectionService
	Inventory        product.InventoryService

	// Store 大类
	Store store.Service

	// Marketing 大类
	Discount marketing.DiscountService

	// Online Store 大类
	Theme     onlinestore.ThemeService
	Page      onlinestore.PageService
	ScriptTag onlinestore.ScriptTagService

	// Webhook 大类
	Webhook webhook.Service

	// Access 大类
	StorefrontAccessToken access.StorefrontAccessTokenService

	// Market 大类
	Market      market.MarketService
	Location    market.LocationService
	Publication market.PublicationService
	GiftCard    market.GiftCardService

	// Localizations 大类
	Localizations localizations.Service

	// Sales Channel 大类
	SalesChannel saleschannel.Service

	// Metafield 大类
	MetafieldDefinition metafield.DefinitionService
	MetafieldResource   metafield.ResourceService
	MetafieldStore      metafield.StoreService

	// Bulk Operations 大类
	BulkOperation bulk.Service

	// SHOPLINE Payments 大类
	ShoplinePayments shoplinepay.Service

	// Payments APP API 大类
	PaymentsApp paymentsapp.Service

	// App OpenAPI 大类
	SizeChart    appopenapi.SizeChartService
	CDP          appopenapi.CDPService
	VariantImage appopenapi.VariantImageService
}

// NewClient creates a new Shopline API client.
//
// Parameters:
//   - app: Application credentials
//   - handle: Store handle (e.g. "open001" for open001.myshopline.com)
//   - token: Bearer access token
//   - opts: Optional configuration (WithVersion, WithRetry, etc.)
func NewClient(app App, handle, token string, opts ...Option) (*Client, error) {
	baseURL, err := url.Parse(fmt.Sprintf("https://%s.myshopline.com", handle))
	if err != nil {
		return nil, fmt.Errorf("shopline: invalid handle %q: %w", handle, err)
	}

	c := &Client{
		app:        app,
		handle:     handle,
		token:      token,
		apiVersion: DefaultAPIVersion,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		baseURL:    baseURL,
		maxRetries: 0,
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	// Handle base URL override (for testing)
	if c.baseURLOverride != "" {
		overrideURL, err := url.Parse(c.baseURLOverride)
		if err != nil {
			return nil, fmt.Errorf("shopline: invalid base URL %q: %w", c.baseURLOverride, err)
		}
		c.baseURL = overrideURL
	}

	// Initialize all services
	c.Order = order.NewService(c)
	c.DraftOrder = order.NewDraftOrderService(c)
	c.Fulfillment = order.NewFulfillmentService(c)
	c.CarrierService = order.NewCarrierServiceService(c)
	c.FulfillmentSvcDef = order.NewFulfillmentServiceDefService(c)
	c.Payment = order.NewPaymentService(c)
	c.AbandonedCheckout = order.NewAbandonedCheckoutService(c)
	c.Subscription = order.NewSubscriptionService(c)
	c.Tax = order.NewTaxService(c)
	c.Return = order.NewReturnService(c)
	c.OrderArchive = order.NewArchiveService(c)
	c.OrderEdit = order.NewEditService(c)

	c.Customer = customer.NewService(c)

	c.Product = product.NewService(c)
	c.Collection = product.NewCollectionService(c)
	c.SmartCollection = product.NewSmartCollectionService(c)
	c.ManualCollection = product.NewManualCollectionService(c)
	c.Inventory = product.NewInventoryService(c)

	c.Store = store.NewService(c)

	c.Discount = marketing.NewDiscountService(c)

	c.Theme = onlinestore.NewThemeService(c)
	c.Page = onlinestore.NewPageService(c)
	c.ScriptTag = onlinestore.NewScriptTagService(c)

	c.Webhook = webhook.NewService(c)

	c.StorefrontAccessToken = access.NewStorefrontAccessTokenService(c)

	c.Market = market.NewMarketService(c)
	c.Location = market.NewLocationService(c)
	c.Publication = market.NewPublicationService(c)
	c.GiftCard = market.NewGiftCardService(c)

	c.Localizations = localizations.NewService(c)

	c.SalesChannel = saleschannel.NewService(c)

	c.MetafieldDefinition = metafield.NewDefinitionService(c)
	c.MetafieldResource = metafield.NewResourceService(c)
	c.MetafieldStore = metafield.NewStoreService(c)

	c.BulkOperation = bulk.NewService(c)

	c.ShoplinePayments = shoplinepay.NewService(c)

	c.PaymentsApp = paymentsapp.NewService(c)

	c.SizeChart = appopenapi.NewSizeChartService(c)
	c.CDP = appopenapi.NewCDPService(c)
	c.VariantImage = appopenapi.NewVariantImageService(c)

	return c, nil
}

// GetHandle returns the store handle.
func (c *Client) GetHandle() string {
	return c.handle
}

// GetAPIVersion returns the API version in use.
func (c *Client) GetAPIVersion() string {
	return c.apiVersion
}

// GetBaseURL returns the base URL.
func (c *Client) GetBaseURL() *url.URL {
	return c.baseURL
}

// TokenManager returns the TokenManager if one was configured via WithTokenManager.
// Returns nil if no TokenManager is set (i.e., static token mode).
//
// Use this to seed an initial token after OAuth, or to invalidate a token:
//
//	client.TokenManager().SetInitialToken(ctx, accessToken, expireAt, scope)
//	client.TokenManager().InvalidateToken(ctx)
func (c *Client) TokenManager() *TokenManager {
	return c.tokenManager
}

// logDebugf logs a debug message if a logger is set.
func (c *Client) logDebugf(format string, args ...interface{}) {
	if c.log != nil {
		c.log.Debugf(format, args...)
	}
}

// logErrorf logs an error message if a logger is set.
func (c *Client) logErrorf(format string, args ...interface{}) {
	if c.log != nil {
		c.log.Errorf(format, args...)
	}
}

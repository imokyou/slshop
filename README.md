# slshop — Shopline Admin REST API SDK for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/imokyou/slshop.svg)](https://pkg.go.dev/github.com/imokyou/slshop)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

A comprehensive, modular Go SDK for [Shopline](https://www.shopline.com/) Admin REST API, covering **all** documented API categories.

## Features

- 🚀 **Full API Coverage** — 16 sub-packages covering every Shopline Admin API category
- 📦 **Modular Design** — Import only the packages you need
- 🔄 **Auto Retry** — Built-in configurable retry with exponential backoff
- 🔐 **OAuth Support** — Complete OAuth2 authorization flow
- 🏷️ **API Versioning** — Easily switch between API versions
- 🧪 **Test Friendly** — Interface-based design for easy mocking

## Installation

```bash
go get github.com/imokyou/slshop
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    shopline "github.com/imokyou/slshop"
    "github.com/imokyou/slshop/order"
)

func main() {
    app := shopline.App{
        AppKey:    "your-app-key",
        AppSecret: "your-app-secret",
    }

    client, err := shopline.NewClient(app, "your-store-handle", "your-access-token",
        shopline.WithVersion(shopline.APIVersion20251201),
        shopline.WithRetry(3),
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // List orders
    orders, err := client.Order.List(ctx, &order.ListOptions{Status: "any"})
    if err != nil {
        log.Fatal(err)
    }
    for _, o := range orders {
        fmt.Printf("Order %s: %s %s\n", o.Name, o.Currency, o.TotalPrice)
    }
}
```

## Package Structure

```
slshop/
├── core/               # Requester interface & shared types
├── access/             # Storefront Access Token
├── order/              # Order, DraftOrder, Fulfillment, Payment, Return
├── customer/           # Customer CRUD, Groups, Addresses, Social Login
├── product/            # Product, Collection, Inventory
├── store/              # Store Info, Staff, Operation Logs, Subscription
├── marketing/          # PriceRule, Discount Code
├── online_store/       # Theme, Page, ScriptTag
├── webhook/            # Webhook CRUD
├── market/             # Market, Location, Publication, GiftCard
├── localizations/      # Languages & Translations
├── sales_channel/      # Product & Collection Listings
├── metafield/          # Metafield Definitions, Resource & Store Metafields
├── bulk/               # Bulk Query & Mutation Operations
├── shopline_payments/  # Balance, Payouts, Billing, Transactions
├── payments_app/       # Payments APP Notifications
└── app_openapi/        # Size Chart, CDP, Variant Images
```

## Available Services

| Service | Access via `client.` | Endpoints |
|---------|---------------------|-----------|
| Order | `Order` | List, Get, Create, Update, Delete, Close, Open, Cancel, Count |
| Draft Order | `DraftOrder` | Create, Update, Get, Delete, Complete, Count, SendInvoice |
| Fulfillment | `Fulfillment` | List, Create, Cancel, UpdateTracking, and more |
| Payment | `Payment` | CreateSlip, GetSlip, ListTransactions, ListPayments |
| Customer | `Customer` | List, Get, Create, Update, Delete, Search, Groups, Addresses |
| Product | `Product` | List, Get, Create, Update, Delete, Count |
| Collection | `Collection` | List, Get, Create, Update, Delete, Count |
| Store | `Store` | GetShop, GetCurrency, ListStaff, ListOperationLogs |
| Discount | `Discount` | PriceRule CRUD, DiscountCode CRUD |
| Theme | `Theme` | List, Get |
| Page | `Page` | List, Get, Create, Update, Delete |
| Webhook | `Webhook` | List, Get, Create, Update, Delete, Count |
| Market | `Market` | List, Get |
| Localizations | `Localizations` | Languages, Translations |
| Sales Channel | `SalesChannel` | Product/Collection Listings |
| Metafield Def | `MetafieldDefinition` | Create, Update, List, Get, Delete, Count |
| Metafield | `MetafieldStore` | Create, Update, List, Get, Delete, Count |
| Bulk Operations | `BulkOperation` | GetCurrent, CreateQuery, CreateMutation, Cancel |
| SHOPLINE Payments | `ShoplinePayments` | Balance, Payouts, Billing, Transactions |
| Payments App | `PaymentsApp` | Activation, Payment, Refund, Device Binding |
| Size Chart | `SizeChart` | Batch query/create/delete product sizes |
| CDP | `CDP` | Report events, Report identity |
| Variant Images | `VariantImage` | Query, Batch update |

## Configuration Options

```go
// API version
shopline.WithVersion(shopline.APIVersion20251201)

// Retry with exponential backoff
shopline.WithRetry(3)

// Custom HTTP client
shopline.WithHTTPClient(customClient)

// Custom logger
shopline.WithLogger(myLogger)

// Base URL override (for testing)
shopline.WithBaseURL("http://localhost:8080")
```

## OAuth Authorization

```go
app := shopline.App{
    AppKey:      "your-app-key",
    AppSecret:   "your-app-secret",
    RedirectURL: "https://your-app.com/callback",
    Scope:       "read_products,read_orders",
}

// Generate authorization URL
authURL := shopline.AuthorizeURL(app, "your-store-handle", "random-nonce")

// Exchange code for token
token, err := shopline.GetAccessToken(app, "your-store-handle", code)
```

## License

This project is licensed under the [GNU General Public License v3.0](LICENSE).

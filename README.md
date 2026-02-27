# slshop — Shopline Admin REST API Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/imokyou/slshop.svg)](https://pkg.go.dev/github.com/imokyou/slshop)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

一个全面、模块化的 [Shopline](https://www.shopline.com/) Admin REST API Go SDK，覆盖**所有**已文档化的 API 类别。

## 特性

- 🚀 **全 API 覆盖** — 16 个子包，涵盖 Shopline Admin API 的每一个类别
- 📦 **模块化设计** — 按需引入，只导入你需要的包
- 🔄 **自动重试** — 内置指数退避 + 随机抖动（jitter）重试机制
- 🔐 **OAuth 支持** — 完整的 OAuth2 授权流程 + Token 自动管理
- 🏷️ **API 版本管理** — 轻松切换 API 版本
- 🔑 **Token 自动管理** — 持久化存储 + 并发安全无感刷新
- 🧪 **测试友好** — 基于接口设计，方便 Mock；40+ 测试用例含 race 检测
- 🛡️ **生产级品质** — 响应大小限制、Context 取消支持、连接池优化

## 安装

```bash
go get github.com/imokyou/slshop
```

零外部依赖，仅使用 Go 标准库。

## 快速开始

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

    // 获取订单列表
    orders, err := client.Order.List(ctx, &order.ListOptions{Status: "any"})
    if err != nil {
        log.Fatal(err)
    }
    for _, o := range orders {
        fmt.Printf("订单 %s: %s %s\n", o.Name, o.Currency, o.TotalPrice)
    }
}
```

## Token 自动管理

SDK 提供开箱即用的 Token 生命周期管理，支持**持久化**和**并发安全无感刷新**：

```go
// 创建文件存储（开发环境）或实现 TokenStore 接口对接 Redis（生产环境）
store := shopline.NewFileTokenStore("./tokens")

client, _ := shopline.NewClient(app, "myshop", "",
    shopline.WithTokenManager(store),
    shopline.WithRetry(3),
)

// 首次 OAuth 后种入初始 Token
client.TokenManager().SetInitialToken(ctx, accessToken, expireAt, scope)

// 之后所有 API 调用自动管理 Token，无需关心刷新！
products, _ := client.Product.List(ctx, nil)
```

**并发安全**：当 Token 即将过期时，只有一个 goroutine 会执行刷新，其他 goroutine 等待结果（singleflight 模式）。

## 示例

| 示例 | 说明 | 路径 |
|------|------|------|
| 基础用法 | 商品/订单/客户 CRUD | [examples/basic/](examples/basic/) |
| OAuth 流程 | 完整授权 + 回调 + Token 获取 | [examples/oauth/](examples/oauth/) |
| Token 管理 | 持久化 + 自动刷新 + 并发演示 | [examples/token_manager/](examples/token_manager/) |
| Webhook 处理 | 签名验证 + 事件路由 | [examples/webhook/](examples/webhook/) |

## 文档

- 📖 [**使用指南**](docs/guide.md) — 完整使用手册（OAuth、API 调用、TokenManager、Webhook、错误处理、最佳实践）
- ❓ [**FAQ**](docs/faq.md) — 常见问题解答（Token、重试、并发、调试等）

## 包结构

```
slshop/
├── core/               # 核心接口与共享类型
├── access/             # 店面访问令牌
├── order/              # 订单、草稿订单、履约、支付、退货
├── customer/           # 客户管理、分组、地址、社交登录
├── product/            # 商品、集合、库存
├── store/              # 店铺信息、员工、操作日志、订阅
├── marketing/          # 价格规则、折扣码
├── online_store/       # 主题、页面、脚本标签
├── webhook/            # Webhook 管理
├── market/             # 市场、位置、发布、礼品卡
├── localizations/      # 多语言与翻译
├── sales_channel/      # 商品与集合上架
├── metafield/          # 元字段定义、资源与店铺元字段
├── bulk/               # 批量查询与批量变更操作
├── shopline_payments/  # 余额、提现、账单、交易
├── payments_app/       # 支付应用通知
├── app_openapi/        # 尺码表、CDP、变体图片
├── docs/               # 使用指南、FAQ 文档
└── examples/           # 示例代码
```

## 可用服务

| 服务 | 通过 `client.` 访问 | 接口方法 |
|------|---------------------|----------|
| 订单 | `Order` | List, Get, Create, Update, Delete, Close, Open, Cancel, Count |
| 草稿订单 | `DraftOrder` | Create, Update, Get, Delete, Complete, Count, SendInvoice |
| 履约 | `Fulfillment` | List, Create, Cancel, UpdateTracking 等 |
| 支付 | `Payment` | CreateSlip, GetSlip, ListTransactions, ListPayments |
| 客户 | `Customer` | List, Get, Create, Update, Delete, Search, Groups, Addresses |
| 商品 | `Product` | List, Get, Create, Update, Delete, Count |
| 集合 | `Collection` | List, Get, Create, Update, Delete, Count |
| 店铺 | `Store` | GetShop, GetCurrency, ListStaff, ListOperationLogs |
| 折扣 | `Discount` | PriceRule CRUD, DiscountCode CRUD |
| 主题 | `Theme` | List, Get |
| 页面 | `Page` | List, Get, Create, Update, Delete |
| Webhook | `Webhook` | List, Get, Create, Update, Delete, Count |
| 市场 | `Market` | List, Get |
| 多语言 | `Localizations` | Languages, Translations |
| 销售渠道 | `SalesChannel` | 商品/集合上架 |
| 元字段定义 | `MetafieldDefinition` | Create, Update, List, Get, Delete, Count |
| 元字段 | `MetafieldStore` | Create, Update, List, Get, Delete, Count |
| 批量操作 | `BulkOperation` | GetCurrent, CreateQuery, CreateMutation, Cancel |
| Shopline 支付 | `ShoplinePayments` | Balance, Payouts, Billing, Transactions |
| 支付应用 | `PaymentsApp` | Activation, Payment, Refund, Device Binding |
| 尺码表 | `SizeChart` | 批量查询/创建/删除商品尺码 |
| CDP | `CDP` | 上报事件、上报身份 |
| 变体图片 | `VariantImage` | 查询、批量更新 |

## 配置选项

```go
// API 版本
shopline.WithVersion(shopline.APIVersion20251201)

// 指数退避重试（推荐生产环境设 2-3）
shopline.WithRetry(3)

// 自定义 HTTP 客户端
shopline.WithHTTPClient(customClient)

// 自定义日志记录器
shopline.WithLogger(myLogger)

// Token 自动管理
shopline.WithTokenManager(store)

// 自定义 Base URL（用于测试）
shopline.WithBaseURL("http://localhost:8080")
```

## OAuth 授权

```go
app := shopline.App{
    AppKey:      "your-app-key",
    AppSecret:   "your-app-secret",
    RedirectURL: "https://your-app.com/callback",
    Scope:       "read_products,read_orders",
}

// 生成授权链接
authURL := app.AuthorizeURL("store-handle", "random-nonce")

// 验证回调签名
valid := app.VerifySignature(r.URL.Query())

// 用授权码换取访问令牌
token, err := app.GetAccessToken(ctx, "store-handle", code)

// 刷新令牌
newToken, err := app.RefreshAccessToken(ctx, "store-handle")
```

## 开源协议

本项目基于 [GNU General Public License v3.0](LICENSE) 开源。

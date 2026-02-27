# slshop 使用指南

本文档是 [slshop](https://github.com/imokyou/slshop) SDK 的完整使用手册，面向需要集成 Shopline Admin API 的 Go 开发者。

## 目录

- [安装](#安装)
- [核心概念](#核心概念)
- [OAuth 授权流程](#oauth-授权流程)
- [基本 API 调用](#基本-api-调用)
- [Token 自动管理](#token-自动管理)
- [Webhook 处理](#webhook-处理)
- [错误处理](#错误处理)
- [高级配置](#高级配置)
- [最佳实践](#最佳实践)

---

## 安装

```bash
go get github.com/imokyou/slshop
```

要求 Go 1.21 或更高版本。零外部依赖，仅使用 Go 标准库。

---

## 核心概念

### App — 应用凭证

```go
app := shopline.App{
    AppKey:      "your-app-key",      // 从 Shopline 开发者中心获取
    AppSecret:   "your-app-secret",   // 用于签名和 Webhook 验证
    RedirectURL: "https://your.app/callback", // OAuth 回调地址
    Scope:       "read_products,read_orders",  // 请求的权限范围
}
```

### Client — API 客户端

```go
client, err := shopline.NewClient(app, "store-handle", "access-token",
    shopline.WithVersion(shopline.APIVersion20251201),  // API 版本
    shopline.WithRetry(3),                              // 自动重试次数
    shopline.WithLogger(myLogger),                      // 日志记录
)
```

Client 通过 `client.Product`、`client.Order` 等字段访问各个子服务。

### API 版本

SDK 内置所有 Shopline API 版本常量：

```go
shopline.APIVersion20251201  // 默认 (Stable)
shopline.APIVersion20260301  // Release Candidate
shopline.APIVersion20260601  // Unstable
```

建议生产环境使用 Stable 版本。

---

## OAuth 授权流程

Shopline 使用标准 OAuth 2.0 流程。完整示例见 [examples/oauth/main.go](../examples/oauth/main.go)。

### Step 1: 生成授权链接

```go
authURL := app.AuthorizeURL("store-handle", "random-nonce-for-csrf")
// 引导商家打开此 URL 授权你的应用
```

### Step 2: 接收回调

商家授权后，Shopline 会跳转到你的 `RedirectURL`，携带 `code` 和 `sign` 参数。

```go
// 验证回调签名（防止伪造）
if !app.VerifySignature(r.URL.Query()) {
    http.Error(w, "Invalid signature", 403)
    return
}
code := r.URL.Query().Get("code")
```

### Step 3: 换取 Access Token

```go
tokenResp, err := app.GetAccessToken(ctx, "store-handle", code)
if err != nil {
    log.Fatal(err)
}
fmt.Println(tokenResp.Data.AccessToken) // 拿到 token
fmt.Println(tokenResp.Data.ExpireTime)  // 过期时间
```

### Step 4: 刷新 Token

Shopline 的 Access Token 有效期约 10 小时，到期前需刷新：

```go
newToken, err := app.RefreshAccessToken(ctx, "store-handle")
```

> **推荐**：使用 [TokenManager](#token-自动管理) 自动处理刷新，无需手动刷新。

---

## 基本 API 调用

完整示例见 [examples/basic/main.go](../examples/basic/main.go)。

### 商品

```go
// 列表
products, err := client.Product.List(ctx, nil)

// 获取单个
product, err := client.Product.Get(ctx, 12345)

// 创建
newProduct, err := client.Product.Create(ctx, product.Product{
    Title: "新商品",
})

// 更新
updated, err := client.Product.Update(ctx, 12345, product.Product{
    Title: "更新后的商品名",
})

// 删除
err := client.Product.Delete(ctx, 12345)

// 计数
count, err := client.Product.Count(ctx, nil)
```

### 订单

```go
// 带筛选条件列表
orders, err := client.Order.List(ctx, &order.ListOptions{
    Status: "open",
})

// 获取
order, err := client.Order.Get(ctx, 67890)

// 关闭订单
err := client.Order.Close(ctx, 67890)
```

### 客户

```go
customers, err := client.Customer.List(ctx, &customer.ListOptions{})
customer, err := client.Customer.Get(ctx, 11111)
```

### 店铺信息

```go
shop, err := client.Store.GetShop(ctx)
fmt.Printf("店铺: %s (%s)\n", shop.Name, shop.Domain)
```

---

## Token 自动管理

完整示例见 [examples/token_manager/main.go](../examples/token_manager/main.go)。

### 为什么需要 TokenManager

| 场景 | 手动管理 | TokenManager |
|------|---------|-------------|
| 进程重启 | Token 丢失，需重新 OAuth | 从持久化存储自动恢复 |
| Token 过期 | 请求失败，需手动刷新 | 过期前 5 分钟自动刷新 |
| 高并发 | 多个 goroutine 同时刷新造成竞争 | Singleflight 保证只刷新一次 |

### 基本用法

```go
// 1. 创建 TokenStore（文件存储）
store := shopline.NewFileTokenStore("./tokens")

// 2. 创建 Client 时传入 TokenManager
client, _ := shopline.NewClient(app, "myshop", "",
    shopline.WithTokenManager(store),
)

// 3. 首次使用前需要种入初始 Token（从 OAuth 获得）
client.TokenManager().SetInitialToken(ctx, accessToken, expireAt, scope)

// 4. 之后所有 API 调用自动管理 Token
products, _ := client.Product.List(ctx, nil)  // 自动使用有效 token
orders, _ := client.Order.List(ctx, nil)       // 快过期时自动刷新
```

### 自定义 TokenStore（Redis 示例）

```go
type RedisTokenStore struct {
    client *redis.Client
}

func (s *RedisTokenStore) Get(ctx context.Context, key string) (*shopline.ManagedToken, error) {
    data, err := s.client.Get(ctx, "shopline:token:"+key).Bytes()
    if err == redis.Nil {
        return nil, nil  // 未找到不是错误
    }
    if err != nil {
        return nil, err
    }
    var token shopline.ManagedToken
    json.Unmarshal(data, &token)
    return &token, nil
}

func (s *RedisTokenStore) Set(ctx context.Context, key string, token *shopline.ManagedToken) error {
    data, _ := json.Marshal(token)
    ttl := time.Until(token.ExpireAt)
    return s.client.Set(ctx, "shopline:token:"+key, data, ttl).Err()
}

func (s *RedisTokenStore) Delete(ctx context.Context, key string) error {
    return s.client.Del(ctx, "shopline:token:"+key).Err()
}

// 使用自定义 store
store := &RedisTokenStore{client: rdb}
client, _ := shopline.NewClient(app, "myshop", "",
    shopline.WithTokenManager(store),
)
```

### 配置选项

```go
shopline.WithTokenManager(store,
    shopline.WithRefreshBuffer(10 * time.Minute), // 提前 10 分钟刷新（默认 5 分钟）
    shopline.WithTokenManagerLogger(myLogger),     // 独立日志
)
```

---

## Webhook 处理

完整示例见 [examples/webhook/main.go](../examples/webhook/main.go)。

### 验证签名

```go
func webhookHandler(w http.ResponseWriter, r *http.Request) {
    // 验证 HMAC-SHA256 签名（body 会自动恢复，后续可继续读取）
    if !app.VerifyWebhookRequest(r) {
        http.Error(w, "Unauthorized", 401)
        return
    }

    // 读取 body（验证后仍可用）
    body, _ := io.ReadAll(r.Body)

    // 根据 topic 处理
    topic := r.Header.Get("X-Shopline-Topic")
    switch topic {
    case "orders/create":
        // 处理新订单...
    }

    w.WriteHeader(200)
}
```

---

## 错误处理

### ResponseError

所有 API 错误都会返回 `*shopline.ResponseError`：

```go
products, err := client.Product.List(ctx, nil)
if err != nil {
    var respErr *shopline.ResponseError
    if errors.As(err, &respErr) {
        fmt.Printf("HTTP %d: %s (traceId: %s)\n",
            respErr.Status, respErr.Message, respErr.TraceID)
        fmt.Printf("详细错误: %s\n", respErr.GetErrors())
    }
}
```

### RateLimitError

429 限流错误会返回 `*shopline.RateLimitError`，包含 `RetryAfter` 字段：

```go
var rlErr *shopline.RateLimitError
if errors.As(err, &rlErr) {
    fmt.Printf("限流！请 %s 后重试\n", rlErr.RetryAfter)
}
```

> 如果配置了 `WithRetry(n)`，SDK 会自动在限流时使用指数退避重试，大多数情况无需手动处理。

---

## 高级配置

### 自定义 HTTP Client

```go
httpClient := &http.Client{
    Timeout: 60 * time.Second,
    Transport: &http.Transport{
        MaxIdleConnsPerHost: 20,
        TLSHandshakeTimeout: 10 * time.Second,
    },
}
client, _ := shopline.NewClient(app, handle, token,
    shopline.WithHTTPClient(httpClient),
)
```

### 自定义日志

实现 `shopline.Logger` 接口：

```go
type Logger interface {
    Debugf(format string, args ...interface{})
    Infof(format string, args ...interface{})
    Errorf(format string, args ...interface{})
}
```

可以用 `zap`、`logrus`、`slog` 等包装：

```go
type ZapLogger struct { log *zap.SugaredLogger }
func (l *ZapLogger) Debugf(f string, a ...interface{}) { l.log.Debugf(f, a...) }
func (l *ZapLogger) Infof(f string, a ...interface{})  { l.log.Infof(f, a...) }
func (l *ZapLogger) Errorf(f string, a ...interface{}) { l.log.Errorf(f, a...) }
```

---

## 最佳实践

### 1. 始终使用 Context

```go
// ✅ 设置超时
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
products, err := client.Product.List(ctx, nil)

// ❌ 不要用 context.Background() 到处传
```

### 2. 复用 Client

```go
// ✅ 全局创建一个 Client，在所有 handler 中复用
var client *shopline.Client

func init() {
    client, _ = shopline.NewClient(app, handle, token)
}

// ❌ 不要每次请求都创建新 Client
```

### 3. 配置重试

```go
// ✅ 生产环境建议配置 2-3 次重试
client, _ := shopline.NewClient(app, handle, token,
    shopline.WithRetry(3),
)
```

### 4. 使用 TokenManager

```go
// ✅ 使用 TokenManager 自动管理 Token
client, _ := shopline.NewClient(app, handle, "",
    shopline.WithTokenManager(store),
)

// ❌ 不要手动管理 Token 刷新
```

### 5. Webhook 签名验证

```go
// ✅ 始终验证 Webhook 签名
if !app.VerifyWebhookRequest(r) {
    http.Error(w, "Unauthorized", 401)
    return
}

// ❌ 不要跳过签名验证
```

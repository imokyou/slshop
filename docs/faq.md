# FAQ — 常见问题

## 基础问题

### Q: slshop 支持哪些 Go 版本？

要求 **Go 1.21+**。SDK 使用纯标准库，零外部依赖。

### Q: 如何获取 AppKey 和 AppSecret？

1. 登录 [Shopline 开发者中心](https://developers.shopline.com/)
2. 创建应用 → 获取 App Key 和 App Secret
3. 配置回调 URL 和权限范围

### Q: 支持哪些 API 版本？

SDK 内置所有版本常量，从 `v20210901` 到 `v20260601`。默认使用 `v20251201`（Stable）。

```go
shopline.WithVersion(shopline.APIVersion20251201)
```

---

## Token 相关

### Q: Access Token 有效期多长？

Shopline Access Token 默认有效期约 **10 小时**。到期前需要调用 `RefreshAccessToken` 刷新。

### Q: 推荐的 Token 管理方案是什么？

使用 `TokenManager`，可以：
- 自动在过期前 5 分钟刷新
- 高并发安全（singleflight 模式）  
- 支持持久化（进程重启不丢失）

```go
store := shopline.NewFileTokenStore("./tokens")
client, _ := shopline.NewClient(app, handle, "",
    shopline.WithTokenManager(store),
)
```

### Q: FileTokenStore 适合生产环境吗？

`FileTokenStore` 适合 **单进程本地开发**。生产环境建议实现 `TokenStore` 接口，使用 Redis 或数据库。详见 [使用指南 — Token 自动管理](guide.md#token-自动管理)。

### Q: 多个进程/实例如何共享 Token？

实现 `TokenStore` 接口，使用 Redis 等共享存储：

```go
type RedisTokenStore struct { client *redis.Client }
// 实现 Get/Set/Delete ...
```

每个实例各自 `WithTokenManager(redisStore)`，token 通过 Redis 共享。

### Q: Token 刷新失败怎么办？

`TokenManager.GetToken()` 会返回错误。建议：
1. 配置日志 (`WithTokenManagerLogger`) 监控刷新状态
2. 在上层添加重试逻辑
3. 如果持续失败，可能需要重新走 OAuth 流程

---

## 重试与限流

### Q: SDK 的重试机制是怎样的？

配置 `WithRetry(n)` 后，SDK 会对以下响应自动重试：
- **HTTP 429** (Too Many Requests) — 限流
- **HTTP 503** (Service Unavailable) — 服务暂不可用
- **网络错误** — 连接超时、DNS 失败等

重试使用 **指数退避 + 随机抖动**（jitter），避免重试风暴。

### Q: 退避策略是什么？

`base × 2^attempt`，上限 30 秒，外加 ±25% 随机抖动：

| 尝试次数 | 基础退避 | 实际范围（含抖动） |
|---------|---------|-----------------|
| 第 1 次重试 | 1s | 0.75s ~ 1.25s |
| 第 2 次重试 | 2s | 1.5s ~ 2.5s |
| 第 3 次重试 | 4s | 3s ~ 5s |

### Q: 如何处理 Retry-After 头？

SDK 自动解析 `Retry-After` 响应头（支持秒数和 HTTP-date 两种格式）。如果服务端返回了此字段，SDK 会优先使用它作为退避时间，否则使用指数退避。

### Q: 重试时 Context 取消会怎样？

退避等待期间会监听 `ctx.Done()`。一旦 context 被取消或超时，SDK 立即返回，**不会** goroutine 泄漏。

---

## Webhook

### Q: 如何验证 Webhook 签名？

```go
if !app.VerifyWebhookRequest(r) {
    http.Error(w, "Unauthorized", 401)
    return
}
```

### Q: 验证后还能读取 Body 吗？

**可以**。`VerifyWebhookRequest` 读取 body 后会自动恢复 `r.Body`，后续 handler 可正常读取。

### Q: Webhook 应该如何快速响应？

Shopline 要求在 5 秒内返回 2xx 响应，否则会重试。建议：

```go
// 快速响应
w.WriteHeader(200)

// 异步处理耗时业务逻辑
go processWebhook(payload)
```

---

## 性能与并发

### Q: Client 是线程安全的吗？

**是的**。`Client` 和 `TokenManager` 都设计为并发安全，可以在多个 goroutine 中共享同一个 Client 实例。

### Q: 应该如何管理 Client 生命周期？

- **全局创建一个 Client**，在所有 handler 中复用
- 不要每次请求都创建新 Client（会浪费连接池）
- SDK 默认配置的 HTTP Transport 已针对高并发优化（`MaxIdleConnsPerHost=10`）

### Q: 响应 Body 有大小限制吗？

SDK 限制响应体最大 **10MB**（`io.LimitReader`），防止异常响应导致内存溢出。

---

## 调试

### Q: 如何启用调试日志？

```go
type debugLogger struct{}
func (l *debugLogger) Debugf(f string, a ...interface{}) { log.Printf("[DEBUG] "+f, a...) }
func (l *debugLogger) Infof(f string, a ...interface{})  { log.Printf("[INFO]  "+f, a...) }
func (l *debugLogger) Errorf(f string, a ...interface{}) { log.Printf("[ERROR] "+f, a...) }

client, _ := shopline.NewClient(app, handle, token,
    shopline.WithLogger(&debugLogger{}),
)
```

### Q: 如何获取请求的 traceId？

所有 API 错误响应都包含 `TraceID`：

```go
var respErr *shopline.ResponseError
if errors.As(err, &respErr) {
    fmt.Println("traceId:", respErr.TraceID)
}
```

将 traceId 提供给 Shopline 技术支持可快速定位问题。

### Q: 如何模拟 API 进行测试？

使用 `httptest.Server` + `WithBaseURL`：

```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]interface{}{
        "products": []map[string]interface{}{
            {"id": 1, "title": "Test"},
        },
    })
}))
defer server.Close()

client, _ := shopline.NewClient(app, "test", "token",
    shopline.WithBaseURL(server.URL),
)
// client 的所有请求都会发到 mock server
```

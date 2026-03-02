# 生产环境部署指南

本文档介绍在生产环境中使用 `slshop` SDK 的最佳实践，包括 Token 持久化存储、多副本部署、监控配置和安全加固。

---

## 一、Token 存储：生产级 Redis 实现

SDK 内置的 `FileTokenStore` 仅适合单进程开发环境。生产环境中多副本（容器/K8s）共享状态，必须使用 Redis 等分布式存储。

### 1.1 完整 Redis TokenStore 参考实现

```go
package tokenstore

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
    shopline "github.com/imokyou/slshop"
)

// RedisTokenStore implements shopline.TokenStore backed by Redis.
// Token JSON 存储在 Redis 中，key 格式为 "slshop:{handle}:{appkey}"
type RedisTokenStore struct {
    client *redis.Client
    ttl    time.Duration // Redis key TTL，建议设为 Token 有效期 + 1小时
}

// NewRedisTokenStore 创建 Redis Token 存储。
//   - addr: Redis 地址，如 "localhost:6379"
//   - password: Redis 密码（无密码传 ""）
//   - db: Redis DB 编号（通常为 0）
func NewRedisTokenStore(addr, password string, db int) *RedisTokenStore {
    rdb := redis.NewClient(&redis.Options{
        Addr:         addr,
        Password:     password,
        DB:           db,
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
        PoolSize:     10,
    })
    return &RedisTokenStore{
        client: rdb,
        ttl:    12 * time.Hour, // Shopline Token 有效期 10 小时，预留 2 小时缓冲
    }
}

func (s *RedisTokenStore) Get(ctx context.Context, key string) (*shopline.ManagedToken, error) {
    data, err := s.client.Get(ctx, s.redisKey(key)).Bytes()
    if err == redis.Nil {
        return nil, nil // key 不存在
    }
    if err != nil {
        return nil, fmt.Errorf("redis: failed to get token: %w", err)
    }

    var token shopline.ManagedToken
    if err := json.Unmarshal(data, &token); err != nil {
        return nil, fmt.Errorf("redis: failed to unmarshal token: %w", err)
    }
    return &token, nil
}

func (s *RedisTokenStore) Set(ctx context.Context, key string, token *shopline.ManagedToken) error {
    data, err := json.Marshal(token)
    if err != nil {
        return fmt.Errorf("redis: failed to marshal token: %w", err)
    }

    // 用 Token 实际过期时间计算 TTL，避免存储永不过期的 key
    ttl := s.ttl
    if !token.ExpireAt.IsZero() {
        remaining := time.Until(token.ExpireAt) + 2*time.Hour
        if remaining > 0 {
            ttl = remaining
        }
    }

    return s.client.Set(ctx, s.redisKey(key), data, ttl).Err()
}

func (s *RedisTokenStore) Delete(ctx context.Context, key string) error {
    return s.client.Del(ctx, s.redisKey(key)).Err()
}

func (s *RedisTokenStore) redisKey(key string) string {
    return "slshop:" + key
}
```

### 1.2 使用方式

```go
import (
    shopline "github.com/imokyou/slshop"
    "yourapp/tokenstore"
)

func NewShoplineClient(appKey, appSecret, storeHandle string) (*shopline.Client, error) {
    app := shopline.App{
        AppKey:      appKey,
        AppSecret:   appSecret,
        RedirectURL: "https://your-app.com/oauth/callback",
        Scope:       "read_products,read_orders,write_inventory",
    }

    store := tokenstore.NewRedisTokenStore(
        os.Getenv("REDIS_ADDR"),
        os.Getenv("REDIS_PASSWORD"),
        0,
    )

    return shopline.NewClient(app, storeHandle, "",
        shopline.WithTokenManager(store),
        shopline.WithRetry(3),
        shopline.WithCircuitBreaker(5, 30*time.Second),
        shopline.WithLogger(logger),
    )
}
```

---

## 二、Docker / Kubernetes 多副本部署

### 2.1 关键原则

> **所有副本必须共享同一个 TokenStore（Redis）。**
> 如果每个 Pod 各自用 FileTokenStore，Token 刷新会被多个 Pod 同时触发，导致 Shopline 平台产生多余的 Token 刷新请求，甚至导致旧 Token 被提前失效。

### 2.2 环境变量最佳实践

**不要** 硬编码任何密钥。使用环境变量或 K8s Secrets：

```yaml
# k8s-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: shopline-app
spec:
  replicas: 3
  template:
    spec:
      containers:
        - name: app
          image: your-app:latest
          env:
            - name: SHOPLINE_APP_KEY
              valueFrom:
                secretKeyRef:
                  name: shopline-secrets
                  key: app-key
            - name: SHOPLINE_APP_SECRET
              valueFrom:
                secretKeyRef:
                  name: shopline-secrets
                  key: app-secret
            - name: REDIS_ADDR
              value: "redis-service:6379"
            - name: REDIS_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: redis-secrets
                  key: password
```

```go
// 从环境变量读取配置
app := shopline.App{
    AppKey:    os.Getenv("SHOPLINE_APP_KEY"),
    AppSecret: os.Getenv("SHOPLINE_APP_SECRET"),
}
```

### 2.3 健康检查推荐

```go
// 在 /healthz 端点添加 Shopline API 连通性检查
func healthHandler(client *shopline.Client) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
        defer cancel()

        _, err := client.Store.GetShop(ctx)
        if err != nil {
            http.Error(w, "shopline api unreachable", http.StatusServiceUnavailable)
            return
        }
        w.WriteHeader(http.StatusOK)
    }
}
```

---

## 三、断路器与限速配置

### 3.1 推荐生产配置

```go
client, _ := shopline.NewClient(app, handle, "",
    shopline.WithRetry(3),                              // 最多重试 3 次
    shopline.WithCircuitBreaker(5, 30*time.Second),    // 5次失败后熔断 30 秒
    shopline.WithTimeout(30*time.Second),               // 普通 API 30s 超时
)

// 批量操作使用独立的长超时 Client
bulkClient, _ := shopline.NewClient(app, handle, "",
    shopline.WithTimeout(10*time.Minute),               // Bulk 操作允许更长时间
    shopline.WithRetry(1),                              // Bulk 操作重试代价高，只重试一次
)
```

### 3.2 限速最佳实践

Shopline API 对请求频率有限制（通常每秒 2 个请求/店铺）。SDK 提供指数退避，但建议在应用层增加主动限速：

```go
import "golang.org/x/time/rate"

// 每店铺限速 2 req/s，允许瞬时 burst 5 个请求
limiter := rate.NewLimiter(rate.Limit(2), 5)

// 在每次 API 调用前等待许可
if err := limiter.Wait(ctx); err != nil {
    return err
}
products, err := client.Product.List(ctx, nil)
```

---

## 四、多租户架构

每个商家（店铺）应该有独立的 Client 实例，通过连接池管理：

```go
// ClientPool 管理多商家 Client 实例
type ClientPool struct {
    mu      sync.RWMutex
    clients map[string]*shopline.Client // key: storeHandle
    app     shopline.App
    store   shopline.TokenStore
}

func (p *ClientPool) Get(handle string) (*shopline.Client, error) {
    p.mu.RLock()
    if c, ok := p.clients[handle]; ok {
        p.mu.RUnlock()
        return c, nil
    }
    p.mu.RUnlock()

    p.mu.Lock()
    defer p.mu.Unlock()

    // Double-check after acquiring write lock
    if c, ok := p.clients[handle]; ok {
        return c, nil
    }

    c, err := shopline.NewClient(p.app, handle, "",
        shopline.WithTokenManager(p.store),
        shopline.WithRetry(3),
        shopline.WithCircuitBreaker(5, 30*time.Second),
    )
    if err != nil {
        return nil, err
    }

    p.clients[handle] = c
    return c, nil
}
```

**Token 隔离**：`TokenManager` 使用 `"{handle}:{appKey}"` 作为 Redis key，不同商家的 Token 天然隔离，无需额外处理。

---

## 五、监控与可观测性

### 5.1 结构化日志接入

实现 `shopline.Logger` 接口接入 `zap` / `zerolog`：

```go
// ZapLogger adapts uber-go/zap to shopline.Logger
type ZapLogger struct{ *zap.SugaredLogger }

func (l *ZapLogger) Debugf(format string, args ...interface{}) { l.SugaredLogger.Debugf(format, args...) }
func (l *ZapLogger) Infof(format string, args ...interface{})  { l.SugaredLogger.Infof(format, args...) }
func (l *ZapLogger) Errorf(format string, args ...interface{}) { l.SugaredLogger.Errorf(format, args...) }

// 使用
logger, _ := zap.NewProduction()
client, _ := shopline.NewClient(app, handle, "",
    shopline.WithLogger(&ZapLogger{logger.Sugar()}),
)
```

### 5.2 推荐 Prometheus 指标（参考结构）

在应用层包装 Service 调用时可收集以下指标：

| 指标名 | 类型 | 标签 | 说明 |
|--------|------|------|------|
| `shopline_api_requests_total` | Counter | `method`, `service`, `status` | API 请求总数 |
| `shopline_api_request_duration_seconds` | Histogram | `method`, `service` | 请求耗时分布 |
| `shopline_api_retries_total` | Counter | `method` | 重试次数 |
| `shopline_token_refresh_total` | Counter | `handle`, `result` | Token 刷新次数 |
| `shopline_circuit_breaker_state` | Gauge | `handle`, `state` | 断路器状态（0=closed,1=open,2=half-open） |

---

## 六、安全加固清单

| 项目 | 建议 |
|------|------|
| AppSecret 存储 | 使用 K8s Secret 或 Vault，禁止写入代码或 .env 文件 |
| Token 存储 | Redis 启用 AUTH 密码 + TLS（`rediss://`） |
| 网络隔离 | SDK 部署在私有子网，Redis 不对公网暴露 |
| 日志脱敏 | Authorization 头只打印前 8 位（`Bearer eyJ...`） |
| 请求审计 | 记录所有写操作（POST/PUT/DELETE）的 TraceID 到数据库 |
| 错误告警 | 当 `RateLimitError` 频率 > 10/min 或断路器 Open 时触发告警 |

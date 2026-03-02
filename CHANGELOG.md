# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## [0.2.0] - 2026-03-02

### Added

**弹性与容错（审计报告 §6.2）**
- 新增 `CircuitBreaker` 断路器（三态状态机：Closed → Open → Half-Open），防止持续故障的上游服务被反复请求
- 新增 `WithCircuitBreaker(threshold int, cooldown time.Duration)` 配置选项，启用断路器功能
- 新增 `WithTimeout(d time.Duration)` 配置选项，支持分操作类型设置 HTTP 超时（如批量操作使用更长超时）
- 断路器已集成到 `Do()` 重试循环：请求前检查状态，成功/失败后自动通知断路器

**查询参数（审计报告 §7.2）**
- `buildQueryString` 现支持 slice 类型字段（`[]string`、`[]int64`、`[]int` 等），展开为重复参数（`ids=1&ids=2&ids=3`）
- `buildQueryString` 现支持嵌套/嵌入 struct（如 `order.ListOptions` 中嵌入 `core.ListOptions`），递归提取 url 标签字段

**子包测试覆盖（审计报告 §8.2）**
- 新增 `order/order_test.go`：11 个用例，覆盖 List / Count / Get / Create / Update / Delete / Cancel / Close / Open / ListRefunds / ListTransactions
- 新增 `customer/customer_test.go`：9 个用例，覆盖 List / Get / Create / Update / Delete / Count / Search / ListGroups / CreateGroup
- 新增 `webhook/webhook_test.go`：5 个用例，覆盖 List / Get / Create / Update / Delete
- 新增 `marketing/discount_test.go`：7 个用例，覆盖 PriceRule CRUD + DiscountCode List/Create
- 新增 `store/store_test.go`：6 个用例，覆盖 GetShop / GetSettlementCurrency / ListStaffMembers / GetStaffMember / ListOperationLogs / GetInfo
- 主包新增 14 个用例：断路器状态机测试（4）、断路器集成测试（1）、slice QueryString 测试（4）、嵌入 struct QueryString 测试（1）、`WithTimeout` 测试（1）、`WithCircuitBreaker` 测试（1）

**文档（审计报告 §9.2）**
- 新增 `docs/production.md`：生产环境部署完整指南
  - Redis `TokenStore` 完整参考实现（含 TTL 自适应策略）
  - Docker / Kubernetes 多副本部署注意事项
  - 多租户 `ClientPool` 参考实现
  - 断路器与限速配置推荐
  - Prometheus 监控指标建议
  - 安全加固清单（密钥管理、TLS、日志脱敏、操作审计）

### Changed

- `http.go`：`Do()` 方法在失败路径（网络错误、429、503）增加断路器 `RecordFailure()` 通知；成功时增加 `RecordSuccess()` 通知
- `http.go`：重构 `buildQueryString` 为 `buildQueryString` + `buildQueryStringFromStruct` 两层函数，支持递归处理嵌入 struct

### Fixed

- `buildQueryString` 对 `order.ListOptions` 等包含嵌入 struct 的选项类型，现在可以正确提取父级 struct 的 `url` 标签字段（原来只取最外层字段）

---

## [0.1.0] - 2026-02-27

### Added

**核心 SDK**
- `shopline.Client`：门面 Client，支持 35+ Service 接口
- `auth.go`：完整 OAuth2 流程（AuthorizeURL / GetAccessToken / RefreshAccessToken / VerifySignature / VerifyWebhookRequest）
- `token_manager.go`：Singleflight 并发安全 Token 刷新（5 分钟预刷新缓冲）
- `token_store.go`：`TokenStore` 接口 + `FileTokenStore` 原子写入实现
- `http.go`：指数退避（+Jitter）重试、Retry-After 解析、10MB 响应体限制、Context 取消

**配置选项**
- `WithVersion(v string)`：指定 API 版本
- `WithRetry(n int)`：设置最大重试次数
- `WithHTTPClient(c *http.Client)`：自定义 HTTP Client
- `WithLogger(l Logger)`：接入自定义日志
- `WithBaseURL(url string)`：覆盖 BaseURL（用于测试）
- `WithTokenManager(store TokenStore)`：启用自动 Token 管理

**API 覆盖（17 个子包，35+ Service）**
- `order`：订单、草稿订单、履约、支付、退货、归档、编辑、税费、承运商
- `product`：商品、集合（普通/智能/手动）、库存
- `customer`：客户、分组、地址、社交登录
- `store`：店铺信息、员工、操作日志、订阅
- `marketing`：价格规则、折扣码
- `online_store`：主题、页面、脚本标签
- `webhook`：Webhook 管理
- `access`：店面访问令牌
- `market`：市场、位置、发布、礼品卡
- `localizations`：多语言与翻译
- `sales_channel`：商品/集合上架
- `metafield`：元字段定义、资源、店铺元字段
- `bulk`：批量查询与变更操作
- `shopline_payments`：余额、提现、账单、交易
- `payments_app`：支付应用激活/通知/退款/设备绑定
- `app_openapi`：尺码表、CDP、变体图片

**测试**
- 40 个测试用例（主包），Race Detector 全部通过
- `go vet` 零告警

**文档**
- `README.md`：快速开始、包结构、可用服务表、配置选项说明
- `docs/guide.md`：完整使用指南
- `docs/faq.md`：常见问题解答
- `examples/`：基础用法、OAuth 流程、Token 管理、Webhook 处理

[0.2.0]: https://github.com/imokyou/slshop/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/imokyou/slshop/releases/tag/v0.1.0

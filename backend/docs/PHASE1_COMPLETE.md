# Gemini/Antigravity 平台裁剪 — 项目进度

> 项目路径：`/data/sub2api/backend`
> 当前状态：**PHASE 1 完成**（编译通过）
> 验证命令：`cd /data/sub2api/backend && go build ./cmd/server`

---

## 变更记录

| 日期 | 阶段 | 变更 |
|------|------|------|
| 2025-04-28 | PHASE 1 | 完成编译层面清理：wire_gen.go 重写、gateway_handler.go 清理、wire.go 清理 |
| 2025-04-29 | PHASE 1 收尾 | go build ./cmd/server exit 0 ✅ 编译通过 |
| 2025-04-29 | PHASE 2 | 创建本文档；完成清理规划 |
| 2025-04-29 | PHASE 2 执行 | 完成所有三批源码注释清理（编译始终 exit 0） |
| 2025-04-29 | PHASE 2 日志清理 | 注释 `selectAccountWithMixedScheduling` 函数体（行 3120-3379，248行），移除其中 10+ 条 `logger.LegacyPrintf`；编译 exit 0，binary 107MB（减少 4MB） |

---

## 已完成工作

### 编译层面的完全清理

| 文件 | 操作 | 状态 |
|------|------|------|
| `cmd/server/wire_gen.go` | 完全重写：移除所有 Gemini/Antigravity 构造函数及依赖链；添加 `provideTokenCacheStub()` 填充被删除的 TokenCache | ✅ |
| `cmd/server/wire.go` | 移除 `geminiOAuth`/`antigravityOAuth` 参数及对应 cleanup 步骤 | ✅ |
| `internal/handler/gateway_handler.go` | 删除 `PromptTooLongError` 处理块（46行）；删除未使用的 `fallbackGroupID` 变量和 `ctxkey` import | ✅ |
| `internal/server/routes/admin.go` | 删除 `BatchRefreshTier` 路由注册 | ✅ |
| `internal/domain/constants.go` | 注释掉 `PlatformGemini`/`PlatformAntigravity` 常量 | ✅ |

### wire_gen.go 关键设计决策

- **`provideTokenCacheStub()`**：内联实现 `service.TokenCache` 接口，用于填充被删除的 `GeminiTokenCache`/`ClaudeTokenCache` 等依赖
- **`ProvideRateLimitService`**：9参数版本（无 `geminiQuotaService`）
- **`ProvideOpenAITokenProvider`**：4参数版本
- **`NewAccountTestService`**：3参数版本，无 `claudeTokenProvider`
- **`NewChannelService`**：第五参数是 `*PricingService`

### 已删除的文件（36个）

```
internal/handler/admin/gemini_oauth_handler.go
internal/handler/admin/antigravity_oauth_handler.go
internal/handler/gemini_v1beta_handler.go
internal/pkg/antigravity/*.go          (12个文件)
internal/pkg/gemini/*.go              (5个文件)
internal/pkg/geminicli/*.go            (7个文件)
internal/repository/gemini_drive_client.go
internal/repository/gemini_oauth_client.go
internal/repository/gemini_token_cache.go
internal/repository/geminicli_codeassist_client.go
internal/service/antigravity_*.go      (11个文件)
internal/service/gemini_*.go          (6个文件)
```

---

## 待清理文件详细分析

> **原则**：不删除代码，先全部注释掉，确保每次改动都能一次编译成功。

---

### 文件 1：`internal/service/gateway_service.go`（9188行，最核心）

#### 1.1 `IsSingleAntigravityAccountGroup` 函数（行 2193–2201）

**当前代码**：
```go
// IsSingleAntigravityAccountGroup 检查指定分组是否只有一个 antigravity 平台的可调度账号。
// 用于 Handler 层在首次请求时提前设置 SingleAccountRetry context，
// 避免单账号分组收到 503 时错误地设置模型限流标记导致后续请求连续快速失败。
func (s *GatewayService) IsSingleAntigravityAccountGroup(ctx context.Context, groupID *int64) bool {
    accounts, _, err := s.listSchedulableAccounts(ctx, groupID, "antigravity", true)
    if err != nil {
        return false
    }
    return len(accounts) == 1
}
```

**引用位置**：
- `gateway_handler.go:293`（已注释掉）
- `gateway_handler.go:519`：`if h.gatewayService.IsSingleAntigravityAccountGroup(...)` — 这行是**活代码**

**处理方式**：将函数体替换为直接返回 `false`，保留函数签名（gateway_handler.go:519 调用处不改）

**改动**：
```go
// IsSingleAntigravityAccountGroup 检查指定分组是否只有一个 antigravity 平台的可调度账号。
// ❌ REMOVED: Antigravity 平台已删除，始终返回 false
func (s *GatewayService) IsSingleAntigravityAccountGroup(ctx context.Context, groupID *int64) bool {
    return false
}
```

---

#### 1.2 `SelectAccountForModelWithExclusions` 函数（行 1323–1372）

**当前关键行**：
```go
// 行1354：
if (platform == "anthropic" || platform == "gemini") && !hasForcePlatform {
    account, err := s.selectAccountWithMixedScheduling(...)
    return s.hydrateSelectedAccount(ctx, account)
}
// 行1362：
account, err := s.selectAccountForModelWithPlatform(...)
```

**分析**：`platform == "gemini"` 的分组已不存在，可简化为：
```go
if platform == "anthropic" && !hasForcePlatform {
    // 注释说明：antigravity 混合调度已移除
}
account, err := s.selectAccountForModelWithPlatform(...)
```

**处理方式**：注释掉 `selectAccountWithMixedScheduling` 调用分支

---

#### 1.3 `listSchedulableAccounts` 函数（行 2096–2122）

**当前关键行**（行2117）：
```go
useMixed := (platform == "anthropic" || platform == "gemini") && !hasForcePlatform
if useMixed {
    platforms := []string{platform, "antigravity"}
    ...
}
```

**处理方式**：改为 `useMixed := platform == "anthropic" && !hasForcePlatform`（删除 `|| platform == "gemini"`）

---

#### 1.4 `selectAccountForModelWithPlatform` 函数（行 2865+）

**残留 `gemini` 引用**：
- 行 2867：`preferOAuth := platform == "gemini"` — 当 platform=="gemini" 时永远走不到这里（已在调用处删除），但仍可改为 `false`
- 行 3240：`if preferOAuth && acc.Platform == "gemini"` — 同上

**处理方式**：注释掉相关行

---

#### 1.5 `resolveAccountUpstreamModel` 函数（行 ~8340）

**当前代码**：
```go
func resolveAccountUpstreamModel(account *Account, requestedModel string) string {
    if account.Platform == "antigravity" {
        return mapAntigravityModel(account, requestedModel)
    }
    return account.GetMappedModel(requestedModel)
}
```

**处理方式**：注释掉 antigravity 分支
```go
func resolveAccountUpstreamModel(account *Account, requestedModel string) string {
    // if account.Platform == "antigravity" {
    //     return mapAntigravityModel(account, requestedModel)
    // }
    return account.GetMappedModel(requestedModel)
}
```

---

#### 1.6 `CountTokens` 相关（行 ~8425）

**当前代码**：
```go
// Antigravity 账户不支持 count_tokens，返回 404 让客户端 fallback 到本地估算。
if account.Platform == "antigravity" {
    s.countTokensError(c, http.StatusNotFound, "not_found_error", "count_tokens endpoint is not supported for this platform")
    return nil
}
```

**处理方式**：注释掉

---

#### 1.7 `mapAntigravityModel` 函数（行 9115–9118）

**当前代码**：
```go
// mapAntigravityModel maps a requested model to the Antigravity upstream model
func mapAntigravityModel(account *Account, requestedModel string) string {
    // Antigravity platform removed
    return ""
}
```

**状态**：已 stub 为返回空字符串 ✅（无需处理）

---

### 文件 2：`internal/service/gateway_request.go`

**残留引用**：
- 行 197：`case "gemini":` — 这个 case 分支**永远不会被执行**（所有调用处只传 `"anthropic"`/`"chat_completions"`/`"responses"`）

**处理方式**：注释掉整个 `case "gemini":` 分支

---

### 文件 3：`internal/service/model_rate_limit.go`

**残留引用**：
- 行 37、59：`resolveFinalAntigravityModelKey` 调用
- 行 68：`resolveFinalAntigravityModelKey` 函数定义

**`resolveFinalAntigravityModelKey` 函数**：
```go
func resolveFinalAntigravityModelKey(ctx context.Context, account *Account, requestedModel string) string {
    modelKey := mapAntigravityModel(account, requestedModel)  // 已返回 ""
    if modelKey == "" {
        return ""
    }
    if enabled, ok := ThinkingEnabledFromContext(ctx); ok {
        modelKey = applyThinkingModelSuffix(modelKey, enabled)
    }
    return modelKey
}
```

**分析**：由于 `mapAntigravityModel` 已返回 `""`，此函数始终返回 `""`。但修改 `resolveFinalAntigravityModelKey` 会影响 `Account.RateLimitRemaining` 等方法的执行逻辑。

**处理方式**：
- 注释掉函数体中的 `modelKey := mapAntigravityModel(...)` 行，改为直接 `modelKey := ""`
- 或注释掉整个函数并在调用处处理

**谨慎处理**，建议放在最后。

---

### 文件 4：`internal/server/middleware/middleware.go`

**残留**：`ForcePlatform` 中间件（行 25–41）

```go
// ContextKeyForcePlatform 强制平台（用于 /antigravity 路由）
ContextKeyForcePlatform ContextKey = "force_platform"

// ForcePlatform 返回设置强制平台的中间件
// ❌ REMOVED: /antigravity 路由已删除，保留此中间件仅用于兼容
func ForcePlatform(platform string) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := context.WithValue(c.Request.Context(), ctxkey.ForcePlatform, platform)
        c.Request = c.Request.WithContext(ctx)
        c.Set(string(ContextKeyForcePlatform), platform)
        c.Next()
    }
}
```

**引用**：仅 `routes/gateway.go` 中的已注释路由使用

**处理方式**：保留此函数（兼容/无风险），无需修改

---

### 文件 5：`internal/server/routes/gateway.go`

**残留**：被注释掉的 gemini/antigravity 路由块（行 117–213），代码块中全是注释，无实际编译影响

**处理方式**：空白区域清理，不影响编译，可忽略

---

## 执行计划（按安全顺序）

### 第一批：零风险改动（编译必过）

| # | 文件 | 改动 | 风险 |
|---|------|------|------|
| 1 | `gateway_service.go` | `mapAntigravityModel` → 已 stub ✅ | 无 |
| 2 | `gateway_service.go` | `IsSingleAntigravityAccountGroup` → 返回 `false` | 低（调用处已清理） |
| 3 | `gateway_service.go` | 注释掉 `resolveAccountUpstreamModel` 中的 antigravity 分支 | 低 |
| 4 | `gateway_service.go` | 注释掉 `CountTokens` 中的 antigravity 分支 | 低 |
| 5 | `gateway_request.go` | 注释掉 `case "gemini":` 分支 | 零（永不执行） |

### 第二批：需验证的改动

| # | 文件 | 改动 | 风险 |
|---|------|------|------|
| 6 | `gateway_service.go` | 简化 `SelectAccountForModelWithExclusions` 中的混合调度条件 | 中 |
| 7 | `gateway_service.go` | 简化 `listSchedulableAccounts` 中的 `useMixed` 条件 | 中 |
| 8 | `gateway_service.go` | 注释掉 `selectAccountForModelWithPlatform` 中的 `preferOAuth` 和 `acc.Platform == "gemini"` 判断 | 中 |

### 第三批：敏感区域（谨慎处理）

| # | 文件 | 改动 | 风险 |
|---|------|------|------|
| 9 | `model_rate_limit.go` | `resolveFinalAntigravityModelKey` stub | 高（影响限流逻辑） |

---

## 编译验证命令

```bash
cd /data/sub2api/backend && go build ./cmd/server
```

每次改动后必须运行此命令，确保 exit 0。

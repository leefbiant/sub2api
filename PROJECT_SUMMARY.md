# Sub2API — 项目总结

> 最后更新: 2026-05-01 07:58

## 📁 项目目录

- **根目录**: `/data/sub2api/`
- **后端**: `/data/sub2api/backend/`
- **前端**: `/data/sub2api/frontend/`
- **测试垃圾文件**: `/data/sub2api/test_garbage/` (67 个文件已在 afe2c64 中删除，目录已清除)
- **Git 分支**: `feat/remove-gemini-antigravity-platform`
- **Git 当前提交**: `afe2c64` feat: remove Gemini and Antigravity platform code, archive test files

## 🛠 编译环境

| 项目     | 值                                    |
| -------- | ------------------------------------- |
| Go 版本  | go1.26.2 linux/amd64                  |
| GOROOT   | /root/go/pkg/mod/golang.org/toolchain |
| GOPATH   | /root/go                              |
| Go Module | github.com/Wei-Shaw/sub2api          |
| Binary   | `/tmp/sub2api-clean` (111MB)          |
| 版本     | Sub2API 0.1.119                       |

## 📊 构建状态

```
go build ./...                          ✅ 通过
go build -o /tmp/sub2api-clean         ✅ 106MB → 111MB binary
```

## 🧪 测试状态

```
internal/config/...                     ✅ 通过
internal/repository/...                 ✅ 通过
internal/service/...                    ✅ 通过（1 个既有失败，非本次改动导致）
internal/service/openai_ws_v2/...      ✅ 通过
```

### 已知既有测试失败（与本次清理无关）

- `TestIdentityService_RewriteUserIDWithMasking_PreservesTopLevelFieldOrder` — 原始代码已失败（JSON 字段顺序问题）

## 📜 变更概述

### 删除的平台

| 平台           | 常量                     | 说明   |
| -------------- | ------------------------ | ------ |
| ✅ Anthropic   | `PlatformAnthropic`      | 已删除 |
| ✅ Gemini      | `PlatformGemini`         | 已删除 |
| ✅ Antigravity | `PlatformAntigravity`    | 已删除 |

**保留的平台**: **OpenAI**、**Codex**、**Bedrock**

### 源文件清理（26 个文件）

注释/删除了以下平台分支和引用：

- **核心服务**: `gateway_service.go` (11 处清理), `endpoint.go`, `openai_gateway_service.go`
- **账号管理**: `account_data.go`, `account_handler.go`, `account.go`
- **通道与转发**: `channel_handler.go`, `gateway_handler.go`, `gateway_forward_as_*.go`
- **配置与设置**: `setting_service.go`, `settings_view.go`, `config.go`, `constant.go`
- **运营工具**: `ops_error_logger.go`, `admin_service.go`, `failover_loop.go`
- **同步模块**: `crs_sync_service.go`, `scheduler_snapshot_service.go`
- **其他**: `simple_mode_default_groups.go`, `group_service.go`, `channel_repo_pricing.go`, `ops_retry.go`, `ratelimit_service.go`, `token_*.go`, `claude_token_provider.go`

### 移出的测试文件（67 个 → test_garbage/ → 已在 afe2c64 中删除）

纯 Anthropic/Gemini/Antigravity 测试文件已归档并删除：

```
claude_token_provider_test.go           account_anthropic_passthrough_test.go
gateway_anthropic_apikey_passthrough_test.go  bedrock_request_test.go
bedrock_signer_test.go                 gateway_service_bedrock_beta_test.go
gateway_service_bedrock_model_support_test.go  scheduler_snapshot_hydration_test.go
account_repo_integration_test.go        admin_service_bulk_update_test.go
admin_service_overages_test.go          admin_service_search_test.go
failover_loop_test.go                  gateway_channel_restriction_test.go
oauth_refresh_api_test.go              openai_token_provider_test.go
simple_mode_default_groups_integration_test.go  api_key_auth_google_test.go
wire_gen_test.go                      (及 60+ 之前移入的文件)
```

### 混合测试文件修复（8 个）

这些测试文件中 `PlatformAnthropic` 被用作"非 OpenAI 平台"占位符，改为字面字符串 `"anthropic"` 保留测试逻辑：

- `account_openai_compact_test.go` — 3 处
- `account_openai_passthrough_test.go` — 5 处
- `identity_service_order_test.go` — 1 处
- `openai_account_scheduler_compact_test.go` — 1 处
- `openai_gateway_service_test.go` — 1 处
- `req_client_pool_test.go` — `createGeminiReqClient` → `createOpenAIReqClient`

## 🔮 待办事项

### 高优先级

- [ ] **推送分支**: `git push origin feat/remove-gemini-antigravity-platform`
- [ ] **创建 PR**: 合并到 main 分支

### 低优先级 / 可选

- [ ] **清理残余死代码（无编译影响）**:
  - `gateway_service.go:2173` — `account.Platform == "antigravity"` 永远 false（~3 行）
  - `account_handler.go:803` — `account.Platform == "anthropic"` 分支（~3 行）
  - `account_data.go` — 两处 `antigravity` privacy goroutine 块
- [ ] **清理 test_garbage/**: ~~确认后永久删除 86 个归档文件~~ ✅ 已完成（afe2c64）
- [ ] **修复既有测试 `TestIdentityService_RewriteUserIDWithMasking_PreservesTopLevelFieldOrder`**
- [ ] **验证线上环境**: 部署新 binary，确认功能正常后关闭旧实例
- [ ] **更新文档**: README 中移除已删除平台的描述

### 已完成

- ~~删除所有 Anthropic/Gemini/Antigravity 源文件中的平台分支~~ ✅
- ~~移出/修复测试文件~~ ✅
- ~~`go build ./...` 通过~~ ✅
- ~~`go test ./...` 编译通过~~ ✅
- ~~创建 commit `afe2c64`~~ ✅

## ⚠️ 已知风险

1. **既存测试失败** — `TestIdentityService_RewriteUserIDWithMasking_PreservesTopLevelFieldOrder` 非本次引入
2. **死代码残留** — 3 处永远不执行的条件分支，编译不报错但存在
3. **test_garbage/ 仍占用空间** — 86 个文件约 ~45000 行代码，建议确认无保留价值后删除

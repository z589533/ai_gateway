# AI Gateway MVP

这是一个面试交付用的 AI Gateway MVP，后端使用 Go + Gin + GORM + MySQL + Redis，实现租户、API Key、OpenAI 风格代理、用量追踪和 OpenAPI 文档；管理台使用 `pure-admin-thin`（Vue 3 + Vite + Element Plus）实现，并由 nginx 托管。

## 功能范围

| 模块 | 说明 |
|------|------|
| 租户管理 | 创建租户、列表查询、启用 / 禁用 |
| API Key 管理 | 一个租户多个 Key；支持 scope、启用 / 禁用、过期时间、软删除 |
| OpenAI 兼容代理 | `/v1/chat/completions`、`/v1/models`，请求和响应保持 OpenAI 风格 |
| 用量追踪 | 按 tenant + key + model + token + timestamp 记录和查询 |
| OpenAPI | `api/openapi.yaml`，运行时可通过 `/openapi.yaml` 访问 |
| 管理台 | 基于 `pure-admin-thin` 的后台管理界面，支持租户、Key、用量、代理测试 |

## 架构说明

```text
浏览器 / 管理员
  │
  ├─ http://localhost:8848 ── pure-admin-thin 管理台 nginx
  │                            ├─ /api/* 反代到 app:8080/api/*
  │                            └─ /v1/*  反代到 app:8080/v1/*
  │
  └─ curl / SDK ────────────── AI Gateway Gin 服务 :8080
                               ├─ MySQL：tenant、api_keys、usage_records
                               ├─ Redis：API Key 鉴权缓存
                               └─ Mock upstream：模拟模型返回
```

## 快速启动

```bash
docker compose up --build
```

启动后访问：

| 服务 | 地址 |
|------|------|
| 后端 API | http://localhost:18080 |
| 管理台 | http://localhost:8848 |
| 健康检查 | http://localhost:18080/health |
| OpenAPI | http://localhost:18080/openapi.yaml |

如果提示 Docker daemon 连接失败，意思是 Docker Desktop 的后台引擎没有启动。你本机装的是 Docker Desktop，但 `docker compose` 真正工作时要连接它的后台服务；打开 Docker Desktop，等左下角显示 Running 后再执行命令即可。

## 默认配置

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `APP_PORT` | `8080` | 后端 HTTP 端口 |
| `APP_HOST_PORT` | `18080` | compose 映射到宿主机的后端端口 |
| `ADMIN_HOST_PORT` | `8848` | compose 映射到宿主机的管理台端口 |
| `ADMIN_TOKEN` | `admin-dev-token` | 管理 API Token |
| `MYSQL_DSN` | compose 内置 DSN | MySQL 连接串 |
| `REDIS_ADDR` | `redis:6379` | Redis 地址 |
| `PROXY_TIMEOUT_SEC` | `30` | 代理超时时间 |
| `MOCK_LATENCY_MS` | `0` | Mock 延迟 |
| `MOCK_FAIL` | `false` | 强制 mock 返回 502 |
| `KEY_CACHE_TTL_SEC` | `300` | API Key 鉴权缓存 TTL |
| `RATE_LIMIT_GLOBAL_QPS` | `100` | 全局 QPS |
| `RATE_LIMIT_KEY_QPS` | `20` | 单 Key QPS |
| `RATE_LIMIT_TENANT_QPS` | `50` | 单租户 QPS |

管理台默认 Admin Token 是 `admin-dev-token`。在 compose 场景下 API Base 留空即可，因为 nginx 已经把 `/api` 和 `/v1` 反代到后端。

MySQL 和 Redis 只在 compose 网络内暴露给后端服务，不映射到宿主机端口，避免和本机已有的 MySQL / Redis 冲突。

## curl 自测

创建租户：

```bash
curl -s -X POST http://localhost:18080/api/v1/tenants \
  -H "Authorization: Bearer admin-dev-token" \
  -H "Content-Type: application/json" \
  -d '{"name":"demo"}'
```

创建 API Key，记录返回里的 `data.secret_key`：

```bash
curl -s -X POST http://localhost:18080/api/v1/tenants/1/keys \
  -H "Authorization: Bearer admin-dev-token" \
  -H "Content-Type: application/json" \
  -d '{"name":"default","scopes":["chat:completions","models:read"]}'
```

调用 OpenAI 风格代理：

```bash
curl -s -X POST http://localhost:18080/v1/chat/completions \
  -H "Authorization: Bearer sk-ag-REPLACE_ME" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hi"}]}'
```

查询模型：

```bash
curl -s http://localhost:18080/v1/models \
  -H "Authorization: Bearer sk-ag-REPLACE_ME"
```

查询用量：

```bash
curl -s "http://localhost:18080/api/v1/usage?tenant_id=1" \
  -H "Authorization: Bearer admin-dev-token"
```

禁用 Key 并验证 403：

```bash
curl -s -X PATCH http://localhost:18080/api/v1/tenants/1/keys/1 \
  -H "Authorization: Bearer admin-dev-token" \
  -H "Content-Type: application/json" \
  -d '{"status":0}'
```

## 管理台

管理台已经换成 `pure-admin-thin`，不是手写静态页面。入口在 `web/`：

| 文件 | 说明 |
|------|------|
| `web/src/views/gateway/index.vue` | AI Gateway 管理台页面 |
| `web/src/api/gateway.ts` | 管理 API、代理 API 封装 |
| `web/src/router/modules/home.ts` | 首页路由指向管理台 |
| `web/nginx.conf` | nginx 托管和反代配置 |
| `Dockerfile.web` | pure-admin-thin 构建并复制 dist 到 nginx |

当前管理台是 MVP 内网模式：不做登录页，页面里配置 Admin Token 后调用管理 API。Token 会保存在浏览器 localStorage，只适合 Demo / 内网环境，不适合公网暴露。

## OpenAPI

OpenAPI 3.0.3 文件：

- 本地文件：`api/openapi.yaml`
- 运行地址：http://localhost:18080/openapi.yaml

覆盖管理面接口、OpenAI 兼容代理接口、鉴权 scheme、主要错误码和响应 schema。

## 测试

后端：

```bash
go test ./...
go vet ./...
```

前端管理台：

```bash
cd web
pnpm install --frozen-lockfile
pnpm test
pnpm typecheck
pnpm build
```

Compose 配置检查：

```bash
docker compose config
```

## 设计决策

- Scope 使用字符串列表，例如 `chat:completions`、`models:read`，比 bitmask 更直观，后续扩展不需要改表结构。
- API Key 明文只在创建时返回一次，数据库只保存 SHA-256 hash 和展示前缀。
- Key 删除采用软删除，避免历史 usage 失去引用。
- 代理层当前只做 mock，不调用真实 OpenAI 或其他模型厂商。
- 成功代理和已识别 Key 的 502 / 504 会写 usage；401、403 不写 usage。
- 管理 API 使用单一 `ADMIN_TOKEN`，满足 MVP 和面试演示，不做 RBAC。
- 限流使用 sentinel-golang 进程内规则，适合单实例 MVP。

## 已知限制

- 没有真实上游模型厂商接入。
- 不支持 `stream=true`，传入时返回 `400 stream_not_supported`。
- Token 数是估算值，不是 tiktoken 精确计算。
- 限流是单进程内存规则，多实例不共享配额。
- 管理台没有登录和 RBAC，Admin Token 存在 localStorage，仅限内网 / Demo。

# AI Gateway MVP

Go + Gin + GORM + MySQL + Redis implementation of an AI Gateway MVP. It manages tenants and API keys, exposes an OpenAI-style `/v1/chat/completions` proxy backed by a local mock model, and records token usage by tenant and key.

## Architecture

- Management API: `/api/v1/*`, protected by `Authorization: Bearer <ADMIN_TOKEN>`.
- Data API: `/v1/chat/completions` and `/v1/models`, protected by tenant API keys.
- MySQL: stores tenants, API keys, and usage records.
- Redis: caches API key auth metadata for 5 minutes by default.
- Mock proxy: returns OpenAI-compatible JSON and estimated token usage without external network calls.

## Quick Start

```bash
docker compose up --build
```

The API listens on `http://localhost:8080`.

Health check:

```bash
curl -s http://localhost:8080/health
```

## Configuration

| Env | Default | Description |
|-----|---------|-------------|
| `APP_PORT` | `8080` | HTTP port |
| `ADMIN_TOKEN` | `admin-dev-token` | Management API bearer token |
| `MYSQL_DSN` | compose DSN | MySQL connection string |
| `REDIS_ADDR` | `redis:6379` | Redis address |
| `PROXY_TIMEOUT_SEC` | `30` | Mock/upstream timeout |
| `MOCK_LATENCY_MS` | `0` | Artificial mock latency |
| `MOCK_FAIL` | `false` | Force 502 from mock proxy |
| `KEY_CACHE_TTL_SEC` | `300` | API key cache TTL |
| `RATE_LIMIT_GLOBAL_QPS` | `100` | Global proxy QPS |
| `RATE_LIMIT_KEY_QPS` | `20` | Per-key proxy QPS |
| `RATE_LIMIT_TENANT_QPS` | `50` | Per-tenant proxy QPS |

## Curl Walkthrough

Create a tenant:

```bash
curl -s -X POST http://localhost:8080/api/v1/tenants \
  -H "Authorization: Bearer admin-dev-token" \
  -H "Content-Type: application/json" \
  -d '{"name":"demo"}'
```

Create an API key and save `data.secret_key`:

```bash
curl -s -X POST http://localhost:8080/api/v1/tenants/1/keys \
  -H "Authorization: Bearer admin-dev-token" \
  -H "Content-Type: application/json" \
  -d '{"name":"default","scopes":["chat:completions","models:read"]}'
```

Call the OpenAI-style proxy:

```bash
curl -s -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-ag-REPLACE_ME" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hi"}]}'
```

List models:

```bash
curl -s http://localhost:8080/v1/models \
  -H "Authorization: Bearer sk-ag-REPLACE_ME"
```

Query usage:

```bash
curl -s "http://localhost:8080/api/v1/usage?tenant_id=1" \
  -H "Authorization: Bearer admin-dev-token"
```

Disable a key and verify 403:

```bash
curl -s -X PATCH http://localhost:8080/api/v1/tenants/1/keys/1 \
  -H "Authorization: Bearer admin-dev-token" \
  -H "Content-Type: application/json" \
  -d '{"status":0}'
```

## API Documentation

OpenAPI 3.0.3 spec:

- Local file: `api/openapi.yaml`
- Runtime URL: `http://localhost:8080/openapi.yaml`

## Testing

```bash
go test ./...
go test ./... -cover
```

## Design Decisions

- Scope model: string scopes such as `chat:completions` and `models:read`; empty scopes mean no permission.
- API keys: plaintext secret is returned once, then only SHA-256 hash and display prefix are stored.
- Deletion: API keys are soft-deleted so historical usage remains traceable.
- Proxy: mock upstream only; no external LLM calls.
- Token usage: estimated from message character count and mock completion size.
- Errors: OpenAI-compatible error body for data API; management API uses `{code,message,data}` envelope.
- Rate limiting: sentinel-golang in-process rules, suitable for single-instance MVP.

## Known Limits

- No real OpenAI or vendor upstream integration.
- `stream=true` is rejected with `400 stream_not_supported`.
- Token counts are approximate.
- Rate limits are process-local and not shared across replicas.
- Management API uses one static token, not RBAC.
- Dashboard is intentionally left as P1 optional per the technical design.

import { readFileSync } from "node:fs";
import test from "node:test";
import assert from "node:assert/strict";

const page = readFileSync(new URL("../src/views/gateway/index.vue", import.meta.url), "utf8");
const api = readFileSync(new URL("../src/api/gateway.ts", import.meta.url), "utf8");
const route = readFileSync(new URL("../src/router/modules/home.ts", import.meta.url), "utf8");
const nginx = readFileSync(new URL("../nginx.conf", import.meta.url), "utf8");

test("pure-admin route opens AI Gateway dashboard by default", () => {
  assert.match(route, /redirect: "\/gateway\/index"/);
  assert.match(route, /name: "GatewayAdmin"/);
});

test("dashboard covers tenants, keys, usage and playground", () => {
  for (const text of ["租户管理", "API Key", "用量统计", "代理测试"]) {
    assert.ok(page.includes(text), `missing ${text}`);
  }
});

test("gateway api wraps required backend endpoints", () => {
  for (const endpoint of ["/api/v1/tenants", "/keys", "/api/v1/usage", "/v1/models", "/v1/chat/completions"]) {
    assert.ok(api.includes(endpoint), `missing ${endpoint}`);
  }
});

test("nginx proxies management and gateway routes", () => {
  assert.match(nginx, /proxy_pass http:\/\/app:8080\/api\//);
  assert.match(nginx, /proxy_pass http:\/\/app:8080\/v1\//);
});

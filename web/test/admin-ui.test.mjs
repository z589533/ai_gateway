import { readFileSync } from "node:fs";
import test from "node:test";
import assert from "node:assert/strict";

const html = readFileSync(new URL("../index.html", import.meta.url), "utf8");
const js = readFileSync(new URL("../assets/app.js", import.meta.url), "utf8");
const css = readFileSync(new URL("../assets/styles.css", import.meta.url), "utf8");
const nginx = readFileSync(new URL("../nginx.conf", import.meta.url), "utf8");

test("admin UI exposes all required views", () => {
  for (const id of ["overviewView", "tenantsView", "keysView", "usageView", "playgroundView"]) {
    assert.match(html, new RegExp(`id="${id}"`));
  }
});

test("admin UI wires management API operations", () => {
  for (const path of ["/api/v1/tenants", "/keys", "/api/v1/usage", "/v1/chat/completions", "/v1/models"]) {
    assert.ok(js.includes(path), `missing ${path}`);
  }
});

test("nginx proxies API traffic to backend app service", () => {
  assert.match(nginx, /proxy_pass http:\/\/app:8080\/api\//);
  assert.match(nginx, /proxy_pass http:\/\/app:8080\/v1\//);
});

test("dashboard styling uses responsive layout constraints", () => {
  assert.match(css, /@media \(max-width: 1080px\)/);
  assert.match(css, /border-radius: 8px/);
});

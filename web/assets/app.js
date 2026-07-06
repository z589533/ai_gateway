const state = {
  apiBase: localStorage.getItem("ag.apiBase") || defaultApiBase(),
  adminToken: localStorage.getItem("ag.adminToken") || "admin-dev-token",
  view: "overview",
  tenants: [],
  keys: [],
  allKeys: [],
  usage: null,
  selectedTenantId: "",
};

const titles = {
  overview: "概览",
  tenants: "租户管理",
  keys: "API Key",
  usage: "用量统计",
  playground: "代理测试",
};

const els = {};

document.addEventListener("DOMContentLoaded", () => {
  bindElements();
  bindEvents();
  hydrateSettings();
  switchView("overview");
  bootstrap();
});

function bindElements() {
  [
    "apiBaseInput",
    "adminTokenInput",
    "saveSettingsBtn",
    "alertHost",
    "healthDot",
    "healthText",
    "viewTitle",
    "metricTenants",
    "metricKeys",
    "metricTokens",
    "metricErrors",
    "recentUsageRows",
    "tenantRows",
    "createTenantForm",
    "keyTenantSelect",
    "createKeyForm",
    "secretPanel",
    "secretValue",
    "copySecretBtn",
    "keyRows",
    "usageFilterForm",
    "usageTenantSelect",
    "usageKeySelect",
    "usagePrompt",
    "usageCompletion",
    "usageTotal",
    "usageSuccess",
    "usageError",
    "usageRows",
    "refreshOverviewBtn",
    "loadModelsBtn",
    "playgroundForm",
    "playgroundOutput",
  ].forEach((id) => {
    els[id] = document.getElementById(id);
  });
}

function bindEvents() {
  document.querySelectorAll(".nav-item").forEach((button) => {
    button.addEventListener("click", () => switchView(button.dataset.view));
  });
  document.querySelectorAll("[data-jump]").forEach((button) => {
    button.addEventListener("click", () => switchView(button.dataset.jump));
  });

  els.saveSettingsBtn.addEventListener("click", saveSettings);
  els.refreshOverviewBtn.addEventListener("click", bootstrap);
  els.createTenantForm.addEventListener("submit", createTenant);
  els.keyTenantSelect.addEventListener("change", () => {
    state.selectedTenantId = els.keyTenantSelect.value;
    loadKeys();
  });
  els.createKeyForm.addEventListener("submit", createKey);
  els.copySecretBtn.addEventListener("click", copySecret);
  els.usageFilterForm.addEventListener("submit", (event) => {
    event.preventDefault();
    loadUsage();
  });
  els.usageTenantSelect.addEventListener("change", refreshUsageKeyOptions);
  els.loadModelsBtn.addEventListener("click", loadModels);
  els.playgroundForm.addEventListener("submit", sendPlaygroundRequest);
}

function hydrateSettings() {
  els.apiBaseInput.value = state.apiBase;
  els.adminTokenInput.value = state.adminToken;
}

async function bootstrap() {
  await checkHealth();
  await loadTenants();
  await loadAllKeys();
  await loadKeys();
  await loadUsage();
  renderOverview();
}

function defaultApiBase() {
  if (window.location.protocol === "file:") {
    return "http://localhost:8080";
  }
  return "";
}

function saveSettings() {
  state.apiBase = trimSlash(els.apiBaseInput.value.trim());
  state.adminToken = els.adminTokenInput.value.trim() || "admin-dev-token";
  localStorage.setItem("ag.apiBase", state.apiBase);
  localStorage.setItem("ag.adminToken", state.adminToken);
  notify("设置已保存", "ok");
  bootstrap();
}

function switchView(view) {
  state.view = view;
  document.querySelectorAll(".view").forEach((section) => {
    section.classList.toggle("active", section.id === `${view}View`);
  });
  document.querySelectorAll(".nav-item").forEach((button) => {
    button.classList.toggle("active", button.dataset.view === view);
  });
  els.viewTitle.textContent = titles[view] || "AI Gateway";
}

async function checkHealth() {
  try {
    await rawFetch("/health");
    els.healthDot.className = "status-dot ok";
    els.healthText.textContent = "api online";
  } catch (error) {
    els.healthDot.className = "status-dot fail";
    els.healthText.textContent = "api offline";
  }
}

async function loadTenants() {
  try {
    const data = await adminFetch("/api/v1/tenants?page_size=100");
    state.tenants = data.items || [];
    if (!state.selectedTenantId && state.tenants.length > 0) {
      state.selectedTenantId = String(state.tenants[0].id);
    }
    renderTenants();
    renderTenantSelects();
  } catch (error) {
    notify(error.message, "err");
  }
}

async function createTenant(event) {
  event.preventDefault();
  const form = new FormData(event.currentTarget);
  const name = String(form.get("name") || "").trim();
  if (!name) return;
  try {
    const tenant = await adminFetch("/api/v1/tenants", {
      method: "POST",
      body: { name },
    });
    state.selectedTenantId = String(tenant.id);
    event.currentTarget.reset();
    notify("租户已创建", "ok");
    await loadTenants();
    await loadAllKeys();
    await loadKeys();
  } catch (error) {
    notify(error.message, "err");
  }
}

async function updateTenantStatus(tenantId, status) {
  try {
    await adminFetch(`/api/v1/tenants/${tenantId}`, {
      method: "PATCH",
      body: { status },
    });
    notify("租户状态已更新", "ok");
    await loadTenants();
  } catch (error) {
    notify(error.message, "err");
  }
}

function renderTenants() {
  if (state.tenants.length === 0) {
    els.tenantRows.innerHTML = emptyRow(5, "暂无租户");
    return;
  }
  els.tenantRows.innerHTML = state.tenants
    .map((tenant) => {
      const nextStatus = tenant.status === 1 ? 0 : 1;
      const action = tenant.status === 1 ? "禁用" : "启用";
      return `<tr>
        <td>${tenant.id}</td>
        <td>${escapeHTML(tenant.name)}</td>
        <td>${statusPill(tenant.status)}</td>
        <td>${formatDate(tenant.created_at)}</td>
        <td><button class="button secondary" type="button" onclick="updateTenantStatus(${tenant.id}, ${nextStatus})">${action}</button></td>
      </tr>`;
    })
    .join("");
}

function renderTenantSelects() {
  const options = state.tenants
    .map((tenant) => `<option value="${tenant.id}">${escapeHTML(tenant.name)} (#${tenant.id})</option>`)
    .join("");
  const placeholder = `<option value="">请选择租户</option>`;
  els.keyTenantSelect.innerHTML = placeholder + options;
  els.usageTenantSelect.innerHTML = `<option value="">全部租户</option>${options}`;
  els.keyTenantSelect.value = state.selectedTenantId || "";
  els.usageTenantSelect.value = state.selectedTenantId || "";
  refreshUsageKeyOptions();
}

async function loadKeys() {
  const tenantId = state.selectedTenantId;
  if (!tenantId) {
    state.keys = [];
    renderKeys();
    refreshUsageKeyOptions();
    renderOverview();
    return;
  }
  try {
    const data = await adminFetch(`/api/v1/tenants/${tenantId}/keys?page_size=100`);
    state.keys = data.items || [];
    renderKeys();
    refreshUsageKeyOptions();
    renderOverview();
  } catch (error) {
    state.keys = [];
    renderKeys();
    notify(error.message, "err");
  }
}

async function loadAllKeys() {
  if (state.tenants.length === 0) {
    state.allKeys = [];
    return;
  }
  const lists = await Promise.all(
    state.tenants.map((tenant) =>
      adminFetch(`/api/v1/tenants/${tenant.id}/keys?page_size=100`).catch(() => ({ items: [] })),
    ),
  );
  state.allKeys = lists.flatMap((item) => item.items || []);
}

async function createKey(event) {
  event.preventDefault();
  const tenantId = state.selectedTenantId;
  if (!tenantId) {
    notify("请先选择租户", "err");
    return;
  }
  const form = new FormData(event.currentTarget);
  const scopes = form.getAll("scope");
  const body = {
    name: String(form.get("name") || "").trim(),
    scopes,
  };
  const expiresAt = localDateTimeToISO(form.get("expires_at"));
  if (expiresAt) body.expires_at = expiresAt;
  try {
    const created = await adminFetch(`/api/v1/tenants/${tenantId}/keys`, {
      method: "POST",
      body,
    });
    els.secretValue.textContent = created.secret_key;
    els.secretPanel.classList.remove("hidden");
    event.currentTarget.reset();
    event.currentTarget.querySelector('input[value="chat:completions"]').checked = true;
    notify("API Key 已创建，明文只展示一次", "ok");
    await loadKeys();
    await loadAllKeys();
  } catch (error) {
    notify(error.message, "err");
  }
}

async function updateKey(keyId) {
  const scopes = Array.from(document.querySelectorAll(`[data-key-scope="${keyId}"]:checked`)).map((input) => input.value);
  const status = Number(document.querySelector(`[data-key-status="${keyId}"]`).value);
  const expiresRaw = document.querySelector(`[data-key-expires="${keyId}"]`).value;
  const body = { scopes, status, expires_at: localDateTimeToISO(expiresRaw) || null };
  try {
    await adminFetch(`/api/v1/tenants/${state.selectedTenantId}/keys/${keyId}`, {
      method: "PATCH",
      body,
    });
    notify("Key 已更新", "ok");
    await loadKeys();
    await loadAllKeys();
  } catch (error) {
    notify(error.message, "err");
  }
}

async function deleteKey(keyId) {
  if (!window.confirm("确认删除这个 Key？历史用量仍会保留。")) return;
  try {
    await adminFetch(`/api/v1/tenants/${state.selectedTenantId}/keys/${keyId}`, {
      method: "DELETE",
    });
    notify("Key 已删除", "ok");
    await loadKeys();
    await loadAllKeys();
  } catch (error) {
    notify(error.message, "err");
  }
}

function renderKeys() {
  if (!state.selectedTenantId) {
    els.keyRows.innerHTML = emptyRow(7, "请先选择租户");
    return;
  }
  if (state.keys.length === 0) {
    els.keyRows.innerHTML = emptyRow(7, "当前租户暂无 Key");
    return;
  }
  els.keyRows.innerHTML = state.keys
    .map((key) => {
      const scopes = Array.isArray(key.scopes) ? key.scopes : [];
      return `<tr>
        <td>${key.id}</td>
        <td>${escapeHTML(key.name)}</td>
        <td><code>${escapeHTML(key.key_prefix || "")}</code></td>
        <td>
          <label class="check"><input data-key-scope="${key.id}" type="checkbox" value="chat:completions" ${scopes.includes("chat:completions") ? "checked" : ""}> chat</label>
          <label class="check"><input data-key-scope="${key.id}" type="checkbox" value="models:read" ${scopes.includes("models:read") ? "checked" : ""}> models</label>
        </td>
        <td>
          <select data-key-status="${key.id}">
            <option value="1" ${key.status === 1 ? "selected" : ""}>enabled</option>
            <option value="0" ${key.status === 0 ? "selected" : ""}>disabled</option>
          </select>
        </td>
        <td><input data-key-expires="${key.id}" type="datetime-local" value="${isoToLocalInput(key.expires_at)}"></td>
        <td>
          <div class="row-actions">
            <button class="button secondary" type="button" onclick="updateKey(${key.id})">保存</button>
            <button class="button danger" type="button" onclick="deleteKey(${key.id})">删除</button>
          </div>
        </td>
      </tr>`;
    })
    .join("");
}

async function loadUsage() {
  const form = new FormData(els.usageFilterForm);
  const params = new URLSearchParams({ page_size: "100" });
  const tenantId = form.get("tenant_id");
  const apiKeyId = form.get("api_key_id");
  const from = localDateTimeToISO(form.get("from"));
  const to = localDateTimeToISO(form.get("to"));
  if (tenantId) params.set("tenant_id", tenantId);
  if (apiKeyId) params.set("api_key_id", apiKeyId);
  if (from) params.set("from", from);
  if (to) params.set("to", to);
  try {
    state.usage = await adminFetch(`/api/v1/usage?${params.toString()}`);
    renderUsage();
    renderOverview();
  } catch (error) {
    notify(error.message, "err");
  }
}

function refreshUsageKeyOptions() {
  const tenantId = els.usageTenantSelect.value || state.selectedTenantId;
  const options = state.allKeys
    .filter((key) => !tenantId || String(key.tenant_id) === String(tenantId))
    .map((key) => `<option value="${key.id}">${escapeHTML(key.name)} (#${key.id})</option>`)
    .join("");
  els.usageKeySelect.innerHTML = `<option value="">全部 Key</option>${options}`;
}

function renderUsage() {
  const usage = state.usage || { items: [], summary: {} };
  const summary = usage.summary || {};
  els.usagePrompt.textContent = number(summary.prompt_tokens);
  els.usageCompletion.textContent = number(summary.completion_tokens);
  els.usageTotal.textContent = number(summary.total_tokens);
  els.usageSuccess.textContent = number(summary.success_count);
  els.usageError.textContent = number(summary.error_count);
  const items = usage.items || [];
  if (items.length === 0) {
    els.usageRows.innerHTML = emptyRow(8, "暂无用量记录");
    return;
  }
  els.usageRows.innerHTML = items
    .map((item) => `<tr>
      <td>${formatDate(item.requested_at)}</td>
      <td>${item.tenant_id}</td>
      <td>${item.api_key_id}</td>
      <td>${escapeHTML(item.model)}</td>
      <td>${number(item.prompt_tokens)}</td>
      <td>${number(item.completion_tokens)}</td>
      <td>${number(item.total_tokens)}</td>
      <td>${usageStatusPill(item.status)}</td>
    </tr>`)
    .join("");
}

function renderOverview() {
  const usage = state.usage || { items: [], summary: {} };
  const summary = usage.summary || {};
  els.metricTenants.textContent = number(state.tenants.length);
  els.metricKeys.textContent = number(state.allKeys.length);
  els.metricTokens.textContent = number(summary.total_tokens);
  els.metricErrors.textContent = number(summary.error_count);

  const recent = usage.items || [];
  if (recent.length === 0) {
    els.recentUsageRows.innerHTML = emptyRow(6, "暂无用量记录");
    return;
  }
  els.recentUsageRows.innerHTML = recent
    .slice(0, 8)
    .map((item) => `<tr>
      <td>${formatDate(item.requested_at)}</td>
      <td>${item.tenant_id}</td>
      <td>${item.api_key_id}</td>
      <td>${escapeHTML(item.model)}</td>
      <td>${number(item.total_tokens)}</td>
      <td>${usageStatusPill(item.status)}</td>
    </tr>`)
    .join("");
}

async function loadModels() {
  const secret = els.playgroundForm.elements.secret.value.trim();
  if (!secret) {
    notify("请输入 API Key Secret", "err");
    return;
  }
  try {
    const models = await gatewayFetch("/v1/models", secret);
    els.playgroundOutput.textContent = JSON.stringify(models, null, 2);
  } catch (error) {
    els.playgroundOutput.textContent = error.message;
  }
}

async function sendPlaygroundRequest(event) {
  event.preventDefault();
  const form = new FormData(event.currentTarget);
  const secret = String(form.get("secret") || "").trim();
  const model = String(form.get("model") || "").trim();
  const prompt = String(form.get("prompt") || "").trim();
  try {
    const result = await gatewayFetch("/v1/chat/completions", secret, {
      method: "POST",
      body: {
        model,
        messages: [{ role: "user", content: prompt }],
      },
    });
    els.playgroundOutput.textContent = JSON.stringify(result, null, 2);
    await loadUsage();
  } catch (error) {
    els.playgroundOutput.textContent = error.message;
  }
}

async function adminFetch(path, options = {}) {
  const payload = await rawFetch(path, {
    ...options,
    headers: {
      Authorization: `Bearer ${state.adminToken}`,
      ...(options.headers || {}),
    },
  });
  return payload && Object.prototype.hasOwnProperty.call(payload, "data") ? payload.data : payload;
}

async function gatewayFetch(path, secret, options = {}) {
  return rawFetch(path, {
    ...options,
    headers: {
      Authorization: `Bearer ${secret}`,
      ...(options.headers || {}),
    },
  });
}

async function rawFetch(path, options = {}) {
  const init = { method: options.method || "GET", headers: options.headers || {} };
  if (options.body !== undefined) {
    init.headers = { "Content-Type": "application/json", ...init.headers };
    init.body = JSON.stringify(options.body);
  }
  const response = await fetch(`${state.apiBase}${path}`, init);
  const contentType = response.headers.get("content-type") || "";
  const payload = contentType.includes("application/json") ? await response.json() : await response.text();
  if (!response.ok) {
    throw new Error(extractError(payload, response.status));
  }
  return payload;
}

function extractError(payload, status) {
  if (payload && payload.error) return `${status} ${payload.error.code}: ${payload.error.message}`;
  if (payload && payload.message) return `${status}: ${payload.message}`;
  return `${status}: request failed`;
}

function copySecret() {
  const value = els.secretValue.textContent;
  if (!value) return;
  navigator.clipboard.writeText(value).then(
    () => notify("Secret 已复制", "ok"),
    () => notify("复制失败，请手动选择", "err"),
  );
}

function notify(message, type = "ok") {
  const div = document.createElement("div");
  div.className = `alert ${type}`;
  div.textContent = message;
  els.alertHost.appendChild(div);
  window.setTimeout(() => div.remove(), 4200);
}

function emptyRow(colspan, message) {
  return `<tr><td class="empty" colspan="${colspan}">${message}</td></tr>`;
}

function statusPill(status) {
  return status === 1 ? '<span class="pill ok">active</span>' : '<span class="pill off">inactive</span>';
}

function usageStatusPill(status) {
  return status === "success" ? '<span class="pill ok">success</span>' : '<span class="pill err">error</span>';
}

function localDateTimeToISO(value) {
  if (!value) return "";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "" : date.toISOString();
}

function isoToLocalInput(value) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const offset = date.getTimezoneOffset() * 60000;
  return new Date(date.getTime() - offset).toISOString().slice(0, 16);
}

function formatDate(value) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return date.toLocaleString();
}

function number(value) {
  return Number(value || 0).toLocaleString();
}

function trimSlash(value) {
  return value.replace(/\/+$/, "");
}

function escapeHTML(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

window.updateTenantStatus = updateTenantStatus;
window.updateKey = updateKey;
window.deleteKey = deleteKey;

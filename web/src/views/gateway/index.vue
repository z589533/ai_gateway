<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  gatewayApi,
  getAdminToken,
  getApiBase,
  health,
  setAdminToken,
  setApiBase,
  type ApiKey,
  type Tenant,
  type UsageRecord,
  type UsageSummary
} from "@/api/gateway";

defineOptions({
  name: "GatewayAdmin"
});

const loading = ref(false);
const apiOnline = ref(false);
const tenants = ref<Tenant[]>([]);
const selectedTenantId = ref<number | undefined>();
const keys = ref<ApiKey[]>([]);
const allKeys = ref<ApiKey[]>([]);
const usage = ref<UsageRecord[]>([]);
const summary = ref<UsageSummary>({
  prompt_tokens: 0,
  completion_tokens: 0,
  total_tokens: 0,
  success_count: 0,
  error_count: 0
});
const secretVisible = ref(false);
const createdSecret = ref("");
const playgroundOutput = ref("");

const settings = reactive({
  apiBase: getApiBase(),
  adminToken: getAdminToken()
});

const tenantForm = reactive({ name: "" });
const keyForm = reactive({
  name: "",
  scopes: ["chat:completions"],
  expires_at: ""
});
const usageFilter = reactive({
  tenant_id: "",
  api_key_id: "",
  from: "",
  to: ""
});
const playground = reactive({
  secret: "",
  model: "gpt-4o-mini",
  prompt: "hi"
});

const metrics = computed(() => [
  { label: "租户", value: tenants.value.length },
  { label: "API Keys", value: allKeys.value.length },
  { label: "Total Tokens", value: summary.value.total_tokens },
  { label: "错误请求", value: summary.value.error_count }
]);

const tenantOptions = computed(() =>
  tenants.value.map(item => ({ label: `${item.name} (#${item.id})`, value: item.id }))
);

const usageKeyOptions = computed(() => {
  const tenantId = usageFilter.tenant_id ? Number(usageFilter.tenant_id) : undefined;
  return allKeys.value
    .filter(item => !tenantId || item.tenant_id === tenantId)
    .map(item => ({ label: `${item.name} (#${item.id})`, value: item.id }));
});

async function bootstrap() {
  loading.value = true;
  try {
    await checkHealth();
    await loadTenants();
    await loadAllKeys();
    await loadKeys();
    await loadUsage();
  } finally {
    loading.value = false;
  }
}

async function checkHealth() {
  try {
    await health();
    apiOnline.value = true;
  } catch {
    apiOnline.value = false;
  }
}

function saveSettings() {
  setApiBase(settings.apiBase);
  setAdminToken(settings.adminToken);
  ElMessage.success("设置已保存");
  bootstrap();
}

async function loadTenants() {
  const data = await gatewayApi.listTenants();
  tenants.value = data.items || [];
  if (!selectedTenantId.value && tenants.value.length > 0) {
    selectedTenantId.value = tenants.value[0].id;
  }
  if (!usageFilter.tenant_id && selectedTenantId.value) {
    usageFilter.tenant_id = String(selectedTenantId.value);
  }
}

async function createTenant() {
  if (!tenantForm.name.trim()) return;
  const tenant = await gatewayApi.createTenant(tenantForm.name.trim());
  tenantForm.name = "";
  selectedTenantId.value = tenant.id;
  usageFilter.tenant_id = String(tenant.id);
  ElMessage.success("租户已创建");
  await bootstrap();
}

async function toggleTenant(row: Tenant) {
  await gatewayApi.updateTenant(row.id, { status: row.status === 1 ? 0 : 1 });
  ElMessage.success("租户状态已更新");
  await loadTenants();
}

async function loadAllKeys() {
  const results = await Promise.all(
    tenants.value.map(tenant =>
      gatewayApi.listKeys(tenant.id).catch(() => ({ items: [], total: 0, page: 1, page_size: 100 }))
    )
  );
  allKeys.value = results.flatMap(item => item.items || []);
}

async function loadKeys() {
  if (!selectedTenantId.value) {
    keys.value = [];
    return;
  }
  const data = await gatewayApi.listKeys(selectedTenantId.value);
  keys.value = data.items || [];
}

async function createKey() {
  if (!selectedTenantId.value) {
    ElMessage.error("请先选择租户");
    return;
  }
  const created = await gatewayApi.createKey(selectedTenantId.value, {
    name: keyForm.name.trim(),
    scopes: keyForm.scopes,
    expires_at: toISO(keyForm.expires_at)
  });
  createdSecret.value = created.secret_key;
  secretVisible.value = true;
  keyForm.name = "";
  keyForm.expires_at = "";
  keyForm.scopes = ["chat:completions"];
  ElMessage.success("API Key 已创建，Secret 只展示一次");
  await loadKeys();
  await loadAllKeys();
}

async function saveKey(row: ApiKey) {
  if (!selectedTenantId.value) return;
  await gatewayApi.updateKey(selectedTenantId.value, row.id, {
    scopes: row.scopes || [],
    status: row.status,
    expires_at: row.expires_at || null
  });
  ElMessage.success("Key 已保存");
  await loadKeys();
  await loadAllKeys();
}

async function deleteKey(row: ApiKey) {
  if (!selectedTenantId.value) return;
  await ElMessageBox.confirm("确认删除这个 Key？历史用量仍会保留。", "删除确认", {
    type: "warning"
  });
  await gatewayApi.deleteKey(selectedTenantId.value, row.id);
  ElMessage.success("Key 已删除");
  await loadKeys();
  await loadAllKeys();
}

async function copySecret() {
  await navigator.clipboard.writeText(createdSecret.value);
  ElMessage.success("Secret 已复制");
}

async function loadUsage() {
  const data = await gatewayApi.queryUsage({
    tenant_id: usageFilter.tenant_id || undefined,
    api_key_id: usageFilter.api_key_id || undefined,
    from: toISO(usageFilter.from),
    to: toISO(usageFilter.to)
  });
  usage.value = data.items || [];
  summary.value = data.summary || summary.value;
}

async function loadModels() {
  if (!playground.secret.trim()) {
    ElMessage.error("请输入 API Key Secret");
    return;
  }
  const data = await gatewayApi.listModels(playground.secret.trim());
  playgroundOutput.value = JSON.stringify(data, null, 2);
}

async function sendChat() {
  if (!playground.secret.trim()) {
    ElMessage.error("请输入 API Key Secret");
    return;
  }
  const data = await gatewayApi.chat(playground.secret.trim(), {
    model: playground.model,
    messages: [{ role: "user", content: playground.prompt }]
  });
  playgroundOutput.value = JSON.stringify(data, null, 2);
  await loadUsage();
}

function toISO(value?: string) {
  if (!value) return undefined;
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? undefined : date.toISOString();
}

function formatDate(value?: string | null) {
  if (!value) return "-";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "-" : date.toLocaleString();
}

onMounted(bootstrap);
</script>

<template>
  <div v-loading="loading" class="gateway-page">
    <section class="hero">
      <div>
        <p class="eyebrow">AI Gateway MVP</p>
        <h1>统一租户 Key、模型代理与用量追踪</h1>
      </div>
      <div class="settings">
        <el-input v-model="settings.apiBase" clearable placeholder="API Base，compose 下可留空" />
        <el-input v-model="settings.adminToken" show-password placeholder="Admin Token" />
        <el-button type="primary" @click="saveSettings">保存设置</el-button>
      </div>
    </section>

    <el-alert
      :title="apiOnline ? '后端 API 在线' : '后端 API 未连接'"
      :type="apiOnline ? 'success' : 'error'"
      show-icon
      :closable="false"
      class="mb16"
    />

    <div class="metric-grid">
      <el-card v-for="item in metrics" :key="item.label" shadow="never" class="metric-card">
        <span>{{ item.label }}</span>
        <strong>{{ item.value.toLocaleString() }}</strong>
      </el-card>
    </div>

    <el-tabs type="border-card" class="workspace">
      <el-tab-pane label="租户管理">
        <div class="toolbar">
          <el-input v-model="tenantForm.name" placeholder="tenant name" />
          <el-button type="primary" @click="createTenant">新增租户</el-button>
          <el-button @click="loadTenants">刷新</el-button>
        </div>
        <el-table :data="tenants" border>
          <el-table-column prop="id" label="ID" width="90" />
          <el-table-column prop="name" label="名称" min-width="180" />
          <el-table-column label="状态" width="120">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'info'">
                {{ row.status === 1 ? "active" : "inactive" }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="创建时间" min-width="180">
            <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="150">
            <template #default="{ row }">
              <el-button size="small" @click="toggleTenant(row)">
                {{ row.status === 1 ? "禁用" : "启用" }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="API Key">
        <div class="toolbar key-toolbar">
          <el-select v-model="selectedTenantId" placeholder="选择租户" @change="loadKeys">
            <el-option v-for="item in tenantOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
          <el-input v-model="keyForm.name" placeholder="Key 名称" />
          <el-date-picker v-model="keyForm.expires_at" type="datetime" value-format="YYYY-MM-DDTHH:mm:ss" placeholder="过期时间" />
          <el-checkbox-group v-model="keyForm.scopes">
            <el-checkbox label="chat:completions" />
            <el-checkbox label="models:read" />
          </el-checkbox-group>
          <el-button type="primary" @click="createKey">创建 Key</el-button>
        </div>
        <el-alert v-if="secretVisible" type="warning" show-icon class="mb16" :closable="false">
          <template #title>
            Secret Key 只展示一次：
            <code>{{ createdSecret }}</code>
            <el-button size="small" class="ml8" @click="copySecret">复制</el-button>
          </template>
        </el-alert>
        <el-table :data="keys" border>
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="name" label="名称" min-width="150" />
          <el-table-column prop="key_prefix" label="Prefix" min-width="140" />
          <el-table-column label="Scopes" min-width="240">
            <template #default="{ row }">
              <el-checkbox-group v-model="row.scopes">
                <el-checkbox label="chat:completions" />
                <el-checkbox label="models:read" />
              </el-checkbox-group>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="150">
            <template #default="{ row }">
              <el-select v-model="row.status">
                <el-option label="enabled" :value="1" />
                <el-option label="disabled" :value="0" />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="过期时间" min-width="190">
            <template #default="{ row }">
              <el-date-picker v-model="row.expires_at" type="datetime" value-format="YYYY-MM-DDTHH:mm:ssZ" placeholder="永不过期" />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="160" fixed="right">
            <template #default="{ row }">
              <el-button size="small" type="primary" @click="saveKey(row)">保存</el-button>
              <el-button size="small" type="danger" @click="deleteKey(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="用量统计">
        <div class="toolbar usage-toolbar">
          <el-select v-model="usageFilter.tenant_id" clearable placeholder="全部租户">
            <el-option v-for="item in tenantOptions" :key="item.value" :label="item.label" :value="String(item.value)" />
          </el-select>
          <el-select v-model="usageFilter.api_key_id" clearable placeholder="全部 Key">
            <el-option v-for="item in usageKeyOptions" :key="item.value" :label="item.label" :value="String(item.value)" />
          </el-select>
          <el-date-picker v-model="usageFilter.from" type="datetime" value-format="YYYY-MM-DDTHH:mm:ss" placeholder="开始时间" />
          <el-date-picker v-model="usageFilter.to" type="datetime" value-format="YYYY-MM-DDTHH:mm:ss" placeholder="结束时间" />
          <el-button type="primary" @click="loadUsage">查询</el-button>
        </div>
        <div class="summary-grid">
          <el-statistic title="Prompt Tokens" :value="summary.prompt_tokens" />
          <el-statistic title="Completion Tokens" :value="summary.completion_tokens" />
          <el-statistic title="Total Tokens" :value="summary.total_tokens" />
          <el-statistic title="Success" :value="summary.success_count" />
          <el-statistic title="Error" :value="summary.error_count" />
        </div>
        <el-table :data="usage" border>
          <el-table-column label="时间" min-width="180">
            <template #default="{ row }">{{ formatDate(row.requested_at) }}</template>
          </el-table-column>
          <el-table-column prop="tenant_id" label="Tenant" width="100" />
          <el-table-column prop="api_key_id" label="Key" width="90" />
          <el-table-column prop="model" label="Model" min-width="150" />
          <el-table-column prop="prompt_tokens" label="Prompt" width="100" />
          <el-table-column prop="completion_tokens" label="Completion" width="120" />
          <el-table-column prop="total_tokens" label="Total" width="100" />
          <el-table-column label="Status" width="120">
            <template #default="{ row }">
              <el-tag :type="row.status === 'success' ? 'success' : 'danger'">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="代理测试">
        <div class="playground">
          <el-form label-position="top">
            <el-form-item label="API Key Secret">
              <el-input v-model="playground.secret" show-password placeholder="sk-ag-..." />
            </el-form-item>
            <el-form-item label="Model">
              <el-input v-model="playground.model" />
            </el-form-item>
            <el-form-item label="Prompt">
              <el-input v-model="playground.prompt" type="textarea" :rows="5" />
            </el-form-item>
            <el-space>
              <el-button @click="loadModels">加载模型</el-button>
              <el-button type="primary" @click="sendChat">发送请求</el-button>
            </el-space>
          </el-form>
          <pre>{{ playgroundOutput }}</pre>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<style scoped>
.gateway-page {
  padding: 20px;
}

.hero {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
  margin-bottom: 16px;
}

.hero h1 {
  margin: 4px 0 0;
  font-size: 28px;
  font-weight: 750;
}

.eyebrow {
  margin: 0;
  color: #0f766e;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.settings {
  display: grid;
  grid-template-columns: 260px 220px auto;
  gap: 10px;
  align-items: center;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
  margin-bottom: 16px;
}

.metric-card span,
.metric-card strong {
  display: block;
}

.metric-card span {
  color: #64748b;
  font-size: 13px;
  font-weight: 700;
}

.metric-card strong {
  margin-top: 8px;
  font-size: 30px;
}

.workspace {
  border-radius: 8px;
}

.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
  margin-bottom: 16px;
}

.toolbar .el-input,
.toolbar .el-select {
  width: 220px;
}

.key-toolbar .el-date-editor,
.usage-toolbar .el-date-editor {
  width: 210px;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}

.summary-grid :deep(.el-statistic) {
  padding: 14px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}

.playground {
  display: grid;
  grid-template-columns: minmax(320px, 420px) 1fr;
  gap: 16px;
}

.playground pre {
  min-height: 320px;
  margin: 0;
  padding: 14px;
  overflow: auto;
  border-radius: 8px;
  background: #111827;
  color: #d1fae5;
}

.mb16 {
  margin-bottom: 16px;
}

.ml8 {
  margin-left: 8px;
}

@media (max-width: 1100px) {
  .hero,
  .playground {
    display: grid;
  }

  .settings,
  .metric-grid,
  .summary-grid {
    grid-template-columns: 1fr;
  }
}
</style>

import axios, { type AxiosRequestConfig } from "axios";

const DEFAULT_ADMIN_TOKEN = "admin-dev-token";

export type Tenant = {
  id: number;
  name: string;
  status: number;
  created_at: string;
  updated_at: string;
};

export type ApiKey = {
  id: number;
  tenant_id: number;
  name: string;
  key_prefix: string;
  scopes: string[];
  status: number;
  expires_at?: string | null;
  created_at: string;
  updated_at: string;
};

export type CreatedApiKey = ApiKey & {
  secret_key: string;
};

export type UsageRecord = {
  id: number;
  tenant_id: number;
  api_key_id: number;
  model: string;
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
  latency_ms: number;
  status: string;
  requested_at: string;
};

export type UsageSummary = {
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
  success_count: number;
  error_count: number;
};

export type PageResult<T> = {
  items: T[];
  total: number;
  page: number;
  page_size: number;
};

export type UsageResult = PageResult<UsageRecord> & {
  summary: UsageSummary;
};

export type ChatCompletionRequest = {
  model: string;
  messages: Array<{ role: string; content: string }>;
};

export function getApiBase() {
  return (localStorage.getItem("ag.apiBase") || "").replace(/\/+$/, "");
}

export function setApiBase(value: string) {
  localStorage.setItem("ag.apiBase", value.replace(/\/+$/, ""));
}

export function getAdminToken() {
  return localStorage.getItem("ag.adminToken") || DEFAULT_ADMIN_TOKEN;
}

export function setAdminToken(value: string) {
  localStorage.setItem("ag.adminToken", value || DEFAULT_ADMIN_TOKEN);
}

async function request<T>(config: AxiosRequestConfig): Promise<T> {
  const response = await axios.request({
    baseURL: getApiBase(),
    timeout: 10000,
    ...config,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${getAdminToken()}`,
      ...(config.headers || {})
    }
  });
  const payload = response.data;
  return Object.prototype.hasOwnProperty.call(payload, "data")
    ? payload.data
    : payload;
}

export function health() {
  return axios.get(`${getApiBase()}/health`, { timeout: 5000 });
}

export const gatewayApi = {
  listTenants: () =>
    request<PageResult<Tenant>>({
      method: "GET",
      url: "/api/v1/tenants",
      params: { page_size: 100 }
    }),
  createTenant: (name: string) =>
    request<Tenant>({
      method: "POST",
      url: "/api/v1/tenants",
      data: { name }
    }),
  updateTenant: (tenantId: number, data: Partial<Pick<Tenant, "name" | "status">>) =>
    request<Tenant>({
      method: "PATCH",
      url: `/api/v1/tenants/${tenantId}`,
      data
    }),
  listKeys: (tenantId: number) =>
    request<PageResult<ApiKey>>({
      method: "GET",
      url: `/api/v1/tenants/${tenantId}/keys`,
      params: { page_size: 100 }
    }),
  createKey: (
    tenantId: number,
    data: { name: string; scopes: string[]; expires_at?: string | null }
  ) =>
    request<CreatedApiKey>({
      method: "POST",
      url: `/api/v1/tenants/${tenantId}/keys`,
      data
    }),
  updateKey: (
    tenantId: number,
    keyId: number,
    data: { scopes?: string[]; status?: number; expires_at?: string | null }
  ) =>
    request<ApiKey>({
      method: "PATCH",
      url: `/api/v1/tenants/${tenantId}/keys/${keyId}`,
      data
    }),
  deleteKey: (tenantId: number, keyId: number) =>
    request<{ deleted: boolean }>({
      method: "DELETE",
      url: `/api/v1/tenants/${tenantId}/keys/${keyId}`
    }),
  queryUsage: (params: Record<string, string | number | undefined>) =>
    request<UsageResult>({
      method: "GET",
      url: "/api/v1/usage",
      params: { page_size: 100, ...params }
    }),
  listModels: (secret: string) =>
    axios
      .get(`${getApiBase()}/v1/models`, {
        headers: { Authorization: `Bearer ${secret}` }
      })
      .then(res => res.data),
  chat: (secret: string, data: ChatCompletionRequest) =>
    axios
      .post(`${getApiBase()}/v1/chat/completions`, data, {
        headers: { Authorization: `Bearer ${secret}` }
      })
      .then(res => res.data)
};

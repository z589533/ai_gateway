import axios, { type AxiosRequestConfig } from "axios";

export const DEFAULT_MODEL = "gpt5.5";
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

const SECRET_STORE_KEY = "ag.keySecrets";
export const API_KEY_PREFIX = "sk-ag-";
export const API_KEY_FULL_LENGTH = API_KEY_PREFIX.length + 48;

export function saveKeySecret(keyId: number, secret: string) {
  const map = getKeySecrets();
  map[String(keyId)] = secret;
  localStorage.setItem(SECRET_STORE_KEY, JSON.stringify(map));
}

export function getKeySecrets(): Record<string, string> {
  try {
    return JSON.parse(localStorage.getItem(SECRET_STORE_KEY) || "{}") as Record<string, string>;
  } catch {
    return {};
  }
}

export function getKeySecret(keyId: number) {
  return getKeySecrets()[String(keyId)];
}

/** Prefix 仅用于展示，完整 Secret 为 sk-ag- + 48 位 hex */
export function isLikelyKeyPrefix(value: string) {
  const secret = value.trim();
  return secret.startsWith(API_KEY_PREFIX) && secret.length < API_KEY_FULL_LENGTH;
}

export function formatSecretHint(prefix: string) {
  return `完整 Secret 约 ${API_KEY_FULL_LENGTH} 位，列表 Prefix「${prefix}」不可用于调用`;
}

function extractErrorMessage(error: unknown) {
  if (axios.isAxiosError(error)) {
    const payload = error.response?.data as { message?: string } | undefined;
    return payload?.message || error.message || "请求失败";
  }
  if (error instanceof Error) return error.message;
  return "请求失败";
}

async function request<T>(config: AxiosRequestConfig): Promise<T> {
  try {
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
  } catch (error) {
    throw new Error(extractErrorMessage(error));
  }
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
  listModels: async (secret: string) => {
    try {
      const res = await axios.get(`${getApiBase()}/v1/models`, {
        headers: { Authorization: `Bearer ${secret}` }
      });
      return res.data;
    } catch (error) {
      throw new Error(extractErrorMessage(error));
    }
  },
  chat: async (secret: string, data: ChatCompletionRequest) => {
    try {
      const res = await axios.post(`${getApiBase()}/v1/chat/completions`, data, {
        headers: {
          Authorization: `Bearer ${secret}`,
          "Content-Type": "application/json"
        }
      });
      return res.data;
    } catch (error) {
      if (axios.isAxiosError(error)) {
        const payload = error.response?.data as { error?: { message?: string }; message?: string } | undefined;
        throw new Error(payload?.error?.message || payload?.message || extractErrorMessage(error));
      }
      throw new Error(extractErrorMessage(error));
    }
  }
};

const TOKEN_KEY = "csa_token";

export function getToken(): string {
  return localStorage.getItem(TOKEN_KEY) || "";
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token);
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const token = getToken();
  const res = await fetch(path, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
      ...options?.headers,
    },
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(body.error || `HTTP ${res.status}`);
  }
  return res.json();
}

// --- Subscriptions ---

export interface SubInfo {
  id: string;
  name: string;
  url: string;
  proxy_count: number;
  updated_at: string;
}

export function listSubscriptions() {
  return request<{ subscriptions: SubInfo[] }>("/api/subscriptions");
}

export function addSubscription(name: string, url: string) {
  return request<{ id: string; message: string; proxy_count: number }>(
    "/api/subscriptions",
    { method: "POST", body: JSON.stringify({ name, url }) }
  );
}

export function deleteSubscription(id: string) {
  return request<{ message: string }>(`/api/subscriptions/${id}`, {
    method: "DELETE",
  });
}

export function refreshAllSubscriptions() {
  return request<{ message: string; proxy_count: number }>(
    "/api/subscriptions/refresh",
    { method: "POST" }
  );
}

export function refreshOneSubscription(id: string) {
  return request<{ message: string; proxy_count: number }>(
    `/api/subscriptions/${id}/refresh`,
    { method: "POST" }
  );
}

// --- Proxies ---

export interface ProxyGroup {
  name: string;
  type: string;
  now: string;
  all: string[];
}

export function listProxies() {
  return request<{ proxies: Record<string, ProxyGroup> }>("/api/proxies");
}

export function switchProxy(group: string, name: string) {
  return request<{ message: string }>(
    `/api/proxies/${encodeURIComponent(group)}/${encodeURIComponent(name)}`,
    { method: "PUT" }
  );
}

export function testDelay(name: string) {
  return request<{ delay?: number; message?: string }>(
    `/api/proxies/${encodeURIComponent(name)}/delay`
  );
}

// --- Status & Health ---

export interface ServiceStatus {
  mihomo_running: boolean;
  controller: string;
}

export function getStatus() {
  return request<ServiceStatus>("/api/status");
}

export interface HealthStatus {
  blacklist_count: number;
  blacklist: { name: string; since: string }[];
  checking: boolean;
  last_check_at?: string;
  last_check_cost?: string;
}

export function getHealth() {
  return request<HealthStatus>("/api/health");
}

export function triggerHealthCheck() {
  return request<{ message: string }>("/api/health/check", { method: "POST" });
}

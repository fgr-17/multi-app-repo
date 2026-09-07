import type { GreetingRecord, User } from "./types";

export function defaultApiUrl(): string {
  return (
    (typeof process !== "undefined" && process.env.EXPO_PUBLIC_API_URL) ||
    (typeof process !== "undefined" && process.env.NEXT_PUBLIC_API_URL) ||
    (typeof process !== "undefined" && process.env.API_URL) ||
    "http://localhost:8080"
  );
}

export function apiRoot(apiBaseUrl: string): string {
  return apiBaseUrl.replace(/\/$/, "");
}

export async function fetchHello(
  apiBaseUrl: string = defaultApiUrl(),
): Promise<GreetingRecord> {
  const url = `${apiRoot(apiBaseUrl)}/api/hello`;
  const res = await fetch(url);
  if (!res.ok) {
    throw new Error(`API ${res.status} al pedir ${url}`);
  }
  return (await res.json()) as GreetingRecord;
}

export async function putHello(
  record: Pick<GreetingRecord, "name" | "updatedAt" | "updatedBy">,
  apiBaseUrl: string = defaultApiUrl(),
): Promise<{ record: GreetingRecord; accepted: boolean }> {
  const url = `${apiRoot(apiBaseUrl)}/api/hello`;
  const res = await fetch(url, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(record),
  });
  if (!res.ok && res.status !== 409) {
    throw new Error(`API ${res.status} al guardar ${url}`);
  }
  return {
    record: (await res.json()) as GreetingRecord,
    accepted: res.status === 200,
  };
}

export async function fetchUsers(
  apiBaseUrl: string = defaultApiUrl(),
): Promise<User[]> {
  const url = `${apiRoot(apiBaseUrl)}/api/users`;
  const res = await fetch(url);
  if (!res.ok) {
    throw new Error(`API ${res.status} al pedir ${url}`);
  }
  return (await res.json()) as User[];
}

export async function pingHealth(apiBaseUrl: string = defaultApiUrl()): Promise<boolean> {
  const ctrl = new AbortController();
  const timer = setTimeout(() => ctrl.abort(), 2500);
  try {
    const res = await fetch(`${apiRoot(apiBaseUrl)}/health`, { signal: ctrl.signal });
    return res.ok;
  } catch {
    return false;
  } finally {
    clearTimeout(timer);
  }
}

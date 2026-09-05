export type HelloResponse = {
  name: string;
};

export function defaultApiUrl(): string {
  return (
    (typeof process !== "undefined" && process.env.EXPO_PUBLIC_API_URL) ||
    (typeof process !== "undefined" && process.env.NEXT_PUBLIC_API_URL) ||
    (typeof process !== "undefined" && process.env.API_URL) ||
    "http://localhost:8080"
  );
}

export async function fetchHello(
  apiBaseUrl: string = defaultApiUrl(),
): Promise<HelloResponse> {
  const url = `${apiBaseUrl.replace(/\/$/, "")}/api/hello`;
  const res = await fetch(url);
  if (!res.ok) {
    throw new Error(`API ${res.status} al pedir ${url}`);
  }
  return (await res.json()) as HelloResponse;
}

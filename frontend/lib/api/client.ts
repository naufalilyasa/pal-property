import { normalizeApiResponse, type NormalizedApiResponse } from "@/lib/api/envelope";

const DEFAULT_API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? process.env.API_BASE_URL ?? "http://127.0.0.1:8080";

export type ApiRequestOptions = Omit<RequestInit, "credentials"> & {
  baseUrl?: string;
  fetch?: typeof fetch;
};

function getApiBaseUrl(baseUrl?: string): string {
  return (baseUrl ?? DEFAULT_API_BASE_URL).replace(/\/$/, "");
}

function buildApiUrl(path: string, baseUrl?: string): string {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  return `${getApiBaseUrl(baseUrl)}${normalizedPath}`;
}

export async function apiRequest<T>(
  path: string,
  options: ApiRequestOptions = {},
): Promise<NormalizedApiResponse<T>> {
  const { baseUrl, fetch: fetchImpl = fetch, headers, ...init } = options;
  const requestHeaders = new Headers(headers);

  if (!requestHeaders.has("Accept")) {
    requestHeaders.set("Accept", "application/json");
  }

  const response = await fetchImpl(buildApiUrl(path, baseUrl), {
    ...init,
    credentials: "include",
    headers: requestHeaders,
  });

  return normalizeApiResponse<T>(response);
}

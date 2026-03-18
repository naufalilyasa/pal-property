import "server-only";

import { normalizeApiResponse, type NormalizedApiResponse } from "@/lib/api/envelope";
import { buildApiUrl } from "@/lib/api/shared";
import { serverEnv } from "@/lib/env/server";

type ServerFetchOptions = RequestInit & {
  baseUrl?: string;
  cookieHeader?: string;
  fetch?: typeof fetch;
};

export async function serverFetch<T>(
  path: string,
  options: ServerFetchOptions = {},
): Promise<NormalizedApiResponse<T>> {
  const { baseUrl = (serverEnv.API_BASE_URL || serverEnv.NEXT_PUBLIC_API_BASE_URL).replace(/\/$/, ""), cookieHeader, fetch: fetchImpl = fetch, headers, ...init } = options;
  const requestHeaders = new Headers(headers);

  if (!requestHeaders.has("Accept")) {
    requestHeaders.set("Accept", "application/json");
  }

  if (cookieHeader) {
    requestHeaders.set("Cookie", cookieHeader);
  }

  const response = await fetchImpl(buildApiUrl(path, baseUrl), {
    ...init,
    credentials: "include",
    headers: requestHeaders,
  });

  return normalizeApiResponse<T>(response);
}

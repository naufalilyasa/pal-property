import { normalizeApiResponse, type NormalizedApiResponse } from "@/lib/api/envelope";
import { buildApiUrl, getBrowserApiBaseUrl } from "@/lib/api/shared";

type BrowserFetchOptions = Omit<RequestInit, "credentials"> & {
  baseUrl?: string;
  fetch?: typeof fetch;
  retryOnUnauthorized?: boolean;
};

async function refreshAccessToken(baseUrl: string, fetchImpl: typeof fetch): Promise<boolean> {
  const response = await fetchImpl(buildApiUrl("/auth/refresh", baseUrl), {
    method: "POST",
    credentials: "include",
    headers: {
      Accept: "application/json",
    },
  });

  return response.ok;
}

export async function browserFetch<T>(
  path: string,
  options: BrowserFetchOptions = {},
): Promise<NormalizedApiResponse<T>> {
  const { baseUrl = getBrowserApiBaseUrl(), fetch: fetchImpl = fetch, headers, retryOnUnauthorized = true, ...init } = options;
  const requestHeaders = new Headers(headers);

  if (!requestHeaders.has("Accept")) {
    requestHeaders.set("Accept", "application/json");
  }

  const response = await fetchImpl(buildApiUrl(path, baseUrl), {
    ...init,
    credentials: "include",
    headers: requestHeaders,
  });

  if (response.status === 401 && retryOnUnauthorized) {
    const refreshed = await refreshAccessToken(baseUrl, fetchImpl);

    if (refreshed) {
      return browserFetch<T>(path, {
        ...options,
        retryOnUnauthorized: false,
      });
    }
  }

  return normalizeApiResponse<T>(response);
}

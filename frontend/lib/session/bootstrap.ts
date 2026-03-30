import { browserFetch } from "@/lib/api/browser-fetch";
import { ApiError } from "@/lib/api/envelope";
import { serverFetch } from "@/lib/api/server-fetch";
import type { SellerCapabilityInfo } from "@/features/auth/auth-destination";

export type SellerSessionUser = {
  id: string;
  name: string;
  email: string;
  avatar_url?: string | null;
  role: string;
  created_at: string;
  seller_capabilities?: SellerCapabilityInfo;
};

export type SellerSession =
  | {
      status: "authenticated";
      user: SellerSessionUser;
      traceId: string;
    }
  | {
      status: "unauthenticated";
      user: null;
    };

export type BootstrapSellerSessionOptions = {
  baseUrl?: string;
  cookieHeader?: string;
  fetch?: typeof fetch;
};

export async function bootstrapSellerSession(
  options: BootstrapSellerSessionOptions = {},
): Promise<SellerSession> {
  const headers = new Headers();

  if (!headers.has("Accept")) {
    headers.set("Accept", "application/json");
  }

  if (options.cookieHeader) {
    headers.set("Cookie", options.cookieHeader);
  }

  try {
    const response = options.cookieHeader
      ? await fetchSellerSessionOnServer(options, headers)
      : await browserFetch<SellerSessionUser>("/auth/me", {
          method: "GET",
          cache: "no-store",
          baseUrl: options.baseUrl,
          fetch: options.fetch,
          headers,
        });

    return {
      status: "authenticated",
      user: response.data,
      traceId: response.traceId,
    };
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      return {
        status: "unauthenticated",
        user: null,
      };
    }

    throw error;
  }
}

async function fetchSellerSessionOnServer(
  options: BootstrapSellerSessionOptions,
  headers: Headers,
) {
  return serverFetch<SellerSessionUser>("/auth/me", {
    method: "GET",
    cache: "no-store",
    baseUrl: options.baseUrl,
    fetch: options.fetch,
    headers,
    cookieHeader: options.cookieHeader,
  });
}

import { apiRequest } from "@/lib/api/client";
import { ApiError } from "@/lib/api/envelope";

export type SellerSessionUser = {
  id: string;
  name: string;
  email: string;
  avatar_url?: string | null;
  role: string;
  created_at: string;
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

  if (options.cookieHeader) {
    headers.set("Cookie", options.cookieHeader);
  }

  try {
    const response = await apiRequest<SellerSessionUser>("/auth/me", {
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

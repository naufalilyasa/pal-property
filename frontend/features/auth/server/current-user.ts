import "server-only";

import type { SellerCapabilityInfo } from "@/features/auth/auth-destination";
import { serverFetch } from "@/lib/api/server-fetch";
import { ApiError } from "@/lib/api/envelope";
import { getRequestCookieHeader } from "@/lib/server/cookies";

export type CurrentUser = {
  id: string;
  name: string;
  email: string;
  avatar_url?: string | null;
  role: string;
  seller_capabilities: SellerCapabilityInfo;
  created_at: string;
};

export async function getOptionalUser(): Promise<CurrentUser | null> {
  try {
    const response = await serverFetch<CurrentUser>("/auth/me", {
      method: "GET",
      cache: "no-store",
      cookieHeader: await getRequestCookieHeader(),
    });

    return response.data;
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      return null;
    }

    throw error;
  }
}

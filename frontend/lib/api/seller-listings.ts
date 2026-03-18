import { ApiError } from "@/lib/api/envelope";
import { browserFetch } from "@/lib/api/browser-fetch";
import { serverFetch } from "@/lib/api/server-fetch";

export type SellerListingImage = {
  id: string;
  url: string;
  format?: string | null;
  bytes?: number | null;
  width?: number | null;
  height?: number | null;
  original_filename?: string | null;
  is_primary: boolean;
  sort_order: number;
  created_at: string;
};

export type SellerListingCategory = {
  id: string;
  name: string;
  slug: string;
  icon_url?: string | null;
};

export type SellerListing = {
  id: string;
  user_id: string;
  category_id?: string | null;
  category?: SellerListingCategory | null;
  title: string;
  slug: string;
  description?: string | null;
  price: number;
  currency: string;
  location_city?: string | null;
  location_district?: string | null;
  address_detail?: string | null;
  status: string;
  is_featured: boolean;
  specifications: unknown;
  view_count: number;
  images: SellerListingImage[];
  created_at: string;
  updated_at: string;
};

export type SellerListingsPage = {
  data: SellerListing[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
};

export type GetSellerListingsOptions = {
  baseUrl?: string;
  cookieHeader?: string;
  fetch?: typeof fetch;
};

export async function getSellerListings(
  options: GetSellerListingsOptions = {},
): Promise<SellerListingsPage> {
  const response = await getSellerListingsResponse(options);

  return response.data;
}

export async function getSellerListingById(
  listingId: string,
  options: GetSellerListingsOptions = {},
): Promise<SellerListing> {
  const response = await getSellerListingsResponse(options, { limit: 1000 });
  const listing = response.data.data.find((candidate) => candidate.id === listingId);

  if (!listing) {
    throw new ApiError("Listing not found.", {
      status: 404,
      traceId: response.traceId,
    });
  }

  return listing;
}

async function getSellerListingsResponse(
  options: GetSellerListingsOptions,
  params?: Record<string, string | number>,
) {
  const headers = new Headers();
  const searchParams = new URLSearchParams();

  if (options.cookieHeader) {
    headers.set("Cookie", options.cookieHeader);
  }

  if (params) {
    for (const [key, value] of Object.entries(params)) {
      searchParams.set(key, String(value));
    }
  }

  const path = searchParams.size > 0 ? `/auth/me/listings?${searchParams.toString()}` : "/auth/me/listings";

  if (options.cookieHeader) {
    return serverFetch<SellerListingsPage>(path, {
      method: "GET",
      cache: "no-store",
      baseUrl: options.baseUrl,
      fetch: options.fetch,
      headers,
      cookieHeader: options.cookieHeader,
    });
  }

  return browserFetch<SellerListingsPage>(path, {
    method: "GET",
    cache: "no-store",
    baseUrl: options.baseUrl,
    fetch: options.fetch,
    headers,
  });
}

export function isUnauthenticatedSellerListingsError(error: unknown): boolean {
  return error instanceof ApiError && error.status === 401;
}

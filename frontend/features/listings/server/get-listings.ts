import "server-only";

import { serverFetch } from "@/lib/api/server-fetch";
import type { SellerListingsPage } from "@/lib/api/seller-listings";

export type ListingQuery = {
  page?: string;
  limit?: string;
  city?: string;
  category_id?: string;
  price_min?: string;
  price_max?: string;
  status?: string;
};

const DEFAULT_PAGE = "1";
const DEFAULT_LIMIT = "10";

export async function getListings(query: ListingQuery = {}) {
  const searchParams = new URLSearchParams();
  searchParams.set("page", query.page ?? DEFAULT_PAGE);
  searchParams.set("limit", query.limit ?? DEFAULT_LIMIT);

  for (const key of ["city", "category_id", "price_min", "price_max", "status"] as const) {
    const value = query[key];
    if (value) {
      searchParams.set(key, value);
    }
  }

  const response = await serverFetch<SellerListingsPage>(`/api/listings?${searchParams.toString()}`, {
    method: "GET",
    cache: "no-store",
  });

  return response.data;
}

import "server-only";

import { serverFetch } from "@/lib/api/server-fetch";

export type SearchListingsQuery = {
  q?: string;
  transaction_type?: string;
  category_id?: string;
  location_province?: string;
  location_city?: string;
  price_min?: string;
  price_max?: string;
  sort?: string;
  page?: string;
  limit?: string;
};

export type SearchCategory = {
  id: string;
  name: string;
  slug: string;
};

export type SearchListingCard = {
  id: string;
  category_id?: string;
  category?: SearchCategory;
  title: string;
  slug: string;
  description_excerpt?: string;
  transaction_type: string;
  price: number;
  currency: string;
  location_province?: string;
  location_city?: string;
  location_district?: string;
  location_village?: string;
  bedroom_count?: number;
  bathroom_count?: number;
  land_area_sqm?: number;
  building_area_sqm?: number;
  status: string;
  is_featured: boolean;
  primary_image_url?: string;
  image_urls?: string[];
  created_at: string;
  updated_at: string;
};

export type SearchListingsPage = {
  items: SearchListingCard[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
};

const DEFAULT_PAGE = "1";
const DEFAULT_LIMIT = "12";

export async function getSearchListings(query: SearchListingsQuery = {}) {
  const searchParams = new URLSearchParams();
  searchParams.set("page", query.page ?? DEFAULT_PAGE);
  searchParams.set("limit", query.limit ?? DEFAULT_LIMIT);

  for (const key of [
    "q",
    "transaction_type",
    "category_id",
    "location_province",
    "location_city",
    "price_min",
    "price_max",
    "sort",
  ] as const) {
    const value = query[key];
    if (value) {
      searchParams.set(key, value);
    }
  }

  const response = await serverFetch<SearchListingsPage>(`/api/search/listings?${searchParams.toString()}`, {
    method: "GET",
    cache: "no-store",
  });

  return response.data;
}

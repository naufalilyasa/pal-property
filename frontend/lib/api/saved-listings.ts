import { browserFetch } from "@/lib/api/browser-fetch";
import { serverFetch } from "@/lib/api/server-fetch";
import type { SellerListing, SellerListingsPage } from "@/lib/api/seller-listings";

export {
  removeSavedListing,
  saveListing,
  type SaveListingOptions,
  type SavedListingToggle,
} from "@/lib/api/saved-listings-client";

const SAVED_LISTINGS_BASE = "/api/me/saved-listings";
const SAVED_LISTINGS_CONTAINS = `${SAVED_LISTINGS_BASE}/contains`;

export type SavedListing = SellerListing & {
  transaction_type: string;
};

export type SavedListingsPage = Omit<SellerListingsPage, "data"> & {
  data: SavedListing[];
};

export type SavedListingsQuery = {
  page?: string;
  limit?: string;
};

export type SavedListingsOptions = {
  baseUrl?: string;
  fetch?: typeof fetch;
  cookieHeader?: string;
};

export type SavedListingContainsResult = {
  listingIds: string[];
};

type SavedListingContainsDto = {
  listing_ids: string[];
};

export async function getSavedListings(
  query: SavedListingsQuery = {},
  options: SavedListingsOptions = {},
): Promise<SavedListingsPage> {
  const response = await fetchSavedListings(query, options);

  return response.data;
}

export async function getSavedListingIds(
  listingIds: string[],
  options: SavedListingsOptions = {},
): Promise<SavedListingContainsResult> {
  if (listingIds.length === 0) {
    return { listingIds: [] };
  }

  const params = new URLSearchParams();
  params.set("listing_ids", listingIds.join(","));
  const response = await fetchSavedListingContains(`${SAVED_LISTINGS_CONTAINS}?${params.toString()}`, options);

  return { listingIds: response.data.listing_ids ?? [] };
}

async function fetchSavedListings(query: SavedListingsQuery, options: SavedListingsOptions) {
  const searchParams = new URLSearchParams();
  if (query.page) {
    searchParams.set("page", query.page);
  }
  if (query.limit) {
    searchParams.set("limit", query.limit);
  }

  const path = searchParams.toString()
    ? `${SAVED_LISTINGS_BASE}?${searchParams.toString()}`
    : SAVED_LISTINGS_BASE;

  return requestAuthFetch<SavedListingsPage>(path, options);
}

function fetchSavedListingContains(path: string, options: SavedListingsOptions) {
  return requestAuthFetch<SavedListingContainsDto>(path, options);
}

async function requestAuthFetch<T>(path: string, options: SavedListingsOptions) {
  const headers = new Headers();

  const fetchOptions = {
    method: "GET",
    cache: "no-store",
    baseUrl: options.baseUrl,
    fetch: options.fetch,
    headers,
  } as const;

  if (options.cookieHeader) {
    return serverFetch<T>(path, {
      ...fetchOptions,
      cookieHeader: options.cookieHeader,
    });
  }

  return browserFetch<T>(path, fetchOptions);
}

import { browserFetch } from "@/lib/api/browser-fetch";
import { ApiError } from "@/lib/api/envelope";

export type ListingSpecifications = {
  bedrooms: number;
  bathrooms: number;
  land_area_sqm: number;
  building_area_sqm: number;
};

export type ListingStatus = "active" | "inactive" | "sold" | "draft" | "archived";
export type ListingTransactionType = "sale" | "rent";

export type ListingCategory = {
  id: string;
  name: string;
  slug: string;
  parent_id?: string | null;
  icon_url?: string | null;
  children?: Array<{
    id: string;
    name: string;
    slug: string;
    icon_url?: string | null;
  }>;
};

export type ListingCategoryOption = {
  id: string;
  name: string;
  slug: string;
  label: string;
};

export type ListingFormRequest = {
  category_id: string | null;
  title: string;
  description: string | null;
  transaction_type: ListingTransactionType;
  price: number;
  currency: string;
  is_negotiable: boolean;
  special_offers: string[];
  location_province: string | null;
  location_city: string | null;
  location_district: string | null;
  address_detail: string | null;
  latitude: number | null;
  longitude: number | null;
  bedroom_count: number | null;
  bathroom_count: number | null;
  floor_count: number | null;
  carport_capacity: number | null;
  land_area_sqm: number | null;
  building_area_sqm: number | null;
  certificate_type: string | null;
  condition: string | null;
  furnishing: string | null;
  electrical_power_va: number | null;
  facing_direction: string | null;
  year_built: number | null;
  facilities: string[];
  status: ListingStatus;
  specifications: ListingSpecifications;
};

export type ListingRecord = {
  id: string;
  user_id: string;
  category_id?: string | null;
  category?: {
    id: string;
    name: string;
    slug: string;
    icon_url?: string | null;
  } | null;
  title: string;
  slug: string;
  description?: string | null;
  transaction_type?: ListingTransactionType;
  price: number;
  currency: string;
  is_negotiable?: boolean;
  special_offers?: string[] | null;
  location_province?: string | null;
  location_city?: string | null;
  location_district?: string | null;
  address_detail?: string | null;
  latitude?: number | null;
  longitude?: number | null;
  bedroom_count?: number | null;
  bathroom_count?: number | null;
  floor_count?: number | null;
  carport_capacity?: number | null;
  land_area_sqm?: number | null;
  building_area_sqm?: number | null;
  certificate_type?: string | null;
  condition?: string | null;
  furnishing?: string | null;
  electrical_power_va?: number | null;
  facing_direction?: string | null;
  year_built?: number | null;
  facilities?: string[] | null;
  status: string;
  is_featured: boolean;
  specifications: unknown;
  view_count: number;
  images: ListingImageRecord[];
  created_at: string;
  updated_at: string;
};

export type ListingImageRecord = {
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

export type ListingFormApiOptions = {
  baseUrl?: string;
  fetch?: typeof fetch;
};

export async function getListingCategories(
  options: ListingFormApiOptions = {},
): Promise<ListingCategoryOption[]> {
  const response = await browserFetch<ListingCategory[]>("/api/categories", {
    method: "GET",
    cache: "no-store",
    baseUrl: options.baseUrl,
    fetch: options.fetch,
  });

  return flattenCategoryOptions(response.data);
}

export async function getListingById(
  listingId: string,
  options: ListingFormApiOptions = {},
): Promise<ListingRecord> {
  const response = await browserFetch<ListingRecord>(`/api/listings/${listingId}`, {
    method: "GET",
    cache: "no-store",
    baseUrl: options.baseUrl,
    fetch: options.fetch,
  });

  return response.data;
}

export async function createSellerListing(
  payload: ListingFormRequest,
  options: ListingFormApiOptions = {},
): Promise<ListingRecord> {
  const response = await browserFetch<ListingRecord>("/api/listings", {
    method: "POST",
    cache: "no-store",
    baseUrl: options.baseUrl,
    fetch: options.fetch,
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });

  return response.data;
}

export async function updateSellerListing(
  listingId: string,
  payload: ListingFormRequest,
  options: ListingFormApiOptions = {},
): Promise<ListingRecord> {
  const response = await browserFetch<ListingRecord>(`/api/listings/${listingId}`, {
    method: "PUT",
    cache: "no-store",
    baseUrl: options.baseUrl,
    fetch: options.fetch,
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });

  return response.data;
}

export async function uploadListingImage(
  listingId: string,
  file: File,
  options: ListingFormApiOptions = {},
): Promise<ListingRecord> {
  const body = new FormData();
  body.set("file", file);

  const response = await browserFetch<ListingRecord>(`/api/listings/${listingId}/images`, {
    method: "POST",
    cache: "no-store",
    baseUrl: options.baseUrl,
    fetch: options.fetch,
    body,
  });

  return response.data;
}

export async function deleteListingImage(
  listingId: string,
  imageId: string,
  options: ListingFormApiOptions = {},
): Promise<ListingRecord> {
  const response = await browserFetch<ListingRecord>(`/api/listings/${listingId}/images/${imageId}`, {
    method: "DELETE",
    cache: "no-store",
    baseUrl: options.baseUrl,
    fetch: options.fetch,
  });

  return response.data;
}

export async function setPrimaryListingImage(
  listingId: string,
  imageId: string,
  options: ListingFormApiOptions = {},
): Promise<ListingRecord> {
  const response = await browserFetch<ListingRecord>(`/api/listings/${listingId}/images/${imageId}/primary`, {
    method: "PATCH",
    cache: "no-store",
    baseUrl: options.baseUrl,
    fetch: options.fetch,
  });

  return response.data;
}

export async function reorderListingImages(
  listingId: string,
  orderedImageIds: string[],
  options: ListingFormApiOptions = {},
): Promise<ListingRecord> {
  const response = await browserFetch<ListingRecord>(`/api/listings/${listingId}/images/reorder`, {
    method: "PATCH",
    cache: "no-store",
    baseUrl: options.baseUrl,
    fetch: options.fetch,
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ ordered_image_ids: orderedImageIds }),
  });

  return response.data;
}

export function getDefaultSpecifications(): ListingSpecifications {
  return {
    bedrooms: 0,
    bathrooms: 0,
    land_area_sqm: 0,
    building_area_sqm: 0,
  };
}

export function parseListingSpecifications(value: unknown): ListingSpecifications {
  if (!value || typeof value !== "object") {
    return getDefaultSpecifications();
  }

  const candidate = value as Partial<Record<keyof ListingSpecifications, unknown>>;

  return {
    bedrooms: normalizeInteger(candidate.bedrooms),
    bathrooms: normalizeInteger(candidate.bathrooms),
    land_area_sqm: normalizeInteger(candidate.land_area_sqm),
    building_area_sqm: normalizeInteger(candidate.building_area_sqm),
  };
}

export function parseStringList(value: unknown): string[] {
  if (!Array.isArray(value)) {
    return [];
  }

  return value.filter((item): item is string => typeof item === "string" && item.trim().length > 0);
}

export function formatListingFormError(error: unknown): string {
  if (error instanceof ApiError) {
    return error.traceId ? `${error.message} (trace ${error.traceId})` : error.message;
  }

  if (error instanceof Error && error.message) {
    return error.message;
  }

  return "We could not complete the listing request.";
}

function flattenCategoryOptions(categories: ListingCategory[]): ListingCategoryOption[] {
  return categories.flatMap((category) => {
    const rootOption: ListingCategoryOption = {
      id: category.id,
      name: category.name,
      slug: category.slug,
      label: category.name,
    };

    const childOptions = (category.children ?? []).map((child) => ({
      id: child.id,
      name: child.name,
      slug: child.slug,
      label: `${category.name} / ${child.name}`,
    }));

    return [rootOption, ...childOptions];
  });
}

function normalizeInteger(value: unknown): number {
  if (typeof value === "number" && Number.isFinite(value)) {
    return Math.max(0, Math.trunc(value));
  }

  return 0;
}

export function normalizeNullableNumber(value: string | undefined) {
  const parsed = Number.parseFloat(value ?? "");
  return Number.isFinite(parsed) ? parsed : null;
}

export function normalizeOptionalIntegerOrNull(value: string | undefined) {
  const parsed = Number.parseInt(value ?? "", 10);
  return !Number.isNaN(parsed) && parsed >= 0 ? parsed : null;
}

export function normalizeStringList(value: string | undefined) {
  return (value ?? "")
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
}

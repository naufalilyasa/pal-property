import "server-only";

import { serverFetch } from "@/lib/api/server-fetch";
import type { ListingRecord } from "@/lib/api/listing-form";

export async function getListingBySlug(slug: string) {
  const response = await serverFetch<ListingRecord>(`/api/listings/slug/${slug}`, {
    method: "GET",
    cache: "no-store",
  });

  return response.data;
}

import "server-only";

import { serverFetch } from "@/lib/api/server-fetch";
import type { ListingCategory } from "@/lib/api/listing-form";

export async function getCategories() {
  const response = await serverFetch<ListingCategory[]>("/api/categories", {
    method: "GET",
    cache: "no-store",
  });

  return response.data;
}

import "server-only";

import { getSavedListings, type SavedListingsQuery } from "@/lib/api/saved-listings";
import { getRequestCookieHeader } from "@/lib/server/cookies";

export async function getSavedListingsPage(query: SavedListingsQuery = {}) {
  const cookieHeader = await getRequestCookieHeader();

  if (!cookieHeader) {
    throw new Error("Missing authentication cookie for saved listings.");
  }

  return getSavedListings(query, {
    cookieHeader,
  });
}

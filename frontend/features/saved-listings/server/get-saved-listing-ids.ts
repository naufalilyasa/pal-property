import "server-only";

import { getSavedListingIds, type SavedListingContainsResult } from "@/lib/api/saved-listings";
import { getRequestCookieHeader } from "@/lib/server/cookies";

export async function getSavedListingIdsForListings(
  listingIds: string[],
): Promise<SavedListingContainsResult> {
  if (listingIds.length === 0) {
    return { listingIds: [] };
  }

  const cookieHeader = await getRequestCookieHeader();

  if (!cookieHeader) {
    return { listingIds: [] };
  }

  return getSavedListingIds(listingIds, {
    cookieHeader,
  });
}

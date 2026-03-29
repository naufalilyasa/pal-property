import { browserFetch } from "@/lib/api/browser-fetch";

const SAVED_LISTINGS_BASE = "/api/me/saved-listings";

type SavedListingToggleDto = {
  listing_id: string;
  saved: boolean;
};

export type SavedListingToggle = {
  listingId: string;
  saved: boolean;
};

export type SaveListingOptions = {
  baseUrl?: string;
  fetch?: typeof fetch;
};

export async function saveListing(listingId: string, options: SaveListingOptions = {}): Promise<SavedListingToggle> {
  const response = await browserFetch<SavedListingToggleDto>(SAVED_LISTINGS_BASE, {
    method: "POST",
    cache: "no-store",
    baseUrl: options.baseUrl,
    fetch: options.fetch,
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ listing_id: listingId }),
  });

  return mapToggleResponse(response.data);
}

export async function removeSavedListing(listingId: string, options: SaveListingOptions = {}): Promise<SavedListingToggle> {
  const response = await browserFetch<SavedListingToggleDto>(`${SAVED_LISTINGS_BASE}/${listingId}`, {
    method: "DELETE",
    cache: "no-store",
    baseUrl: options.baseUrl,
    fetch: options.fetch,
  });

  return mapToggleResponse(response.data);
}

function mapToggleResponse(data: SavedListingToggleDto): SavedListingToggle {
  return {
    listingId: data.listing_id,
    saved: data.saved,
  };
}

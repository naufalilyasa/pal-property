import { ApiError } from "@/lib/api/envelope";
import {
  getSellerListingById,
  isUnauthenticatedSellerListingsError,
  type SellerListing,
} from "@/lib/api/seller-listings";
import { ListingForm } from "@/app/dashboard/_components/listing-form";
import { getRequestCookieHeader } from "@/lib/server/cookies";
import { notFound, redirect } from "next/navigation";

export default async function EditListingPage({
  params,
}: {
  params: Promise<{ listingId: string }>;
}) {
  const { listingId } = await params;
  let listing: SellerListing;

  try {
    listing = await getSellerListingById(listingId, {
      cookieHeader: await getRequestCookieHeader(),
    });
  } catch (error) {
    if (isUnauthenticatedSellerListingsError(error)) {
      redirect("/");
    }

    if (error instanceof ApiError && error.status === 404) {
      notFound();
    }

    throw error;
  }

  return <ListingForm initialListing={listing} listingId={listingId} mode="edit" />;
}

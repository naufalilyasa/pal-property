import { ApiError } from "@/lib/api/envelope";
import { requireUser } from "@/features/auth/server/require-user";
import { ListingForm } from "@/features/listings/forms/listing-form";
import { getSellerListingById } from "@/features/listings/server/get-seller-listings";
import { notFound, redirect } from "next/navigation";
import type { SellerListing } from "@/lib/api/seller-listings";

export default async function EditListingPage({
  params,
}: {
  params: Promise<{ listingId: string }>;
}) {
  const { listingId } = await params;
  await requireUser();
  let listing: SellerListing;

  try {
    listing = await getSellerListingById(listingId);
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      redirect("/login");
    }

    if (error instanceof ApiError && error.status === 404) {
      notFound();
    }

    throw error;
  }

  return <ListingForm initialListing={listing} listingId={listingId} mode="edit" />;
}

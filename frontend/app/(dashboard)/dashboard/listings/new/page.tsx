import { requireUser } from "@/features/auth/server/require-user";
import { ListingForm } from "@/features/listings/forms/listing-form";

export default async function NewListingPage() {
  await requireUser({ intent: "seller", returnTo: "/dashboard/listings/new" });

  return <ListingForm mode="create" />;
}

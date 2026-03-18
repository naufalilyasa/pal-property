import {
  deleteListingImage,
  reorderListingImages,
  setPrimaryListingImage,
  uploadListingImage,
  type ListingRecord,
} from "@/lib/api/listing-form";

export const listingImageRoutePatterns = {
  upload: "/api/listings/:id/images",
  remove: "/api/listings/:id/images/:imageId",
  primary: "/api/listings/:id/images/:imageId/primary",
  reorder: "/api/listings/:id/images/reorder",
} as const;

export function listingImageEndpoint(listingId: string) {
  return `/api/listings/${listingId}/images`;
}

export function listingImageDeleteEndpoint(listingId: string, imageId: string) {
  return `${listingImageEndpoint(listingId)}/${imageId}`;
}

export function listingImagePrimaryEndpoint(listingId: string, imageId: string) {
  return `${listingImageDeleteEndpoint(listingId, imageId)}/primary`;
}

export function listingImageReorderEndpoint(listingId: string) {
  return `${listingImageEndpoint(listingId)}/reorder`;
}

export function createListingImageFormData(file: File) {
  const formData = new FormData();
  formData.set("file", file);
  return formData;
}

export async function uploadSellerListingImage(listingId: string, file: File): Promise<ListingRecord> {
  createListingImageFormData(file);
  return uploadListingImage(listingId, file);
}

export async function deleteSellerListingImage(listingId: string, imageId: string): Promise<ListingRecord> {
  return deleteListingImage(listingId, imageId);
}

export async function setSellerPrimaryListingImage(listingId: string, imageId: string): Promise<ListingRecord> {
  return setPrimaryListingImage(listingId, imageId);
}

export async function reorderSellerListingImages(listingId: string, orderedImageIds: string[]): Promise<ListingRecord> {
  return reorderListingImages(listingId, orderedImageIds);
}

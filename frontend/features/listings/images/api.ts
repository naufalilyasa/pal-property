import {
  deleteListingImage,
  deleteListingVideo,
  reorderListingImages,
  setPrimaryListingImage,
  uploadListingImage,
  uploadListingImages,
  uploadListingVideo,
  type ListingRecord,
} from "@/lib/api/listing-form";

export const listingImageRoutePatterns = {
  upload: "/api/listings/:id/images",
  remove: "/api/listings/:id/images/:imageId",
  primary: "/api/listings/:id/images/:imageId/primary",
  reorder: "/api/listings/:id/images/reorder",
} as const;

export const listingVideoRoutePatterns = {
  upload: "/api/listings/:id/video",
  remove: "/api/listings/:id/video",
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

export function listingVideoEndpoint(listingId: string) {
  return `/api/listings/${listingId}/video`;
}

export async function uploadSellerListingImage(listingId: string, file: File): Promise<ListingRecord> {
  return uploadListingImage(listingId, file);
}

export async function uploadSellerListingImages(listingId: string, files: File[]): Promise<ListingRecord> {
  return uploadListingImages(listingId, files);
}

export async function uploadSellerListingVideo(listingId: string, file: File): Promise<ListingRecord> {
  return uploadListingVideo(listingId, file);
}

export async function deleteSellerListingVideo(listingId: string): Promise<ListingRecord> {
  return deleteListingVideo(listingId);
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

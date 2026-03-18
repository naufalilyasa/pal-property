"use client";

import Image from "next/image";
import Link from "next/link";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useMemo, useState } from "react";
import { useForm } from "react-hook-form";

import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import {
  createSellerListing,
  formatListingFormError,
  getDefaultSpecifications,
  getListingCategories,
  parseListingSpecifications,
  updateSellerListing,
  type ListingCategoryOption,
  type ListingFormRequest,
  type ListingImageRecord,
  type ListingRecord,
} from "@/lib/api/listing-form";
import { queryKeys } from "@/lib/query/keys";

import {
  deleteSellerListingImage,
  reorderSellerListingImages,
  setSellerPrimaryListingImage,
  uploadSellerListingImage,
} from "@/features/listings/images/api";

import { type ListingFormSchema, listingFormSchema } from "./listing-schema";

type ListingFormMode = "create" | "edit";

type ListingFormProps = {
  initialListing?: ListingRecord | null;
  mode: ListingFormMode;
  listingId?: string;
};

const STATUS_OPTIONS: Array<{ value: ListingFormRequest["status"]; label: string }> = [
  { value: "active", label: "Active" },
  { value: "inactive", label: "Inactive" },
  { value: "sold", label: "Sold" },
];

export function ListingForm({ initialListing = null, mode, listingId }: ListingFormProps) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [listing, setListing] = useState<ListingRecord | null>(initialListing);
  const [selectedImageFile, setSelectedImageFile] = useState<File | null>(null);
  const [formError, setFormError] = useState<string | null>(null);
  const [formSuccess, setFormSuccess] = useState<string | null>(null);
  const [imageMessage, setImageMessage] = useState<string | null>(null);
  const [uploadInputKey, setUploadInputKey] = useState(0);

  const categoriesQuery = useQuery({
    queryKey: queryKeys.categories,
    queryFn: () => getListingCategories(),
  });

  const form = useForm<ListingFormSchema>({
    resolver: zodResolver(listingFormSchema),
    defaultValues: toFormValues(initialListing),
  });

  const orderedImages = useMemo(() => sortListingImages(listing?.images ?? []), [listing]);

  const submitMutation = useMutation({
    mutationFn: async (values: ListingFormSchema) => {
      const payload = toRequestPayload(values);

      if (mode === "create") {
        return createSellerListing(payload);
      }

      if (!listingId) {
        throw new Error("Listing id is required for edit mode.");
      }

      return updateSellerListing(listingId, payload);
    },
    onSuccess: (nextListing) => {
      setFormError(null);
      setFormSuccess(mode === "create" ? null : "Listing changes saved successfully.");
      setListing(nextListing);
      queryClient.invalidateQueries({ queryKey: queryKeys.sellerListings });

      if (mode === "create") {
        router.push(`/dashboard/listings/${nextListing.id}/edit?created=1`);
        return;
      }

      form.reset(toFormValues(nextListing));
      router.refresh();
    },
    onError: (error) => {
      setFormSuccess(null);
      setFormError(formatListingFormError(error));
    },
  });

  const imageMutation = useMutation({
    mutationFn: async (action: { type: "upload" | "set-primary" | "delete" | "reorder"; imageId?: string; direction?: "earlier" | "later" }) => {
      if (!listingId) {
        throw new Error("Save the listing first before updating images.");
      }

      if (action.type === "upload") {
        if (!selectedImageFile) {
          throw new Error("Choose an image file before uploading.");
        }

        return uploadSellerListingImage(listingId, selectedImageFile);
      }

      if (action.type === "set-primary" && action.imageId) {
        return setSellerPrimaryListingImage(listingId, action.imageId);
      }

      if (action.type === "delete" && action.imageId) {
        return deleteSellerListingImage(listingId, action.imageId);
      }

      if (action.type === "reorder" && action.imageId && action.direction) {
        const currentIndex = orderedImages.findIndex((image) => image.id === action.imageId);
        const targetIndex = action.direction === "earlier" ? currentIndex - 1 : currentIndex + 1;

        if (currentIndex < 0 || targetIndex < 0 || targetIndex >= orderedImages.length) {
          return listing ?? initialListing ?? createEmptyListing();
        }

        const reorderedIds = orderedImages.map((image) => image.id);
        const [movedId] = reorderedIds.splice(currentIndex, 1);
        reorderedIds.splice(targetIndex, 0, movedId);

        return reorderSellerListingImages(listingId, reorderedIds);
      }

      throw new Error("Unsupported image action.");
    },
    onSuccess: (nextListing, action) => {
      setListing(nextListing);
      setImageMessage(getImageSuccessMessage(action.type));
      queryClient.invalidateQueries({ queryKey: queryKeys.sellerListings });
      router.refresh();

      if (action.type === "upload") {
        setSelectedImageFile(null);
        setUploadInputKey((current) => current + 1);
      }
    },
    onError: (error) => {
      setImageMessage(formatListingFormError(error));
    },
  });

  if (categoriesQuery.isError) {
    return (
      <section className="rounded-[1.75rem] border border-[var(--line)] bg-white/72 p-8">
        <p className="text-xs uppercase tracking-[0.3em] text-[var(--muted)]">Listing form</p>
        <h2 className="mt-4 text-2xl font-semibold tracking-[-0.03em] text-[var(--ink)]">We could not prepare this listing form</h2>
        <p className="mt-3 text-sm leading-7 text-[var(--muted)]">{formatListingFormError(categoriesQuery.error)}</p>
        <Button className="mt-6" onClick={() => void categoriesQuery.refetch()} type="button" variant="secondary">
          Retry
        </Button>
      </section>
    );
  }

  return (
    <div className="space-y-6">
      <section className="flex flex-col gap-4 rounded-[1.75rem] border border-[var(--line)] bg-white/72 p-6 sm:p-8 lg:flex-row lg:items-end lg:justify-between">
        <div className="space-y-3">
          <p className="text-xs uppercase tracking-[0.3em] text-[var(--muted)]" style={{ fontFamily: "var(--font-mono), monospace" }}>
            {mode === "create" ? "Create listing" : "Edit listing"}
          </p>
          <div className="space-y-2">
            <h2 className="text-3xl font-semibold tracking-[-0.04em] text-[var(--ink)]">
              {mode === "create" ? "Publish a new property draft" : "Refine an existing property record"}
            </h2>
            <p className="max-w-3xl text-sm leading-7 text-[var(--muted)] sm:text-base">
              Use RHF, Zod, and the canonical backend listing contract so every saved field stays aligned with the PAL Property API.
            </p>
          </div>
        </div>

        <Link className="inline-flex items-center rounded-full border border-[var(--line)] bg-[var(--panel)] px-5 py-3 text-sm font-semibold text-[var(--ink)] transition hover:border-[var(--accent)] hover:text-[var(--accent)]" href="/dashboard/listings">
          Back to listings
        </Link>
      </section>

      <Form {...form}>
        <form className="space-y-6" onSubmit={form.handleSubmit((values) => submitMutation.mutate(values))}>
          <section className="grid gap-6 xl:grid-cols-[1.15fr_0.85fr]">
            <div className="space-y-6 rounded-[1.75rem] border border-[var(--line)] bg-white/80 p-6 sm:p-8">
              <div className="space-y-5">
                <div className="space-y-2">
                  <h3 className="text-xl font-semibold tracking-[-0.03em] text-[var(--ink)]">Listing basics</h3>
                  <p className="text-sm leading-7 text-[var(--muted)]">The backend remains the source of truth for create and update payloads.</p>
                </div>

                <FormField
                  control={form.control}
                  name="title"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel htmlFor="listing-title">Title</FormLabel>
                      <FormControl>
                        <Input id="listing-title" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="description"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel htmlFor="listing-description">Description</FormLabel>
                      <FormControl>
                        <Textarea className="min-h-36 resize-y" id="listing-description" {...field} value={field.value ?? ""} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <div className="grid gap-5 sm:grid-cols-2">
                  <FormField
                    control={form.control}
                    name="category_id"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel htmlFor="listing-category">Category</FormLabel>
                        <FormControl>
                          <Select aria-label="Category" id="listing-category" {...field} value={field.value ?? ""}>
                            <option value="">No category selected</option>
                            {(categoriesQuery.data ?? []).map((category: ListingCategoryOption) => (
                              <option key={category.id} value={category.id}>
                                {category.label}
                              </option>
                            ))}
                          </Select>
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="status"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel htmlFor="listing-status">Status</FormLabel>
                        <FormControl>
                          <Select aria-label="Status" id="listing-status" {...field}>
                            {STATUS_OPTIONS.map((status) => (
                              <option key={status.value} value={status.value}>
                                {status.label}
                              </option>
                            ))}
                          </Select>
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>

                <FormField
                  control={form.control}
                  name="price"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel htmlFor="listing-price">Price (IDR)</FormLabel>
                      <FormControl>
                        <Input id="listing-price" inputMode="numeric" min="1" pattern="[0-9]*" {...field} />
                      </FormControl>
                      <FormDescription>Money stays in integer IDR values.</FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            </div>

            <div className="space-y-6 rounded-[1.75rem] border border-[var(--line)] bg-white/80 p-6 sm:p-8">
              <div className="space-y-5">
                <div className="space-y-2">
                  <h3 className="text-xl font-semibold tracking-[-0.03em] text-[var(--ink)]">Location and specifications</h3>
                  <p className="text-sm leading-7 text-[var(--muted)]">Optional location fields and numeric specification values map directly to backend request fields.</p>
                </div>

                {(
                  [
                    ["location_city", "City"],
                    ["location_district", "District"],
                    ["address_detail", "Address detail"],
                    ["bedrooms", "Bedrooms"],
                    ["bathrooms", "Bathrooms"],
                    ["land_area_sqm", "Land area (sqm)"],
                    ["building_area_sqm", "Building area (sqm)"],
                  ] as const
                ).map(([name, label]) => (
                  <FormField
                    control={form.control}
                    key={name}
                    name={name}
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel htmlFor={`field-${name}`}>{label}</FormLabel>
                        <FormControl>
                          {name === "address_detail" ? (
                            <Textarea className="min-h-28 resize-y" id={`field-${name}`} {...field} value={field.value ?? ""} />
                          ) : (
                            <Input id={`field-${name}`} inputMode={name.includes("sqm") || name === "bedrooms" || name === "bathrooms" ? "numeric" : undefined} pattern={name.includes("sqm") || name === "bedrooms" || name === "bathrooms" ? "[0-9]*" : undefined} {...field} value={field.value ?? ""} />
                          )}
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                ))}
              </div>
            </div>
          </section>

          <section className="rounded-[1.75rem] border border-[var(--line)] bg-white/80 p-6 sm:p-8">
            <div className="space-y-3">
              <h3 className="text-xl font-semibold tracking-[-0.03em] text-[var(--ink)]">Listing images</h3>
              <p className="text-sm leading-7 text-[var(--muted)]">Uploads, primary selection, deletion, and ordering all wait for backend-confirmed listing state before the UI changes.</p>
            </div>

            {mode === "edit" && listingId ? (
              <div className="mt-6 space-y-5">
                <div className="rounded-[1.5rem] border border-[var(--line)] bg-[var(--panel)] p-5">
                  <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
                    <div className="space-y-2">
                      <p className="text-sm font-medium text-[var(--ink)]">Upload to this listing</p>
                      <label className="block text-sm font-medium text-[var(--ink)]" htmlFor="listing-image-upload">
                        <span className="sr-only">Choose listing image</span>
                        <input
                          key={uploadInputKey}
                          accept="image/*"
                          className="block w-full text-sm text-[var(--muted)] file:mr-4 file:rounded-full file:border-0 file:bg-[var(--accent)] file:px-4 file:py-2 file:text-sm file:font-semibold file:text-white"
                          data-testid="listing-image-upload"
                          id="listing-image-upload"
                          name="listing_image_upload"
                          onChange={(event) => setSelectedImageFile(event.currentTarget.files?.[0] ?? null)}
                          type="file"
                        />
                      </label>
                      <p className="text-xs uppercase tracking-[0.24em] text-[var(--muted)]" style={{ fontFamily: "var(--font-mono), monospace" }}>
                        {selectedImageFile ? `Ready: ${selectedImageFile.name}` : "No image selected yet"}
                      </p>
                    </div>
                    <Button disabled={imageMutation.isPending} onClick={() => imageMutation.mutate({ type: "upload" })} type="button">
                      {imageMutation.isPending ? "Uploading image..." : "Upload image"}
                    </Button>
                  </div>
                </div>

                {imageMessage ? (
                  <p className={imageMutation.isError ? "text-sm font-medium text-red-700" : "text-sm font-medium text-emerald-700"} data-testid={imageMutation.isError ? "listing-image-error" : undefined} role={imageMutation.isError ? "alert" : undefined}>
                    {imageMessage}
                  </p>
                ) : null}

                {orderedImages.length === 0 ? (
                  <div className="rounded-[1.5rem] border border-dashed border-[var(--line)] bg-white/60 p-5 text-sm leading-7 text-[var(--muted)]">
                    No images yet. Upload the first seller photo to let the backend assign ordering and primary state.
                  </div>
                ) : (
                  <div className="grid gap-4 lg:grid-cols-2">
                    {orderedImages.map((image, index) => (
                      <article className="overflow-hidden rounded-[1.5rem] border border-[var(--line)] bg-white/72" data-testid={`listing-image-card-${image.id}`} key={image.id}>
                        <div data-testid="listing-image-item">
                        <div className="relative aspect-[4/3] bg-[var(--panel)]">
                          <Image alt={image.original_filename ?? `Listing image ${index + 1}`} fill sizes="(min-width: 1024px) 30vw, 100vw" src={image.url} className="object-cover" unoptimized />
                        </div>
                        <div className="space-y-4 p-5">
                          <div className="flex flex-wrap items-center gap-2">
                            <span className="rounded-full bg-[var(--panel)] px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em] text-[var(--muted)]">Order {image.sort_order + 1}</span>
                            {image.is_primary ? <span className="rounded-full bg-emerald-100 px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em] text-emerald-700">Primary</span> : null}
                          </div>
                          <p className="text-sm font-medium text-[var(--ink)]">{image.original_filename ?? `Listing image ${index + 1}`}</p>
                          <div className="flex flex-wrap gap-3">
                            <Button data-testid="listing-image-make-primary" disabled={imageMutation.isPending || image.is_primary} onClick={() => imageMutation.mutate({ type: "set-primary", imageId: image.id })} type="button" variant="secondary">
                              Set primary
                            </Button>
                            <Button
                              aria-label={`Move ${image.original_filename ?? `image ${index + 1}`} earlier`}
                              disabled={imageMutation.isPending || index === 0}
                              onClick={() => imageMutation.mutate({ type: "reorder", imageId: image.id, direction: "earlier" })}
                              type="button"
                              variant="secondary"
                            >
                              Move earlier
                            </Button>
                            <Button
                              aria-label={`Move ${image.original_filename ?? `image ${index + 1}`} later`}
                              disabled={imageMutation.isPending || index === orderedImages.length - 1}
                              onClick={() => imageMutation.mutate({ type: "reorder", imageId: image.id, direction: "later" })}
                              type="button"
                              variant="secondary"
                            >
                              Move later
                            </Button>
                            <Button disabled={imageMutation.isPending} onClick={() => imageMutation.mutate({ type: "delete", imageId: image.id })} type="button" variant="destructive">
                              Delete image
                            </Button>
                          </div>
                        </div>
                        </div>
                      </article>
                    ))}
                  </div>
                )}
              </div>
            ) : (
              <div className="mt-6 rounded-[1.5rem] border border-dashed border-[var(--line)] bg-[var(--panel)] p-5 text-sm leading-7 text-[var(--muted)]">
                Publish the listing first, then return here to upload images, set a primary photo, remove outdated media, and adjust ordering from backend-backed state.
              </div>
            )}
          </section>

          <section className="rounded-[1.75rem] border border-[var(--line)] bg-[var(--panel)] p-6 sm:p-8">
            <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
              <div className="space-y-2">
                <p className="text-xs uppercase tracking-[0.3em] text-[var(--muted)]" style={{ fontFamily: "var(--font-mono), monospace" }}>
                  Submission state
                </p>
                <p className="text-sm leading-7 text-[var(--muted)]">
                  {mode === "create"
                    ? "Creating a listing sends the canonical create payload and redirects into edit mode once the backend returns the new record."
                    : "Saving changes sends the full listing contract back through the update endpoint so seller edits stay explicit."}
                </p>
                {formError ? (
                  <p className="text-sm font-medium text-red-700" data-testid="listing-form-error" role="alert">
                    {formError}
                  </p>
                ) : null}
                {formSuccess ? <p className="text-sm font-medium text-emerald-700">{formSuccess}</p> : null}
              </div>

              <Button data-testid="listing-submit-button" disabled={submitMutation.isPending} type="submit">
                {submitMutation.isPending ? (mode === "create" ? "Create listing..." : "Save changes...") : mode === "create" ? "Create listing" : "Save changes"}
              </Button>
            </div>
          </section>
        </form>
      </Form>
    </div>
  );
}

function toFormValues(listing: ListingRecord | null): ListingFormSchema {
  if (!listing) {
    return {
      category_id: "",
      title: "",
      description: "",
      price: "",
      location_city: "",
      location_district: "",
      address_detail: "",
      status: "active",
      bedrooms: "0",
      bathrooms: "0",
      land_area_sqm: "0",
      building_area_sqm: "0",
    };
  }

  const specifications = parseListingSpecifications(listing.specifications);

  return {
    category_id: listing.category_id ?? "",
    title: listing.title,
    description: listing.description ?? "",
    price: String(listing.price),
    location_city: listing.location_city ?? "",
    location_district: listing.location_district ?? "",
    address_detail: listing.address_detail ?? "",
    status: normalizeStatus(listing.status),
    bedrooms: String(specifications.bedrooms),
    bathrooms: String(specifications.bathrooms),
    land_area_sqm: String(specifications.land_area_sqm),
    building_area_sqm: String(specifications.building_area_sqm),
  };
}

function toRequestPayload(values: ListingFormSchema): ListingFormRequest {
  const defaults = getDefaultSpecifications();

  return {
    category_id: normalizeNullableString(values.category_id),
    title: values.title.trim(),
    description: normalizeNullableString(values.description),
    price: normalizeRequiredInteger(values.price, 1),
    location_city: normalizeNullableString(values.location_city),
    location_district: normalizeNullableString(values.location_district),
    address_detail: normalizeNullableString(values.address_detail),
    status: values.status,
    specifications: {
      bedrooms: normalizeOptionalInteger(values.bedrooms, defaults.bedrooms),
      bathrooms: normalizeOptionalInteger(values.bathrooms, defaults.bathrooms),
      land_area_sqm: normalizeOptionalInteger(values.land_area_sqm, defaults.land_area_sqm),
      building_area_sqm: normalizeOptionalInteger(values.building_area_sqm, defaults.building_area_sqm),
    },
  };
}

function normalizeNullableString(value: string | undefined) {
  const trimmed = value?.trim() ?? "";
  return trimmed ? trimmed : null;
}

function normalizeRequiredInteger(value: string | undefined, fallback: number) {
  const parsed = Number.parseInt(value ?? "", 10);
  return !Number.isNaN(parsed) && parsed >= fallback ? parsed : fallback;
}

function normalizeOptionalInteger(value: string | undefined, fallback: number) {
  const parsed = Number.parseInt(value ?? "", 10);
  return !Number.isNaN(parsed) && parsed >= 0 ? parsed : fallback;
}

function normalizeStatus(status: string): ListingFormRequest["status"] {
  if (status === "inactive" || status === "sold") {
    return status;
  }

  return "active";
}

function sortListingImages(images: ListingImageRecord[]) {
  return [...images].sort((left, right) => {
    if (left.sort_order === right.sort_order) {
      return left.created_at.localeCompare(right.created_at);
    }

    return left.sort_order - right.sort_order;
  });
}

function getImageSuccessMessage(type: "upload" | "set-primary" | "delete" | "reorder") {
  switch (type) {
    case "upload":
      return "Image uploaded. The gallery now reflects the backend response.";
    case "set-primary":
      return "Primary image updated from the backend response.";
    case "delete":
      return "Image removed. Remaining images were refreshed from the backend.";
    case "reorder":
      return "Image order refreshed from the backend response.";
  }
}

function createEmptyListing(): ListingRecord {
  return {
    id: "",
    user_id: "",
    category_id: null,
    category: null,
    title: "",
    slug: "",
    description: null,
    price: 0,
    currency: "IDR",
    location_city: null,
    location_district: null,
    address_detail: null,
    status: "active",
    is_featured: false,
    specifications: {},
    view_count: 0,
    images: [],
    created_at: new Date(0).toISOString(),
    updated_at: new Date(0).toISOString(),
  };
}

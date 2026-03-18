"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";

import {
  createSellerListing,
  deleteListingImage,
  formatListingFormError,
  getDefaultSpecifications,
  getListingCategories,
  parseListingSpecifications,
  reorderListingImages,
  setPrimaryListingImage,
  uploadListingImage,
  updateSellerListing,
  type ListingCategoryOption,
  type ListingFormRequest,
  type ListingImageRecord,
  type ListingRecord,
} from "@/lib/api/listing-form";

type ListingFormMode = "create" | "edit";

type ListingFormProps = {
  initialListing?: ListingRecord | null;
  mode: ListingFormMode;
  listingId?: string;
};

type ListingFormValues = {
  category_id: string;
  title: string;
  description: string;
  price: string;
  location_city: string;
  location_district: string;
  address_detail: string;
  status: ListingFormRequest["status"];
  bedrooms: string;
  bathrooms: string;
  land_area_sqm: string;
  building_area_sqm: string;
};

type BootstrapState =
  | { status: "loading" }
  | { status: "error"; message: string }
  | {
      status: "ready";
      categories: ListingCategoryOption[];
      listing: ListingRecord | null;
    };

const EMPTY_VALUES: ListingFormValues = {
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

const STATUS_OPTIONS: Array<{ value: ListingFormRequest["status"]; label: string }> = [
  { value: "active", label: "Active" },
  { value: "inactive", label: "Inactive" },
  { value: "sold", label: "Sold" },
];

export function ListingForm({ initialListing = null, mode, listingId }: ListingFormProps) {
  const router = useRouter();
  const [bootstrapState, setBootstrapState] = useState<BootstrapState>({ status: "loading" });
  const [values, setValues] = useState<ListingFormValues>(EMPTY_VALUES);
  const [selectedImageFile, setSelectedImageFile] = useState<File | null>(null);
  const [uploadInputKey, setUploadInputKey] = useState(0);
  const [submitState, setSubmitState] = useState<{
    status: "idle" | "submitting" | "success" | "error";
    message?: string;
  }>({ status: "idle" });
  const [imageMutationState, setImageMutationState] = useState<{
    status: "idle" | "uploading" | "mutating" | "success" | "error";
    message?: string;
    targetId?: string;
  }>({ status: "idle" });

  useEffect(() => {
    let isActive = true;

    async function loadForm() {
      setBootstrapState({ status: "loading" });

      try {
        const categories = await getListingCategories();
        const listing = mode === "edit" ? initialListing : null;

        if (!isActive) {
          return;
        }

        if (mode === "edit" && listingId && !listing) {
          throw new Error("We could not load this seller listing.");
        }

        setBootstrapState({
          status: "ready",
          categories,
          listing,
        });
        setValues(listing ? toFormValues(listing) : EMPTY_VALUES);
      } catch (error) {
        if (!isActive) {
          return;
        }

        setBootstrapState({
          status: "error",
          message: formatListingFormError(error),
        });
      }
    }

    void loadForm();

    return () => {
      isActive = false;
    };
  }, [initialListing, listingId, mode]);

  const submitLabel = mode === "create" ? "Create listing" : "Save changes";
  const pageCopy = useMemo(() => getPageCopy(mode), [mode]);
  const readyListing = bootstrapState.status === "ready" ? bootstrapState.listing : null;
  const orderedImages = useMemo(() => sortListingImages(readyListing?.images ?? []), [readyListing]);

  function replaceListingRecord(nextListing: ListingRecord) {
    setBootstrapState((current) =>
      current.status === "ready"
        ? {
            ...current,
            listing: nextListing,
          }
        : current,
    );
  }

  async function handleSubmit(event: { preventDefault: () => void }) {
    event.preventDefault();
    setSubmitState({ status: "submitting" });

    try {
      const payload = toRequestPayload(values);

      if (mode === "create") {
        const listing = await createSellerListing(payload);
        router.push(`/dashboard/listings/${listing.id}/edit?created=1`);
        return;
      }

      if (!listingId) {
        throw new Error("Listing id is required for edit mode.");
      }

      const listing = await updateSellerListing(listingId, payload);
      setValues(toFormValues(listing));
      replaceListingRecord(listing);
      setSubmitState({
        status: "success",
        message: "Listing changes saved successfully.",
      });
      router.refresh();
    } catch (error) {
      setSubmitState({
        status: "error",
        message: formatListingFormError(error),
      });
    }
  }

  function updateValue<Key extends keyof ListingFormValues>(key: Key, value: ListingFormValues[Key]) {
    setValues((current) => ({
      ...current,
      [key]: value,
    }));
    setSubmitState((current) => (current.status === "error" ? { status: "idle" } : current));
  }

  async function handleImageUpload() {
    if (!listingId) {
      setImageMutationState({
        status: "error",
        message: "Save the listing first before uploading images.",
      });
      return;
    }

    if (!selectedImageFile) {
      setImageMutationState({
        status: "error",
        message: "Choose an image file before uploading.",
      });
      return;
    }

    setImageMutationState({
      status: "uploading",
      message: `Uploading ${selectedImageFile.name}...`,
    });

    try {
      const listing = await uploadListingImage(listingId, selectedImageFile);
      replaceListingRecord(listing);
      setSelectedImageFile(null);
      setUploadInputKey((current) => current + 1);
      setImageMutationState({
        status: "success",
        message: "Image uploaded. The gallery now reflects the backend response.",
      });
      router.refresh();
    } catch (error) {
      setImageMutationState({
        status: "error",
        message: formatListingFormError(error),
      });
    }
  }

  async function handleSetPrimaryImage(imageId: string) {
    if (!listingId) {
      return;
    }

    setImageMutationState({
      status: "mutating",
      targetId: imageId,
      message: "Updating primary image...",
    });

    try {
      const listing = await setPrimaryListingImage(listingId, imageId);
      replaceListingRecord(listing);
      setImageMutationState({
        status: "success",
        targetId: imageId,
        message: "Primary image updated from the backend response.",
      });
      router.refresh();
    } catch (error) {
      setImageMutationState({
        status: "error",
        targetId: imageId,
        message: formatListingFormError(error),
      });
    }
  }

  async function handleDeleteImage(imageId: string) {
    if (!listingId) {
      return;
    }

    setImageMutationState({
      status: "mutating",
      targetId: imageId,
      message: "Removing image...",
    });

    try {
      const listing = await deleteListingImage(listingId, imageId);
      replaceListingRecord(listing);
      setImageMutationState({
        status: "success",
        targetId: imageId,
        message: "Image removed. Remaining images were refreshed from the backend.",
      });
      router.refresh();
    } catch (error) {
      setImageMutationState({
        status: "error",
        targetId: imageId,
        message: formatListingFormError(error),
      });
    }
  }

  async function handleMoveImage(imageId: string, direction: "earlier" | "later") {
    if (!listingId) {
      return;
    }

    const currentIndex = orderedImages.findIndex((image) => image.id === imageId);
    const offset = direction === "earlier" ? -1 : 1;
    const targetIndex = currentIndex + offset;

    if (currentIndex < 0 || targetIndex < 0 || targetIndex >= orderedImages.length) {
      return;
    }

    const reorderedIds = orderedImages.map((image) => image.id);
    const [movedId] = reorderedIds.splice(currentIndex, 1);
    reorderedIds.splice(targetIndex, 0, movedId);

    setImageMutationState({
      status: "mutating",
      targetId: imageId,
      message: "Saving image order...",
    });

    try {
      const listing = await reorderListingImages(listingId, reorderedIds);
      replaceListingRecord(listing);
      setImageMutationState({
        status: "success",
        targetId: imageId,
        message: "Image order refreshed from the backend response.",
      });
      router.refresh();
    } catch (error) {
      setImageMutationState({
        status: "error",
        targetId: imageId,
        message: formatListingFormError(error),
      });
    }
  }

  if (bootstrapState.status === "loading") {
    return <ListingFormState eyebrow={pageCopy.eyebrow} title="Loading listing form" body={pageCopy.loadingBody} />;
  }

  if (bootstrapState.status === "error") {
    return (
      <ListingFormState
        eyebrow={pageCopy.eyebrow}
        title="We could not prepare this listing form"
        body={bootstrapState.message}
        action={
          <button
            className="inline-flex items-center rounded-full border border-[var(--line)] bg-[var(--panel-strong)] px-5 py-3 text-sm font-semibold text-[var(--ink)] transition hover:border-[var(--accent)] hover:text-[var(--accent)]"
            onClick={() => window.location.reload()}
            type="button"
          >
            Retry
          </button>
        }
      />
    );
  }

  return (
    <div className="space-y-6">
      <section className="flex flex-col gap-4 rounded-[1.75rem] border border-[var(--line)] bg-white/72 p-6 sm:p-8 lg:flex-row lg:items-end lg:justify-between">
        <div className="space-y-3">
          <p
            className="text-xs uppercase tracking-[0.3em] text-[var(--muted)]"
            style={{ fontFamily: "var(--font-mono), monospace" }}
          >
            {pageCopy.eyebrow}
          </p>
          <div className="space-y-2">
            <h2 className="text-3xl font-semibold tracking-[-0.04em] text-[var(--ink)]">{pageCopy.title}</h2>
            <p className="max-w-3xl text-sm leading-7 text-[var(--muted)] sm:text-base">{pageCopy.body}</p>
          </div>
        </div>

        <Link
          className="inline-flex items-center rounded-full border border-[var(--line)] bg-[var(--panel)] px-5 py-3 text-sm font-semibold text-[var(--ink)] transition hover:border-[var(--accent)] hover:text-[var(--accent)]"
          href="/dashboard"
        >
          Back to listings
        </Link>
      </section>

      <form className="space-y-6" onSubmit={handleSubmit}>
        <section className="grid gap-6 xl:grid-cols-[1.15fr_0.85fr]">
          <div className="space-y-6 rounded-[1.75rem] border border-[var(--line)] bg-white/80 p-6 sm:p-8">
            <FieldGroup title="Listing basics" description="Use the backend listing contract directly so every saved field maps to the PAL Property API.">
              <Field htmlFor="listing-title" label="Title" required>
                <input
                  className={inputClassName}
                  id="listing-title"
                  name="title"
                  onChange={(event) => updateValue("title", event.currentTarget.value)}
                  required
                  value={values.title}
                />
              </Field>

              <Field htmlFor="listing-description" label="Description">
                <textarea
                  className={`${inputClassName} min-h-36 resize-y`}
                  id="listing-description"
                  name="description"
                  onChange={(event) => updateValue("description", event.currentTarget.value)}
                  value={values.description}
                />
              </Field>

              <div className="grid gap-5 sm:grid-cols-2">
                <Field htmlFor="listing-category" label="Category">
                  <select
                    aria-label="Category"
                    className={inputClassName}
                    id="listing-category"
                    name="category_id"
                    onChange={(event) => updateValue("category_id", event.currentTarget.value)}
                    value={values.category_id}
                  >
                    <option value="">No category selected</option>
                    {bootstrapState.categories.map((category) => (
                      <option key={category.id} value={category.id}>
                        {category.label}
                      </option>
                    ))}
                  </select>
                </Field>

                <Field htmlFor="listing-status" label="Status" required>
                  <select
                    aria-label="Status"
                    className={inputClassName}
                    id="listing-status"
                    name="status"
                    onChange={(event) => updateValue("status", event.currentTarget.value as ListingFormValues["status"])}
                    required
                    value={values.status}
                  >
                    {STATUS_OPTIONS.map((status) => (
                      <option key={status.value} value={status.value}>
                        {status.label}
                      </option>
                    ))}
                  </select>
                </Field>
              </div>

              <Field htmlFor="listing-price" label="Price (IDR)" required>
                <input
                  className={inputClassName}
                  id="listing-price"
                  inputMode="numeric"
                  min="1"
                  name="price"
                  onChange={(event) => updateValue("price", event.currentTarget.value)}
                  pattern="[0-9]*"
                  required
                  value={values.price}
                />
              </Field>
            </FieldGroup>
          </div>

          <div className="space-y-6 rounded-[1.75rem] border border-[var(--line)] bg-white/80 p-6 sm:p-8">
            <FieldGroup title="Location" description="These optional fields map to the listing request address fields without extra frontend-only structure.">
              <Field htmlFor="listing-city" label="City">
                <input
                  className={inputClassName}
                  id="listing-city"
                  name="location_city"
                  onChange={(event) => updateValue("location_city", event.currentTarget.value)}
                  value={values.location_city}
                />
              </Field>

              <Field htmlFor="listing-district" label="District">
                <input
                  className={inputClassName}
                  id="listing-district"
                  name="location_district"
                  onChange={(event) => updateValue("location_district", event.currentTarget.value)}
                  value={values.location_district}
                />
              </Field>

              <Field htmlFor="listing-address-detail" label="Address detail">
                <textarea
                  className={`${inputClassName} min-h-28 resize-y`}
                  id="listing-address-detail"
                  name="address_detail"
                  onChange={(event) => updateValue("address_detail", event.currentTarget.value)}
                  value={values.address_detail}
                />
              </Field>
            </FieldGroup>

            <FieldGroup title="Specifications" description="Keep bedroom, bathroom, and size data aligned with the backend specifications object.">
              <div className="grid gap-5 sm:grid-cols-2">
                <Field htmlFor="listing-bedrooms" label="Bedrooms">
                  <input
                    className={inputClassName}
                    id="listing-bedrooms"
                    inputMode="numeric"
                    min="0"
                    name="bedrooms"
                    onChange={(event) => updateValue("bedrooms", event.currentTarget.value)}
                    pattern="[0-9]*"
                    value={values.bedrooms}
                  />
                </Field>

                <Field htmlFor="listing-bathrooms" label="Bathrooms">
                  <input
                    className={inputClassName}
                    id="listing-bathrooms"
                    inputMode="numeric"
                    min="0"
                    name="bathrooms"
                    onChange={(event) => updateValue("bathrooms", event.currentTarget.value)}
                    pattern="[0-9]*"
                    value={values.bathrooms}
                  />
                </Field>

                <Field htmlFor="listing-land-area" label="Land area (sqm)">
                  <input
                    className={inputClassName}
                    id="listing-land-area"
                    inputMode="numeric"
                    min="0"
                    name="land_area_sqm"
                    onChange={(event) => updateValue("land_area_sqm", event.currentTarget.value)}
                    pattern="[0-9]*"
                    value={values.land_area_sqm}
                  />
                </Field>

                <Field htmlFor="listing-building-area" label="Building area (sqm)">
                  <input
                    className={inputClassName}
                    id="listing-building-area"
                    inputMode="numeric"
                    min="0"
                    name="building_area_sqm"
                    onChange={(event) => updateValue("building_area_sqm", event.currentTarget.value)}
                    pattern="[0-9]*"
                    value={values.building_area_sqm}
                  />
                </Field>
              </div>
            </FieldGroup>
          </div>
        </section>

        <section className="rounded-[1.75rem] border border-[var(--line)] bg-white/80 p-6 sm:p-8">
          <FieldGroup
            title="Listing images"
            description={
              mode === "edit"
                ? "Upload, reorder, and curate seller media against the existing listing image endpoints. Every mutation waits for the backend listing response before the UI updates."
                : "Image management starts after the first save so the backend has a listing id for upload, delete, primary, and reorder routes."
            }
          >
            {mode === "edit" && listingId && readyListing ? (
              <ListingImageManager
                imageMutationState={imageMutationState}
                images={orderedImages}
                onDeleteImage={handleDeleteImage}
                onFileSelect={setSelectedImageFile}
                onMoveImage={handleMoveImage}
                onSetPrimaryImage={handleSetPrimaryImage}
                onUpload={handleImageUpload}
                selectedImageFile={selectedImageFile}
                uploadInputKey={uploadInputKey}
              />
            ) : (
              <div className="rounded-[1.5rem] border border-dashed border-[var(--line)] bg-[var(--panel)] p-5 text-sm leading-7 text-[var(--muted)]">
                Publish the listing first, then return here to upload images, set a primary photo, remove outdated media, and adjust ordering from backend-backed state.
              </div>
            )}
          </FieldGroup>
        </section>

        <section className="rounded-[1.75rem] border border-[var(--line)] bg-[var(--panel)] p-6 sm:p-8">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
            <div className="space-y-2">
              <p
                className="text-xs uppercase tracking-[0.3em] text-[var(--muted)]"
                style={{ fontFamily: "var(--font-mono), monospace" }}
              >
                Submission state
              </p>
              <p className="text-sm leading-7 text-[var(--muted)]">
                {mode === "create"
                  ? "Creating a listing sends the canonical create payload and redirects into edit mode once the backend returns the new record."
                  : "Saving changes sends the full listing contract back through the update endpoint so seller edits stay explicit."}
              </p>
              {submitState.status === "error" ? (
                <p aria-live="polite" className="text-sm font-medium text-red-700" role="alert">
                  {submitState.message}
                </p>
              ) : null}
              {submitState.status === "success" && submitState.message ? (
                <p aria-live="polite" className="text-sm font-medium text-emerald-700">
                  {submitState.message}
                </p>
              ) : null}
            </div>

            <button
              className="inline-flex items-center justify-center rounded-full bg-[var(--accent)] px-6 py-3 text-sm font-semibold text-white transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
              disabled={submitState.status === "submitting"}
              type="submit"
            >
              {submitState.status === "submitting" ? `${submitLabel}...` : submitLabel}
            </button>
          </div>
        </section>
      </form>
    </div>
  );
}

function ListingImageManager({
  imageMutationState,
  images,
  onDeleteImage,
  onFileSelect,
  onMoveImage,
  onSetPrimaryImage,
  onUpload,
  selectedImageFile,
  uploadInputKey,
}: {
  imageMutationState: {
    status: "idle" | "uploading" | "mutating" | "success" | "error";
    message?: string;
    targetId?: string;
  };
  images: ListingImageRecord[];
  onDeleteImage: (imageId: string) => Promise<void>;
  onFileSelect: (file: File | null) => void;
  onMoveImage: (imageId: string, direction: "earlier" | "later") => Promise<void>;
  onSetPrimaryImage: (imageId: string) => Promise<void>;
  onUpload: () => Promise<void>;
  selectedImageFile: File | null;
  uploadInputKey: number;
}) {
  const isBusy = imageMutationState.status === "uploading" || imageMutationState.status === "mutating";

  return (
    <div className="space-y-5">
      <div className="rounded-[1.5rem] border border-[var(--line)] bg-[var(--panel)] p-5">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div className="space-y-2">
            <p className="text-sm font-medium text-[var(--ink)]">Upload to this listing</p>
            <p className="text-sm leading-7 text-[var(--muted)]">
              The backend allows up to 10 images. Validation errors such as invalid files or image-limit failures are shown directly from the response message.
            </p>
            <label className="block text-sm font-medium text-[var(--ink)]" htmlFor="listing-image-upload">
              <span className="sr-only">Choose listing image</span>
              <input
                key={uploadInputKey}
                accept="image/*"
                className="block w-full text-sm text-[var(--muted)] file:mr-4 file:rounded-full file:border-0 file:bg-[var(--accent)] file:px-4 file:py-2 file:text-sm file:font-semibold file:text-white"
                id="listing-image-upload"
                name="listing_image_upload"
                onChange={(event) => onFileSelect(event.currentTarget.files?.[0] ?? null)}
                type="file"
              />
            </label>
            <p className="text-xs uppercase tracking-[0.24em] text-[var(--muted)]" style={{ fontFamily: "var(--font-mono), monospace" }}>
              {selectedImageFile ? `Ready: ${selectedImageFile.name}` : "No image selected yet"}
            </p>
          </div>

          <button
            className="inline-flex items-center justify-center rounded-full bg-[var(--accent)] px-5 py-3 text-sm font-semibold text-white transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
            disabled={isBusy}
            onClick={() => void onUpload()}
            type="button"
          >
            {imageMutationState.status === "uploading" ? "Uploading image..." : "Upload image"}
          </button>
        </div>
      </div>

      {imageMutationState.message ? (
        <p
          aria-live="polite"
          className={imageMutationState.status === "error" ? "text-sm font-medium text-red-700" : "text-sm font-medium text-emerald-700"}
          role={imageMutationState.status === "error" ? "alert" : undefined}
        >
          {imageMutationState.message}
        </p>
      ) : null}

      {images.length === 0 ? (
        <div className="rounded-[1.5rem] border border-dashed border-[var(--line)] bg-white/60 p-5 text-sm leading-7 text-[var(--muted)]">
          No images yet. Upload the first seller photo to let the backend assign ordering and primary state.
        </div>
      ) : (
        <div className="grid gap-4 lg:grid-cols-2">
          {images.map((image, index) => {
            const isTarget = imageMutationState.targetId === image.id;

            return (
              <article
                className="overflow-hidden rounded-[1.5rem] border border-[var(--line)] bg-white/72"
                data-testid={`listing-image-card-${image.id}`}
                key={image.id}
              >
                <div
                  aria-label={image.original_filename ?? `Listing image ${index + 1}`}
                  className="aspect-[4/3] bg-[var(--panel)] bg-cover bg-center"
                  role="img"
                  style={{ backgroundImage: `url(${image.url})` }}
                />
                <div className="space-y-4 p-5">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="rounded-full bg-[var(--panel)] px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em] text-[var(--muted)]">
                      Order {image.sort_order + 1}
                    </span>
                    {image.is_primary ? (
                      <span className="rounded-full bg-emerald-100 px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em] text-emerald-700">
                        Primary
                      </span>
                    ) : null}
                    {isTarget && isBusy ? (
                      <span className="rounded-full bg-amber-100 px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em] text-amber-700">
                        Updating
                      </span>
                    ) : null}
                  </div>

                  <div className="space-y-1">
                    <p className="text-sm font-medium text-[var(--ink)]">{image.original_filename ?? `Listing image ${index + 1}`}</p>
                    <p className="text-sm text-[var(--muted)]">Backend sort index {image.sort_order}</p>
                  </div>

                  <div className="flex flex-wrap gap-3">
                    <button
                      className={secondaryButtonClassName}
                      disabled={isBusy || image.is_primary}
                      onClick={() => void onSetPrimaryImage(image.id)}
                      type="button"
                    >
                      Set primary
                    </button>
                    <button
                      aria-label={`Move ${image.original_filename ?? `image ${index + 1}`} earlier`}
                      className={secondaryButtonClassName}
                      disabled={isBusy || index === 0}
                      onClick={() => void onMoveImage(image.id, "earlier")}
                      type="button"
                    >
                      Move earlier
                    </button>
                    <button
                      aria-label={`Move ${image.original_filename ?? `image ${index + 1}`} later`}
                      className={secondaryButtonClassName}
                      disabled={isBusy || index === images.length - 1}
                      onClick={() => void onMoveImage(image.id, "later")}
                      type="button"
                    >
                      Move later
                    </button>
                    <button
                      className="inline-flex items-center justify-center rounded-full border border-red-200 bg-red-50 px-4 py-2 text-sm font-semibold text-red-700 transition hover:border-red-300 hover:bg-red-100 disabled:cursor-not-allowed disabled:opacity-60"
                      disabled={isBusy}
                      onClick={() => void onDeleteImage(image.id)}
                      type="button"
                    >
                      Delete image
                    </button>
                  </div>
                </div>
              </article>
            );
          })}
        </div>
      )}
    </div>
  );
}

function getPageCopy(mode: ListingFormMode) {
  if (mode === "create") {
    return {
      eyebrow: "Create listing",
      title: "Publish a new property draft",
      body: "Start with the listing fields the backend already accepts, then add media management in a later flow.",
      loadingBody: "Category options and the listing contract are loading so the seller form can start from backend-backed data.",
    };
  }

  return {
    eyebrow: "Edit listing",
    title: "Refine an existing property record",
    body: "This editor hydrates from the authenticated seller inventory before posting the update payload back through the seller flow.",
    loadingBody: "Existing seller-owned listing details and category options are loading so the form can hydrate safely.",
  };
}

function toFormValues(listing: ListingRecord): ListingFormValues {
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

function toRequestPayload(values: ListingFormValues): ListingFormRequest {
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

function normalizeNullableString(value: string): string | null {
  const trimmed = value.trim();
  return trimmed ? trimmed : null;
}

function normalizeRequiredInteger(value: string, fallback: number): number {
  const parsed = Number.parseInt(value, 10);

  if (!Number.isNaN(parsed) && parsed >= fallback) {
    return parsed;
  }

  return fallback;
}

function normalizeOptionalInteger(value: string, fallback: number): number {
  const parsed = Number.parseInt(value, 10);

  if (!Number.isNaN(parsed) && parsed >= 0) {
    return parsed;
  }

  return fallback;
}

function normalizeStatus(status: string): ListingFormRequest["status"] {
  if (status === "inactive" || status === "sold") {
    return status;
  }

  return "active";
}

function sortListingImages(images: ListingImageRecord[]): ListingImageRecord[] {
  return [...images].sort((left, right) => {
    if (left.sort_order === right.sort_order) {
      return left.created_at.localeCompare(right.created_at);
    }

    return left.sort_order - right.sort_order;
  });
}

function ListingFormState({
  eyebrow,
  title,
  body,
  action,
}: {
  eyebrow: string;
  title: string;
  body: string;
  action?: React.ReactNode;
}) {
  return (
    <section className="rounded-[1.75rem] border border-[var(--line)] bg-white/72 p-8">
      <p
        className="text-xs uppercase tracking-[0.3em] text-[var(--muted)]"
        style={{ fontFamily: "var(--font-mono), monospace" }}
      >
        {eyebrow}
      </p>
      <h2 className="mt-4 text-2xl font-semibold tracking-[-0.03em] text-[var(--ink)]">{title}</h2>
      <p className="mt-3 max-w-2xl text-sm leading-7 text-[var(--muted)] sm:text-base">{body}</p>
      {action ? <div className="mt-6">{action}</div> : null}
    </section>
  );
}

function FieldGroup({
  title,
  description,
  children,
}: {
  title: string;
  description: string;
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-5">
      <div className="space-y-2">
        <h3 className="text-xl font-semibold tracking-[-0.03em] text-[var(--ink)]">{title}</h3>
        <p className="text-sm leading-7 text-[var(--muted)]">{description}</p>
      </div>
      <div className="space-y-5">{children}</div>
    </div>
  );
}

function Field({
  htmlFor,
  label,
  required,
  children,
}: {
  htmlFor: string;
  label: string;
  required?: boolean;
  children: React.ReactNode;
}) {
  return (
    <label className="block space-y-2 text-sm font-medium text-[var(--ink)]" htmlFor={htmlFor}>
      <span>
        {label}
        {required ? <span className="ml-1 text-[var(--accent)]">*</span> : null}
      </span>
      {children}
    </label>
  );
}

const inputClassName =
  "w-full rounded-[1rem] border border-[var(--line)] bg-[var(--panel)] px-4 py-3 text-sm text-[var(--ink)] outline-none transition placeholder:text-[var(--muted)] focus:border-[var(--accent)] focus:ring-2 focus:ring-[color:var(--accent)]/15";

const secondaryButtonClassName =
  "inline-flex items-center justify-center rounded-full border border-[var(--line)] bg-[var(--panel)] px-4 py-2 text-sm font-semibold text-[var(--ink)] transition hover:border-[var(--accent)] hover:text-[var(--accent)] disabled:cursor-not-allowed disabled:opacity-60";

"use client";

import Image from "next/image";
import Link from "next/link";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { type ChangeEvent, useEffect, useMemo, useState } from "react";
import { useForm, useWatch } from "react-hook-form";

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
  getRegionCities,
  getRegionDistricts,
  getRegionProvinces,
  getRegionVillages,
  parseListingSpecifications,
  parseStringList,
  normalizeNullableNumber,
  normalizeOptionalIntegerOrNull,
  normalizeStringList,
  updateSellerListing,
  type ListingCategoryOption,
  type ListingFormRequest,
  type ListingImageRecord,
  type ListingRecord,
  type RegionOption,
  type ListingTransactionType,
} from "@/lib/api/listing-form";
import { queryKeys } from "@/lib/query/keys";

import {
  deleteSellerListingImage,
  deleteSellerListingVideo,
  reorderSellerListingImages,
  setSellerPrimaryListingImage,
  uploadSellerListingImages,
  uploadSellerListingVideo,
} from "@/features/listings/images/api";

import { type ListingFormSchema, listingFormSchema } from "./listing-schema";
import {
  describeExistingListingVideo,
  describeSelectedImageFiles,
  formatDuration,
  formatVideoBytes,
  inspectListingImageSelection,
  MAX_LISTING_VIDEO_BYTES,
  RECOMMENDED_LISTING_IMAGE_RATIO_LABEL,
  validateListingVideoSelection,
} from "./listing-media";

type ListingFormMode = "create" | "edit";

type ListingFormProps = {
  initialListing?: ListingRecord | null;
  mode: ListingFormMode;
  listingId?: string;
};

type FeedbackMessage = {
  tone: "error" | "info" | "success";
  text: string;
};

const STATUS_OPTIONS: Array<{ value: ListingFormRequest["status"]; label: string }> = [
  { value: "active", label: "Active" },
  { value: "inactive", label: "Inactive" },
  { value: "sold", label: "Sold" },
  { value: "draft", label: "Draft" },
  { value: "archived", label: "Archived" },
];

const TRANSACTION_TYPE_OPTIONS: Array<{ value: ListingTransactionType; label: string }> = [
  { value: "sale", label: "For sale" },
  { value: "rent", label: "For rent" },
];

const CONDITION_OPTIONS = [
  { value: "new", label: "Properti Baru" },
  { value: "second", label: "Properti Second" },
] as const;

const CERTIFICATE_TYPE_OPTIONS = [
  { value: "SHM", label: "SHM" },
  { value: "HGB", label: "HGB" },
  { value: "Hak Pakai", label: "Hak Pakai" },
  { value: "Hak Sewa", label: "Hak Sewa" },
  { value: "HGU", label: "HGU" },
  { value: "Adat", label: "Adat" },
  { value: "Girik", label: "Girik" },
  { value: "PPJB", label: "PPJB" },
  { value: "Strata", label: "Strata" },
  { value: "Lainnya", label: "Lainnya" },
] as const;

const FURNISHING_OPTIONS = [
  { value: "furnished", label: "Furnished" },
  { value: "semi", label: "Semi Furnished" },
  { value: "unfurnished", label: "Unfurnished" },
] as const;

const FACILITY_OPTIONS = ["AC", "Akses Parkir", "CCTV", "Keamanan", "Wifi / Internet"] as const;

export function ListingForm({ initialListing = null, mode, listingId }: ListingFormProps) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [listing, setListing] = useState<ListingRecord | null>(initialListing);
  const [selectedImageFiles, setSelectedImageFiles] = useState<File[]>([]);
  const [selectedVideoFile, setSelectedVideoFile] = useState<File | null>(null);
  const [formError, setFormError] = useState<string | null>(null);
  const [formSuccess, setFormSuccess] = useState<string | null>(null);
  const [imageMessage, setImageMessage] = useState<FeedbackMessage | null>(null);
  const [videoMessage, setVideoMessage] = useState<FeedbackMessage | null>(null);
  const [isImagePrecheckPending, setIsImagePrecheckPending] = useState(false);
  const [isVideoPrecheckPending, setIsVideoPrecheckPending] = useState(false);
  const [uploadInputKey, setUploadInputKey] = useState(0);
  const [videoInputKey, setVideoInputKey] = useState(0);
  const activeListingId = listingId ?? listing?.id ?? null;
  const effectiveMode: ListingFormMode = activeListingId ? "edit" : mode;

  const categoriesQuery = useQuery({
    queryKey: queryKeys.categories,
    queryFn: () => getListingCategories(),
  });

  const form = useForm<ListingFormSchema>({
    resolver: zodResolver(listingFormSchema),
    defaultValues: toFormValues(initialListing),
  });

  const selectedProvinceCode = useWatch({ control: form.control, name: "location_province_code" });
  const selectedCityCode = useWatch({ control: form.control, name: "location_city_code" });
  const selectedDistrictCode = useWatch({ control: form.control, name: "location_district_code" });
  const selectedFacilitiesValue = useWatch({ control: form.control, name: "facilities" });

  const provincesQuery = useQuery({
    queryKey: ["regions", "provinces"],
    queryFn: () => getRegionProvinces(),
  });

  const citiesQuery = useQuery({
    queryKey: ["regions", "cities", selectedProvinceCode],
    queryFn: () => getRegionCities(selectedProvinceCode),
    enabled: Boolean(selectedProvinceCode),
  });

  const districtsQuery = useQuery({
    queryKey: ["regions", "districts", selectedCityCode],
    queryFn: () => getRegionDistricts(selectedCityCode),
    enabled: Boolean(selectedCityCode),
  });

  const villagesQuery = useQuery({
    queryKey: ["regions", "villages", selectedDistrictCode],
    queryFn: () => getRegionVillages(selectedDistrictCode),
    enabled: Boolean(selectedDistrictCode),
  });

  useEffect(() => {
    const matchedProvince = findRegionOptionByName(provincesQuery.data, listing?.location_province);
    if (!matchedProvince || form.getValues("location_province_code")) {
      return;
    }

    form.setValue("location_province_code", matchedProvince.code, { shouldDirty: false });
    form.setValue("location_province", matchedProvince.name, { shouldDirty: false });
  }, [form, listing?.location_province, provincesQuery.data]);

  useEffect(() => {
    const matchedCity = findRegionOptionByName(citiesQuery.data, listing?.location_city);
    if (!matchedCity || form.getValues("location_city_code")) {
      return;
    }

    form.setValue("location_city_code", matchedCity.code, { shouldDirty: false });
    form.setValue("location_city", matchedCity.name, { shouldDirty: false });
  }, [citiesQuery.data, form, listing?.location_city]);

  useEffect(() => {
    const matchedDistrict = findRegionOptionByName(districtsQuery.data, listing?.location_district);
    if (!matchedDistrict || form.getValues("location_district_code")) {
      return;
    }

    form.setValue("location_district_code", matchedDistrict.code, { shouldDirty: false });
    form.setValue("location_district", matchedDistrict.name, { shouldDirty: false });
  }, [districtsQuery.data, form, listing?.location_district]);

  useEffect(() => {
    const matchedVillage = findRegionOptionByName(villagesQuery.data, listing?.location_village);
    if (!matchedVillage || form.getValues("location_village_code")) {
      return;
    }

    form.setValue("location_village_code", matchedVillage.code, { shouldDirty: false });
    form.setValue("location_village", matchedVillage.name, { shouldDirty: false });
  }, [form, listing?.location_village, villagesQuery.data]);

  const orderedImages = useMemo(() => sortListingImages(listing?.images ?? []), [listing]);
  const listingVideo = listing?.video ?? null;

  const submitMutation = useMutation({
    mutationFn: async (values: ListingFormSchema) => {
      const payload = toRequestPayload(values);

      if (effectiveMode === "create") {
        return createSellerListing(payload);
      }

      if (!activeListingId) {
        throw new Error("Listing id is required for edit mode.");
      }

      return updateSellerListing(activeListingId, payload);
    },
    onSuccess: (nextListing) => {
      setFormError(null);
      setFormSuccess(effectiveMode === "create" ? null : "Listing changes saved successfully.");
      setListing(nextListing);
      queryClient.invalidateQueries({ queryKey: queryKeys.sellerListings });

      if (effectiveMode === "create") {
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

  const ensureListingForMedia = async () => {
    if (activeListingId) {
      return activeListingId;
    }

    const isValid = await form.trigger();
    if (!isValid) {
      throw new Error("Complete the required listing fields before uploading media.");
    }

    const createdListing = await createSellerListing({
      ...toRequestPayload(form.getValues()),
      status: "draft",
    });

    setListing(createdListing);
    form.reset(toFormValues(createdListing));
    setFormError(null);
    setFormSuccess("Draft listing created automatically so you can continue uploading media.");
    queryClient.invalidateQueries({ queryKey: queryKeys.sellerListings });

    return createdListing.id;
  };

  const imageMutation = useMutation({
    mutationFn: async (action: {
      type: "upload" | "set-primary" | "delete" | "reorder";
      imageId?: string;
      direction?: "earlier" | "later";
    }) => {
      const ensuredListingId = action.type === "upload" ? await ensureListingForMedia() : activeListingId;

      if (!ensuredListingId) {
        throw new Error("Save the listing first before updating images.");
      }

      if (action.type === "upload") {
        if (selectedImageFiles.length === 0) {
          throw new Error("Choose at least one image before uploading.");
        }

        return uploadSellerListingImages(ensuredListingId, selectedImageFiles);
      }

      if (action.type === "set-primary" && action.imageId) {
        return setSellerPrimaryListingImage(ensuredListingId, action.imageId);
      }

      if (action.type === "delete" && action.imageId) {
        return deleteSellerListingImage(ensuredListingId, action.imageId);
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

        return reorderSellerListingImages(ensuredListingId, reorderedIds);
      }

      throw new Error("Unsupported image action.");
    },
    onSuccess: (nextListing, action) => {
      setListing(nextListing);
      setImageMessage({ tone: "success", text: getImageSuccessMessage(action.type, selectedImageFiles.length) });
      queryClient.invalidateQueries({ queryKey: queryKeys.sellerListings });
      if (mode === "create" && !listingId) {
        router.push(`/dashboard/listings/${nextListing.id}/edit?created=1&media=1`);
      } else {
        router.refresh();
      }

      if (action.type === "upload") {
        setSelectedImageFiles([]);
        setUploadInputKey((current) => current + 1);
      }
    },
    onError: (error) => {
      setImageMessage({ tone: "error", text: formatListingFormError(error) });
    },
  });

  const videoMutation = useMutation({
    mutationFn: async (action: { type: "upload" | "delete" }) => {
      const ensuredListingId = action.type === "upload" ? await ensureListingForMedia() : activeListingId;

      if (!ensuredListingId) {
        throw new Error("Save the listing first before updating media.");
      }

      if (action.type === "delete") {
        return deleteSellerListingVideo(ensuredListingId);
      }

      if (listingVideo) {
        throw new Error("Delete the current video before uploading another.");
      }

      if (!selectedVideoFile) {
        throw new Error("Choose a video file before uploading.");
      }

      return uploadSellerListingVideo(ensuredListingId, selectedVideoFile);
    },
    onSuccess: (nextListing, action) => {
      setListing(nextListing);
      setVideoMessage({
        tone: "success",
        text:
          action.type === "upload"
            ? "Video uploaded. The slot now reflects the backend response."
            : "Video removed. The slot now reflects the backend response.",
      });
      setSelectedVideoFile(null);
      setVideoInputKey((current) => current + 1);
      queryClient.invalidateQueries({ queryKey: queryKeys.sellerListings });
      if (mode === "create" && !listingId) {
        router.push(`/dashboard/listings/${nextListing.id}/edit?created=1&media=1`);
      } else {
        router.refresh();
      }
    },
    onError: (error) => {
      setVideoMessage({ tone: "error", text: formatListingFormError(error) });
    },
  });

  const handleVideoSelection = async (event: ChangeEvent<HTMLInputElement>) => {
    const nextFile = event.currentTarget.files?.[0] ?? null;

    setSelectedVideoFile(null);

    if (!nextFile) {
      setVideoMessage(null);
      return;
    }

    if (listingVideo) {
      setVideoMessage({ tone: "error", text: "Delete the current video before uploading another." });
      setVideoInputKey((current) => current + 1);
      return;
    }

    setIsVideoPrecheckPending(true);
    const precheck = await validateListingVideoSelection(nextFile);
    setIsVideoPrecheckPending(false);

    if (!precheck.ok) {
      setVideoMessage({ tone: "error", text: precheck.message });
      setVideoInputKey((current) => current + 1);
      return;
    }

    setSelectedVideoFile(nextFile);
    setVideoMessage({ tone: "info", text: precheck.message });
  };

  const handleImageSelection = async (event: ChangeEvent<HTMLInputElement>) => {
    const nextFiles = Array.from(event.currentTarget.files ?? []);

    setSelectedImageFiles(nextFiles);

    if (nextFiles.length === 0) {
      setImageMessage(null);
      return;
    }

    setIsImagePrecheckPending(true);
    const precheck = await inspectListingImageSelection(nextFiles);
    setIsImagePrecheckPending(false);
    setImageMessage({ tone: "info", text: precheck.message });
  };

  if (categoriesQuery.isError) {
    return (
      <section className="rounded-[1.75rem] border border-slate-200 bg-white/72 p-8">
        <p className="text-xs uppercase tracking-[0.3em] text-slate-900">Listing form</p>
        <h2 className="mt-4 text-2xl font-semibold tracking-[-0.03em] text-slate-900">We could not prepare this listing form</h2>
        <p className="mt-3 text-sm leading-7 text-slate-900">{formatListingFormError(categoriesQuery.error)}</p>
        <Button className="mt-6" onClick={() => void categoriesQuery.refetch()} type="button" variant="secondary">
          Retry
        </Button>
      </section>
    );
  }

  return (
    <div className="space-y-6">
      <section className="flex flex-col gap-4 rounded-[1.75rem] border border-slate-200 bg-white/72 p-6 sm:p-8 lg:flex-row lg:items-end lg:justify-between">
        <div className="space-y-3">
          <p className="text-xs uppercase tracking-[0.3em] text-slate-900" style={{ fontFamily: "var(--font-mono), monospace" }}>
            {effectiveMode === "create" ? "Buat Listing Baru" : "Edit Listing"}
          </p>
          <div className="space-y-2">
            <h2 className="text-3xl font-semibold tracking-[-0.04em] text-slate-900">
              {effectiveMode === "create" ? "Tambah Properti Baru" : "Edit Data Properti"}
            </h2>
            <p className="max-w-3xl text-sm leading-7 text-slate-900 sm:text-base">
              Lengkapi formulir di bawah ini dengan detail properti yang valid untuk menarik minat pembeli atau penyewa.
            </p>
          </div>
        </div>

        <Link className="inline-flex items-center rounded-full border border-slate-200 bg-slate-50 px-5 py-3 text-sm font-semibold text-slate-900 transition hover:border-slate-900 hover:text-slate-900" href="/dashboard/listings">
          Back to listings
        </Link>
      </section>

      <Form {...form}>
        <form className="space-y-6" onSubmit={form.handleSubmit((values) => submitMutation.mutate(values))}>
          <section className="grid gap-6 xl:grid-cols-[1.15fr_0.85fr]">
            <div className="space-y-6 rounded-[1.75rem] border border-slate-200 bg-white/80 p-6 sm:p-8">
              <div className="space-y-5">
                <div className="space-y-2">
                  <h3 className="text-xl font-semibold tracking-[-0.03em] text-slate-900">Informasi Dasar</h3>
                  <p className="text-sm leading-7 text-slate-900">Informasi utama mengenai properti yang akan ditampilkan ke publik.</p>
                </div>

                <FormField
                  control={form.control}
                  name="title"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel htmlFor="listing-title">Judul Listing</FormLabel>
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
                      <FormLabel htmlFor="listing-description">Deskripsi Properti</FormLabel>
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
                        <FormLabel htmlFor="listing-category">Kategori Properti</FormLabel>
                        <FormControl>
                          <Select aria-label="Category" id="listing-category" {...field} value={field.value ?? ""}>
                            <option value="">Pilih kategori</option>
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
                    name="transaction_type"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel htmlFor="listing-transaction-type">Tipe Transaksi</FormLabel>
                        <FormControl>
                          <Select aria-label="Transaction type" id="listing-transaction-type" {...field}>
                            {TRANSACTION_TYPE_OPTIONS.map((option) => (
                              <option key={option.value} value={option.value}>
                                {option.label}
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
                        <FormLabel htmlFor="listing-status">Status Publikasi</FormLabel>
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

                <div className="grid gap-5 sm:grid-cols-2">
                  <FormField
                    control={form.control}
                    name="currency"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel htmlFor="listing-currency">Mata Uang</FormLabel>
                        <FormControl>
                          <Input id="listing-currency" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="is_negotiable"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel htmlFor="listing-is-negotiable">Harga Dapat Dinegosiasikan</FormLabel>
                        <FormControl>
                          <label className="flex h-10 items-center gap-3 rounded-full border border-slate-200 bg-white px-4 text-sm text-slate-900">
                            <input
                              checked={field.value}
                              id="listing-is-negotiable"
                              onChange={(event) => field.onChange(event.target.checked)}
                              type="checkbox"
                            />
                            <span>Bisa ditiadakan / Nego</span>
                          </label>
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>

                <FormField
                  control={form.control}
                  name="special_offers"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel htmlFor="listing-special-offers">Promo / Penawaran Khusus</FormLabel>
                      <FormControl>
                        <Input id="listing-special-offers" placeholder="Promo, DP_0, Turun_Harga" {...field} value={field.value ?? ""} />
                      </FormControl>
                      <FormDescription>Pisahkan dengan koma (contoh: Promo, DP_0, Turun_Harga).</FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="price"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel htmlFor="listing-price">Harga (IDR)</FormLabel>
                      <FormControl>
                        <Input id="listing-price" inputMode="numeric" min="1" pattern="[0-9]*" {...field} />
                      </FormControl>
                      <FormDescription>Pastikan memasukkan angka saja tanpa titik atau koma.</FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            </div>

            <div className="space-y-6 rounded-[1.75rem] border border-slate-200 bg-white/80 p-6 sm:p-8">
              <div className="space-y-5">
                <div className="space-y-2">
                  <h3 className="text-xl font-semibold tracking-[-0.03em] text-slate-900">Lokasi & Detail Properti</h3>
                  <p className="text-sm leading-7 text-slate-900">Lokasi properti serta spesifikasi ruangan dan bentuk bangunan.</p>
                </div>

                <div className="grid gap-5 sm:grid-cols-2">
                  <FormField
                    control={form.control}
                    name="location_province_code"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel htmlFor="field-location-province">Provinsi</FormLabel>
                        <FormControl>
                          <Select
                            aria-label="Province"
                            id="field-location-province"
                            value={field.value ?? ""}
                            onChange={(event) => {
                              const nextCode = event.target.value;
                              const option = (provincesQuery.data ?? []).find((item) => item.code === nextCode) ?? null;
                              field.onChange(nextCode);
                              form.setValue("location_province", option?.name ?? "", { shouldDirty: true, shouldValidate: true });
                              form.setValue("location_city_code", "", { shouldDirty: true, shouldValidate: true });
                              form.setValue("location_city", "", { shouldDirty: true });
                              form.setValue("location_district_code", "", { shouldDirty: true, shouldValidate: true });
                              form.setValue("location_district", "", { shouldDirty: true });
                              form.setValue("location_village_code", "", { shouldDirty: true, shouldValidate: true });
                              form.setValue("location_village", "", { shouldDirty: true });
                            }}
                          >
                            <option value="">Pilih provinsi</option>
                            {(provincesQuery.data ?? []).map((option) => (
                              <option key={option.code} value={option.code}>
                                {option.name}
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
                    name="location_city_code"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel htmlFor="field-location-city">Kota / Kabupaten</FormLabel>
                        <FormControl>
                          <Select
                            aria-label="City"
                            disabled={!selectedProvinceCode || citiesQuery.isLoading}
                            id="field-location-city"
                            value={field.value ?? ""}
                            onChange={(event) => {
                              const nextCode = event.target.value;
                              const option = (citiesQuery.data ?? []).find((item) => item.code === nextCode) ?? null;
                              field.onChange(nextCode);
                              form.setValue("location_city", option?.name ?? "", { shouldDirty: true, shouldValidate: true });
                              form.setValue("location_district_code", "", { shouldDirty: true, shouldValidate: true });
                              form.setValue("location_district", "", { shouldDirty: true });
                              form.setValue("location_village_code", "", { shouldDirty: true, shouldValidate: true });
                              form.setValue("location_village", "", { shouldDirty: true });
                            }}
                          >
                            <option value="">{selectedProvinceCode ? "Select city" : "Choose province first"}</option>
                            {(citiesQuery.data ?? []).map((option) => (
                              <option key={option.code} value={option.code}>
                                {option.name}
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
                    name="location_district_code"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel htmlFor="field-location-district">Kecamatan</FormLabel>
                        <FormControl>
                          <Select
                            aria-label="District"
                            disabled={!selectedCityCode || districtsQuery.isLoading}
                            id="field-location-district"
                            value={field.value ?? ""}
                            onChange={(event) => {
                              const nextCode = event.target.value;
                              const option = (districtsQuery.data ?? []).find((item) => item.code === nextCode) ?? null;
                              field.onChange(nextCode);
                              form.setValue("location_district", option?.name ?? "", { shouldDirty: true, shouldValidate: true });
                              form.setValue("location_village_code", "", { shouldDirty: true, shouldValidate: true });
                              form.setValue("location_village", "", { shouldDirty: true });
                            }}
                          >
                            <option value="">{selectedCityCode ? "Select district" : "Choose city first"}</option>
                            {(districtsQuery.data ?? []).map((option) => (
                              <option key={option.code} value={option.code}>
                                {option.name}
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
                    name="location_village_code"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel htmlFor="field-location-village">Desa / Kelurahan</FormLabel>
                        <FormControl>
                          <Select
                            aria-label="Village / Kelurahan"
                            disabled={!selectedDistrictCode || villagesQuery.isLoading}
                            id="field-location-village"
                            value={field.value ?? ""}
                            onChange={(event) => {
                              const nextCode = event.target.value;
                              const option = (villagesQuery.data ?? []).find((item) => item.code === nextCode) ?? null;
                              field.onChange(nextCode);
                              form.setValue("location_village", option?.name ?? "", { shouldDirty: true, shouldValidate: true });
                            }}
                          >
                            <option value="">{selectedDistrictCode ? "Select village" : "Choose district first"}</option>
                            {(villagesQuery.data ?? []).map((option) => (
                              <option key={option.code} value={option.code}>
                                {option.name}
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
                  name="condition"
                  render={({ field }) => (
                    <div className="sm:col-span-2">
                      <FormItem>
                      <FormLabel>Kondisi</FormLabel>
                      <FormControl>
                        <div className="flex flex-wrap gap-3">
                          {CONDITION_OPTIONS.map((option) => (
                            <button
                              key={option.value}
                              className={getChoiceButtonClassName(field.value === option.value)}
                              onClick={() => field.onChange(field.value === option.value ? "" : option.value)}
                              type="button"
                            >
                              {option.label}
                            </button>
                          ))}
                        </div>
                      </FormControl>
                      <FormMessage />
                      </FormItem>
                    </div>
                  )}
                />

                <FormField
                  control={form.control}
                  name="certificate_type"
                  render={({ field }) => (
                    <div className="sm:col-span-2">
                      <FormItem>
                      <FormLabel>Sertifikat</FormLabel>
                      <FormControl>
                        <div className="flex flex-wrap gap-3">
                          {CERTIFICATE_TYPE_OPTIONS.map((option) => (
                            <button
                              key={option.value}
                              className={getChoiceButtonClassName(field.value === option.value)}
                              onClick={() => field.onChange(field.value === option.value ? "" : option.value)}
                              type="button"
                            >
                              {option.label}
                            </button>
                          ))}
                        </div>
                      </FormControl>
                      <FormMessage />
                      </FormItem>
                    </div>
                  )}
                />

                <FormField
                  control={form.control}
                  name="furnishing"
                  render={({ field }) => (
                    <div className="sm:col-span-2">
                      <FormItem>
                      <FormLabel>Kondisi Perabotan</FormLabel>
                      <FormControl>
                        <div className="flex flex-wrap gap-3">
                          {FURNISHING_OPTIONS.map((option) => (
                            <button
                              key={option.value}
                              className={getChoiceButtonClassName(field.value === option.value)}
                              onClick={() => field.onChange(field.value === option.value ? "" : option.value)}
                              type="button"
                            >
                              {option.label}
                            </button>
                          ))}
                        </div>
                      </FormControl>
                      <FormMessage />
                      </FormItem>
                    </div>
                  )}
                />

                <FormField
                  control={form.control}
                  name="facilities"
                  render={() => {
                    const selectedFacilities = new Set(parseStringList(selectedFacilitiesValue));

                    return (
                      <div className="sm:col-span-2">
                        <FormItem>
                        <FormLabel>Fasilitas Properti</FormLabel>
                        <FormControl>
                          <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
                            {FACILITY_OPTIONS.map((option) => {
                              const selected = selectedFacilities.has(option);

                              return (
                                <button
                                  key={option}
                                  className={`flex items-center justify-between rounded-2xl border px-4 py-3 text-sm font-medium transition ${selected ? "border-slate-900 bg-slate-900/10 text-slate-900" : "border-slate-200 bg-white text-slate-900 hover:border-slate-900/50"}`}
                                  onClick={() => {
                                    const next = new Set(parseStringList(form.getValues("facilities")));
                                    if (selected) {
                                      next.delete(option);
                                    } else {
                                      next.add(option);
                                    }
                                    form.setValue("facilities", Array.from(next).join(", "), {
                                      shouldDirty: true,
                                      shouldValidate: true,
                                    });
                                  }}
                                  type="button"
                                >
                                  <span>{option}</span>
                                  <span
                                    aria-hidden="true"
                                    className={`h-5 w-5 rounded-md border ${selected ? "border-slate-900 bg-slate-900" : "border-slate-200 bg-transparent"}`}
                                  />
                                </button>
                              );
                            })}
                          </div>
                        </FormControl>
                        <FormMessage />
                        </FormItem>
                      </div>
                    );
                  }}
                />

                {(
                  [
                    ["address_detail", "Alamat Lengkap (Detail Jalan/Blok)"],
                    ["latitude", "Latitude (Garis Lintang)"],
                    ["longitude", "Longitude (Garis Bujur)"],
                    ["bedrooms", "Kamar Tidur"],
                    ["bathrooms", "Kamar Mandi"],
                    ["floor_count", "Jumlah Lantai"],
                    ["carport_capacity", "Kapasitas Carport / Kendaraan"],
                    ["land_area_sqm", "Luas Tanah (m²)"],
                    ["building_area_sqm", "Luas Bangunan (m²)"],
                    ["electrical_power_va", "Daya Listrik (VA)"],
                    ["facing_direction", "Arah Bangunan"],
                    ["year_built", "Tahun Pembangunan"],
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
                            <Input id={`field-${name}`} inputMode={name.includes("sqm") || name === "bedrooms" || name === "bathrooms" || name === "floor_count" || name === "carport_capacity" || name === "year_built" || name === "electrical_power_va" || name === "latitude" || name === "longitude" ? "numeric" : undefined} pattern={name.includes("sqm") || name === "bedrooms" || name === "bathrooms" || name === "floor_count" || name === "carport_capacity" || name === "year_built" || name === "electrical_power_va" ? "[0-9]*" : undefined} {...field} value={field.value ?? ""} />
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

          <section className="rounded-[1.75rem] border border-slate-200 bg-white/80 p-6 sm:p-8">
            <div className="space-y-3">
              <h3 className="text-xl font-semibold tracking-[-0.03em] text-slate-900">Media Properti</h3>
              <p className="text-sm leading-7 text-slate-900">Unggah foto dan video terbaik agar properti lebih menarik. Anda dapat mengatur ulang urutan foto setelah diunggah.</p>
            </div>

            <div className="mt-6 space-y-6">
              {!activeListingId ? (
                <div className="rounded-[1.25rem] border border-dashed border-slate-200 bg-white/70 p-4 text-sm leading-7 text-slate-900">
                  Catatan: Mengubah media otomatis memberlakukan sistem autosave.
                </div>
              ) : null}

              <div className="mt-6 space-y-6">
                <div className="grid gap-4 xl:grid-cols-[1.1fr_0.9fr]">
                  <div className="rounded-3xl border border-slate-200 bg-slate-50 p-5">
                    <div className="flex h-full flex-col gap-4 lg:justify-between">
                      <div className="space-y-2">
                        <p className="text-sm font-medium text-slate-900">Unggah Foto (Bisa Lebih Dari Satu)</p>
                        <p className="text-sm leading-7 text-slate-900">Pilih satu atau beberapa foto sekaligus. Rasio gambar yang disarankan: {RECOMMENDED_LISTING_IMAGE_RATIO_LABEL}.</p>
                      </div>
                      <div className="space-y-3">
                        <label className="block text-sm font-medium text-slate-900" htmlFor="listing-image-upload">
                          <span className="sr-only">Choose listing images</span>
                          <input
                            key={uploadInputKey}
                            accept="image/*"
                            className="block w-full text-sm text-slate-900 file:mr-4 file:rounded-full file:border-0 file:bg-slate-900 file:px-4 file:py-2 file:text-sm file:font-semibold file:text-white"
                            data-testid="listing-image-upload"
                            id="listing-image-upload"
                            multiple
                            name="listing_image_upload"
                            onChange={(event) => {
                              void handleImageSelection(event);
                            }}
                            type="file"
                          />
                        </label>
                        <p className="text-xs uppercase tracking-[0.24em] text-slate-500" style={{ fontFamily: "var(--font-mono), monospace" }}>
                          {describeSelectedImageFiles(selectedImageFiles)}
                        </p>
                        <Button disabled={imageMutation.isPending || isImagePrecheckPending || selectedImageFiles.length === 0} onClick={() => imageMutation.mutate({ type: "upload" })} type="button">
                          {isImagePrecheckPending ? "Memeriksa file..." : imageMutation.isPending ? "Mengunggah gambar..." : "Unggah Gambar"}
                        </Button>
                      </div>
                    </div>
                  </div>

                  <div className="rounded-3xl border border-slate-200 bg-slate-50 p-5">
                    <div className="space-y-4">
                      <div className="space-y-2">
                        <p className="text-sm font-medium text-slate-900">Optional listing video</p>
                        <p className="text-sm leading-7 text-slate-900">Hanya dapat mengunggah maksimal satu video tayangan properti dengan batas upload {formatVideoBytes(MAX_LISTING_VIDEO_BYTES)}.</p>
                      </div>

                      {listingVideo ? (
                        <div className="space-y-4 rounded-[1.25rem] border border-slate-200 bg-white/80 p-4">
                          <div className="overflow-hidden rounded-2xl border border-slate-200 bg-black/90">
                            <video className="aspect-video h-full w-full object-cover" controls preload="metadata" src={listingVideo.url}>
                              <track kind="captions" label="Listing video captions unavailable" />
                            </video>
                          </div>
                          <div className="space-y-3">
                            <div className="flex flex-wrap items-center gap-2">
                              <span className="rounded-full bg-slate-50 px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em] text-slate-900">Single slot</span>
                              {listingVideo.duration_seconds != null ? <span className="rounded-full bg-white px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em] text-slate-900">{formatDuration(listingVideo.duration_seconds)}</span> : null}
                            </div>
                            <p className="text-sm font-medium text-slate-900">{describeExistingListingVideo(listingVideo.original_filename, listingVideo.duration_seconds)}</p>
                            <p className="text-sm leading-7 text-slate-900">Hapus video di atas ini terlebih dahulu jika Anda ingin menggantinya dengan video versi baru.</p>
                            <Button disabled={videoMutation.isPending} onClick={() => videoMutation.mutate({ type: "delete" })} type="button" variant="destructive">
                              {videoMutation.isPending ? "Menghapus video..." : "Hapus Video"}
                            </Button>
                          </div>
                        </div>
                      ) : (
                        <div className="rounded-[1.25rem] border border-dashed border-slate-200 bg-white/70 p-4 text-sm leading-7 text-slate-900">
                          Belum ada video tur properti. Tambahkan sebuah klip bila tersedia.
                        </div>
                      )}

                      <label className="block text-sm font-medium text-slate-900" htmlFor="listing-video-upload">
                        <span className="sr-only">Choose listing video</span>
                        <input
                          key={videoInputKey}
                          accept="video/*,.mp4,.mov,.m4v,.webm,.mkv,.flv,.avi,.mpg,.mpeg,.ogv"
                          className="block w-full text-sm text-slate-900 file:mr-4 file:rounded-full file:border-0 file:bg-slate-900 file:px-4 file:py-2 file:text-sm file:font-semibold file:text-white disabled:cursor-not-allowed disabled:opacity-60"
                          data-testid="listing-video-upload"
                          disabled={Boolean(listingVideo) || isVideoPrecheckPending || videoMutation.isPending}
                          id="listing-video-upload"
                          name="listing_video_upload"
                          onChange={(event) => {
                            void handleVideoSelection(event);
                          }}
                          type="file"
                        />
                      </label>
                      <Button
                        disabled={Boolean(listingVideo) || !selectedVideoFile || isVideoPrecheckPending || videoMutation.isPending}
                        onClick={() => videoMutation.mutate({ type: "upload" })}
                        type="button"
                      >
                        {isVideoPrecheckPending
                          ? "Memeriksa file..."
                          : videoMutation.isPending
                            ? "Uploading video..."
                            : "Upload video"}
                      </Button>
                    </div>
                  </div>
                </div>

                {imageMessage ? (
                  <p
                    className={getFeedbackMessageClassName(imageMessage.tone)}
                    data-testid={imageMessage.tone === "error" ? "listing-image-error" : undefined}
                    role={imageMessage.tone === "error" ? "alert" : undefined}
                  >
                    {imageMessage.text}
                  </p>
                ) : null}

                {videoMessage ? (
                  <p
                    className={getFeedbackMessageClassName(videoMessage.tone)}
                    data-testid={videoMessage.tone === "error" ? "listing-video-error" : undefined}
                    role={videoMessage.tone === "error" ? "alert" : undefined}
                  >
                    {videoMessage.text}
                  </p>
                ) : null}

                {orderedImages.length === 0 ? (
                  <div className="rounded-3xl border border-dashed border-slate-200 bg-white/60 p-5 text-sm leading-7 text-slate-900">
                    Belum ada gambar yang terunggah. Silakan pilih foto dengan tombol di atas.
                  </div>
                ) : (
                  <div className="grid gap-4 lg:grid-cols-2">
                    {orderedImages.map((image, index) => (
                      <article className="overflow-hidden rounded-3xl border border-slate-200 bg-white/72" data-testid={`listing-image-card-${image.id}`} key={image.id}>
                        <div data-testid="listing-image-item">
                          <div className="relative aspect-4/3 bg-slate-50">
                            <Image alt={image.original_filename ?? `Listing image ${index + 1}`} className="object-cover" fill sizes="(min-width: 1024px) 30vw, 100vw" src={image.url} unoptimized />
                          </div>
                          <div className="space-y-4 p-5">
                            <div className="flex flex-wrap items-center gap-2">
                              <span className="rounded-full bg-slate-50 px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em] text-slate-900">Order {image.sort_order + 1}</span>
                              {image.is_primary ? <span className="rounded-full bg-emerald-100 px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em] text-emerald-700">Primary</span> : null}
                            </div>
                            <p className="text-sm font-medium text-slate-900">{image.original_filename ?? `Listing image ${index + 1}`}</p>
                            <div className="flex flex-wrap gap-3">
                              <Button data-testid="listing-image-make-primary" disabled={imageMutation.isPending || image.is_primary} onClick={() => imageMutation.mutate({ type: "set-primary", imageId: image.id })} type="button" variant="secondary">
                                Jadikan Utama
                              </Button>
                              <Button
                                aria-label={`Move ${image.original_filename ?? `image ${index + 1}`} earlier`}
                                disabled={imageMutation.isPending || index === 0}
                                onClick={() => imageMutation.mutate({ type: "reorder", imageId: image.id, direction: "earlier" })}
                                type="button"
                                variant="secondary"
                              >
                                Pindah Kiri
                              </Button>
                              <Button
                                aria-label={`Move ${image.original_filename ?? `image ${index + 1}`} later`}
                                disabled={imageMutation.isPending || index === orderedImages.length - 1}
                                onClick={() => imageMutation.mutate({ type: "reorder", imageId: image.id, direction: "later" })}
                                type="button"
                                variant="secondary"
                              >
                                Pindah Kanan
                              </Button>
                              <Button disabled={imageMutation.isPending} onClick={() => imageMutation.mutate({ type: "delete", imageId: image.id })} type="button" variant="destructive">
                                Hapus Gambar
                              </Button>
                            </div>
                          </div>
                        </div>
                      </article>
                    ))}
                  </div>
                )}
              </div>
              </div>
          </section>

          <section className="rounded-[1.75rem] border border-slate-200 bg-slate-50 p-6 sm:p-8">
            <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
              <div className="space-y-2">
                <p className="text-xs uppercase tracking-[0.3em] text-slate-900" style={{ fontFamily: "var(--font-mono), monospace" }}>
                  Status Pengisian
                </p>
                <p className="text-sm leading-7 text-slate-900">
                  {effectiveMode === "create"
                    ? "Mohon periksa kembali kelengkapan informasi dasar sebelum Anda menyimpan. Saat menekan tombol, properti ini akan masuk ke database."
                    : "Perubahan apa pun yang Anda simpan akan langsung terhubung dengan database utama PAL Property."}
                </p>
                {formError ? (
                  <p className="text-sm font-medium text-red-700" data-testid="listing-form-error" role="alert">
                    {formError}
                  </p>
                ) : null}
                {formSuccess ? <p className="text-sm font-medium text-emerald-700">{formSuccess}</p> : null}
              </div>

              <Button data-testid="listing-submit-button" disabled={submitMutation.isPending} type="submit">
                {submitMutation.isPending ? (effectiveMode === "create" ? "Menyimpan Listing..." : "Menyimpan Perubahan...") : effectiveMode === "create" ? "Buat & Simpan Listing" : "Simpan Perubahan"}
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
      transaction_type: "sale",
      price: "",
      currency: "IDR",
      is_negotiable: false,
      special_offers: "",
      location_province: "",
      location_province_code: "",
      location_city: "",
      location_city_code: "",
      location_district: "",
      location_district_code: "",
      location_village: "",
      location_village_code: "",
      address_detail: "",
      latitude: "",
      longitude: "",
      status: "active",
      bedrooms: "0",
      bathrooms: "0",
      floor_count: "",
      carport_capacity: "",
      land_area_sqm: "0",
      building_area_sqm: "0",
      certificate_type: "",
      condition: "",
      furnishing: "",
      electrical_power_va: "",
      facing_direction: "",
      year_built: "",
      facilities: "",
    };
  }

  const specifications = parseListingSpecifications(listing.specifications);

  return {
    category_id: listing.category_id ?? "",
    title: listing.title,
    description: listing.description ?? "",
    transaction_type: listing.transaction_type ?? "sale",
    price: String(listing.price),
    currency: listing.currency ?? "IDR",
    is_negotiable: listing.is_negotiable ?? false,
    special_offers: parseStringList(listing.special_offers).join(", "),
    location_province: listing.location_province ?? "",
    location_province_code: listing.location_province_code ?? "",
    location_city: listing.location_city ?? "",
    location_city_code: listing.location_city_code ?? "",
    location_district: listing.location_district ?? "",
    location_district_code: listing.location_district_code ?? "",
    location_village: listing.location_village ?? "",
    location_village_code: listing.location_village_code ?? "",
    address_detail: listing.address_detail ?? "",
    latitude: listing.latitude != null ? String(listing.latitude) : "",
    longitude: listing.longitude != null ? String(listing.longitude) : "",
    status: normalizeStatus(listing.status),
    bedrooms: String(listing.bedroom_count ?? specifications.bedrooms),
    bathrooms: String(listing.bathroom_count ?? specifications.bathrooms),
    floor_count: listing.floor_count != null ? String(listing.floor_count) : "",
    carport_capacity: listing.carport_capacity != null ? String(listing.carport_capacity) : "",
    land_area_sqm: String(listing.land_area_sqm ?? specifications.land_area_sqm),
    building_area_sqm: String(listing.building_area_sqm ?? specifications.building_area_sqm),
    certificate_type: listing.certificate_type ?? "",
    condition: listing.condition ?? "",
    furnishing: listing.furnishing ?? "",
    electrical_power_va: listing.electrical_power_va != null ? String(listing.electrical_power_va) : "",
    facing_direction: listing.facing_direction ?? "",
    year_built: listing.year_built != null ? String(listing.year_built) : "",
    facilities: parseStringList(listing.facilities).join(", "),
  };
}

function toRequestPayload(values: ListingFormSchema): ListingFormRequest {
  const defaults = getDefaultSpecifications();

  const bedroomCount = normalizeOptionalIntegerOrNull(values.bedrooms);
  const bathroomCount = normalizeOptionalIntegerOrNull(values.bathrooms);
  const landAreaSqm = normalizeOptionalIntegerOrNull(values.land_area_sqm);
  const buildingAreaSqm = normalizeOptionalIntegerOrNull(values.building_area_sqm);

  return {
    category_id: normalizeNullableString(values.category_id),
    title: values.title.trim(),
    description: normalizeNullableString(values.description),
    transaction_type: values.transaction_type,
    price: normalizeRequiredInteger(values.price, 1),
    currency: values.currency?.trim() || "IDR",
    is_negotiable: values.is_negotiable ?? false,
    special_offers: normalizeStringList(values.special_offers),
    location_province: normalizeNullableString(values.location_province),
    location_province_code: normalizeNullableString(values.location_province_code),
    location_city: normalizeNullableString(values.location_city),
    location_city_code: normalizeNullableString(values.location_city_code),
    location_district: normalizeNullableString(values.location_district),
    location_district_code: normalizeNullableString(values.location_district_code),
    location_village: normalizeNullableString(values.location_village),
    location_village_code: normalizeNullableString(values.location_village_code),
    address_detail: normalizeNullableString(values.address_detail),
    latitude: normalizeNullableNumber(values.latitude),
    longitude: normalizeNullableNumber(values.longitude),
    bedroom_count: bedroomCount,
    bathroom_count: bathroomCount,
    floor_count: normalizeOptionalIntegerOrNull(values.floor_count),
    carport_capacity: normalizeOptionalIntegerOrNull(values.carport_capacity),
    land_area_sqm: landAreaSqm,
    building_area_sqm: buildingAreaSqm,
    certificate_type: normalizeNullableString(values.certificate_type),
    condition: normalizeNullableString(values.condition),
    furnishing: normalizeNullableString(values.furnishing),
    electrical_power_va: normalizeOptionalIntegerOrNull(values.electrical_power_va),
    facing_direction: normalizeNullableString(values.facing_direction),
    year_built: normalizeOptionalIntegerOrNull(values.year_built),
    facilities: normalizeStringList(values.facilities),
    status: values.status,
    specifications: {
      bedrooms: bedroomCount ?? defaults.bedrooms,
      bathrooms: bathroomCount ?? defaults.bathrooms,
      land_area_sqm: landAreaSqm ?? defaults.land_area_sqm,
      building_area_sqm: buildingAreaSqm ?? defaults.building_area_sqm,
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

function normalizeStatus(status: string): ListingFormRequest["status"] {
  if (status === "inactive" || status === "sold" || status === "draft" || status === "archived") {
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

function getImageSuccessMessage(type: "upload" | "set-primary" | "delete" | "reorder", uploadCount = 0) {
  switch (type) {
    case "upload":
      return uploadCount > 1
        ? `Uploaded ${uploadCount} images. The gallery now reflects the backend response.`
        : "Image uploaded. The gallery now reflects the backend response.";
    case "set-primary":
      return "Primary image updated from the backend response.";
    case "delete":
      return "Image removed. Remaining images were refreshed from the backend.";
    case "reorder":
      return "Image order refreshed from the backend response.";
  }
}

function getFeedbackMessageClassName(tone: FeedbackMessage["tone"]) {
  switch (tone) {
    case "error":
      return "text-sm font-medium text-red-700";
    case "success":
      return "text-sm font-medium text-emerald-700";
    case "info":
      return "text-sm font-medium text-slate-900";
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
    transaction_type: "sale",
    price: 0,
    currency: "IDR",
    is_negotiable: false,
    special_offers: [],
    location_province: null,
    location_province_code: null,
    location_city: null,
    location_city_code: null,
    location_district: null,
    location_district_code: null,
    location_village: null,
    location_village_code: null,
    address_detail: null,
    latitude: null,
    longitude: null,
    bedroom_count: null,
    bathroom_count: null,
    floor_count: null,
    carport_capacity: null,
    land_area_sqm: null,
    building_area_sqm: null,
    certificate_type: null,
    condition: null,
    furnishing: null,
    electrical_power_va: null,
    facing_direction: null,
    year_built: null,
    facilities: [],
    status: "active",
    is_featured: false,
    specifications: {},
    view_count: 0,
    images: [],
    video: null,
    created_at: new Date(0).toISOString(),
    updated_at: new Date(0).toISOString(),
  };
}

function findRegionOptionByName(options: RegionOption[] | undefined, value: string | null | undefined) {
  const normalizedValue = value?.trim().toLowerCase();
  if (!normalizedValue) {
    return null;
  }

  return (options ?? []).find((option) => option.name.trim().toLowerCase() === normalizedValue) ?? null;
}

function getChoiceButtonClassName(selected: boolean) {
  return `rounded-full border px-4 py-2 text-sm font-medium transition ${selected ? "border-slate-900 bg-slate-900 text-white" : "border-slate-200 bg-white text-slate-900 hover:border-slate-900/50"}`;
}

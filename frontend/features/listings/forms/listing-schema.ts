import { z } from "zod";

export const listingFormSchema = z.object({
  category_id: z.string().optional(),
  title: z.string().trim().min(5, "Title must be at least 5 characters."),
  description: z.string().optional(),
  price: z.string().trim().min(1, "Price is required."),
  location_city: z.string().optional(),
  location_district: z.string().optional(),
  address_detail: z.string().optional(),
  status: z.enum(["active", "inactive", "sold"]),
  bedrooms: z.string().optional(),
  bathrooms: z.string().optional(),
  land_area_sqm: z.string().optional(),
  building_area_sqm: z.string().optional(),
});

export type ListingFormSchema = z.infer<typeof listingFormSchema>;

import { z } from "zod";

export const listingFormSchema = z.object({
  category_id: z.string().optional(),
  title: z.string().trim().min(5, "Title must be at least 5 characters."),
  description: z.string().optional(),
  transaction_type: z.enum(["sale", "rent"]),
  price: z.string().trim().min(1, "Price is required."),
  currency: z.string().optional(),
  is_negotiable: z.boolean().optional(),
  special_offers: z.string().optional(),
  location_province: z.string().optional(),
  location_city: z.string().optional(),
  location_district: z.string().optional(),
  address_detail: z.string().optional(),
  latitude: z.string().optional(),
  longitude: z.string().optional(),
  status: z.enum(["active", "inactive", "sold", "draft", "archived"]),
  bedrooms: z.string().optional(),
  bathrooms: z.string().optional(),
  floor_count: z.string().optional(),
  carport_capacity: z.string().optional(),
  land_area_sqm: z.string().optional(),
  building_area_sqm: z.string().optional(),
  certificate_type: z.string().optional(),
  condition: z.string().optional(),
  furnishing: z.string().optional(),
  electrical_power_va: z.string().optional(),
  facing_direction: z.string().optional(),
  year_built: z.string().optional(),
  facilities: z.string().optional(),
});

export type ListingFormSchema = z.infer<typeof listingFormSchema>;

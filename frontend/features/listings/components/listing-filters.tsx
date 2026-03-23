"use client";

import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";

import { browserFetch } from "@/lib/api/browser-fetch";
import type { ListingCategory } from "@/lib/api/listing-form";
import { queryKeys } from "@/lib/query/keys";

type ListingFilterValues = {
  city?: string;
  category_id?: string;
  price_min?: string;
  price_max?: string;
  status?: string;
  limit?: string;
};

async function getCategoryOptions() {
  const response = await browserFetch<ListingCategory[]>("/api/categories", {
    method: "GET",
    cache: "no-store",
  });

  return response.data;
}

export function ListingFilters({ values, total, visibleCount }: { values: ListingFilterValues; total: number; visibleCount: number }) {
  const categoriesQuery = useQuery({
    queryKey: queryKeys.categories,
    queryFn: () => getCategoryOptions(),
  });

  const quickStats = useMemo(
    () => [
      { label: "Listings", value: total.toLocaleString("en-US") },
      { label: "Visible now", value: visibleCount.toLocaleString("en-US") },
      { label: "Categories", value: (categoriesQuery.data?.length ?? 0).toLocaleString("en-US") },
    ],
    [categoriesQuery.data?.length, total, visibleCount],
  );

  return (
    <section className="space-y-3" data-testid="listing-filters">
      <form action="/listings" className="grid gap-2 xl:grid-cols-[minmax(0,1.8fr)_repeat(5,minmax(0,1fr))]">
        <label className="flex min-h-12 flex-col justify-center gap-1 rounded-md border border-[#ddd] bg-white px-3 py-2">
          <span className="text-[10px] font-semibold uppercase tracking-[0.18em] text-[#8a8a86]">Enter property address</span>
          <input
            className="bg-transparent text-[13px] text-[var(--ink)] outline-none placeholder:text-[#9a978f]"
            defaultValue={values.city ?? ""}
            name="city"
            placeholder="Jakarta, Bandung, Surabaya"
            type="text"
          />
        </label>

        <label className="flex min-h-12 flex-col justify-center gap-1 rounded-md border border-[#ddd] bg-white px-3 py-2">
          <span className="text-[10px] font-semibold uppercase tracking-[0.18em] text-[#8a8a86]">Neighborhoods</span>
          <select className="bg-transparent text-[13px] text-[var(--ink)] outline-none" defaultValue={values.category_id ?? ""} name="category_id">
            <option value="">All categories</option>
            {(categoriesQuery.data ?? []).map((category) => (
              <option key={category.id} value={category.id}>
                {category.name}
              </option>
            ))}
          </select>
        </label>

        <label className="flex min-h-12 flex-col justify-center gap-1 rounded-md border border-[#ddd] bg-white px-3 py-2">
          <span className="text-[10px] font-semibold uppercase tracking-[0.18em] text-[#8a8a86]">For sale</span>
          <select className="bg-transparent text-[13px] text-[var(--ink)] outline-none" defaultValue={values.status ?? "active"} name="status">
            <option value="">Any status</option>
            <option value="active">Active</option>
            <option value="inactive">Inactive</option>
            <option value="sold">Sold</option>
          </select>
        </label>

        <label className="flex min-h-12 flex-col justify-center gap-1 rounded-md border border-[#ddd] bg-white px-3 py-2">
          <span className="text-[10px] font-semibold uppercase tracking-[0.18em] text-[#8a8a86]">Price</span>
          <input className="bg-transparent text-[13px] text-[var(--ink)] outline-none placeholder:text-[#9a978f]" defaultValue={values.price_min ?? ""} inputMode="numeric" name="price_min" placeholder="Min budget" type="text" />
        </label>

        <label className="flex min-h-12 flex-col justify-center gap-1 rounded-md border border-[#ddd] bg-white px-3 py-2">
          <span className="text-[10px] font-semibold uppercase tracking-[0.18em] text-[#8a8a86]">Residential</span>
          <input className="bg-transparent text-[13px] text-[var(--ink)] outline-none placeholder:text-[#9a978f]" defaultValue={values.price_max ?? ""} inputMode="numeric" name="price_max" placeholder="Max budget" type="text" />
        </label>

        <div className="flex min-h-12 items-stretch gap-2 xl:justify-end">
          <input name="limit" type="hidden" value={values.limit ?? "12"} />
          <button className="inline-flex flex-1 items-center justify-center rounded-md border border-[#ddd] bg-white px-3 py-2 text-[13px] font-medium text-[var(--ink)] transition hover:border-[var(--accent)] hover:text-[var(--accent)] xl:flex-none" type="reset">
            Clear
          </button>
          <button className="inline-flex flex-[1.1] items-center justify-center rounded-md bg-[var(--ink)] px-4 py-2 text-[13px] font-medium text-white transition hover:bg-[var(--accent)] xl:flex-none" type="submit">
            Search
          </button>
        </div>
      </form>

      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-[#ece9e2] pb-3 text-[11px] font-medium uppercase tracking-[0.16em] text-[#77746d]">
        <div className="flex flex-wrap gap-2">
          {quickStats.map((item) => (
            <span key={item.label} className="rounded-full bg-[#f2f1ed] px-2.5 py-1">
              {item.label}: {item.value}
            </span>
          ))}
        </div>
        <span>{categoriesQuery.isLoading ? "Loading filters" : `${categoriesQuery.data?.length ?? 0} category filters ready`}</span>
      </div>
    </section>
  );
}

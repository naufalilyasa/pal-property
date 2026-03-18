"use client";

import { useQuery } from "@tanstack/react-query";

import { browserFetch } from "@/lib/api/browser-fetch";
import type { ListingCategory } from "@/lib/api/listing-form";
import { queryKeys } from "@/lib/query/keys";

async function getCategoryOptions() {
  const response = await browserFetch<ListingCategory[]>("/api/categories", {
    method: "GET",
    cache: "no-store",
  });

  return response.data;
}

export function ListingFilters() {
  const categoriesQuery = useQuery({
    queryKey: queryKeys.categories,
    queryFn: () => getCategoryOptions(),
  });

  return (
    <section className="rounded-[1.75rem] border border-[var(--line)] bg-white/80 p-5" data-testid="listing-filters">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <p className="text-xs uppercase tracking-[0.24em] text-[var(--accent)]">Filters</p>
          <h2 className="mt-2 text-xl font-semibold text-[var(--ink)]">Server-first browse, client filter helpers</h2>
        </div>
        <p className="text-sm text-[var(--muted)]">
          {categoriesQuery.isLoading
            ? "Loading category options..."
            : `${categoriesQuery.data?.length ?? 0} category options ready for interactive filters.`}
        </p>
      </div>
    </section>
  );
}

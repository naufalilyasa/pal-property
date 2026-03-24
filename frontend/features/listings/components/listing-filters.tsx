"use client";

import { useQuery } from "@tanstack/react-query";
import { useRouter, useSearchParams } from "next/navigation";

import { browserFetch } from "@/lib/api/browser-fetch";
import type { ListingCategory } from "@/lib/api/listing-form";
import { queryKeys } from "@/lib/query/keys";

export function ListingFilters({ view }: { view: "map" | "list" }) {
  const router = useRouter();
  const searchParams = useSearchParams();

  const categoriesQuery = useQuery({
    queryKey: queryKeys.categories,
    queryFn: async () => {
      const response = await browserFetch<ListingCategory[]>("/api/categories", {
        method: "GET",
        cache: "no-store",
      });
      return response.data;
    },
  });

  const updateView = (newView: "map" | "list") => {
    const params = new URLSearchParams(searchParams.toString());
    params.set("view", newView);
    router.push(`/listings?${params.toString()}`);
  };

  return (
    <div
      className="flex w-full items-center justify-between overflow-x-auto border-b border-gray-200 bg-white px-4 py-3 sm:px-6"
      data-testid="listing-filters"
    >
      {/* Left side: filters */}
      <div className="flex items-center gap-3">
        {/* Search Input */}
        <div className="relative flex items-center">
          <svg
            className="absolute left-3 text-gray-400"
            xmlns="http://www.w3.org/2000/svg"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <circle cx="11" cy="11" r="8" />
            <path d="m21 21-4.3-4.3" />
          </svg>
          <input
            className="w-64 rounded-full border border-gray-300 bg-white py-1.5 pr-4 pl-9 text-sm text-[#111] outline-none transition focus:border-black placeholder:text-gray-500"
            placeholder="Address, neighborhood, city, ZIP"
            type="text"
            defaultValue={searchParams.get("city") ?? ""}
          />
        </div>

        {/* Dropdowns */}
        <select className="w-auto cursor-pointer appearance-none rounded-full border border-gray-300 bg-white px-4 py-1.5 pr-8 text-sm font-medium text-[#111] outline-none hover:border-black bg-[url('data:image/svg+xml;charset=US-ASCII,%3Csvg%20xmlns%3D%22http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg%22%20width%3D%2214%22%20height%3D%2214%22%20viewBox%3D%220%200%2024%2024%22%20fill%3D%22none%22%20stroke%3D%22currentColor%22%20stroke-width%3D%222%22%20stroke-linecap%3D%22round%22%20stroke-linejoin%3D%22round%22%3E%3Cpath%20d%3D%22m6%209%206%206%206-6%22%2F%3E%3C%2Fsvg%3E')] bg-position-[calc(100%-12px)_center] bg-no-repeat">
          <option value="">Neighborhood</option>
          {categoriesQuery.data?.map((c) => (
            <option key={c.id} value={c.id}>
              {c.name}
            </option>
          ))}
        </select>

        <select className="w-auto cursor-pointer appearance-none rounded-full border border-gray-300 bg-white px-4 py-1.5 pr-8 text-sm font-medium text-[#111] outline-none hover:border-black bg-[url('data:image/svg+xml;charset=US-ASCII,%3Csvg%20xmlns%3D%22http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg%22%20width%3D%2214%22%20height%3D%2214%22%20viewBox%3D%220%200%2024%2024%22%20fill%3D%22none%22%20stroke%3D%22currentColor%22%20stroke-width%3D%222%22%20stroke-linecap%3D%22round%22%20stroke-linejoin%3D%22round%22%3E%3Cpath%20d%3D%22m6%209%206%206%206-6%22%2F%3E%3C%2Fsvg%3E')] bg-position-[calc(100%-12px)_center] bg-no-repeat">
          <option value="">Property</option>
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
        </select>

        <select className="w-auto cursor-pointer appearance-none rounded-full border border-gray-300 bg-white px-4 py-1.5 pr-8 text-sm font-medium text-[#111] outline-none hover:border-black bg-[url('data:image/svg+xml;charset=US-ASCII,%3Csvg%20xmlns%3D%22http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg%22%20width%3D%2214%22%20height%3D%2214%22%20viewBox%3D%220%200%2024%2024%22%20fill%3D%22none%22%20stroke%3D%22currentColor%22%20stroke-width%3D%222%22%20stroke-linecap%3D%22round%22%20stroke-linejoin%3D%22round%22%3E%3Cpath%20d%3D%22m6%209%206%206%206-6%22%2F%3E%3C%2Fsvg%3E')] bg-position-[calc(100%-12px)_center] bg-no-repeat">
          <option value="">Price</option>
        </select>

        <select className="w-auto cursor-pointer appearance-none rounded-full border border-gray-300 bg-white px-4 py-1.5 pr-8 text-sm font-medium text-[#111] outline-none hover:border-black bg-[url('data:image/svg+xml;charset=US-ASCII,%3Csvg%20xmlns%3D%22http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg%22%20width%3D%2214%22%20height%3D%2214%22%20viewBox%3D%220%200%2024%2024%22%20fill%3D%22none%22%20stroke%3D%22currentColor%22%20stroke-width%3D%222%22%20stroke-linecap%3D%22round%22%20stroke-linejoin%3D%22round%22%3E%3Cpath%20d%3D%22m6%209%206%206%206-6%22%2F%3E%3C%2Fsvg%3E')] bg-position-[calc(100%-12px)_center] bg-no-repeat">
          <option value="">Bedrooms</option>
        </select>

        <button className="flex h-[34px] items-center gap-1 rounded-full border border-gray-300 bg-white px-4 text-sm font-medium text-[#111] outline-none hover:border-black">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M21 4H3" />
            <path d="M21 12H3" />
            <path d="M21 20H3" />
          </svg>
          <span className="mt-px leading-none">More details</span>
        </button>
      </div>

      {/* Right side: Map/List toggle */}
      <div className="flex shrink-0 items-center justify-center rounded-full border border-gray-300 bg-white p-[3px]">
        <button
          onClick={() => updateView("map")}
          className={`rounded-full px-4 py-1 text-sm font-semibold transition ${view === "map" ? "bg-black text-white" : "text-gray-600 hover:text-black"}`}
        >
          Map
        </button>
        <button
          onClick={() => updateView("list")}
          className={`rounded-full px-4 py-1 text-sm font-semibold transition ${view === "list" ? "bg-black text-white" : "text-gray-600 hover:text-black"}`}
        >
          List
        </button>
      </div>
    </div>
  );
}

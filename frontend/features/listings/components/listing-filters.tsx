"use client";

import { useQuery } from "@tanstack/react-query";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { useEffect, useState } from "react";
import CurrencyInput from "react-currency-input-field";

import { browserFetch } from "@/lib/api/browser-fetch";
import type { ListingCategory, RegionOption } from "@/lib/api/listing-form";
import { queryKeys } from "@/lib/query/keys";

export function ListingFilters({ view }: { view: "map" | "list" }) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [selectedProvince, setSelectedProvince] = useState(searchParams.get("location_province") ?? "");
  const [selectedProvinceCode, setSelectedProvinceCode] = useState("");
  const [selectedCity, setSelectedCity] = useState(searchParams.get("location_city") ?? "");
  const [priceMin, setPriceMin] = useState(searchParams.get("price_min") ?? "");
  const [priceMax, setPriceMax] = useState(searchParams.get("price_max") ?? "");

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

  const provincesQuery = useQuery({
    queryKey: ["regions", "provinces", "public-filters"],
    queryFn: async () => {
      const response = await browserFetch<RegionOption[]>("/api/regions/provinces", {
        method: "GET",
        cache: "no-store",
      });
      return response.data;
    },
  });

  const citiesQuery = useQuery({
    queryKey: ["regions", "cities", "public-filters", selectedProvinceCode],
    queryFn: async () => {
      const response = await browserFetch<RegionOption[]>(`/api/regions/cities?province_code=${selectedProvinceCode}`, {
        method: "GET",
        cache: "no-store",
      });
      return response.data;
    },
    enabled: Boolean(selectedProvinceCode),
  });

  const selectedProvinceOption = provincesQuery.data?.find((province) => province.name === selectedProvince) ?? null;

  useEffect(() => {
    if (selectedProvince && selectedProvinceCode === "" && selectedProvinceOption) {
      setSelectedProvinceCode(selectedProvinceOption.code);
    }
  }, [selectedProvince, selectedProvinceCode, selectedProvinceOption]);

  const updateView = (newView: "map" | "list") => {
    const params = new URLSearchParams(searchParams.toString());
    params.set("view", newView);
    router.push(`/listings?${params.toString()}`);
  };

  const handleSubmit = (formData: FormData) => {
    const params = new URLSearchParams();
    for (const [key, value] of formData.entries()) {
      const normalized = String(value).trim();
      if (normalized) {
        params.set(key, normalized);
      }
    }
    params.set("view", view);
    router.push(`/listings?${params.toString()}`);
  };

  return (
    <div
      className="flex w-full items-center justify-between overflow-x-auto border-b border-gray-200 bg-white px-4 py-3 sm:px-6"
      data-testid="listing-filters"
    >
      {/* Left side: filters */}
      <form
        className="flex items-center gap-3"
        onSubmit={(event) => {
          event.preventDefault();
          handleSubmit(new FormData(event.currentTarget));
        }}
      >
        {/* Search Input */}
        <div className="relative flex items-center">
          <svg
            aria-hidden="true"
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
            name="q"
            className="w-64 rounded-full border border-gray-300 bg-white py-1.5 pr-4 pl-9 text-sm text-[#111] outline-none transition focus:border-black placeholder:text-gray-500"
            placeholder="Search title, city, province"
            type="text"
            defaultValue={searchParams.get("q") ?? ""}
          />
        </div>

        {/* Dropdowns */}
        <select name="category_id" className="w-auto cursor-pointer appearance-none rounded-full border border-gray-300 bg-white px-4 py-1.5 pr-8 text-sm font-medium text-[#111] outline-none hover:border-black bg-[url('data:image/svg+xml;charset=US-ASCII,%3Csvg%20xmlns%3D%22http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg%22%20width%3D%2214%22%20height%3D%2214%22%20viewBox%3D%220%200%2024%2024%22%20fill%3D%22none%22%20stroke%3D%22currentColor%22%20stroke-width%3D%222%22%20stroke-linecap%3D%22round%22%20stroke-linejoin%3D%22round%22%3E%3Cpath%20d%3D%22m6%209%206%206%206-6%22%2F%3E%3C%2Fsvg%3E')] bg-position-[calc(100%-12px)_center] bg-no-repeat" defaultValue={searchParams.get("category_id") ?? ""}>
          <option value="">Category</option>
          {categoriesQuery.data?.map((c) => (
            <option key={c.id} value={c.id}>
              {c.name}
            </option>
          ))}
        </select>

        <select name="transaction_type" className="w-auto cursor-pointer appearance-none rounded-full border border-gray-300 bg-white px-4 py-1.5 pr-8 text-sm font-medium text-[#111] outline-none hover:border-black bg-[url('data:image/svg+xml;charset=US-ASCII,%3Csvg%20xmlns%3D%22http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg%22%20width%3D%2214%22%20height%3D%2214%22%20viewBox%3D%220%200%2024%2024%22%20fill%3D%22none%22%20stroke%3D%22currentColor%22%20stroke-width%3D%222%22%20stroke-linecap%3D%22round%22%20stroke-linejoin%3D%22round%22%3E%3Cpath%20d%3D%22m6%209%206%206%206-6%22%2F%3E%3C%2Fsvg%3E')] bg-position-[calc(100%-12px)_center] bg-no-repeat" defaultValue={searchParams.get("transaction_type") ?? ""}>
          <option value="">Transaction</option>
          <option value="sale">Sale</option>
          <option value="rent">Rent</option>
        </select>

        <select
          className="w-44 cursor-pointer appearance-none rounded-full border border-gray-300 bg-white px-4 py-1.5 pr-8 text-sm text-[#111] outline-none hover:border-black bg-[url('data:image/svg+xml;charset=US-ASCII,%3Csvg%20xmlns%3D%22http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg%22%20width%3D%2214%22%20height%3D%2214%22%20viewBox%3D%220%200%2024%2024%22%20fill%3D%22none%22%20stroke%3D%22currentColor%22%20stroke-width%3D%222%22%20stroke-linecap%3D%22round%22%20stroke-linejoin%3D%22round%22%3E%3Cpath%20d%3D%22m6%209%206%206%206-6%22%2F%3E%3C%2Fsvg%3E')] bg-position-[calc(100%-12px)_center] bg-no-repeat"
          name="location_province"
          onChange={(event) => {
            const nextProvince = event.target.value;
            const option = provincesQuery.data?.find((province) => province.name === nextProvince) ?? null;
            setSelectedProvince(nextProvince);
            setSelectedProvinceCode(option?.code ?? "");
            setSelectedCity("");
          }}
          value={selectedProvince}
        >
          <option value="">Province</option>
          {(provincesQuery.data ?? []).map((province) => (
            <option key={province.code} value={province.name}>
              {province.name}
            </option>
          ))}
        </select>

        <select
          className="w-44 cursor-pointer appearance-none rounded-full border border-gray-300 bg-white px-4 py-1.5 pr-8 text-sm text-[#111] outline-none hover:border-black disabled:cursor-not-allowed disabled:bg-gray-50 disabled:text-gray-400 bg-[url('data:image/svg+xml;charset=US-ASCII,%3Csvg%20xmlns%3D%22http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg%22%20width%3D%2214%22%20height%3D%2214%22%20viewBox%3D%220%200%2024%2024%22%20fill%3D%22none%22%20stroke%3D%22currentColor%22%20stroke-width%3D%222%22%20stroke-linecap%3D%22round%22%20stroke-linejoin%3D%22round%22%3E%3Cpath%20d%3D%22m6%209%206%206%206-6%22%2F%3E%3C%2Fsvg%3E')] bg-position-[calc(100%-12px)_center] bg-no-repeat"
          disabled={!selectedProvinceCode}
          name="location_city"
          onChange={(event) => setSelectedCity(event.target.value)}
          value={selectedCity}
        >
          <option value="">City</option>
          {(citiesQuery.data ?? []).map((city) => (
            <option key={city.code} value={city.name}>
              {city.name}
            </option>
          ))}
        </select>

        <div className="w-40 rounded-full border border-gray-300 bg-white px-4 py-1.5 text-sm text-[#111] transition focus-within:border-black">
          <CurrencyInput
            allowNegativeValue={false}
            className="w-full bg-transparent outline-none placeholder:text-gray-500"
            decimalsLimit={0}
            defaultValue={undefined}
            groupSeparator="."
            inputMode="numeric"
            intlConfig={{ locale: "id-ID", currency: "IDR" }}
            name="price_min"
            onValueChange={(value) => setPriceMin(value ?? "")}
            placeholder="Min price"
            prefix="Rp "
            value={priceMin}
          />
        </div>

        <div className="w-40 rounded-full border border-gray-300 bg-white px-4 py-1.5 text-sm text-[#111] transition focus-within:border-black">
          <CurrencyInput
            allowNegativeValue={false}
            className="w-full bg-transparent outline-none placeholder:text-gray-500"
            decimalsLimit={0}
            defaultValue={undefined}
            groupSeparator="."
            inputMode="numeric"
            intlConfig={{ locale: "id-ID", currency: "IDR" }}
            name="price_max"
            onValueChange={(value) => setPriceMax(value ?? "")}
            placeholder="Max price"
            prefix="Rp "
            value={priceMax}
          />
        </div>

        <select name="sort" className="w-auto cursor-pointer appearance-none rounded-full border border-gray-300 bg-white px-4 py-1.5 pr-8 text-sm font-medium text-[#111] outline-none hover:border-black bg-[url('data:image/svg+xml;charset=US-ASCII,%3Csvg%20xmlns%3D%22http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg%22%20width%3D%2214%22%20height%3D%2214%22%20viewBox%3D%220%200%2024%2024%22%20fill%3D%22none%22%20stroke%3D%22currentColor%22%20stroke-width%3D%222%22%20stroke-linecap%3D%22round%22%20stroke-linejoin%3D%22round%22%3E%3Cpath%20d%3D%22m6%209%206%206%206-6%22%2F%3E%3C%2Fsvg%3E')] bg-position-[calc(100%-12px)_center] bg-no-repeat" defaultValue={searchParams.get("sort") ?? ""}>
          <option value="">Sort</option>
          <option value="newest">Newest</option>
          <option value="price_asc">Price: Low to High</option>
          <option value="price_desc">Price: High to Low</option>
          <option value="relevance">Relevance</option>
        </select>

        <button className="flex h-[34px] items-center gap-1 rounded-full border border-gray-900 bg-[#111] px-4 text-sm font-medium text-white outline-none hover:bg-black" type="submit">
          Apply
        </button>

        <Link href={`/listings?view=${view}`} className="text-sm font-medium text-[#111] underline-offset-4 hover:underline">
          Clear
        </Link>
      </form>

      {/* Right side: Map/List toggle */}
      <div className="flex shrink-0 items-center justify-center rounded-full border border-gray-300 bg-white p-[3px]">
        <button
          onClick={() => updateView("map")}
          className={`rounded-full px-4 py-1 text-sm font-semibold transition ${view === "map" ? "bg-black text-white" : "text-gray-600 hover:text-black"}`}
          type="button"
        >
          Map
        </button>
        <button
          onClick={() => updateView("list")}
          className={`rounded-full px-4 py-1 text-sm font-semibold transition ${view === "list" ? "bg-black text-white" : "text-gray-600 hover:text-black"}`}
          type="button"
        >
          List
        </button>
      </div>
    </div>
  );
}

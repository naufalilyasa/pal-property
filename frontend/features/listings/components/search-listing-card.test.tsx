import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { SearchListingCardItem } from "./search-listing-card";

vi.mock("next/image", () => ({
  default: (props: React.ImgHTMLAttributes<HTMLImageElement>) => {
    const { alt = "", ...rest } = props;
    return <img alt={alt} {...rest} />;
  },
}));

vi.mock("@/features/saved-listings/components/save-listing-button", () => ({
  SaveListingButton: () => <button data-testid="save-listing-button" type="button">save</button>,
}));

function buildListing(overrides: Partial<import("@/features/listings/server/get-search-listings").SearchListingCard> = {}) {
  return {
    id: "listing-1",
    title: "Jakarta River House",
    slug: "jakarta-river-house",
    transaction_type: "sale",
    price: 3250000000,
    currency: "IDR",
    location_province: "DKI Jakarta",
    location_city: "Jakarta Timur",
    location_district: "Cibubur",
    location_village: "Ciracas",
    bedroom_count: 4,
    bathroom_count: 3,
    land_area_sqm: 152,
    building_area_sqm: 166,
    status: "active",
    is_featured: false,
    created_at: "2026-03-17T00:00:00Z",
    updated_at: "2026-03-17T00:00:00Z",
    primary_image_url: "https://images.example/primary.jpg",
    image_urls: [
      "https://images.example/primary.jpg",
      "https://images.example/2.jpg",
      "https://images.example/3.jpg",
      "https://images.example/4.jpg",
      "https://images.example/5.jpg",
      "https://images.example/6.jpg",
    ],
    ...overrides,
  };
}

describe("SearchListingCardItem", () => {
  it("renders the primary image first and uses contain styling", () => {
    render(<SearchListingCardItem href="/listings/jakarta-river-house" listing={buildListing()} />);

    const image = screen.getByAltText(/jakarta river house/i);
    expect(image).toHaveAttribute("src", "https://images.example/primary.jpg");
    expect(image).toHaveClass("object-contain");
  });

  it("slides through up to five images and shows more-photos overlay on the fifth frame", () => {
    render(<SearchListingCardItem href="/listings/jakarta-river-house" listing={buildListing()} />);

    const nextButton = screen.getByRole("button", { name: /show next image/i });
    const image = screen.getByAltText(/jakarta river house/i);

    fireEvent.click(nextButton);
    fireEvent.click(nextButton);
    fireEvent.click(nextButton);
    fireEvent.click(nextButton);

    expect(image).toHaveAttribute("src", "https://images.example/5.jpg");
    expect(screen.getByText("+1")).toBeInTheDocument();
    expect(screen.getByText(/more photos/i)).toBeInTheDocument();
    expect(nextButton).toBeDisabled();
  });

  it("renders denser location and property metrics without changing the image shell", () => {
    render(<SearchListingCardItem href="/listings/jakarta-river-house" listing={buildListing()} />);

    expect(screen.getByText(/ciracas, cibubur, jakarta timur/i)).toBeInTheDocument();
    expect(screen.getByText(/dki jakarta/i)).toBeInTheDocument();
    expect(screen.getByText("KT 4")).toBeInTheDocument();
    expect(screen.getByText("KM 3")).toBeInTheDocument();
    expect(screen.getByText("LT 152 m²")).toBeInTheDocument();
    expect(screen.getByText("LB 166 m²")).toBeInTheDocument();
  });
});

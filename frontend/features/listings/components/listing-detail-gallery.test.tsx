import type { ImgHTMLAttributes } from "react";

import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ListingDetailGallery } from "./listing-detail-gallery";

vi.mock("next/image", () => ({
  default: ({ fill: _fill, priority: _priority, unoptimized: _unoptimized, ...props }: ImgHTMLAttributes<HTMLImageElement> & { fill?: boolean; priority?: boolean; unoptimized?: boolean }) => (
    <img {...props} alt={props.alt ?? ""} />
  ),
}));

const images = [
  "https://images.example/one.jpg",
  "https://images.example/two.jpg",
  "https://images.example/three.jpg",
];

describe("ListingDetailGallery", () => {
  it("moves between photos with the inline arrow controls", () => {
    render(<ListingDetailGallery address="Jakarta River House" images={images} />);

    expect(screen.getByAltText("Jakarta River House")).toHaveAttribute("src", images[0]);

    fireEvent.click(screen.getByTestId("listing-detail-gallery-next"));
    expect(screen.getByAltText("Jakarta River House")).toHaveAttribute("src", images[1]);

    fireEvent.click(screen.getByTestId("listing-detail-gallery-prev"));
    expect(screen.getByAltText("Jakarta River House")).toHaveAttribute("src", images[0]);
  });

  it("opens a lightbox when the main photo is clicked", () => {
    render(<ListingDetailGallery address="Jakarta River House" images={images} />);

    fireEvent.click(screen.getByTestId("listing-detail-gallery-open"));

    expect(screen.getByTestId("listing-detail-gallery-lightbox")).toBeInTheDocument();
    expect(screen.getByAltText("Jakarta River House detail 1")).toHaveAttribute("src", images[0]);

    fireEvent.click(screen.getByRole("button", { name: /foto berikutnya di galeri/i }));
    expect(screen.getByAltText("Jakarta River House detail 2")).toHaveAttribute("src", images[1]);
  });

  it("supports zoom controls inside the lightbox", () => {
    render(<ListingDetailGallery address="Jakarta River House" images={images} />);

    fireEvent.click(screen.getByTestId("listing-detail-gallery-open"));

    expect(screen.getByRole("button", { name: /reset zoom foto/i })).toHaveTextContent("100%");

    fireEvent.click(screen.getByRole("button", { name: /zoom in foto/i }));
    expect(screen.getByRole("button", { name: /reset zoom foto/i })).toHaveTextContent("125%");

    fireEvent.click(screen.getByRole("button", { name: /zoom out foto/i }));
    expect(screen.getByRole("button", { name: /reset zoom foto/i })).toHaveTextContent("100%");
  });
});

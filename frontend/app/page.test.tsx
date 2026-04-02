import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import Home from "./page";

describe("Home", () => {
  it("renders the Pal Property public landing shell", () => {
    render(<Home />);

    expect(
      screen.getByRole("heading", {
        level: 1,
        name: /jual beli properti mewah & eksklusif di indonesia\./i,
      }),
    ).toBeInTheDocument();
    expect(screen.getByText(/agen properti tepercaya/i)).toBeInTheDocument();
    expect(screen.getByText(/jual beli mudah/i)).toBeInTheDocument();
    expect(screen.getByText(/properti pilihan/i)).toBeInTheDocument();
    expect(screen.getByTestId("home-shell")).toBeInTheDocument();
    expect(screen.getAllByRole("link", { name: /cari properti/i })[0]).toHaveAttribute("href", "/listings");
    expect(screen.getByRole("link", { name: /^login$/i })).toHaveAttribute("href", "/login");
  });
});

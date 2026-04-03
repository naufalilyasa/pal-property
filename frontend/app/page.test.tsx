import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

vi.mock("@/features/listings/components/top-nav", () => ({
  TopNav: () => <div data-testid="top-nav" />,
}));

vi.mock("@/features/listings/components/footer", () => ({
  Footer: () => <div data-testid="footer" />,
}));

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
    expect(screen.getByTestId("top-nav")).toBeInTheDocument();
    expect(screen.getByTestId("home-shell")).toBeInTheDocument();
    expect(screen.getAllByRole("link", { name: /cari properti/i })[0]).toHaveAttribute("href", "/listings");
    expect(screen.getByTestId("footer")).toBeInTheDocument();
  });
});

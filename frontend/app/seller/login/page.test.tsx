import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { parseAuthIntentState } from "@/features/auth/auth-intent";

vi.mock("@/features/listings/components/top-nav", () => ({
  TopNav: () => <div data-testid="top-nav" />,
}));

vi.mock("@/features/listings/components/footer", () => ({
  Footer: () => <div data-testid="footer" />,
}));

import SellerLoginPage from "./page";

describe("SellerLoginPage", () => {
  it("renders the seller login experience with the seller auth intent", async () => {
    render(await SellerLoginPage({}));

    expect(
      screen.getByRole("heading", {
        level: 1,
        name: /manajemen listing anda/i,
      }),
    ).toBeInTheDocument();
    expect(screen.getByText(/portal agen \/ seller/i)).toBeInTheDocument();
    expect(screen.getByText(/menambah, mengedit, dan mengelola daftar properti/i)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /bukan agen\? masuk publik/i })).toHaveAttribute("href", "/login");

    const googleButton = screen.getByTestId("login-google-button");
    const href = googleButton.getAttribute("href");

    expect(href).toBeTruthy();

    const url = new URL(href!);
    const state = url.searchParams.get("state");

    expect(url.pathname).toBe("/auth/oauth/google");
    expect(state).toBeTruthy();
    expect(parseAuthIntentState(state!)).toEqual(
      expect.objectContaining({
        intent: "seller",
        returnTo: "/dashboard",
      }),
    );
  });

  it("passes returnTo via the seller auth intent state when requested", async () => {
    const returnTo = "/dashboard/listings/new";
    render(await SellerLoginPage({ searchParams: Promise.resolve({ returnTo }) }));

    const googleButton = screen.getByTestId("login-google-button");
    const href = googleButton.getAttribute("href");

    expect(href).toBeTruthy();

    const url = new URL(href!);
    const state = url.searchParams.get("state");

    expect(state).toBeTruthy();
    expect(parseAuthIntentState(state!).returnTo).toBe(returnTo);
  });

  it("shows the seller session-expired banner when requested", async () => {
    render(await SellerLoginPage({ searchParams: Promise.resolve({ reason: "session-expired" }) }));

    expect(screen.getByTestId("auth-status-banner")).toHaveTextContent(
      /sesi anda telah berakhir\. silakan masuk kembali\./i,
    );
  });
});

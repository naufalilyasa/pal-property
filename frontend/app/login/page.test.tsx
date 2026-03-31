import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { parseAuthIntentState } from "@/features/auth/auth-intent";

import LoginPage from "./page";

describe("LoginPage", () => {
  it("renders the public login experience with the shared auth intent", async () => {
    render(await LoginPage({}));

    expect(
      screen.getByRole("heading", {
        level: 1,
        name: /selamat datang di pal property/i,
      }),
    ).toBeInTheDocument();
    expect(screen.getByText(/akses publik/i)).toBeInTheDocument();
    expect(screen.getByText(/dapatkan properti impian anda dengan lebih mudah dan cepat/i)).toBeInTheDocument();

    const googleButton = screen.getByTestId("login-google-button");
    const href = googleButton.getAttribute("href");

    expect(href).toBeTruthy();

    const url = new URL(href!);
    const state = url.searchParams.get("state");

    expect(url.pathname).toBe("/auth/oauth/google");
    expect(state).toBeTruthy();
    expect(parseAuthIntentState(state!)).toEqual(
      expect.objectContaining({
        intent: "public",
        returnTo: "/",
      }),
    );
  });

  it("passes returnTo through the auth intent state when provided", async () => {
    const returnTo = "/saved-listings";
    render(await LoginPage({ searchParams: Promise.resolve({ returnTo }) }));

    const googleButton = screen.getByTestId("login-google-button");
    const href = googleButton.getAttribute("href");

    expect(href).toBeTruthy();

    const url = new URL(href!);
    const state = url.searchParams.get("state");

    expect(state).toBeTruthy();
    expect(parseAuthIntentState(state!).returnTo).toBe(returnTo);
  });

  it("shows the general session-expired banner when requested", async () => {
    render(await LoginPage({ searchParams: Promise.resolve({ reason: "session-expired" }) }));

    expect(screen.getByTestId("auth-status-banner")).toHaveTextContent(/sesi anda telah berakhir\. silakan masuk kembali\./i);
  });
});

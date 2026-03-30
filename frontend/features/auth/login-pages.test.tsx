import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import LoginPage from "@/app/login/page";
import SellerLoginPage from "@/app/seller/login/page";
import { parseAuthIntentState } from "@/features/auth/auth-intent";

describe("login entry experiences", () => {
  it("shows the public login shell with its copy and public intent", async () => {
    const element = await LoginPage({
      searchParams: Promise.resolve({ reason: "session-expired", returnTo: "/listings" }),
    });

    render(element);

    expect(screen.getByText(/general access/i)).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: /sign in across find/i })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /back to home/i })).toHaveAttribute("href", "/");
    expect(screen.getByTestId("auth-status-banner")).toHaveTextContent(/Your session expired/i);

    const googleButton = screen.getByTestId("login-google-button");
    const href = googleButton.getAttribute("href");
    expect(href).toBeTruthy();

    const parsedState = parseAuthIntentState(new URL(href!).searchParams.get("state")!);
    expect(parsedState.intent).toBe("public");
    expect(parsedState.returnTo).toBe("/listings");
  });

  it("shows the seller login shell with seller copy and seller intent", async () => {
    const element = await SellerLoginPage({
      searchParams: Promise.resolve({ reason: "session-expired", returnTo: "/dashboard/listings" }),
    });

    render(element);

    expect(screen.getByText(/seller workspace/i)).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: /access your listing desk/i })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /need the general login page\?/i })).toHaveAttribute(
      "href",
      "/login",
    );
    expect(screen.getByTestId("auth-status-banner")).toHaveTextContent(/Your seller session expired/i);

    const googleButton = screen.getByTestId("login-google-button");
    const href = googleButton.getAttribute("href");
    expect(href).toBeTruthy();

    const parsedState = parseAuthIntentState(new URL(href!).searchParams.get("state")!);
    expect(parsedState.intent).toBe("seller");
    expect(parsedState.returnTo).toBe("/dashboard/listings");
  });
});

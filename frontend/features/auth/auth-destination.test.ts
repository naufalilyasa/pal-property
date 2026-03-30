import { describe, expect, it } from "vitest";

import { AuthIntent } from "@/features/auth/auth-intent";
import {
  getLoginPathForIntent,
  PUBLIC_HOME_PATH,
  resolveAuthIntentDestination,
  resolveSellerDestination,
  SELLER_DASHBOARD_PATH,
  SELLER_ONBOARDING_PATH,
} from "./auth-destination";

describe("auth destination helpers", () => {
  it("defaults seller intent to the dashboard path", () => {
    const destination = resolveSellerDestination();

    expect(destination).toBe(SELLER_DASHBOARD_PATH);
  });

  it("routes sellers requiring onboarding to the onboarding path", () => {
    const destination = resolveSellerDestination({ requiresOnboarding: true });

    expect(destination).toBe(SELLER_ONBOARDING_PATH);
  });

  it("routes sellers without dashboard access to onboarding", () => {
    const destination = resolveSellerDestination({ canAccessDashboard: false });

    expect(destination).toBe(SELLER_ONBOARDING_PATH);
  });

  it("respects seller intent when resolving auth destinations", () => {
    const destination = resolveAuthIntentDestination("seller", { requiresOnboarding: true });

    expect(destination).toBe(SELLER_ONBOARDING_PATH);
  });

  it("resolves public intent to the public home path", () => {
    const destination = resolveAuthIntentDestination("public");

    expect(destination).toBe(PUBLIC_HOME_PATH);
  });

  it("returns the correct login path for each intent", () => {
    expect(getLoginPathForIntent("seller" as AuthIntent)).toBe("/seller/login");
    expect(getLoginPathForIntent("public" as AuthIntent)).toBe("/login");
  });
});

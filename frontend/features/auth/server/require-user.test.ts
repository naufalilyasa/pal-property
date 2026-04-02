import { beforeEach, describe, expect, it, vi } from "vitest";

import type { CurrentUser } from "./current-user";

const { redirectMock, getOptionalUserMock } = vi.hoisted(() => ({
  redirectMock: vi.fn<(target: string) => never>(),
  getOptionalUserMock: vi.fn<Promise<CurrentUser | null>, []>(),
}));

vi.mock("next/navigation", () => ({
  redirect: (target: string) => {
    redirectMock(target);
    throw new Error("redirect called");
  },
}));

vi.mock("./current-user", () => ({
  getOptionalUser: getOptionalUserMock,
}));

import { requireUser } from "./require-user";

describe("requireUser", () => {
  beforeEach(() => {
    redirectMock.mockReset();
    getOptionalUserMock.mockReset();
  });

  it("redirects unauthenticated sellers to the seller login path with returnTo", async () => {
    getOptionalUserMock.mockResolvedValue(null);
    const returnTo = "/dashboard/listings";

    await expect(requireUser({ intent: "seller", returnTo })).rejects.toThrow("redirect called");

    expect(redirectMock).toHaveBeenCalledWith(
      `/seller/login?returnTo=${encodeURIComponent(returnTo)}`,
    );
  });

  it("returns the current admin when seller dashboard access is allowed", async () => {
    const user: CurrentUser = {
      id: "admin-1",
      name: "Admin One",
      email: "admin@example.com",
      avatar_url: null,
      role: "admin",
      seller_capabilities: {
        canAccessDashboard: true,
        requiresOnboarding: false,
      },
      created_at: "2026-03-18T00:00:00Z",
    };
    getOptionalUserMock.mockResolvedValue(user);

    const result = await requireUser({ intent: "seller" });

    expect(result).toBe(user);
    expect(redirectMock).not.toHaveBeenCalled();
  });

  it("redirects authenticated non-admin users away from seller routes", async () => {
    getOptionalUserMock.mockResolvedValue({
      id: "user-1",
      name: "Basic User",
      email: "user@example.com",
      avatar_url: null,
      role: "user",
      seller_capabilities: {
        canAccessDashboard: false,
        requiresOnboarding: false,
      },
      created_at: "2026-03-18T00:00:00Z",
    });

    await expect(requireUser({ intent: "seller", returnTo: "/dashboard" })).rejects.toThrow("redirect called");
    expect(redirectMock).toHaveBeenCalledWith("/seller/onboarding");
  });
});

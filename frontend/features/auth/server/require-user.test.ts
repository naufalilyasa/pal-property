import { beforeEach, describe, expect, it, vi } from "vitest";

import type { CurrentUser } from "./current-user";

const redirectMock = vi.fn<(target: string) => never>();
const getOptionalUserMock = vi.fn<Promise<CurrentUser | null>, []>();

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

  it("returns the current user when already authenticated", async () => {
    const user: CurrentUser = {
      id: "seller-1",
      name: "Seller One",
      email: "seller@example.com",
      avatar_url: null,
      role: "seller",
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
});

import { expect, test } from "@playwright/test";

function decodeStateParam(href: string) {
  const url = new URL(href, "http://127.0.0.1:3100");
  const state = url.searchParams.get("state");

  if (!state) {
    throw new Error("Missing OAuth state query parameter");
  }

  const normalized = state.replace(/-/g, "+").replace(/_/g, "/");
  const padded = normalized + "=".repeat((4 - (normalized.length % 4)) % 4);
  return JSON.parse(Buffer.from(padded, "base64").toString("utf-8"));
}

test.describe("login ux split", () => {
  test("public login keeps public intent and general copy", async ({ page }) => {
    await page.goto("/login");

    await expect(page.getByRole("heading", { name: /sign in across find/i })).toBeVisible();
    await expect(page.getByText(/search, save, and return whenever you are ready/i)).toBeVisible();

    const googleButton = page.getByTestId("login-google-button");
    await expect(googleButton).toBeVisible();
    const href = await googleButton.getAttribute("href");
    expect(href).toBeTruthy();

    const state = decodeStateParam(href!);
    expect(state.intent).toBe("public");
    expect(state.returnTo).toBe("/");
  });

  test("seller login keeps seller intent and seller-focused copy", async ({ page }) => {
    await page.goto("/seller/login");

    await expect(page.getByRole("heading", { name: /access your listing desk/i })).toBeVisible();
    await expect(page.getByText(/manage drafts, gallery updates, and publishing tasks/i)).toBeVisible();

    const googleButton = page.getByTestId("login-google-button");
    await expect(googleButton).toBeVisible();
    const href = await googleButton.getAttribute("href");
    expect(href).toBeTruthy();

    const state = decodeStateParam(href!);
    expect(state.intent).toBe("seller");
    expect(state.returnTo).toBe("/dashboard");
  });
});

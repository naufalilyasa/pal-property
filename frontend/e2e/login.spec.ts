import { expect, test } from "@playwright/test";

import { parseAuthIntentStateFromHref } from "./helpers/auth-state";

test.describe("login journeys", () => {
  test("public login experience surfaces the general copy and public intent", async ({ page }) => {
    await page.goto("/login");

    await expect(page.getByRole("heading", { level: 1, name: /sign in across find/i })).toBeVisible();
    await expect(page.getByText(/general access/i)).toBeVisible();
    await expect(page.getByText(/Keep your account session in the backend while this page simply starts the OAuth flow\./i)).toBeVisible();

    const googleButton = page.getByTestId("login-google-button");
    await expect(googleButton).toHaveAttribute("href", /\/auth\/oauth\/google\?/);

    const href = await googleButton.getAttribute("href");
    const state = parseAuthIntentStateFromHref(href!);

    expect(state.intent).toBe("public");
    expect(state.returnTo).toBe("/");
  });

  test("seller login experience surfaces the seller copy and intent metadata", async ({ page }) => {
    await page.goto("/seller/login");

    await expect(page.getByRole("heading", { level: 1, name: /access your listing desk/i })).toBeVisible();
    await expect(page.getByText(/seller workspace/i)).toBeVisible();
    await expect(page.getByText(/return to the seller workspace intent without changing the backend OAuth entrypoint\./i)).toBeVisible();

    const googleButton = page.getByTestId("login-google-button");
    await expect(googleButton).toHaveAttribute("href", /\/auth\/oauth\/google\?/);

    const href = await googleButton.getAttribute("href");
    const state = parseAuthIntentStateFromHref(href!);

    expect(state.intent).toBe("seller");
    expect(state.returnTo).toBe("/dashboard");

    await expect(page.getByRole("link", { name: /Need the general login page\?/i })).toBeVisible();
  });
});

import { createServer } from "node:http";

import { expect, test } from "@playwright/test";

const MOCK_API_ORIGIN = "http://127.0.0.1:45731";

const backendEnvelope = <T,>(data: T, message = "Success") => ({
  success: true,
  message,
  data,
  trace_id: "trace-e2e",
});

const buyerUser = {
  id: "buyer-1",
  name: "Buyer One",
  email: "buyer@example.com",
  avatar_url: null,
  role: "user",
  created_at: "2026-03-17T00:00:00Z",
};

const SSR_AUTH_COOKIE = "access_token=playwright-test-token";

type MockRequest = {
  method: string;
  path: string;
  searchParams: URLSearchParams;
  headers: Record<string, string | string[] | undefined>;
};

type MockResponse = {
  status: number;
  body: unknown;
};

let responder: (request: MockRequest) => MockResponse;
let savedListingPageRequests: Array<Record<string, string>> = [];

const server = createServer(async (request, response) => {
  response.setHeader("Access-Control-Allow-Origin", request.headers.origin ?? "http://127.0.0.1:3100");
  response.setHeader("Access-Control-Allow-Credentials", "true");
  response.setHeader("Access-Control-Allow-Headers", "Content-Type, Accept, Cookie");
  response.setHeader("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS");

  if ((request.method ?? "GET") === "OPTIONS") {
    response.statusCode = 204;
    response.end();
    return;
  }

  const url = new URL(request.url ?? "/", MOCK_API_ORIGIN);
  const result = responder({
    method: request.method ?? "GET",
    path: url.pathname,
    searchParams: url.searchParams,
    headers: request.headers,
  });

  response.statusCode = result.status;
  response.setHeader("Content-Type", "application/json");
  response.end(JSON.stringify(result.body));
});

test.describe.configure({ mode: "serial" });

test.beforeAll(async () => {
  responder = () => ({
    status: 404,
    body: { success: false, message: "not mocked", data: null, trace_id: "trace-e2e-missing" },
  });

  await new Promise<void>((resolve, reject) => {
    server.once("error", reject);
    server.listen(45731, "127.0.0.1", () => {
      server.off("error", reject);
      resolve();
    });
  });
});

test.afterAll(async () => {
  await new Promise<void>((resolve, reject) => {
    server.close((error) => {
      if (error) {
        reject(error);
        return;
      }

      resolve();
    });
  });
});

test.beforeEach(() => {
  savedListingPageRequests = [];

  let savedListingsState = [
    buildSavedListing({
      id: "listing-2",
      title: "Newest Saved Loft",
      slug: "newest-saved-loft",
      created_at: "2026-03-18T00:00:00Z",
    }),
    buildSavedListing({
      id: "listing-1",
      title: "Earlier Garden Home",
      slug: "earlier-garden-home",
      created_at: "2026-03-17T00:00:00Z",
    }),
  ];

  responder = (request) => {
    if (request.path === "/auth/me") {
      expect(request.method).toBe("GET");
      return { status: 200, body: backendEnvelope(buyerUser) };
    }

    if (request.path === "/api/me/saved-listings") {
      expect(request.method).toBe("GET");
      savedListingPageRequests.push(Object.fromEntries(request.searchParams.entries()));

      return {
        status: 200,
        body: backendEnvelope({
          data: savedListingsState,
          total: savedListingsState.length,
          page: Number(request.searchParams.get("page") ?? "1"),
          limit: Number(request.searchParams.get("limit") ?? "12"),
          total_pages: savedListingsState.length > 0 ? 1 : 0,
        }),
      };
    }

    if (request.path === "/api/me/saved-listings/listing-2") {
      expect(request.method).toBe("DELETE");
      savedListingsState = savedListingsState.filter((listing) => listing.id !== "listing-2");
      return { status: 200, body: backendEnvelope({ listing_id: "listing-2", saved: false }) };
    }

    if (request.path === "/api/me/saved-listings/listing-1") {
      expect(request.method).toBe("DELETE");
      savedListingsState = savedListingsState.filter((listing) => listing.id !== "listing-1");
      return { status: 200, body: backendEnvelope({ listing_id: "listing-1", saved: false }) };
    }

    return {
      status: 404,
      body: { success: false, message: `Unhandled ${request.path}`, data: null, trace_id: "trace-e2e-404" },
    };
  };
});

test("saved listings redirects anonymous visitors to login", async ({ page }) => {
  responder = (request) => {
    if (request.path === "/auth/me") {
      return {
        status: 401,
        body: { success: false, message: "unauthenticated", data: null, trace_id: "trace-e2e-401" },
      };
    }

    return {
      status: 404,
      body: { success: false, message: `Unhandled ${request.path}`, data: null, trace_id: "trace-e2e-404" },
    };
  };

  await page.goto("/saved-listings");

  await expect(page).toHaveURL(/\/login$/);
  await expect(page.getByTestId("login-google-button")).toBeVisible();
});

test("saved listings renders newest-first cards and removes them through the save button", async ({ page }) => {
  await page.context().addCookies([
    {
      name: "access_token",
      value: "playwright-test-token",
      url: "http://127.0.0.1:3100",
    },
    {
      name: "access_token",
      value: "playwright-test-token",
      url: "http://127.0.0.1:45731",
    },
  ]);

  const forwardedCookies: string[] = [];
  const activeResponder = responder;
  responder = (request) => {
    if (request.path === "/auth/me" || request.path === "/api/me/saved-listings") {
      forwardedCookies.push(String(request.headers.cookie ?? ""));
    }

    return activeResponder(request);
  };

  await page.goto("/saved-listings");

  await expect(page.getByRole("heading", { level: 1, name: /saved listings, ready when you are/i })).toBeVisible();
  await expect(page.getByTestId("listing-card")).toHaveCount(2);
  await expect(page.getByTestId("listing-card").nth(0)).toContainText("Newest Saved Loft");
  await expect(page.getByTestId("listing-card").nth(1)).toContainText("Earlier Garden Home");
  expect(savedListingPageRequests).toContainEqual(expect.objectContaining({ page: "1", limit: "12" }));
  expect(forwardedCookies).toEqual(
    expect.arrayContaining([expect.stringContaining(SSR_AUTH_COOKIE), expect.stringContaining(SSR_AUTH_COOKIE)]),
  );

  await page.getByTestId("save-listing-button-listing-2").click();

  await expect(page.getByTestId("listing-card")).toHaveCount(1);
  await expect(page.getByText("Newest Saved Loft")).toHaveCount(0);
  await expect(page.getByTestId("listing-card").first()).toContainText("Earlier Garden Home");

  await page.getByTestId("save-listing-button-listing-1").click();

  await expect(page.getByTestId("saved-listings-empty")).toBeVisible();
  await expect(page.getByRole("link", { name: /explore listings/i })).toHaveAttribute("href", "/listings");
});

function buildSavedListing(
  overrides: Partial<{
    id: string;
    title: string;
    slug: string;
    created_at: string;
  }> = {},
) {
  return {
    id: "listing-1",
    user_id: "buyer-1",
    category_id: "category-1",
    category: {
      id: "category-1",
      name: "House",
      slug: "house",
      icon_url: null,
    },
    title: "Saved Listing",
    slug: "saved-listing",
    description: "Quiet street with a bright living room.",
    price: 2850000000,
    currency: "IDR",
    location_city: "Jakarta",
    location_district: "Menteng",
    address_detail: "Jl. Example 12",
    status: "active",
    is_featured: false,
    specifications: {},
    view_count: 100,
    images: [
      {
        id: "image-1",
        url: "https://images.example/saved-listing.jpg",
        format: "jpg",
        bytes: 1024,
        width: 1200,
        height: 800,
        original_filename: "saved-listing.jpg",
        is_primary: true,
        sort_order: 0,
        created_at: "2026-03-17T00:00:00Z",
      },
    ],
    created_at: "2026-03-17T00:00:00Z",
    updated_at: "2026-03-17T00:00:00Z",
    ...overrides,
  };
}

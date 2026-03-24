import { createServer } from "node:http";

import { expect, test } from "@playwright/test";

const MOCK_API_ORIGIN = "http://127.0.0.1:45731";

const backendEnvelope = <T,>(data: T, message = "Success") => ({
  success: true,
  message,
  data,
  trace_id: "trace-e2e",
});

const categoriesResponse = [
  {
    id: "cat-house",
    name: "House",
    slug: "house",
    parent_id: null,
    icon_url: null,
    created_at: "2026-03-17T00:00:00Z",
    parent: null,
    children: [],
  },
];

const listingPage = {
  data: [
    {
      id: "listing-1",
      user_id: "seller-1",
      category_id: "cat-house",
      category: { id: "cat-house", name: "House", slug: "house", icon_url: null },
      title: "Jakarta River House",
      slug: "jakarta-river-house",
      description: "Wide river view with compact urban access.",
      price: 3250000000,
      currency: "IDR",
      location_city: "Jakarta",
      location_district: "Menteng",
      address_detail: "Jl. Ciliwung 9",
      status: "active",
      is_featured: true,
      specifications: { bedrooms: 4, bathrooms: 3, land_area_sqm: 200, building_area_sqm: 180 },
      view_count: 42,
      images: [
        {
          id: "image-1",
          url: "https://images.example/river-house.jpg",
          format: "jpg",
          bytes: 1200,
          width: 1200,
          height: 800,
          original_filename: "river-house.jpg",
          is_primary: true,
          sort_order: 0,
          created_at: "2026-03-17T00:00:00Z",
        },
      ],
      created_at: "2026-03-17T00:00:00Z",
      updated_at: "2026-03-17T00:00:00Z",
    },
    {
      id: "listing-2",
      user_id: "seller-2",
      category_id: "cat-house",
      category: { id: "cat-house", name: "House", slug: "house", icon_url: null },
      title: "Garden Court Residence",
      slug: "garden-court-residence",
      description: "Private courtyard home with bright main living room.",
      price: 2890000000,
      currency: "IDR",
      location_city: "Jakarta",
      location_district: "Kebayoran",
      address_detail: "Jl. Melati 5",
      status: "active",
      is_featured: false,
      specifications: { bedrooms: 3, bathrooms: 2, land_area_sqm: 160, building_area_sqm: 140 },
      view_count: 18,
      images: [
        {
          id: "image-2",
          url: "https://images.example/garden-court.jpg",
          format: "jpg",
          bytes: 1250,
          width: 1200,
          height: 800,
          original_filename: "garden-court.jpg",
          is_primary: true,
          sort_order: 0,
          created_at: "2026-03-17T00:00:00Z",
        },
      ],
      created_at: "2026-03-17T00:00:00Z",
      updated_at: "2026-03-17T00:00:00Z",
    },
  ],
  total: 2,
  page: 1,
  limit: 12,
  total_pages: 8,
};

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
let listingsRequests: Array<Record<string, string>> = [];

const server = createServer(async (request, response) => {
  response.setHeader("Access-Control-Allow-Origin", "http://127.0.0.1:3000");
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
  listingsRequests = [];

  responder = (request) => {
    if (request.path === "/api/categories") {
      expect(request.method).toBe("GET");
      return { status: 200, body: backendEnvelope(categoriesResponse) };
    }

    if (request.path === "/api/listings") {
      expect(request.method).toBe("GET");
      listingsRequests.push(Object.fromEntries(request.searchParams.entries()));
      return { status: 200, body: backendEnvelope(listingPage) };
    }

    return {
      status: 404,
      body: { success: false, message: `Unhandled ${request.path}`, data: null, trace_id: "trace-e2e-404" },
    };
  };
});

test("desktop listings shell keeps map-left and results-right layout", async ({ page }) => {
  await page.setViewportSize({ width: 1440, height: 1200 });
  await page.goto("/listings?city=Jakarta&status=active&limit=12");

  await expect(page.getByTestId("listing-filters")).toBeVisible();
  await expect(page.getByTestId("listing-map-panel")).toBeVisible();
  await expect(page.getByTestId("listing-pagination")).toBeVisible();
  await expect(page.getByTestId("listing-card")).toHaveCount(2);

  const boxes = await Promise.all([
    page.getByTestId("listing-map-panel").boundingBox(),
    page.getByTestId("listing-card").first().boundingBox(),
  ]);

  expect(boxes[0]).not.toBeNull();
  expect(boxes[1]).not.toBeNull();
  expect(boxes[0]!.x).toBeLessThan(boxes[1]!.x);

  const overflowX = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth);
  expect(overflowX).toBe(false);
  expect(listingsRequests).toContainEqual(
    expect.objectContaining({ city: "Jakarta", status: "active", limit: "12" }),
  );
});

test("clear flow removes stale query params from the listings shell", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 900 });
  await page.goto("/listings?city=Jakarta&status=active&price_min=500000000&limit=12");

  await page.getByRole("link", { name: /clear/i }).click();

  await expect(page).toHaveURL(/\/listings$/);
  await expect(page.getByTestId("listing-filters")).toBeVisible();
  await expect(page.locator('input[name="city"]')).toHaveValue("");
  await expect(page.locator('select[name="status"]')).toHaveValue("");
  await expect(page.locator('input[name="price_min"]')).toHaveValue("");
});

test("mobile listings shell stacks toolbar, map, and results deterministically", async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto("/listings");

  await expect(page.getByTestId("listing-filters")).toBeVisible();
  await expect(page.getByTestId("listing-map-panel")).toBeVisible();
  await expect(page.getByTestId("listing-card")).toHaveCount(2);

  const positions = await Promise.all([
    page.getByTestId("listing-filters").boundingBox(),
    page.getByTestId("listing-map-panel").boundingBox(),
    page.getByTestId("listing-card").first().boundingBox(),
  ]);

  expect(positions[0]).not.toBeNull();
  expect(positions[1]).not.toBeNull();
  expect(positions[2]).not.toBeNull();
  expect(positions[0]!.y).toBeLessThan(positions[1]!.y);
  expect(positions[1]!.y).toBeLessThan(positions[2]!.y);

  const overflowX = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth);
  expect(overflowX).toBe(false);
});

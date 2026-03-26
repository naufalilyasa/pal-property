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
  items: [
    {
      id: "listing-1",
      category_id: "cat-house",
      category: { id: "cat-house", name: "House", slug: "house" },
      title: "Jakarta River House",
      slug: "jakarta-river-house",
      description_excerpt: "Wide river view with compact urban access.",
      transaction_type: "sale",
      price: 3250000000,
      currency: "IDR",
      location_province: "DKI Jakarta",
      location_city: "Jakarta",
      location_district: "Menteng",
      status: "active",
      is_featured: true,
      primary_image_url: "https://images.example/river-house.jpg",
      image_urls: ["https://images.example/river-house.jpg"],
      created_at: "2026-03-17T00:00:00Z",
      updated_at: "2026-03-17T00:00:00Z",
    },
    {
      id: "listing-2",
      category_id: "cat-house",
      category: { id: "cat-house", name: "House", slug: "house" },
      title: "Garden Court Residence",
      slug: "garden-court-residence",
      description_excerpt: "Private courtyard home with bright main living room.",
      transaction_type: "sale",
      price: 2890000000,
      currency: "IDR",
      location_province: "DKI Jakarta",
      location_city: "Jakarta",
      location_district: "Kebayoran",
      status: "active",
      is_featured: false,
      primary_image_url: "https://images.example/garden-court.jpg",
      image_urls: ["https://images.example/garden-court.jpg"],
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

    if (request.path === "/api/search/listings") {
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
  await page.goto("/listings?q=jakarta&transaction_type=sale&location_city=Jakarta&limit=12&sort=newest");

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
    expect.objectContaining({ q: "jakarta", transaction_type: "sale", location_city: "Jakarta", limit: "12", sort: "newest" }),
  );
});

test("search-backed listings shell tolerates query params without crashing", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 900 });
  await page.goto("/listings?q=jakarta&transaction_type=sale&location_city=Jakarta&price_min=500000000&limit=12");

  await expect(page.getByTestId("listing-filters")).toBeVisible();
  await expect(page).toHaveURL(/\/listings\?/);
  await expect(page.getByTestId("listing-card")).toHaveCount(2);
});

test("mobile listings shell stacks toolbar and results deterministically", async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto("/listings");

  await expect(page.getByTestId("listing-filters")).toBeVisible();
  await expect(page.getByTestId("listing-map-panel")).toBeHidden();
  await expect(page.getByTestId("listing-card")).toHaveCount(2);

  const positions = await Promise.all([
    page.getByTestId("listing-filters").boundingBox(),
    page.getByTestId("listing-card").first().boundingBox(),
  ]);

	 expect(positions[0]).not.toBeNull();
	 expect(positions[1]).not.toBeNull();
	 expect(positions[0]!.y).toBeLessThan(positions[1]!.y);

	 const overflowX = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth);
	 expect(overflowX).toBe(false);
});

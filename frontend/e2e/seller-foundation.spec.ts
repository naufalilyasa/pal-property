import { createServer } from "node:http";

import { expect, test } from "@playwright/test";

const backendEnvelope = <T,>(data: T, message = "Success") => ({
  success: true,
  message,
  data,
  trace_id: "trace-e2e",
});

const MOCK_API_ORIGIN = "http://127.0.0.1:45731";

const sellerUser = {
  id: "seller-1",
  name: "Seller One",
  email: "seller@example.com",
  avatar_url: null,
  role: "seller",
  created_at: "2026-03-17T00:00:00Z",
};

const categoriesResponse = [
  {
    id: "cat-root",
    name: "House",
    slug: "house",
    parent_id: null,
    icon_url: null,
    created_at: "2026-03-17T00:00:00Z",
    parent: null,
    children: [{ id: "cat-child", name: "Villa", slug: "villa", icon_url: null }],
  },
];

function buildListing(
  overrides: Partial<{
    id: string;
    category_id: string;
    category: { id: string; name: string; slug: string; icon_url: null };
    title: string;
    description: string;
    price: number;
    location_city: string;
    location_district: string;
    address_detail: string;
    status: string;
    specifications: {
      bedrooms: number;
      bathrooms: number;
      land_area_sqm: number;
      building_area_sqm: number;
    };
    images: Array<{
      id: string;
      url: string;
      original_filename: string;
      is_primary: boolean;
      sort_order: number;
      created_at: string;
    }>;
  }> = {},
) {
  return {
    id: "listing-7",
    user_id: "seller-1",
    category_id: "cat-child",
    category: { id: "cat-child", name: "Villa", slug: "villa", icon_url: null },
    title: "Existing Residence",
    slug: "existing-residence",
    description: "Fresh paint and pool.",
    price: 3150000000,
    currency: "IDR",
    location_city: "Bandung",
    location_district: "Cidadap",
    address_detail: "Jl. Setiabudi 10",
    status: "inactive",
    is_featured: false,
    specifications: {
      bedrooms: 5,
      bathrooms: 3,
      land_area_sqm: 240,
      building_area_sqm: 180,
    },
    view_count: 12,
    images: [],
    created_at: "2026-03-17T00:00:00Z",
    updated_at: "2026-03-17T00:00:00Z",
    ...overrides,
  };
}

type MockRequest = {
  method: string;
  path: string;
  headers: Record<string, string | string[] | undefined>;
  bodyText: string;
};

type MockResponse = {
  status: number;
  body: unknown;
};

let responder: (request: MockRequest) => MockResponse;

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

  const chunks: Buffer[] = [];
  for await (const chunk of request) {
    chunks.push(typeof chunk === "string" ? Buffer.from(chunk) : chunk);
  }

  const url = new URL(request.url ?? "/", MOCK_API_ORIGIN);
  const result = responder({
    method: request.method ?? "GET",
    path: url.pathname,
    headers: request.headers,
    bodyText: Buffer.concat(chunks).toString("utf-8"),
  });

  response.statusCode = result.status;
  response.setHeader("Content-Type", "application/json");
  response.end(JSON.stringify(result.body));
});

test.describe.configure({ mode: "serial" });

test.beforeAll(async () => {
  responder = () => ({
    status: 404,
    body: {
      success: false,
      message: "not mocked",
      data: null,
      trace_id: "trace-e2e-missing",
    },
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

test("seller foundation shell loads", async ({ page }) => {
  await page.goto("/");

  await expect(page).toHaveTitle(/pal property seller/i);
  await expect(
    page.getByRole("heading", {
      level: 1,
      name: /a calm workspace for sellers to prepare listing operations/i,
    }),
  ).toBeVisible();
  await expect(page.getByText(/seller app foundation/i)).toBeVisible();
});

test("unauthenticated users are redirected away from the dashboard subtree", async ({ page }) => {
  responder = (request) => {
    if (request.path === "/auth/me" || request.path === "/auth/me/listings") {
      expect(request.method).toBe("GET");
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

  await page.goto("/dashboard");

  await expect(page).toHaveURL(/\/login$/);
  await expect(page.getByTestId("login-google-button")).toBeVisible();
  await expect(page.getByTestId("dashboard-shell")).toHaveCount(0);
});

test("authenticated dashboard listings route renders seller inventory", async ({ page }) => {
  responder = (request) => {
    if (request.path === "/auth/me") {
      expect(request.method).toBe("GET");
      return { status: 200, body: backendEnvelope(sellerUser) };
    }

    if (request.path === "/auth/me/listings") {
      expect(request.method).toBe("GET");
      return {
        status: 200,
        body: backendEnvelope({
          data: [
            buildListing({
              id: "listing-1",
              title: "Garden Residence",
              status: "published",
              price: 2750000000,
              location_city: "Jakarta",
              images: [
                {
                  id: "image-1",
                  url: "https://images.example/listing-1.jpg",
                  original_filename: "listing-1.jpg",
                  is_primary: true,
                  sort_order: 0,
                  created_at: "2026-03-17T00:00:00Z",
                },
              ],
            }),
          ],
          total: 1,
          page: 1,
          limit: 10,
          total_pages: 1,
        }),
      };
    }

    return {
      status: 404,
      body: { success: false, message: `Unhandled ${request.path}`, data: null, trace_id: "trace-e2e-404" },
    };
  };

  await page.goto("/dashboard/listings");

  await expect(page).toHaveTitle(/pal property seller/i);
  await expect(page.getByRole("heading", { level: 1, name: /seller inventory/i })).toBeVisible();
  await expect(page.getByTestId("dashboard-listings-table")).toBeVisible();
  await expect(page.getByText(/seller@example.com/i)).toBeVisible();
  await expect(page.getByText(/garden residence/i)).toBeVisible();
  await expect(page.getByTestId("dashboard-refresh-button")).toBeVisible();
});

test("create listing route accepts the expanded listing fields", async ({ page }) => {
  responder = (request) => {
    if (request.path === "/auth/me") {
      expect(request.method).toBe("GET");
      return { status: 200, body: backendEnvelope(sellerUser) };
    }

    if (request.path === "/api/categories") {
      expect(request.method).toBe("GET");
      return { status: 200, body: backendEnvelope(categoriesResponse) };
    }

    if (request.path === "/auth/me/listings") {
      expect(request.method).toBe("GET");
      return {
        status: 200,
        body: backendEnvelope({
          data: [buildListing({ id: "listing-99", title: "Sunset Loft" })],
          total: 1,
          page: 1,
          limit: 1000,
          total_pages: 1,
        }),
      };
    }

    return {
      status: 404,
      body: { success: false, message: `Unhandled ${request.path}`, data: null, trace_id: "trace-e2e-404" },
    };
  };

  await page.goto("/dashboard/listings/new");

  await page.getByLabel(/^title/i).fill("  Sunset Loft  ");
  await page.getByLabel(/^price/i).fill("0");
  await page.getByLabel(/^city/i).fill("  Surabaya ");
  await page.getByLabel(/^district/i).fill("   ");
  await page.getByLabel(/^address detail/i).fill("  Tower A  ");
  await page.getByLabel(/^status/i).selectOption("sold");
  await page.getByLabel(/^transaction type/i).selectOption("sale");
  await page.getByLabel(/^bedrooms/i).fill("0");
  await page.getByLabel(/^bathrooms/i).fill("0");
  await page.getByLabel(/^land area/i).fill("120");

  await expect(page.getByLabel(/^title/i)).toHaveValue("  Sunset Loft  ");
  await expect(page.getByLabel(/^price/i)).toHaveValue("0");
  await expect(page.getByLabel(/^city/i)).toHaveValue("  Surabaya ");
  await expect(page.getByLabel(/^status/i)).toHaveValue("sold");
  await expect(page.getByLabel(/^transaction type/i)).toHaveValue("sale");
  await expect(page.getByLabel(/^land area/i)).toHaveValue("120");
});

test("edit flow saves changes and keeps image controls available", async ({ page }) => {
  const initialListing = buildListing({
    images: [
      {
        id: "image-1",
        url: "https://images.example/1.jpg",
        original_filename: "front.jpg",
        is_primary: true,
        sort_order: 0,
        created_at: "2026-03-17T00:00:00Z",
      },
      {
        id: "image-2",
        url: "https://images.example/2.jpg",
        original_filename: "pool.jpg",
        is_primary: false,
        sort_order: 1,
        created_at: "2026-03-17T00:00:01Z",
      },
    ],
  });
  let currentListing = initialListing;

  responder = (request) => {
    if (request.path === "/auth/me") {
      expect(request.method).toBe("GET");
      return { status: 200, body: backendEnvelope(sellerUser) };
    }

    if (request.path === "/auth/me/listings") {
      expect(request.method).toBe("GET");
      return {
        status: 200,
        body: backendEnvelope({
          data: [currentListing],
          total: 1,
          page: 1,
          limit: 1000,
          total_pages: 1,
        }),
      };
    }

    if (request.path === "/api/categories") {
      return { status: 200, body: backendEnvelope(categoriesResponse) };
    }

    if (request.path === "/api/listings/listing-7" && request.method === "PUT") {
      const payload = JSON.parse(request.bodyText);
      expect(payload).toEqual({
        category_id: "cat-root",
        title: "Existing Residence Updated",
        description: "Fresh paint and pool.",
        price: 3300000000,
        location_city: "Bandung",
        location_district: "Cidadap",
        address_detail: "Jl. Setiabudi 10",
        status: "inactive",
        specifications: {
          bedrooms: 5,
          bathrooms: 3,
          land_area_sqm: 240,
          building_area_sqm: 180,
        },
      });

      currentListing = buildListing({
        title: "Existing Residence Updated",
        category_id: "cat-root",
        category: { id: "cat-root", name: "House", slug: "house", icon_url: null },
        price: 3300000000,
        images: currentListing.images,
      });

      return {
        status: 200,
        body: backendEnvelope(currentListing),
      };
    }

    if (request.path === "/api/listings/listing-7/images" && request.method === "POST") {
      expect(request.headers["content-type"] ?? "").toContain("multipart/form-data");
      expect(request.bodyText.length).toBeGreaterThan(0);

      currentListing = buildListing({
        ...currentListing,
        images: [
          ...currentListing.images,
          {
            id: "image-3",
            url: "https://images.example/3.jpg",
            original_filename: "garden.png",
            is_primary: false,
            sort_order: 2,
            created_at: "2026-03-17T00:00:02Z",
          },
        ],
      });

      return {
        status: 200,
        body: backendEnvelope(currentListing),
      };
    }

    if (request.path === "/api/listings/listing-7/images/reorder" && request.method === "PATCH") {
      expect(JSON.parse(request.bodyText)).toEqual({
        ordered_image_ids: ["image-2", "image-1", "image-3"],
      });

      currentListing = buildListing({
        ...currentListing,
        images: [
          {
            id: "image-2",
            url: "https://images.example/2.jpg",
            original_filename: "pool.jpg",
            is_primary: false,
            sort_order: 0,
            created_at: "2026-03-17T00:00:01Z",
          },
          {
            id: "image-1",
            url: "https://images.example/1.jpg",
            original_filename: "front.jpg",
            is_primary: true,
            sort_order: 1,
            created_at: "2026-03-17T00:00:00Z",
          },
          {
            id: "image-3",
            url: "https://images.example/3.jpg",
            original_filename: "garden.png",
            is_primary: false,
            sort_order: 2,
            created_at: "2026-03-17T00:00:02Z",
          },
        ],
      });

      return {
        status: 200,
        body: backendEnvelope(currentListing),
      };
    }

    if (request.path === "/api/listings/listing-7/images/image-2/primary" && request.method === "PATCH") {
      currentListing = buildListing({
        ...currentListing,
        images: [
          {
            id: "image-2",
            url: "https://images.example/2.jpg",
            original_filename: "pool.jpg",
            is_primary: true,
            sort_order: 0,
            created_at: "2026-03-17T00:00:01Z",
          },
          {
            id: "image-1",
            url: "https://images.example/1.jpg",
            original_filename: "front.jpg",
            is_primary: false,
            sort_order: 1,
            created_at: "2026-03-17T00:00:00Z",
          },
          {
            id: "image-3",
            url: "https://images.example/3.jpg",
            original_filename: "garden.png",
            is_primary: false,
            sort_order: 2,
            created_at: "2026-03-17T00:00:02Z",
          },
        ],
      });

      return {
        status: 200,
        body: backendEnvelope(currentListing),
      };
    }

    if (request.path === "/api/listings/listing-7/images/image-3" && request.method === "DELETE") {
      currentListing = buildListing({
        ...currentListing,
        images: [
          {
            id: "image-2",
            url: "https://images.example/2.jpg",
            original_filename: "pool.jpg",
            is_primary: true,
            sort_order: 0,
            created_at: "2026-03-17T00:00:01Z",
          },
          {
            id: "image-1",
            url: "https://images.example/1.jpg",
            original_filename: "front.jpg",
            is_primary: false,
            sort_order: 1,
            created_at: "2026-03-17T00:00:00Z",
          },
        ],
      });

      return {
        status: 200,
        body: backendEnvelope(currentListing),
      };
    }

    return {
      status: 404,
      body: { success: false, message: `Unhandled ${request.path}`, data: null, trace_id: "trace-e2e-404" },
    };
  };

  await page.goto("/dashboard/listings/listing-7/edit");

  await page.getByLabel(/^title/i).fill("Existing Residence Updated");
  await page.getByLabel(/^price/i).fill("3300000000");
  await page.getByRole("button", { name: /save changes/i }).click();

  await expect(page.getByRole("button", { name: /upload image/i })).toBeVisible();
  await expect(page.locator("[data-testid='listing-image-card-image-1']")).toBeVisible();
  await expect(page.locator("[data-testid='listing-image-card-image-2']")).toBeVisible();
});

import { createServer } from "node:http";
import { Socket } from "node:net";

import { expect, test } from "@playwright/test";

const MOCK_API_ORIGIN = "http://127.0.0.1:45731";

const backendEnvelope = <T,>(data: T, message = "Success") => ({
  success: true,
  message,
  data,
  trace_id: "trace-e2e-chat",
});

const chatAnswer = `### Rekomendasi Properti
Temukan rumah impian dengan cepat: cek [Highland Villa](/listings/highland-villa) dan hubungi agen kami.`;
const degradedAnswer = "Maaf, saya belum menemukan properti aktif yang relevan.";

const chatRecommendation = {
  listing_id: "listing-chat-1",
  title: "Highland Villa",
  slug: "highland-villa",
  price: 4250000000,
  currency: "IDR",
  location_city: "Bandung",
  location_district: "Lembang",
  location_province: "Jawa Barat",
  primary_image_url: "https://images.example/highland-villa.jpg",
  category: {
    id: "cat-house",
    name: "House",
    slug: "house",
  },
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

const sockets = new Set<Socket>();
let responder: (request: MockRequest) => MockResponse;

const server = createServer(async (request, response) => {
  response.setHeader("Access-Control-Allow-Origin", request.headers.origin ?? "http://127.0.0.1:3000");
  response.setHeader("Access-Control-Allow-Credentials", "true");
  response.setHeader("Access-Control-Allow-Headers", "Content-Type, Accept, Cookie");
  response.setHeader("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS");
  response.setHeader("Connection", "close");

  if ((request.method ?? "GET") === "OPTIONS") {
    response.statusCode = 204;
    response.setHeader("Connection", "close");
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
  response.setHeader("Connection", "close");
  response.end(JSON.stringify(result.body));
});

server.on("connection", (socket) => {
  sockets.add(socket);
  socket.once("close", () => sockets.delete(socket));
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
    for (const socket of sockets) {
      socket.destroy();
    }
    sockets.clear();
    server.closeAllConnections?.();
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
  responder = (request) => {
    if (request.path === "/auth/me") {
      expect(request.method).toBe("GET");
      return {
        status: 401,
        body: { success: false, message: "unauthenticated", data: null, trace_id: "trace-e2e-401" },
      };
    }

    if (request.path === "/api/chat/messages") {
      expect(request.method).toBe("POST");
      return {
        status: 200,
        body: backendEnvelope({
          session_id: "session-chat-1",
          answer: chatAnswer,
          answer_format: "markdown",
          grounding: { is_degraded: false, listing_slugs: [chatRecommendation.slug] },
          recommendations: [chatRecommendation],
        }),
      };
    }

    return {
      status: 404,
      body: { success: false, message: `Unhandled ${request.path}`, data: null, trace_id: "trace-e2e-404" },
    };
  };
});

test("chat widget renders markdown answer with inline link and recommendation card", async ({ page }) => {
  await page.goto("/");

  const openChatButton = page.getByRole("button", { name: /open chatbot/i });
  await expect(openChatButton).toBeVisible();
  await openChatButton.click();

  const input = page.getByPlaceholder("Ketik pesan Anda...");
  await expect(input).toBeVisible();
  await input.fill("Cari rumah mewah 4 kamar");
  await input.press("Enter");

  const markdownHeading = page.getByRole("heading", { level: 3, name: /rekomendasi properti/i });
  await expect(markdownHeading).toBeVisible();

  const markdownContainer = page.locator("div:has(h3:has-text('Rekomendasi Properti'))").first();
  const inlineLink = markdownContainer.getByRole("link", { name: /highland villa/i }).first();
  await expect(inlineLink).toBeVisible();
  await expect(inlineLink).toHaveAttribute("href", "/listings/highland-villa");

  const cardCta = page.getByText(/Lihat detail/i).first();
  await expect(cardCta).toBeVisible();
  const cardHref = await cardCta.evaluate((element) => element.closest("a")?.getAttribute("href"));
  expect(cardHref).toBe("/listings/highland-villa");

  const headingBox = await markdownHeading.boundingBox();
  const cardBox = await cardCta.boundingBox();
  expect(headingBox).not.toBeNull();
  expect(cardBox).not.toBeNull();
  expect(cardBox!.y).toBeGreaterThan(headingBox!.y + headingBox!.height);
});


test("chat widget renders degraded answer without recommendation card", async ({ page }) => {
  responder = (request) => {
    if (request.path === "/auth/me") {
      expect(request.method).toBe("GET");
      return {
        status: 401,
        body: { success: false, message: "unauthenticated", data: null, trace_id: "trace-e2e-401" },
      };
    }

    if (request.path === "/api/chat/messages") {
      expect(request.method).toBe("POST");
      return {
        status: 200,
        body: backendEnvelope({
          session_id: "session-chat-degraded",
          answer: degradedAnswer,
          answer_format: "text",
          grounding: {
            is_degraded: true,
            degraded_reason: "no_active_grounded_results",
            listing_slugs: [],
          },
          recommendations: [],
        }),
      };
    }

    return {
      status: 404,
      body: { success: false, message: `Unhandled ${request.path}`, data: null, trace_id: "trace-e2e-404" },
    };
  };

  await page.goto("/");

  const openChatButton = page.getByRole("button", { name: /open chatbot/i });
  await expect(openChatButton).toBeVisible();
  await openChatButton.click();

  const input = page.getByPlaceholder("Ketik pesan Anda...");
  await expect(input).toBeVisible();
  await input.fill("Tampilkan fallback chat");
  await input.press("Enter");

  const botMessage = page.locator(`div:has-text("${degradedAnswer}")`).first();
  await expect(botMessage).toBeVisible();

  const cardCta = botMessage.locator("text=/Lihat detail/i");
  await expect(cardCta).toHaveCount(0);
});

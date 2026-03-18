import { describe, expect, it, vi } from "vitest";

import { bootstrapSellerSession } from "@/lib/session/bootstrap";

describe("bootstrapSellerSession", () => {
  it("returns an authenticated session from /auth/me and forwards cookies when provided", async () => {
    const fetchMock = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          success: true,
          message: "Success",
          data: {
            id: "seller-1",
            name: "Seller One",
            email: "seller@example.com",
            avatar_url: null,
            role: "seller",
            created_at: "2026-03-17T00:00:00Z",
          },
          trace_id: "trace-auth",
        }),
        {
          status: 200,
          headers: { "Content-Type": "application/json" },
        },
      ),
    );

    const session = await bootstrapSellerSession({
      fetch: fetchMock,
      baseUrl: "http://127.0.0.1:8080",
      cookieHeader: "access_token=abc; refresh_token=def",
    });

    expect(session).toEqual({
      status: "authenticated",
      user: {
        id: "seller-1",
        name: "Seller One",
        email: "seller@example.com",
        avatar_url: null,
        role: "seller",
        created_at: "2026-03-17T00:00:00Z",
      },
      traceId: "trace-auth",
    });

    const [, init] = fetchMock.mock.calls[0];
    expect(fetchMock).toHaveBeenCalledWith(
      "http://127.0.0.1:8080/auth/me",
      expect.objectContaining({
        method: "GET",
        cache: "no-store",
        credentials: "include",
        headers: expect.any(Headers),
      }),
    );
    expect(new Headers(init?.headers).get("Cookie")).toBe("access_token=abc; refresh_token=def");
    expect(new Headers(init?.headers).get("Accept")).toBe("application/json");
  });

  it("returns a defined unauthenticated session result for 401 responses", async () => {
    const fetchMock = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          success: false,
          message: "missing access token",
          data: null,
          trace_id: "trace-unauth",
        }),
        {
          status: 401,
          headers: { "Content-Type": "application/json" },
        },
      ),
    );

    const session = await bootstrapSellerSession({
      fetch: fetchMock,
      baseUrl: "http://127.0.0.1:8080",
    });

    expect(session).toEqual({
      status: "unauthenticated",
      user: null,
    });
  });
});

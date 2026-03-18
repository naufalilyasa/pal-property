import { describe, expect, it, vi } from "vitest";

import { apiRequest } from "@/lib/api/client";

describe("apiRequest", () => {
  it("normalizes backend envelopes and sends credentialed requests", async () => {
    const fetchMock = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          success: true,
          message: "Success",
          data: { id: "seller-1" },
          trace_id: "trace-123",
        }),
        {
          status: 200,
          headers: { "Content-Type": "application/json" },
        },
      ),
    );

    const result = await apiRequest<{ id: string }>("/auth/me", {
      fetch: fetchMock,
      baseUrl: "http://127.0.0.1:8080",
    });

    expect(result).toEqual({
      data: { id: "seller-1" },
      message: "Success",
      traceId: "trace-123",
    });
    expect(fetchMock).toHaveBeenCalledWith(
      "http://127.0.0.1:8080/auth/me",
      expect.objectContaining({
        credentials: "include",
        headers: expect.any(Headers),
      }),
    );

    const [, init] = fetchMock.mock.calls[0];
    expect(new Headers(init?.headers).get("Accept")).toBe("application/json");
  });

  it("surfaces backend failures as ApiError with envelope metadata", async () => {
    const fetchMock = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(
        JSON.stringify({
          success: false,
          message: "missing access token",
          data: null,
          trace_id: "trace-401",
        }),
        {
          status: 401,
          headers: { "Content-Type": "application/json" },
        },
      ),
    );

    await expect(
      apiRequest("/auth/me", {
        fetch: fetchMock,
        baseUrl: "http://127.0.0.1:8080",
      }),
    ).rejects.toMatchObject({
      name: "ApiError",
      message: "missing access token",
      status: 401,
      traceId: "trace-401",
      data: null,
    });
  });
});

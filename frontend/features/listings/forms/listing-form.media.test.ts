import { describe, expect, it, vi } from "vitest";

import {
  describeSelectedImageFiles,
  MAX_LISTING_VIDEO_BYTES,
  validateListingVideoSelection,
} from "./listing-media";

describe("listing media helpers", () => {
  it("summarizes multi-image selections for the seller form", () => {
    const files = [
      new File(["image-1"], "front.png", { type: "image/png" }),
      new File(["image-2"], "pool.png", { type: "image/png" }),
      new File(["image-3"], "garden.png", { type: "image/png" }),
    ];

    expect(describeSelectedImageFiles([])).toBe("No images selected yet");
    expect(describeSelectedImageFiles(files.slice(0, 1))).toBe("Ready: front.png");
    expect(describeSelectedImageFiles(files)).toBe("Ready: 3 images selected (front.png, pool.png, ...)");
  });

  it("rejects videos that exceed the client-side size hint before reading duration", async () => {
    const readDuration = vi.fn().mockResolvedValue(30);
    const file = new File(["video"], "tour.mp4", { type: "video/mp4" });

    Object.defineProperty(file, "size", {
      value: MAX_LISTING_VIDEO_BYTES + 1,
      configurable: true,
    });

    await expect(
      validateListingVideoSelection(file, {
        readDuration,
      }),
    ).resolves.toMatchObject({
      ok: false,
      durationSeconds: null,
      message: expect.stringMatching(/under 100 mb/i),
    });
    expect(readDuration).not.toHaveBeenCalled();
  });

  it("rejects videos that exceed the client-side duration hint before upload", async () => {
    const file = new File(["video"], "tour.mp4", { type: "video/mp4" });

    await expect(
      validateListingVideoSelection(file, {
        readDuration: vi.fn().mockResolvedValue(61),
      }),
    ).resolves.toMatchObject({
      ok: false,
      durationSeconds: 61,
      message: expect.stringMatching(/under 1m/i),
    });
  });
});

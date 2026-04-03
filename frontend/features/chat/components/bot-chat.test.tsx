import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { BotChat } from "./bot-chat";

const { browserFetchMock } = vi.hoisted(() => ({
  browserFetchMock: vi.fn(),
}));

vi.mock("@/lib/api/browser-fetch", () => ({
  browserFetch: browserFetchMock,
}));

vi.mock("next/image", () => ({
  default: (props: React.ImgHTMLAttributes<HTMLImageElement>) => <img {...props} alt={props.alt ?? ""} />,
}));

describe("BotChat", () => {
  beforeEach(() => {
    browserFetchMock.mockReset();
    window.HTMLElement.prototype.scrollIntoView = vi.fn();
  });

  it("renders a natural bot response with recommendation cards", async () => {
    browserFetchMock.mockResolvedValue({
      data: {
        answer: "Saya menemukan beberapa properti yang cukup cocok untuk Anda di Jakarta Selatan.",
        recommendations: [
          {
            listing_id: "listing-1",
            title: "Rumah Modern Senopati",
            slug: "rumah-modern-senopati",
            price: 4500000000,
            currency: "IDR",
            location_district: "Kebayoran Baru",
            location_city: "Jakarta Selatan",
            location_province: "DKI Jakarta",
            primary_image_url: "https://images.example/senopati.jpg",
          },
        ],
      },
    });

    render(<BotChat />);

    fireEvent.click(screen.getByRole("button", { name: /open chatbot/i }));
    expect(screen.getByText(/Halo! Saya asisten pintar dari PAL Property/i)).toBeInTheDocument();

    fireEvent.change(screen.getByPlaceholderText(/Ketik pesan Anda/i), {
      target: { value: "rekomendasi rumah di jakarta?" },
    });
    fireEvent.submit(screen.getByPlaceholderText(/Ketik pesan Anda/i).closest("form")!);

    await waitFor(() => {
      expect(screen.getByText(/Saya menemukan beberapa properti/i)).toBeInTheDocument();
    });

    expect(screen.getByText(/Rumah Modern Senopati/i)).toBeInTheDocument();
    expect(screen.getByText(/Rp 4,5 Miliar/i)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /Rumah Modern Senopati/i })).toHaveAttribute(
      "href",
      "/listings/rumah-modern-senopati",
    );
  });
});

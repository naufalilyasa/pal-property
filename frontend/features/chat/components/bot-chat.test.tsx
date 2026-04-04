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
        answer_format: "text",
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
    expect(screen.getByText(/Lihat detail/i)).toBeInTheDocument();
  });

  it("renders markdown bot answers with trusted listing links while keeping cards below the answer", async () => {
    browserFetchMock.mockResolvedValue({
      data: {
        answer: "### Pilihan cocok\n\nLihat [Rumah Modern Senopati](/listings/rumah-modern-senopati) untuk detail lebih lanjut.",
        answer_format: "markdown",
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
    fireEvent.change(screen.getByPlaceholderText(/Ketik pesan Anda/i), {
      target: { value: "rekomendasi rumah di jakarta?" },
    });
    fireEvent.submit(screen.getByPlaceholderText(/Ketik pesan Anda/i).closest("form")!);

    const answerHeading = await screen.findByRole("heading", { name: /Pilihan cocok/i, level: 3 });
    const inlineListingLink = screen.getByText("Rumah Modern Senopati", { selector: "a" });
    const cardCta = screen.getByText(/Lihat detail/i);

    expect(inlineListingLink).toHaveAttribute("href", "/listings/rumah-modern-senopati");
    expect(cardCta).toBeInTheDocument();
    expect(answerHeading.compareDocumentPosition(cardCta) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });

  it("renders unsupported markdown links as plain text instead of active navigation", async () => {
    browserFetchMock.mockResolvedValue({
      data: {
        answer: "Buka [dashboard](/dashboard) atau [tautan luar](https://example.com) untuk melihat perbedaannya.",
        answer_format: "markdown",
        recommendations: [
          {
            listing_id: "listing-3",
            title: "Rumah Aman Sadar",
            slug: "rumah-aman-sadar",
            price: 5200000000,
            currency: "IDR",
            location_district: "Kebayoran Baru",
            location_city: "Jakarta Selatan",
            location_province: "DKI Jakarta",
            primary_image_url: "https://images.example/aman-sadar.jpg",
          },
        ],
      },
    });

    render(<BotChat />);

    fireEvent.click(screen.getByRole("button", { name: /open chatbot/i }));
    fireEvent.change(screen.getByPlaceholderText(/Ketik pesan Anda/i), {
      target: { value: "uji tautan" },
    });
    fireEvent.submit(screen.getByPlaceholderText(/Ketik pesan Anda/i).closest("form")!);

    await waitFor(() => {
      expect(screen.getByText(/dashboard/i)).toBeInTheDocument();
      expect(screen.getByText(/tautan luar/i)).toBeInTheDocument();
    });


    const cardTitle = screen.getByText(/Rumah Aman Sadar/i);
    const cardCta = screen.getByText(/Lihat detail/i);
    expect(cardTitle).toBeInTheDocument();
    expect(cardCta).toBeInTheDocument();
    expect(cardCta.closest("a")).toHaveAttribute("href", "/listings/rumah-aman-sadar");

    expect(screen.queryByRole("link", { name: /^dashboard$/i })).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: /^tautan luar$/i })).not.toBeInTheDocument();
  });
});

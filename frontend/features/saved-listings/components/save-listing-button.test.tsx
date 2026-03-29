import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ApiError } from "@/lib/api/envelope";

import { SaveListingButton } from "./save-listing-button";

const { bootstrapSellerSessionMock, pushMock, refreshMock, removeSavedListingMock, saveListingMock } = vi.hoisted(() => ({
  bootstrapSellerSessionMock: vi.fn(),
  pushMock: vi.fn(),
  refreshMock: vi.fn(),
  removeSavedListingMock: vi.fn(),
  saveListingMock: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: pushMock,
    refresh: refreshMock,
  }),
}));

vi.mock("@/lib/session/bootstrap-client", () => ({
  bootstrapSellerSession: bootstrapSellerSessionMock,
}));

vi.mock("@/lib/api/saved-listings-client", () => ({
  removeSavedListing: removeSavedListingMock,
  saveListing: saveListingMock,
}));

describe("SaveListingButton", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("optimistically saves for authenticated users and keeps the saved state on success", async () => {
    const pendingSave = createDeferred<{ listingId: string; saved: boolean }>();

    bootstrapSellerSessionMock.mockResolvedValue({
      status: "authenticated",
      traceId: "trace-auth",
      user: {
        id: "user-1",
        name: "Buyer",
        email: "buyer@example.com",
        role: "user",
        created_at: "2026-03-29T00:00:00Z",
      },
    });
    saveListingMock.mockReturnValueOnce(pendingSave.promise);

    render(<SaveListingButton initialSaved={false} listingId="listing-1" variant="cta" />);

    const button = screen.getByTestId("save-listing-button");

    fireEvent.click(button);

    await waitFor(() => expect(saveListingMock).toHaveBeenCalledWith("listing-1"));
    await waitFor(() => expect(button).toHaveAttribute("aria-pressed", "true"));
    expect(button).toHaveTextContent(/^saved$/i);
    expect(button).toBeDisabled();

    pendingSave.resolve({ listingId: "listing-1", saved: true });

    await waitFor(() => expect(button).not.toBeDisabled());
    expect(button).toHaveTextContent(/^saved$/i);
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("optimistically unsaves for authenticated users, keeps the unsaved state on success, and can refresh the page", async () => {
    const pendingRemove = createDeferred<{ listingId: string; saved: boolean }>();

    bootstrapSellerSessionMock.mockResolvedValue({
      status: "authenticated",
      traceId: "trace-auth",
      user: {
        id: "user-1",
        name: "Buyer",
        email: "buyer@example.com",
        role: "user",
        created_at: "2026-03-29T00:00:00Z",
      },
    });
    removeSavedListingMock.mockReturnValueOnce(pendingRemove.promise);

    render(<SaveListingButton initialSaved listingId="listing-2" refreshOnRemove variant="icon" />);

    const button = screen.getByTestId("save-listing-button");

    fireEvent.click(button);

    await waitFor(() => expect(removeSavedListingMock).toHaveBeenCalledWith("listing-2"));
    await waitFor(() => expect(button).toHaveAttribute("aria-pressed", "false"));
    expect(button).toBeDisabled();

    pendingRemove.resolve({ listingId: "listing-2", saved: false });

    await waitFor(() => expect(button).not.toBeDisabled());
    expect(button).toHaveAttribute("aria-pressed", "false");
    expect(refreshMock).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("rolls back optimistic state and surfaces the failure when the mutation fails", async () => {
    bootstrapSellerSessionMock.mockResolvedValue({
      status: "authenticated",
      traceId: "trace-auth",
      user: {
        id: "user-1",
        name: "Buyer",
        email: "buyer@example.com",
        role: "user",
        created_at: "2026-03-29T00:00:00Z",
      },
    });
    saveListingMock.mockRejectedValueOnce(
      new ApiError("save failed", {
        status: 500,
        traceId: "trace-save-fail",
      }),
    );

    render(<SaveListingButton initialSaved={false} listingId="listing-3" variant="cta" />);

    const button = screen.getByTestId("save-listing-button");

    fireEvent.click(button);

    await waitFor(() => expect(saveListingMock).toHaveBeenCalledWith("listing-3"));
    await waitFor(() => expect(button).toHaveAttribute("aria-pressed", "false"));
    expect(await screen.findByRole("alert")).toHaveTextContent(/save failed \(trace trace-save-fail\)/i);
    expect(button).toHaveTextContent(/^save$/i);
  });

  it("redirects unauthenticated users to login without firing save mutations", async () => {
    bootstrapSellerSessionMock.mockResolvedValue({
      status: "unauthenticated",
      user: null,
    });

    render(<SaveListingButton initialSaved={false} listingId="listing-4" variant="icon" />);

    const button = screen.getByTestId("save-listing-button");

    fireEvent.click(button);

    await waitFor(() => expect(pushMock).toHaveBeenCalledWith("/login"));
    expect(saveListingMock).not.toHaveBeenCalled();
    expect(removeSavedListingMock).not.toHaveBeenCalled();
    expect(button).toHaveAttribute("aria-pressed", "false");
  });

  it("uses stable data test ids for single and repeated contexts", () => {
    render(
      <div>
        <SaveListingButton initialSaved={false} listingId="listing-5" variant="icon" />
        <SaveListingButton
          initialSaved
          listingId="listing-6"
          scope="repeated"
          variant="cta"
        />
      </div>,
    );

    expect(screen.getByTestId("save-listing-button")).toBeInTheDocument();
    expect(screen.getByTestId("save-listing-button-listing-6")).toBeInTheDocument();
  });
});

function createDeferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (error?: unknown) => void;

  const promise = new Promise<T>((innerResolve, innerReject) => {
    resolve = innerResolve;
    reject = innerReject;
  });

  return { promise, resolve, reject };
}

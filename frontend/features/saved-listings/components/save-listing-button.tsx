"use client";

import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";

import { ApiError } from "@/lib/api/envelope";
import { removeSavedListing, saveListing } from "@/lib/api/saved-listings-client";
import { bootstrapSellerSession } from "@/lib/session/bootstrap-client";

export type SaveListingButtonVariant = "icon" | "cta";
export type SaveListingButtonScope = "single" | "repeated";

type SaveListingButtonProps = {
  listingId: string;
  initialSaved: boolean;
  variant?: SaveListingButtonVariant;
  scope?: SaveListingButtonScope;
  refreshOnRemove?: boolean;
  className?: string;
};

export function SaveListingButton({
  listingId,
  initialSaved,
  variant = "icon",
  scope = "single",
  refreshOnRemove = false,
  className = "",
}: SaveListingButtonProps) {
  const router = useRouter();
  const [saved, setSaved] = useState(initialSaved);
  const [isPending, setIsPending] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [authState, setAuthState] = useState<"unknown" | "authenticated" | "unauthenticated">("unknown");

  useEffect(() => {
    setSaved(initialSaved);
    setErrorMessage(null);
  }, [initialSaved, listingId]);

  const testId = useMemo(
    () => (scope === "repeated" ? `save-listing-button-${listingId}` : "save-listing-button"),
    [listingId, scope],
  );

  async function handleClick() {
    if (isPending) {
      return;
    }

    setIsPending(true);
    setErrorMessage(null);

    try {
      if (authState !== "authenticated") {
        const session = await bootstrapSellerSession();

        if (session.status !== "authenticated") {
          setAuthState("unauthenticated");
          router.push("/login");
          return;
        }

        setAuthState("authenticated");
      }

      const previousSaved = saved;
      const nextSaved = !previousSaved;

      setSaved(nextSaved);

      const response = nextSaved
        ? await saveListing(listingId)
        : await removeSavedListing(listingId);

      setSaved(response.saved);

      if (refreshOnRemove && previousSaved && !response.saved) {
        router.refresh();
      }
    } catch (error) {
      setSaved(saved);
      setErrorMessage(formatSaveListingError(error));
    } finally {
      setIsPending(false);
    }
  }

  return (
    <div className="flex flex-col items-start gap-2">
      <button
        aria-label={saved ? "Remove saved listing" : "Save listing"}
        aria-pressed={saved}
        className={buildButtonClassName({ className, saved, variant })}
        data-testid={testId}
        disabled={isPending}
        onClick={handleClick}
        type="button"
      >
        <HeartIcon saved={saved} variant={variant} />
        {variant === "cta" ? <span>{saved ? "Saved" : "Save"}</span> : null}
      </button>

      {errorMessage ? (
        <p className="text-sm font-medium text-red-700" role="alert">
          {errorMessage}
        </p>
      ) : null}
    </div>
  );
}

function buildButtonClassName({
  className,
  saved,
  variant,
}: {
  className: string;
  saved: boolean;
  variant: SaveListingButtonVariant;
}) {
  const baseClassName =
    "inline-flex items-center justify-center rounded-full border transition-colors disabled:pointer-events-none disabled:opacity-50";

  if (variant === "cta") {
    return [
      baseClassName,
      "h-10 gap-2 px-4 text-sm font-semibold shadow-sm",
      saved
        ? "border-slate-900 bg-slate-900 text-white hover:bg-slate-900/90"
        : "border-slate-200 bg-white text-slate-900 hover:bg-slate-50",
      className,
    ]
      .filter(Boolean)
      .join(" ");
  }

  return [
    baseClassName,
    "h-10 w-10 bg-white text-slate-900 shadow-sm",
    saved
      ? "border-slate-900 bg-slate-900 text-white hover:bg-slate-900/90"
      : "border-slate-200 hover:bg-slate-50",
    className,
  ]
    .filter(Boolean)
    .join(" ");
}

function formatSaveListingError(error: unknown) {
  if (error instanceof ApiError && error.traceId) {
    return `${error.message} (trace ${error.traceId})`;
  }

  if (error instanceof Error && error.message) {
    return error.message;
  }

  return "Unable to update saved listing right now.";
}

function HeartIcon({
  saved,
  variant,
}: {
  saved: boolean;
  variant: SaveListingButtonVariant;
}) {
  const size = variant === "cta" ? 16 : 14;

  return (
    <svg
      aria-hidden="true"
      fill={saved ? "currentColor" : "none"}
      height={size}
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="2"
      viewBox="0 0 24 24"
      width={size}
      xmlns="http://www.w3.org/2000/svg"
    >
      <path d="M19 14c1.49-1.46 3-3.21 3-5.5A5.5 5.5 0 0 0 16.5 3c-1.76 0-3 .5-4.5 2-1.5-1.5-2.74-2-4.5-2A5.5 5.5 0 0 0 2 8.5c0 2.3 1.5 4.05 3 5.5l7 7Z" />
    </svg>
  );
}

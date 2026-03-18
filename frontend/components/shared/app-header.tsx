import Link from "next/link";

type AppHeaderProps = {
  sellerName: string;
};

export function AppHeader({ sellerName }: AppHeaderProps) {
  return (
    <header className="flex flex-col gap-4 border-b border-[var(--line)] pb-6 lg:flex-row lg:items-center lg:justify-between" data-testid="app-header">
      <div className="space-y-2">
        <p className="text-xs font-semibold uppercase tracking-[0.3em] text-[var(--accent)]" style={{ fontFamily: "var(--font-mono), monospace" }}>
          Seller dashboard
        </p>
        <div>
          <h1 className="text-3xl font-semibold tracking-[-0.04em] text-[var(--ink)] sm:text-4xl">PAL Property seller workspace</h1>
          <p className="mt-2 max-w-2xl text-sm leading-7 text-[var(--muted)] sm:text-base">
            Review inventory, refine property records, and keep listing media synchronized with the backend.
          </p>
        </div>
      </div>

      <div className="flex items-center gap-3">
        <button
          aria-label="Toggle dashboard navigation"
          className="inline-flex items-center justify-center rounded-full border border-[var(--line)] bg-[var(--panel)] px-4 py-2 text-sm font-semibold text-[var(--ink)] lg:hidden"
          data-testid="dashboard-nav-toggle"
          type="button"
        >
          Menu
        </button>
        <div className="rounded-[1.5rem] border border-[var(--line)] bg-white/72 px-5 py-4 text-sm text-[var(--muted)]">
          <p className="text-[11px] uppercase tracking-[0.28em]" style={{ fontFamily: "var(--font-mono), monospace" }}>
            Active session
          </p>
          <p className="mt-2 font-medium text-[var(--ink)]" data-testid="current-user-email">{sellerName}</p>
        </div>
        <Link className="hidden rounded-full border border-[var(--line)] bg-[var(--panel)] px-4 py-2 text-sm font-semibold text-[var(--ink)] transition hover:border-[var(--accent)] hover:text-[var(--accent)] lg:inline-flex" href="/dashboard/listings/new">
          New listing
        </Link>
      </div>
    </header>
  );
}

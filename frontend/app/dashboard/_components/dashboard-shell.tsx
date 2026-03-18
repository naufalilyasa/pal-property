type DashboardShellProps = {
  sellerName?: string;
  children: React.ReactNode;
};

export function DashboardShell({ sellerName, children }: DashboardShellProps) {
  return (
    <main className="min-h-screen px-6 py-8 sm:px-10 lg:px-12">
      <div className="mx-auto flex min-h-[calc(100vh-4rem)] w-full max-w-6xl flex-col rounded-[2rem] border border-white/60 bg-[var(--panel)] p-6 shadow-[0_30px_100px_rgba(15,23,42,0.18)] backdrop-blur sm:p-8">
        <header className="flex flex-col gap-6 border-b border-[var(--line)] pb-8 lg:flex-row lg:items-end lg:justify-between">
          <div className="space-y-3">
            <p
              className="text-xs font-semibold uppercase tracking-[0.3em] text-[var(--accent)]"
              style={{ fontFamily: "var(--font-mono), monospace" }}
            >
              Seller dashboard
            </p>
            <div className="space-y-2">
              <h1 className="text-3xl font-semibold tracking-[-0.04em] text-[var(--ink)] sm:text-4xl">
                Your listings
              </h1>
              <p className="max-w-2xl text-sm leading-7 text-[var(--muted)] sm:text-base">
                Review inventory, scan current status, and spot which properties need attention next.
              </p>
            </div>
          </div>

          <div className="rounded-[1.5rem] border border-[var(--line)] bg-white/72 px-5 py-4 text-sm text-[var(--muted)]">
            <p
              className="text-[11px] uppercase tracking-[0.28em]"
              style={{ fontFamily: "var(--font-mono), monospace" }}
            >
              Active session
            </p>
            <p className="mt-2 font-medium text-[var(--ink)]">{sellerName ?? "Guest seller"}</p>
          </div>
        </header>

        <section className="flex-1 py-8">{children}</section>
      </div>
    </main>
  );
}

export default function Home() {
  return (
    <main className="min-h-screen px-6 py-10 sm:px-10 lg:px-12">
      <div className="mx-auto flex min-h-[calc(100vh-5rem)] w-full max-w-6xl flex-col justify-between rounded-4xl border border-white/60 bg-(--panel) p-8 shadow-[0_30px_100px_rgba(15,23,42,0.18)] backdrop-blur sm:p-10">
        <section className="grid gap-12 lg:grid-cols-[1.25fr_0.75fr] lg:items-end">
          <div className="space-y-8">
            <div
              className="inline-flex items-center gap-3 rounded-full border border-(--line) bg-white/70 px-4 py-2 text-xs font-semibold uppercase tracking-[0.3em] text-(--muted)"
              style={{ fontFamily: "var(--font-mono), monospace" }}
            >
              Seller app foundation
            </div>
            <div className="space-y-4">
              <p className="text-sm font-medium uppercase tracking-[0.3em] text-(--accent)">
                PAL Property
              </p>
              <h1 className="max-w-3xl text-4xl font-semibold tracking-[-0.04em] text-(--ink) sm:text-5xl lg:text-6xl">
                A calm workspace for sellers to prepare listing operations.
              </h1>
            <p className="max-w-2xl text-base leading-8 text-(--muted) sm:text-lg">
                The seller workspace already covers dashboard review, listing create and edit flows, and backend-authoritative media management.
            </p>
            </div>
          </div>

          <aside className="rounded-[1.75rem] border border-(--line) bg-(--panel-strong) p-6">
            <p
              className="text-xs uppercase tracking-[0.3em] text-(--muted)"
              style={{ fontFamily: "var(--font-mono), monospace" }}
            >
              Next steps
            </p>
            <ul className="mt-5 space-y-4 text-sm text-(--muted)">
              <li>Dashboard access stays behind seller session checks.</li>
              <li>Auth-aware API flows stay cookie-based.</li>
              <li>Listing create, edit, and image tools are live.</li>
            </ul>
          </aside>
        </section>

        <section className="grid gap-4 pt-12 md:grid-cols-3">
          <article className="rounded-3xl border border-(--line) bg-white/72 p-5">
            <p
              className="text-xs uppercase tracking-[0.3em] text-(--accent)"
              style={{ fontFamily: "var(--font-mono), monospace" }}
            >
              Routing
            </p>
            <h2 className="mt-3 text-xl font-semibold text-(--ink)">
              Dashboard-ready shell
            </h2>
            <p className="mt-2 text-sm leading-7 text-(--muted)">
              A focused home surface that can expand into seller navigation without reworking the layout foundation.
            </p>
          </article>
          <article className="rounded-3xl border border-(--line) bg-white/72 p-5">
            <p
              className="text-xs uppercase tracking-[0.3em] text-(--accent)"
              style={{ fontFamily: "var(--font-mono), monospace" }}
            >
              Testing
            </p>
            <h2 className="mt-3 text-xl font-semibold text-(--ink)">
              Minimal confidence layers
            </h2>
            <p className="mt-2 text-sm leading-7 text-(--muted)">
              Unit coverage protects the shell markup while a deterministic smoke flow confirms the app boots in a browser.
            </p>
          </article>
          <article className="rounded-3xl border border-(--line) bg-white/72 p-5">
            <p
              className="text-xs uppercase tracking-[0.3em] text-(--accent)"
              style={{ fontFamily: "var(--font-mono), monospace" }}
            >
              Scope
            </p>
            <h2 className="mt-3 text-xl font-semibold text-(--ink)">
              Seller-only baseline
            </h2>
            <p className="mt-2 text-sm leading-7 text-(--muted)">
              Marketplace flows remain out of scope, but seller listing forms and dashboard operations are already wired to the API.
            </p>
          </article>
        </section>
      </div>
    </main>
  );
}

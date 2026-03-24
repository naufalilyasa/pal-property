import Link from "next/link";
import { TopNav } from "@/features/listings/components/top-nav";
import { Footer } from "@/features/listings/components/footer";

export default function Home() {
  return (
    <div className="flex min-h-screen flex-col bg-background font-sans text-foreground">
      <TopNav />
      
      {/* Main Hero Content */}
      <main className="flex-1" data-testid="home-shell">
        <div className="mx-auto w-full max-w-[1580px] px-6 py-16 sm:px-10 sm:py-24 lg:px-12">
          <section className="grid gap-16 lg:grid-cols-[1.2fr_0.8fr] lg:items-center">
            <div className="space-y-8">
              <div
                className="inline-flex items-center gap-3 rounded-full border border-border bg-muted/30 px-4 py-2 text-xs font-semibold uppercase tracking-[0.25em] text-muted-foreground"
                style={{ fontFamily: "var(--font-mono)" }}
              >
                Seller app foundation
              </div>
              <div className="space-y-6">
                <p className="text-sm font-bold uppercase tracking-[0.3em] text-primary/80">
                  PAL Property
                </p>
                <h1 className="max-w-3xl text-4xl font-extrabold tracking-tight text-foreground sm:text-6xl lg:text-7xl">
                  A calm workspace for sellers to prepare listing operations.
                </h1>
                <p className="max-w-2xl text-lg text-muted-foreground sm:text-xl">
                  The seller workspace already covers dashboard review, listing create and edit flows, and backend-authoritative media management.
                </p>
              </div>
              
              <div className="flex flex-wrap gap-4 pt-4">
                <Link className="inline-flex h-12 items-center justify-center rounded-full bg-primary px-8 text-sm font-semibold text-primary-foreground shadow transition hover:bg-primary/90" href="/login">
                  Go to login
                </Link>
                <Link className="inline-flex h-12 items-center justify-center rounded-full border border-border bg-background px-8 text-sm font-semibold text-foreground shadow-sm transition hover:bg-accent hover:text-accent-foreground" href="/listings">
                  Browse public listings
                </Link>
              </div>
            </div>

            <aside className="rounded-2xl border border-border bg-card p-8 shadow-sm">
              <p
                className="text-xs font-semibold uppercase tracking-[0.25em] text-muted-foreground"
                style={{ fontFamily: "var(--font-mono)" }}
              >
                Next steps
              </p>
              <ul className="mt-6 flex flex-col gap-5 text-sm font-medium text-card-foreground">
                <li className="flex items-start gap-4">
                   <div className="mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary/10 text-primary">
                     <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
                   </div>
                   <span className="leading-relaxed">Dashboard access stays behind seller session checks.</span>
                </li>
                <li className="flex items-start gap-4">
                   <div className="mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary/10 text-primary">
                     <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
                   </div>
                   <span className="leading-relaxed">Auth-aware API flows stay cookie-based.</span>
                </li>
                <li className="flex items-start gap-4">
                   <div className="mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary/10 text-primary">
                     <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
                   </div>
                   <span className="leading-relaxed">Listing create, edit, and image tools are live.</span>
                </li>
              </ul>
            </aside>
          </section>

          <section className="mt-24 grid gap-6 md:grid-cols-3">
            <article className="rounded-2xl border border-border bg-card p-8 shadow-sm transition-colors hover:bg-accent/40">
              <p
                className="text-xs font-semibold uppercase tracking-[0.25em] text-primary"
                style={{ fontFamily: "var(--font-mono)" }}
              >
                Routing
              </p>
              <h2 className="mt-5 text-xl font-bold text-card-foreground">
                Dashboard-ready shell
              </h2>
              <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
                A focused home surface that can expand into seller navigation without reworking the layout foundation.
              </p>
            </article>
            <article className="rounded-2xl border border-border bg-card p-8 shadow-sm transition-colors hover:bg-accent/40">
              <p
                className="text-xs font-semibold uppercase tracking-[0.25em] text-primary"
                style={{ fontFamily: "var(--font-mono)" }}
              >
                Testing
              </p>
              <h2 className="mt-5 text-xl font-bold text-card-foreground">
                Minimal confidence layers
              </h2>
              <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
                Unit coverage protects the shell markup while a deterministic smoke flow confirms the app boots in a browser.
              </p>
            </article>
            <article className="rounded-2xl border border-border bg-card p-8 shadow-sm transition-colors hover:bg-accent/40">
              <p
                className="text-xs font-semibold uppercase tracking-[0.25em] text-primary"
                style={{ fontFamily: "var(--font-mono)" }}
              >
                Scope
              </p>
              <h2 className="mt-5 text-xl font-bold text-card-foreground">
                Seller-only baseline
              </h2>
              <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
                Marketplace flows remain out of scope, but seller listing forms and dashboard operations are already wired to the API.
              </p>
            </article>
          </section>
        </div>
      </main>

      <Footer />
    </div>
  );
}

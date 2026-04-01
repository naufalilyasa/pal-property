import Link from "next/link";
import Image from "next/image";
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
                Agen Properti Tepercaya
              </div>
              <div className="space-y-6">
                <p className="text-sm font-bold uppercase tracking-[0.3em] text-primary/80">
                  PAL PROPERTY
                </p>
                <h1 className="max-w-3xl text-4xl font-extrabold tracking-tight text-foreground sm:text-5xl lg:text-6xl">
                  Jual Beli Properti Mewah & Eksklusif di Indonesia.
                </h1>
                <p className="max-w-2xl text-lg text-muted-foreground sm:text-xl">
                  Kami siap membantu Anda menemukan berbagai macam properti
                  premium dengan layanan agen terbaik dan terpercaya di seluruh
                  penjuru Nusantara.
                </p>
              </div>

              <div className="flex flex-wrap gap-4 pt-4">
                <Link
                  className="inline-flex h-12 items-center justify-center rounded-full bg-primary px-8 text-sm font-semibold text-primary-foreground shadow transition hover:bg-primary/90"
                  href="/listings"
                >
                  Cari Properti
                </Link>
              </div>
            </div>

            <aside className="relative aspect-square w-full overflow-hidden rounded-2xl shadow-xl sm:aspect-video lg:aspect-4/3">
              <Image
                src="/hero-property-3.png"
                alt="Beautiful Indonesian Property"
                fill
                className="object-cover"
                priority
              />
            </aside>
          </section>

          <section className="mt-24 grid gap-6 md:grid-cols-3">
            <article className="rounded-2xl border border-border bg-card p-8 shadow-sm transition-colors hover:bg-accent/40">
              <p
                className="text-xs font-semibold uppercase tracking-[0.25em] text-primary"
                style={{ fontFamily: "var(--font-mono)" }}
              >
                Layanan
              </p>
              <h2 className="mt-5 text-xl font-bold text-card-foreground">
                Jual Beli Mudah
              </h2>
              <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
                Proses pemasaran dan pencarian properti yang transparan, cepat,
                dan aman melalui dukungan agen yang profesional.
              </p>
            </article>
            <article className="rounded-2xl border border-border bg-card p-8 shadow-sm transition-colors hover:bg-accent/40">
              <p
                className="text-xs font-semibold uppercase tracking-[0.25em] text-primary"
                style={{ fontFamily: "var(--font-mono)" }}
              >
                Kualitas
              </p>
              <h2 className="mt-5 text-xl font-bold text-card-foreground">
                Properti Pilihan
              </h2>
              <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
                Kurasi ketat terhadap listing properti untuk memberikan jaminan
                kualitas dan investasi menguntungkan bagi Anda.
              </p>
            </article>
            <article className="rounded-2xl border border-border bg-card p-8 shadow-sm transition-colors hover:bg-accent/40">
              <p
                className="text-xs font-semibold uppercase tracking-[0.25em] text-primary"
                style={{ fontFamily: "var(--font-mono)" }}
              >
                Jaringan
              </p>
              <h2 className="mt-5 text-xl font-bold text-card-foreground">
                Koneksi Luas
              </h2>
              <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
                Didukung oleh jaringan penjual dan pembeli properti yang masif
                di berbagai kota besar seluruh Indonesia.
              </p>
            </article>
          </section>
        </div>
      </main>

      <Footer />
    </div>
  );
}

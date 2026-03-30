import Link from "next/link";

import { Footer } from "@/features/listings/components/footer";
import { TopNav } from "@/features/listings/components/top-nav";

export default function SellerOnboardingPage() {
  return (
    <div className="flex min-h-screen flex-col bg-background font-sans text-foreground">
      <TopNav />
      <main className="flex flex-1 flex-col items-center justify-center px-6 py-20 text-center">
        <p className="text-xs font-semibold uppercase tracking-[0.3em] text-muted-foreground">Seller workspace</p>
        <h1 className="mt-4 text-3xl font-bold text-slate-900">Seller onboarding in progress</h1>
        <p className="mt-4 max-w-2xl text-sm leading-relaxed text-slate-600">
          Your account needs an additional capability check before we can show the full dashboard tools.
          We will guide you through the required setup so you can publish listings and manage media securely.
        </p>
        <div className="mt-8 flex flex-col items-center gap-3 text-sm text-slate-600">
          <Link
            className="font-semibold text-primary transition hover:text-primary/80"
            href="/seller/login"
          >
            Return to seller login
          </Link>
          <Link className="text-slate-500" href="/contact">
            Contact support if you believe this is an error
          </Link>
        </div>
      </main>
      <Footer />
    </div>
  );
}

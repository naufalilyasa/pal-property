import { headers } from "next/headers";
import { AppHeader } from "@/components/shared/app-header";
import { DashboardSidebar } from "@/components/shared/dashboard-sidebar";
import { redirect } from "next/navigation";

import { AuthIntent } from "@/features/auth/auth-intent";
import { requireUser } from "@/features/auth/server/require-user";
import {
  resolveAuthIntentDestination,
  SELLER_DASHBOARD_PATH,
} from "@/features/auth/auth-destination";

export default async function ProtectedDashboardLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const intent: AuthIntent = "seller";
  const headerStore = await headers();
  const returnTo = headerStore.get("x-pathname") ?? SELLER_DASHBOARD_PATH;
  const user = await requireUser({ intent, returnTo });

  const destination = resolveAuthIntentDestination(intent, user.seller_capabilities);

  if (destination !== SELLER_DASHBOARD_PATH) {
    redirect(destination);
  }

  return (
    <div className="flex min-h-screen w-full flex-col bg-gray-50/50 font-sans text-slate-950">
      <AppHeader sellerName={user.email} />
      <div className="flex flex-1">
        <aside className="hidden w-64 shrink-0 flex-col border-r border-gray-200 bg-white md:flex">
          <DashboardSidebar />
        </aside>
        <main className="flex-1 p-6 md:p-8 lg:p-10">
          <div className="mx-auto max-w-6xl">
            {children}
          </div>
        </main>
      </div>
    </div>
  );
}

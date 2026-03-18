import { AppHeader } from "@/components/shared/app-header";
import { DashboardSidebar } from "@/components/shared/dashboard-sidebar";
import { requireUser } from "@/features/auth/server/require-user";

export default async function ProtectedDashboardLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const user = await requireUser();

  return (
    <main className="min-h-screen px-6 py-8 sm:px-10 lg:px-12">
      <div className="mx-auto flex min-h-[calc(100vh-4rem)] w-full max-w-7xl flex-col rounded-[2rem] border border-white/60 bg-[var(--panel)] p-6 shadow-[0_30px_100px_rgba(15,23,42,0.18)] backdrop-blur sm:p-8">
        <AppHeader sellerName={user.email} />
        <div className="mt-8 grid gap-6 lg:grid-cols-[280px_1fr]">
          <DashboardSidebar />
          <section>{children}</section>
        </div>
      </div>
    </main>
  );
}

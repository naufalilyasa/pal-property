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

import { DashboardShell } from "@/app/dashboard/_components/dashboard-shell";
import { getSellerSession } from "@/lib/session/server";
import { redirect } from "next/navigation";

export default async function DashboardLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const session = await getSellerSession();

  if (session.status === "unauthenticated") {
    redirect("/");
  }

  return <DashboardShell sellerName={session.user.name}>{children}</DashboardShell>;
}

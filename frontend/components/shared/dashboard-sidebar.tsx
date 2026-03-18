import Link from "next/link";

const navItems = [
  { href: "/dashboard", label: "Overview" },
  { href: "/dashboard/listings", label: "Listings" },
  { href: "/dashboard/listings/new", label: "Create listing" },
  { href: "/listings", label: "Public listings" },
];

export function DashboardSidebar() {
  return (
    <aside className="rounded-[1.75rem] border border-[var(--line)] bg-white/72 p-5" data-testid="dashboard-sidebar">
      <p className="text-xs uppercase tracking-[0.28em] text-[var(--muted)]" style={{ fontFamily: "var(--font-mono), monospace" }}>
        Navigation
      </p>
      <nav className="mt-4 flex flex-col gap-2">
        {navItems.map((item) => (
          <Link key={item.href} className="rounded-2xl border border-transparent px-4 py-3 text-sm font-semibold text-[var(--ink)] transition hover:border-[var(--line)] hover:bg-[var(--panel)]" href={item.href}>
            {item.label}
          </Link>
        ))}
      </nav>
    </aside>
  );
}

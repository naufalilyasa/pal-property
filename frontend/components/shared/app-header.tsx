import Link from "next/link";

type AppHeaderProps = {
  sellerName: string;
};

export function AppHeader({ sellerName }: AppHeaderProps) {
  return (
    <header className="sticky top-0 z-30 flex h-16 shrink-0 items-center justify-between border-b border-gray-200 bg-white px-4 shadow-sm md:px-8" data-testid="app-header">
      <div className="flex items-center gap-4">
        <Link href="/dashboard" className="text-xl font-bold tracking-tight text-slate-900">
          PAL Workspace
        </Link>
        <span className="hidden rounded-md bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-600 md:inline-block">
          Seller Dashboard
        </span>
      </div>

      <div className="flex items-center gap-4">
        <button
          aria-label="Toggle navigation"
          className="inline-flex h-9 w-9 items-center justify-center rounded-md border border-gray-200 bg-white text-sm font-medium shadow-sm transition-colors hover:bg-gray-100 md:hidden"
          type="button"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="4" x2="20" y1="12" y2="12" /><line x1="4" x2="20" y1="6" y2="6" /><line x1="4" x2="20" y1="18" y2="18" /></svg>
        </button>
        <div className="hidden items-center gap-2 rounded-md border border-gray-200 bg-white px-3 py-1.5 shadow-sm md:flex">
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-gray-500"><path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2" /><circle cx="12" cy="7" r="4" /></svg>
          <span className="text-sm font-medium text-slate-700">{sellerName}</span>
        </div>
        <div
          className="hidden h-9 items-center justify-center rounded-md bg-slate-900 px-4 text-sm font-medium text-slate-50 shadow transition-colors hover:bg-slate-900/90 md:inline-flex"
        >
          <Link
            href="/dashboard/listings/new">
            New listing
          </Link>
        </div>
      </div>
    </header>
  );
}

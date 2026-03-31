import Link from "next/link";

export function TopNav() {
  return (
    <nav className="flex items-center justify-between border-b border-gray-200 bg-white px-4 py-3 sm:px-6">
      <div className="flex items-center gap-8">
        <Link
          href="/"
          className="text-2xl font-black tracking-tighter text-[#111]"
        >
          Pal Property
        </Link>
        <div className="hidden items-center gap-6 text-sm font-medium text-gray-700 md:flex">
          <Link href="/listings" className="hover:text-black">
            Cari Properti
          </Link>
          <Link href="/" className="hover:text-black">
            Tentang Kami
          </Link>
        </div>
      </div>
      <div>
        <Link
          href="/login"
          className="rounded-full bg-[#111] px-5 py-2 text-sm font-semibold text-white transition hover:bg-black/90"
        >
          Login
        </Link>
      </div>
    </nav>
  );
}

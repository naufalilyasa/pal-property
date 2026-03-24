import Link from "next/link";

export function TopNav() {
  return (
    <nav className="flex items-center justify-between border-b border-gray-200 bg-white px-4 py-3 sm:px-6">
      <div className="flex items-center gap-8">
        <Link
          href="/listings"
          className="text-2xl font-black tracking-tighter text-[#111]"
        >
          FIND
        </Link>
        <div className="hidden items-center gap-6 text-sm font-medium text-gray-700 md:flex">
          <Link href="/search" className="hover:text-black">
            Search
          </Link>
          <Link href="/agents" className="hover:text-black">
            Agents
          </Link>
          <Link href="/join" className="hover:text-black">
            Join
          </Link>
          <button className="flex items-center gap-1 hover:text-black">
            My Account
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="m6 9 6 6 6-6" />
            </svg>
          </button>
          <button className="flex items-center gap-1 hover:text-black">
            Company
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="m6 9 6 6 6-6" />
            </svg>
          </button>
          <button className="flex items-center gap-1 hover:text-black">
            About
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="m6 9 6 6 6-6" />
            </svg>
          </button>
        </div>
      </div>
      <div
        className="rounded-full bg-[#111] px-5 py-2 text-sm font-semibold text-white transition hover:bg-black/90"
      >
        <Link
          href="/create"
        >
          Create
        </Link>
      </div>
    </nav>
  );
}

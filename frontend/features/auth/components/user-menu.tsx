"use client";

import { useState, useRef, useEffect } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import type { CurrentUser } from "@/features/auth/server/current-user";
import { browserFetch } from "@/lib/api/browser-fetch";

export function UserMenu({ user }: { user: CurrentUser }) {
  const [isOpen, setIsOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);
  const router = useRouter();

  // Close menu on click outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleLogout = async () => {
    try {
      await browserFetch("/auth/logout", { method: "POST" });
      router.refresh(); // Tells Next.js to re-fetch Server Components (like TopNav)
      // Optional: force reload so client state cleans up entirely
      window.location.reload();
    } catch (e) {
      console.error("Logout failed", e);
    }
  };

  const nameOrEmail = user.name || user.email;
  const initials = nameOrEmail.slice(0, 2).toUpperCase();

  return (
    <div className="relative" ref={menuRef}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-2 rounded-full border border-gray-200 bg-white p-1 pr-3 shadow-sm transition hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-black/5"
      >
        <div className="flex h-8 w-8 items-center justify-center rounded-full bg-[#111] text-xs font-bold text-white">
          {initials}
        </div>
        <span className="max-w-[120px] truncate text-sm font-medium text-gray-700">
          {nameOrEmail}
        </span>
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2.5"
          strokeLinecap="round"
          strokeLinejoin="round"
          className={`text-gray-400 transition-transform ${isOpen ? "rotate-180" : ""}`}
        >
          <path d="m6 9 6 6 6-6" />
        </svg>
      </button>

      {isOpen && (
        <div className="absolute right-0 mt-2 w-56 origin-top-right rounded-xl border border-gray-100 bg-white p-1 shadow-lg ring-1 ring-black/5 focus:outline-none z-50">
          <div className="border-b border-gray-50 px-3 py-3 mb-1">
            <p className="text-sm font-semibold text-gray-900">{user.name || "User"}</p>
            <p className="truncate text-xs text-gray-500">{user.email}</p>
          </div>

          {user.role !== "user" && (
            <Link
              href="/dashboard"
              className="flex items-center rounded-md px-3 py-2 text-sm text-gray-700 transition-colors hover:bg-gray-50 hover:text-gray-900"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                className="mr-2 text-gray-400"
              >
                <path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2" />
                <circle cx="12" cy="7" r="4" />
              </svg>
              Dashboard
            </Link>
          )}

          <div className="my-1 border-t border-gray-50"></div>

          <button
            onClick={handleLogout}
            className="flex w-full items-center rounded-md px-3 py-2 text-sm text-red-600 transition-colors hover:bg-red-50 hover:text-red-700"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
              className="mr-2"
            >
              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
              <polyline points="16 17 21 12 16 7" />
              <line x1="21" x2="9" y1="12" y2="12" />
            </svg>
            Logout
          </button>
        </div>
      )}
    </div>
  );
}

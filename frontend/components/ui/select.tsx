import type { SelectHTMLAttributes } from "react";

export function Select({ className = "", ...props }: SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <select
      className={`w-full rounded-[1rem] border border-[var(--line)] bg-[var(--panel)] px-4 py-3 text-sm text-[var(--ink)] outline-none transition focus:border-[var(--accent)] focus:ring-2 focus:ring-[color:var(--accent)]/15 ${className}`.trim()}
      {...props}
    />
  );
}

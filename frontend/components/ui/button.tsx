import type { ButtonHTMLAttributes } from "react";

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "primary" | "secondary" | "ghost" | "destructive";
};

const variantClassNames: Record<NonNullable<ButtonProps["variant"]>, string> = {
  primary: "bg-[var(--accent)] text-white hover:opacity-90",
  secondary: "border border-[var(--line)] bg-[var(--panel)] text-[var(--ink)] hover:border-[var(--accent)] hover:text-[var(--accent)]",
  ghost: "text-[var(--ink)] hover:bg-[var(--panel)]",
  destructive: "border border-red-200 bg-red-50 text-red-700 hover:border-red-300 hover:bg-red-100",
};

export function Button({ className = "", type = "button", variant = "primary", ...props }: ButtonProps) {
  return (
    <button
      className={`inline-flex items-center justify-center rounded-full px-5 py-3 text-sm font-semibold transition disabled:cursor-not-allowed disabled:opacity-60 ${variantClassNames[variant]} ${className}`.trim()}
      type={type}
      {...props}
    />
  );
}

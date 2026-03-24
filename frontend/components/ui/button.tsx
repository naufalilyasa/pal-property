import type { ButtonHTMLAttributes } from "react";

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "primary" | "secondary" | "outline" | "ghost" | "destructive";
};

const variantClassNames: Record<NonNullable<ButtonProps["variant"]>, string> = {
  primary: "bg-slate-900 text-slate-50 shadow hover:bg-slate-900/90",
  secondary: "bg-slate-100 text-slate-900 shadow-sm hover:bg-slate-100/80",
  outline: "border border-slate-200 bg-white shadow-sm hover:bg-slate-100 hover:text-slate-900",
  ghost: "hover:bg-slate-100 hover:text-slate-900",
  destructive: "bg-red-500 text-slate-50 shadow-sm hover:bg-red-500/90",
};

export function Button({ className = "", type = "button", variant = "primary", ...props }: ButtonProps) {
  // We keep backward compatibility by interpreting 'secondary' as 'outline' if needed,
  // but let's standardise on the shadcn setup.
  const actualVariant = variant === "secondary" && !className.includes("bg-") ? "outline" : variant;

  return (
    <button
      className={`inline-flex h-9 items-center justify-center whitespace-nowrap rounded-md px-4 py-2 text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-slate-950 disabled:pointer-events-none disabled:opacity-50 ${variantClassNames[actualVariant]} ${className}`.trim()}
      type={type}
      {...props}
    />
  );
}

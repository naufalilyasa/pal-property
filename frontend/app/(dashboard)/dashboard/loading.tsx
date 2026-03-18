export default function DashboardLoading() {
  return (
    <div className="grid gap-6">
      <section className="rounded-3xl border border-(--line) bg-white/72 p-6">
        <p
          className="text-xs uppercase tracking-[0.28em] text-(--muted)"
          style={{ fontFamily: "var(--font-mono), monospace" }}
        >
          Loading listings
        </p>
        <div className="mt-5 grid gap-4">
          {["shell-a", "shell-b", "shell-c"].map((item) => (
            <div
              key={item}
              className="h-28 animate-pulse rounded-[1.25rem] border border-(--line) bg-(--panel-strong)"
            />
          ))}
        </div>
      </section>
    </div>
  );
}

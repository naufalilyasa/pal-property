# Learnings

- The seller-dashboard merge covered only part of the approved frontend ruleset; the full plan still needed route groups, feature slices, split browser/server fetch helpers, and public listing pages.
- `server-only` env modules cannot leak through shared browser helpers in Next.js App Router. Browser-safe URL helpers must stay isolated from server env parsing.
- Query-enabled client forms and client islands require test harness updates (`QueryClientProvider`, retry disabled in tests) once TanStack Query becomes part of the architecture.

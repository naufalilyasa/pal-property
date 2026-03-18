# Decisions

- Kept backend Go as the sole auth/session authority.
- Implemented route groups under `frontend/app/(dashboard)` and `frontend/app/(public)` while preserving `/dashboard` and `/listings` URLs.
- Used RHF + Zod + local shadcn-style primitives for the listing form instead of adding a separate UI generator dependency.
- Preserved `browserFetch` 401 refresh retry for client-initiated requests only; SSR helpers continue to redirect on `/auth/me` failure without server-side refresh loops.

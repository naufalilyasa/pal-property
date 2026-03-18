# AGENTS.md — frontend/lib/api/

## OVERVIEW

Frontend transport boundary. Shared envelope parsing plus browser/server fetch helpers live here.

## WHERE TO LOOK

| Need | Location | Notes |
|------|----------|-------|
| Shared envelope/error shape | `envelope.ts` | preserves `trace_id` |
| Browser requests | `browser-fetch.ts` | `credentials: include`, client-only refresh retry |
| Server requests | `server-fetch.ts` | `server-only`, forwards cookies explicitly |
| Feature API helpers | `listing-form.ts`, `seller-listings.ts` | thin wrappers over shared fetchers |

## CONVENTIONS

- Normalize backend `{ success, message, data, trace_id }` in one place.
- Keep browser-safe helpers free of server-only imports.
- Only client-initiated protected requests may auto-refresh on 401.
- Pass backend cookies explicitly on server-side authenticated reads.

## ANTI-PATTERNS

- **NEVER** add Axios here.
- **NEVER** mix `server-only` env/helpers into browser-fetch code paths.
- **NEVER** scatter envelope parsing into feature components.

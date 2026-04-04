# AGENTS.md — frontend/e2e/

## OVERVIEW

Playwright browser coverage. Tests run the real Next app against a local mock backend and assert user-visible flows.

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Public browse/detail | `public-listings.spec.ts` | search params, map panel, detail media |
| Auth/login UX | `login.spec.ts`, `login-ux.spec.ts` | OAuth entry and redirects |
| Dashboard flows | `seller-foundation.spec.ts` | dashboard auth, listings create/edit shells |
| Saved listings | `saved-listings.spec.ts` | protected list + toggle behavior |
| Shared helpers | `helpers/` | fixtures and support utilities |

## CONVENTIONS

- Start from the shared mock-backend pattern on `127.0.0.1:45731`; keep API responses envelope-shaped.
- Prefer real navigation, redirects, and visible assertions over implementation-detail checks.
- Keep specs serial when they share a mock server lifecycle.
- Match the current Playwright config/web-server assumptions before changing ports or env vars.
- Cover both anonymous and authenticated roles when a route changes behavior by auth state.

## ANTI-PATTERNS

- **NEVER** depend on live backend, Cloudinary, Elasticsearch, or OAuth providers in e2e.
- **NEVER** assert internal React implementation details when DOM text, URL, or test IDs can prove behavior.
- **NEVER** introduce a new mock API shape that diverges from `{ success, message, data, trace_id }`.

# AGENTS.md — frontend/app/(dashboard)/

## OVERVIEW

Protected seller routes. Server-first auth gating, seller workspace layouts, and dashboard page entrypoints live here.

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Shared protected shell | `layout.tsx` | must gate with `requireUser()` |
| Seller overview | `dashboard/page.tsx` | server-rendered metrics + client refresh island |
| Listings index | `dashboard/listings/page.tsx` | SSR list, no route-level `useQuery` |
| Create/edit routes | `dashboard/listings/**/page.tsx` | page shells only; feature logic stays elsewhere |

## CONVENTIONS

- Keep route files as Server Components unless an entire page truly needs client state.
- Redirect unauthenticated users to `/login`, not `/`.
- Pull auth/session checks from `features/auth/server`, not ad hoc route code.
- Delegate listing form and image behavior to `features/listings/*`.

## ANTI-PATTERNS

- **NEVER** put RHF or heavy mutation logic directly into route entry files.
- **NEVER** bypass the protected layout with page-local auth hacks.
- **NEVER** fetch seller data from guessed endpoints.

# AGENTS.md — frontend/app/(public)/

## OVERVIEW

Public buyer-facing routes. Server-rendered browse/detail pages should stay aligned with backend query contracts.

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Listing browse | `listings/page.tsx` | reads query params, calls server helper |
| Listing detail | `listings/[slug]/page.tsx` | slug-based server read |

## CONVENTIONS

- Keep first render server-driven.
- Preserve backend filter names: `page`, `limit`, `city`, `category_id`, `price_min`, `price_max`, `status`.
- Use feature server helpers and public listing components instead of inline fetch logic.
- Client islands here should be additive only, not the main data source.

## ANTI-PATTERNS

- **NEVER** move the public pages wholesale to client rendering.
- **NEVER** rename backend filter params in route code without a mapper layer.
- **NEVER** introduce auth/session assumptions in this route group.

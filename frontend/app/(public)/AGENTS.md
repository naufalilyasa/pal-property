# AGENTS.md — frontend/app/(public)/

## OVERVIEW

Public buyer-facing routes. Server-rendered browse/detail pages should stay aligned with the search-backed query contract and slug-based listing detail reads.

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Listing browse | `listings/page.tsx` | reads query params, calls server helper |
| Listing detail | `listings/[slug]/page.tsx` | slug-based server read |

## CONVENTIONS

- Keep first render server-driven.
- Preserve canonical public-search params: `q`, `transaction_type`, `category_id`, `location_province`, `location_city`, `price_min`, `price_max`, `sort`, `page`, and `limit`.
- Treat `view` as a shell-level browse toggle; route data still flows through `getSearchListings()`.
- Use feature server helpers and public listing components instead of inline fetch logic.
- Client islands here should be additive only, not the main data source.

## ANTI-PATTERNS

- **NEVER** move the public pages wholesale to client rendering.
- **NEVER** rename backend filter params in route code without a mapper layer.
- **NEVER** fork the public-search contract away from `features/listings/server/get-search-listings.ts`.
- **NEVER** introduce auth/session assumptions in this route group.

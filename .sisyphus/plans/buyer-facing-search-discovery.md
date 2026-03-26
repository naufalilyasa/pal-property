# Buyer-Facing Search and Discovery

## TL;DR
> **Summary**: Build the first real buyer-facing discovery flow on top of the search-indexing foundation: a public search-read API backed by Elasticsearch and a frontend browse/filter UI that consumes it.
> **Deliverables**:
> - backend `GET /api/search/listings` endpoint backed by Elasticsearch
> - search query parsing, validation, and pagination contract
> - frontend public browse page migrated from OLTP listing reads to the search endpoint
> - deterministic empty/error/filter UX and tests
> **Effort**: Large
> **Parallel**: YES - 2 waves
> **Critical Path**: T1 backend search-read contract -> T2 backend search endpoint -> T3 frontend integration -> T4 UX and verification -> F1-F4

## Context
### Why This Next
- Seller flows are already implemented.
- Listing model expansion is already in place.
- Eventing + indexing write-side foundation already exists.
- The biggest user-facing gap now is buyer/public discovery powered by the new search index.

### Current Repo Reality
- Public browse/detail pages exist in `frontend/app/(public)/listings/*`, but they still read from backend OLTP listing endpoints.
- `backend/search-read-contract.md` already proposes `GET /api/search/listings` with a limited query surface.
- Search indexing currently writes canonical search documents with: `id`, `user_id`, `category_id`, category summary, `title`, `slug`, `description_excerpt`, `transaction_type`, `price`, `currency`, `location_province`, `location_city`, `location_district`, `status`, `is_featured`, `primary_image_url`, `image_urls`, and timestamps.

### Locked Architectural Decisions
- `GET /api/search/listings` must be a separate public read endpoint backed only by Elasticsearch.
- Public listing detail pages remain on the existing Postgres-backed slug/id read path in phase 1.
- Public visibility is enforced server-side; the MVP does **not** expose a public `status` query parameter.
- Search response should stay card-ready and lean; do not expose seller/internal-only fields like `user_id` or `deleted_at` on the public API.
- Default sort behavior:
  - `relevance` only when `q` is present
  - `newest` otherwise

## Work Objectives
### Core Objective
Expose a stable buyer-facing search endpoint and migrate the public browse UI to use it, without leaking seller-only statuses or falling back to direct OLTP listing reads for search behavior.

### Must Have
- Backend search endpoint: `GET /api/search/listings`
- Public-safe results only (`active` and any other explicitly approved public-safe statuses)
- Query params limited to what the search document actually supports:
  - `q`
  - `transaction_type`
  - `category_id`
  - `location_province`
  - `location_city`
  - `price_min`, `price_max`
  - `page`, `limit`
  - `sort` (`relevance`, `newest`, `price_asc`, `price_desc`)
- Backend response still uses `{ success, message, data, trace_id }`
- Frontend public listings page reads from search, not from OLTP listing list
- Search response shape should stay as close as possible to the current paginated browse expectations to minimize frontend contract churn

### Must NOT Have
- No seller/admin-only results in public search
- No public `status` filter in MVP
- No unsupported filters like `bedroom_count` until they are truly indexed and documented
- No frontend-only fake sorting/filtering detached from backend search results
- No replacement of listing detail page read path yet unless explicitly required
- No zero-downtime alias-swap reindexing in MVP
- No faceting, autocomplete, or map bounding-box search in MVP

## Execution Strategy

## TODOs

- [x] T1. Finalize backend search-read contract and DTOs

  **What to do**: Formalize the backend search request/response DTOs around the existing `backend/search-read-contract.md`, making sure they only expose fields that the canonical search document actually contains. Keep the response envelope and pagination shape as close as possible to the current public listings browse expectations, and do not add a public `status` query parameter. Add a small internal search query parser/validator for query params if needed.

- [x] T2. Implement `GET /api/search/listings`

  **What to do**: Add a backend search handler/service/repository slice that queries Elasticsearch, validates query params, applies public-safe visibility semantics, and returns the standard response envelope. Keep the implementation search-index-backed and do not silently fall back to OLTP listing reads. Default to `relevance` only when `q` is present; otherwise sort by `newest`.

- [x] T3. Migrate public listings browse UI to the search endpoint

  **What to do**: Update `frontend/features/listings/server/get-listings.ts` and the public browse page to consume the new search endpoint and query surface. Remove the current dummy filler behavior and make the page reflect real search totals, pagination, and server-driven filters. Keep listing detail reads on the current slug/id path for this phase.

- [x] T4. Add buyer-facing search UX + verification

  **What to do**: Tighten browse UX around empty, invalid-filter, and sort states. Add backend tests for the search endpoint and frontend tests/E2E for public search browsing, filter application, and safe degradation on invalid input.

- [x] T5. Sync docs and contracts

  **What to do**: Update Postman, root/backend/frontend AGENTS/docs where the public listings read path and search-read contract become current reality instead of future intent.

## Final Verification Wave
- [x] F1. Plan Compliance Audit - oracle
- [x] F2. Code Quality Review - reviewer
- [x] F3. Runtime/Test Review - tester
- [x] F4. Scope Fidelity Check - deep

## Success Criteria
- Public browse uses Elasticsearch-backed search results.
- Search contract is narrow, explicit, and matches actual indexed fields.
- Frontend public discovery no longer depends on OLTP list filtering for the main browse experience.

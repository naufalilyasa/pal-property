# Buyer Saved Listings

## Context
- Next feature after seller workspace + public browse/detail
- Scope: first buyer-side post-discovery action with minimal UI drift
- Test strategy: tests-after
- Guardrails: explicit resource endpoints, idempotent save/delete, `page`/`limit` pagination, optimistic UI later, no public listing API contract changes

## Execution Order
1. schema/contracts
2. backend repository/service behavior
3. protected HTTP endpoints
4. frontend API helpers + SSR wrappers
5. reusable optimistic save button
6. public browse/detail wiring
7. protected saved listings page + full-stack verification
8. final review wave

## Tasks
- [x] 1. Add saved-listings schema and backend contracts
- [x] 2. Implement saved-listing repository and service behavior
- [x] 3. Expose protected saved-listing HTTP endpoints
- [x] 4. Add frontend saved-listing API helpers and SSR wrappers
- [x] 5. Build a reusable optimistic SaveListingButton component
- [x] 6. Wire saved-state into public browse cards and listing detail
- [x] 7. Add the protected saved listings page and full-stack verification

## Final Verification Wave
- [x] F1. Plan Compliance Audit — oracle
- [x] F2. Code Quality Review — unspecified-high
- [x] F3. Real Manual QA — unspecified-high (+ playwright if UI)
- [x] F4. Scope Fidelity Check — deep

## Endpoint Contract Targets
- `GET /api/me/saved-listings?page=&limit=`
- `GET /api/me/saved-listings/contains?listing_ids=<uuid,uuid>`
- `POST /api/me/saved-listings`
- `DELETE /api/me/saved-listings/:listingId`

## Success Criteria
- Authenticated user can save from browse card/detail CTA
- Unauthenticated save redirects to `/login`
- `/saved-listings` is protected and sorted newest-saved first
- Repeat save/delete are idempotent
- Failed optimistic mutation rolls state back

# Decisions

- Chose the smallest two-layer test stack: Vitest for component coverage and Playwright for browser smoke coverage; skipped Storybook/Cypress and broader harness setup to keep T1 foundation-only.
- Replaced create-next-app branding with a neutral seller workspace shell that can expand into later dashboard routes without pre-committing navigation, auth, or listing management flows.
- Kept the auth foundation minimal and app-specific: one credentialed `apiRequest` helper, one shared envelope normalizer, and a `/auth/me` bootstrap that returns a non-throwing `unauthenticated` state for future dashboard guards.
- Split generic session bootstrap from the Next.js `cookies()` reader so unit tests can exercise auth behavior without importing server-only request APIs.

- Built `/dashboard` as a tiny nested route shell: `app/dashboard/layout.tsx` owns the reusable seller frame and `app/dashboard/page.tsx` owns session-aware data states, so later create/edit routes can reuse the shell without duplicating listing fetch logic.
- Reused the T2 foundation instead of fetching inside JSX by adding one focused `getSellerListings()` helper plus a shared server cookie-header helper; page components only orchestrate session/listing states and presentation.
- Kept seller create/edit under `app/dashboard/listings/new` and `app/dashboard/listings/[listingId]/edit`, with one reusable client form component plus a focused API helper module, so both routes share category loading, edit hydration, submit state, and backend error formatting without adding a query or form library.
- Chose to submit the full canonical listing payload for both create and edit flows, normalizing empty optional strings to `null` and numeric specification fields to integers, because the backend DTOs are the source of truth and the update endpoint accepts partial PUT semantics without requiring a separate frontend-only diff layer.
- Added listing-image helpers beside the existing listing form API module rather than creating a second client stack, so upload/delete/set-primary/reorder reuse the same envelope normalization, credentialed fetch behavior, and listing hydration types.
- Kept image ordering non-optimistic with explicit `Move earlier` and `Move later` controls; every reorder waits for the backend `ordered_image_ids` response to rehydrate the gallery, which avoids frontend/server drift and stays within the no-drag-drop scope.
- Expanded T6 verification depth by keeping the same Vitest/Playwright stack but increasing contract assertions: unit tests now validate listing-form normalization helpers, and e2e tests assert real request payloads for create/update/image flows instead of route-load visibility alone.
- Synced guidance by updating `frontend/AGENTS.md` and root `AGENTS.md` to reflect accepted seller frontend implementation details, test stack, and cookie-session behavior; left `backend/AGENTS.md` unchanged because its frontend assumptions were not stale.
- Moved seller auth enforcement into `app/dashboard/layout.tsx` and redirect unauthenticated requests to `/`, so every dashboard route inherits one explicit server-side gate instead of per-page guest fallbacks.
- Switched edit-mode bootstrap to server-provided seller-owned listing data from `/auth/me/listings`, leaving the public listing read path out of frontend edit hydration to avoid view-count side effects.

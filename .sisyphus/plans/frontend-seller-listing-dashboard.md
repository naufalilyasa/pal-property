# Frontend Seller Listing Dashboard

## TL;DR
> **Summary**: Build the first real frontend product surface for sellers: authenticated dashboard access, seller listings index, create/edit listing form, category selection, and full image management against the existing backend APIs.
> **Deliverables**:
> - frontend test harness setup
> - cookie-auth/session bootstrap and API client layer
> - seller dashboard routes and listings index
> - create/edit listing form with category selection
> - listing image upload/delete/set-primary/reorder UI
> - seller-flow docs and verification assets
> **Effort**: Large
> **Parallel**: YES - 2 waves
> **Critical Path**: T1 test/app foundation -> T2 auth/API layer -> T3 dashboard shell -> T4 listing form -> T5 image manager -> T6/T7 verification and docs

## Context
### Original Request
Plan the next feature after Cloudinary listing images. Recommended default was a seller-only frontend listing dashboard with create/edit flow and image management.

### Interview Summary
- User accepted the recommended next feature: seller dashboard plus create/edit listing flow.
- Scope is seller-only; no public listing detail/discovery screens in this plan.
- Frontend test tooling must be added now rather than deferred.
- Backend contracts, cookie-auth behavior, and listing/category/image endpoints already exist and must be reused.

### Metis Review (gaps addressed)
- Added strong scope boundaries to prevent public marketplace/admin/design-system overbuild.
- Anchored form, auth, and image behavior to backend DTO/route authority instead of inventing frontend-first contracts.
- Added explicit QA requirements for auth bootstrap, ordered image projection, and frontend test harness execution.
- Preserved `/dashboard` as the seller landing route unless backend OAuth redirect behavior is intentionally changed in implementation.

## Work Objectives
### Core Objective
Create the first usable seller frontend so an authenticated user can land on `/dashboard`, view their own listings, create a listing, edit a listing, and manage listing images end-to-end using the existing backend.

### Deliverables
- Frontend test stack for one unit/component layer and one end-to-end seller flow layer.
- Shared frontend API client/response normalization with credentialed requests.
- Seller dashboard shell and route protection.
- Seller listings index using `/auth/me/listings`.
- Create/edit listing forms using the backend DTO contract.
- Category fetch/select flow using public category endpoints.
- Image manager for upload, delete, set-primary, and reorder.
- Minimal docs/update guidance inside frontend/backend AGENTS if implementation changes conventions.

### Definition of Done (verifiable conditions with commands)
- `cd frontend && npm run lint`
- `cd frontend && npm run build`
- `cd frontend && npm test`
- `cd frontend && npm run test:e2e`
- Seller can reach `/dashboard` only when authenticated.
- Seller listings index renders data from `/auth/me/listings`.
- Create/edit listing flow can submit valid payloads and surface backend validation errors.
- Image manager supports upload, delete, primary, and reorder against backend routes.

### Must Have
- Cookie-auth session bootstrap via `/auth/me`; no client token storage.
- Seller route namespace rooted at `/dashboard`.
- Frontend response normalization for backend `{ success, message, data, trace_id }` envelope.
- Listing form fields derived from backend listing create/update DTOs.
- Category selector driven from backend category reads.
- Image management supports the full accepted backend feature set.
- Tests added from scratch and included in the plan, not deferred.

### Must NOT Have (guardrails, AI slop patterns, scope boundaries)
- No public marketplace pages or search/discovery work.
- No admin category management UI.
- No localStorage/sessionStorage token auth.
- No broad design-system or component-library buildout beyond immediate reuse.
- No drag-and-drop extras, client image editing, or offline drafts/autosave.
- No backend contract rewrites unless implementation uncovers a true blocker.

## Verification Strategy
> ZERO HUMAN INTERVENTION - all verification is agent-executed.
- Test decision: add frontend test tooling now
- Unit/component layer: frontend test runner around route guards, API client normalization, and form/image UI logic
- E2E layer: seller dashboard bootstrap, create/edit form flow, and image management journey
- Backend contract checks: use existing API routes with credentialed requests and deterministic fixtures
- Evidence: `.sisyphus/evidence/task-{N}-{slug}.{ext}`

## Execution Strategy
### Parallel Execution Waves
> Build foundation before page work, then split page UI and verification/documentation.

Wave 1: T1 test/tooling foundation, T2 auth+API client foundation, T3 dashboard shell and seller index
Wave 2: T4 listing form, T5 image manager, T6 frontend tests/e2e, T7 docs + verification sync

### Dependency Matrix (full, all tasks)
- T1 blocks T6
- T2 blocks T3, T4, T5, T6
- T3 blocks T4, T5, T6
- T4 blocks T5, T6, T7
- T5 blocks T6, T7
- T6 and T7 depend on T4 and T5 being stable

### Agent Dispatch Summary (wave -> task count -> categories)
- Wave 1 -> 3 tasks -> `implementation`, `testing`
- Wave 2 -> 4 tasks -> `implementation`, `testing`, `writing`

## TODOs
> Implementation + Test = ONE task. Never separate.
> Every task includes explicit QA scenarios.

<!-- TASKS -->

- [x] T1. Add frontend test and app foundation

  **What to do**: Install and configure the minimum frontend test stack required by this feature: one unit/component test runner and one end-to-end runner, both wired into `frontend/package.json`. Replace create-next-app placeholder app metadata/content with a seller-oriented app shell baseline that still stays small. Add any minimal config files needed for test execution, route aliases, and environment-safe local execution.
  **Must NOT do**: Do not add Storybook, visual regression tooling, Cypress, or a broad component library. Do not implement seller pages yet.

  **Recommended Agent Profile**:
  - Category: `implementation` - Reason: frontend foundation and tooling setup
  - Skills: `[]` - no extra skill required
  - Omitted: `[]` - no omission needed

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: T6 | Blocked By: none

  **References**:
  - Pattern: `frontend/package.json` - current script and dependency baseline
  - Pattern: `frontend/app/layout.tsx` - current app shell starting point
  - Pattern: `frontend/app/page.tsx` - scaffold content to replace
  - Pattern: `frontend/app/globals.css` - current styling baseline
  - Pattern: `frontend/AGENTS.md` - current scaffold guidance

  **Acceptance Criteria**:
  - [x] `frontend/package.json` contains scripts for lint, build, unit/component tests, and end-to-end tests
  - [x] Frontend scaffold landing page/content is replaced with seller-app baseline content, not create-next-app branding
  - [x] `cd frontend && npm run lint`, `cd frontend && npm run build`, and `cd frontend && npm test` all execute successfully

  **QA Scenarios**:
  ```bash
  Scenario: Frontend foundation commands run cleanly
    Tool: Bash
    Steps: cd frontend && npm run lint && npm run build && npm test
    Expected: all commands exit 0
    Evidence: .sisyphus/evidence/task-T1-frontend-foundation.txt

  Scenario: Scaffold branding is removed
    Tool: Bash
    Steps: python3 - <<'PY'
from pathlib import Path
text = Path('frontend/app/page.tsx').read_text() + '\n' + Path('frontend/app/layout.tsx').read_text()
assert 'Create Next App' not in text
assert 'Vercel' not in text
print('ok')
PY
    Expected: script prints `ok`
    Evidence: .sisyphus/evidence/task-T1-no-scaffold-branding.txt
  ```

  **Commit**: YES | Message: `test(frontend): add seller app test foundation` | Files: `frontend/package.json`, frontend test config files, `frontend/app/layout.tsx`, `frontend/app/page.tsx`, `frontend/app/globals.css`

- [x] T2. Build cookie-auth session bootstrap and API client layer

  **What to do**: Add a thin frontend API layer that normalizes the backend envelope and always sends credentialed requests for protected routes. Add session bootstrap around `/auth/me` and shared request helpers for `/auth/me/listings`, `/api/listings`, `/api/categories`, and listing image mutations. Define how unauthenticated access to `/dashboard` is handled and keep it aligned with the backend cookie-auth model.
  **Must NOT do**: Do not introduce bearer-token local storage auth or an overengineered SDK. Do not duplicate envelope parsing in page components.

  **Recommended Agent Profile**:
  - Category: `implementation` - Reason: auth/session and API integration foundation
  - Skills: `[]` - no extra skill required
  - Omitted: `[]` - no omission needed

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: T3, T4, T5, T6 | Blocked By: none

  **References**:
  - Pattern: `backend/internal/router/router.go` - route authority for auth/listing/category/image endpoints
  - Pattern: `backend/pkg/utils/response.go` - backend response envelope shape
  - Pattern: `backend/postman_collection.json` - payload and endpoint examples
  - Pattern: `frontend/AGENTS.md` - cookie-auth expectations

  **Acceptance Criteria**:
  - [x] Protected frontend requests use credentialed fetch behavior
  - [x] `/dashboard` session bootstrap depends on `/auth/me`, not local token state
  - [x] API helpers normalize success/error responses from the backend envelope

  **QA Scenarios**:
  ```bash
  Scenario: Unit tests prove response normalization and credentialed requests
    Tool: Bash
    Steps: cd frontend && npm test
    Expected: tests covering auth/session or API client pass
    Evidence: .sisyphus/evidence/task-T2-api-client-tests.txt

  Scenario: Unauthenticated dashboard access is handled predictably
    Tool: Bash
    Steps: cd frontend && npm run test:e2e
    Expected: e2e output includes an unauthenticated dashboard case that redirects or blocks access as planned
    Evidence: .sisyphus/evidence/task-T2-dashboard-auth.txt
  ```

  **Commit**: YES | Message: `feat(frontend): add seller session and api client` | Files: new frontend API/session files and any minimal dashboard route guard wiring

- [x] T3. Add seller dashboard shell and listings index

  **What to do**: Create the `/dashboard` seller shell and the seller listings index powered by `/auth/me/listings`. Include loading, empty, and error states. Render image thumbnails, listing status, category, and key metadata from the current backend response. Keep the shell intentionally small but reusable for create/edit routes.
  **Must NOT do**: Do not include public marketplace routes or analytics. Do not hide backend errors silently.

  **Recommended Agent Profile**:
  - Category: `implementation` - Reason: first product route and seller index UI
  - Skills: `[]` - no extra skill required
  - Omitted: `[]` - no omission needed

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: T4, T5, T6 | Blocked By: T2

  **References**:
  - Pattern: `backend/internal/router/router.go` - `GET /auth/me/listings`
  - Pattern: `backend/internal/dto/response/listing_response.go` - listing response shape including images
  - Pattern: `backend/postman_collection.json` - seller listings response examples
  - Pattern: `frontend/app/layout.tsx` - seller shell insertion point

  **Acceptance Criteria**:
  - [x] `/dashboard` renders seller listings from `/auth/me/listings`
  - [x] Empty-state and error-state UX exists for seller listings
  - [x] Listings index shows image, title, category, status, and price metadata from backend responses

  **QA Scenarios**:
  ```bash
  Scenario: Seller listings index e2e flow works
    Tool: Bash
    Steps: cd frontend && npm run test:e2e
    Expected: e2e output includes successful authenticated dashboard/listings rendering
    Evidence: .sisyphus/evidence/task-T3-dashboard-listings.txt

  Scenario: Listings index build path is stable
    Tool: Bash
    Steps: cd frontend && npm run build
    Expected: build exits 0 with dashboard route included
    Evidence: .sisyphus/evidence/task-T3-dashboard-build.txt
  ```

  **Commit**: YES | Message: `feat(frontend): add seller dashboard listings index` | Files: dashboard route components and related frontend helpers

- [x] T4. Add create/edit listing form with category selection

  **What to do**: Build seller create and edit routes under the dashboard shell using the backend listing DTO contract as the form source of truth. Include all MVP fields supported by the backend create/update DTOs, backend-aligned validation messaging, category fetching/selection, submit/loading/error states, and edit-mode hydration from existing listing reads.
  **Must NOT do**: Do not invent frontend-only listing fields, and do not collapse backend validation into generic error toasts.

  **Recommended Agent Profile**:
  - Category: `implementation` - Reason: main seller form workflow
  - Skills: `[]` - no extra skill required
  - Omitted: `[]` - no omission needed

  **Parallelization**: Can Parallel: NO | Wave 2 | Blocks: T5, T6, T7 | Blocked By: T2, T3

  **References**:
  - Pattern: `backend/internal/dto/request/listing_request.go` - authoritative create/update field contract
  - Pattern: `backend/internal/dto/response/listing_response.go` - edit-mode listing projection shape
  - Pattern: `backend/postman_collection.json` - create/update payload examples
  - Pattern: `backend/internal/router/router.go` - seller listing create/update/get routes

  **Acceptance Criteria**:
  - [x] Create route can submit a valid backend-aligned listing payload
  - [x] Edit route loads an existing listing and submits backend-aligned partial/full updates as planned
  - [x] Category options are fetched from backend categories and selectable in the form

  **QA Scenarios**:
  ```bash
  Scenario: Seller can create a listing through the UI flow
    Tool: Bash
    Steps: cd frontend && npm run test:e2e
    Expected: e2e output includes a successful create listing journey with backend-aligned fields
    Evidence: .sisyphus/evidence/task-T4-create-listing.txt

  Scenario: Seller can edit an existing listing through the UI flow
    Tool: Bash
    Steps: cd frontend && npm run test:e2e
    Expected: e2e output includes an edit listing journey with prefilled values and successful save
    Evidence: .sisyphus/evidence/task-T4-edit-listing.txt
  ```

  **Commit**: YES | Message: `feat(frontend): add seller listing form flow` | Files: create/edit route components, form helpers, category data hooks/helpers

- [x] T5. Add listing image management UI to create/edit flow

  **What to do**: Build image-management UI that works with the accepted backend image API surface: upload, delete, set-primary, and reorder. Support both create/edit listing contexts as planned, keep image ordering visibly deterministic, and surface backend constraint errors such as invalid file or image-limit issues clearly. Use the backend response shape as the UI source of truth after each mutation.
  **Must NOT do**: Do not add drag-and-drop extras, client-side image editing, or frontend-only optimistic reorder models that can drift from server truth.

  **Recommended Agent Profile**:
  - Category: `implementation` - Reason: image management UI and API mutation integration
  - Skills: `[]` - no extra skill required
  - Omitted: `[]` - no omission needed

  **Parallelization**: Can Parallel: NO | Wave 2 | Blocks: T6, T7 | Blocked By: T2, T3, T4

  **References**:
  - Pattern: `backend/internal/router/router.go` - image upload/delete/primary/reorder routes
  - Pattern: `backend/internal/dto/response/listing_response.go` - image response projection fields
  - Pattern: `backend/postman_collection.json` - image upload/delete/primary/reorder examples
  - Pattern: `backend/internal/service/listing_service.go` - backend constraints to surface accurately (max 10 images, primary, ordering)

  **Acceptance Criteria**:
  - [x] Seller can upload image files and see refreshed server-driven image state
  - [x] Seller can delete images, set primary, and reorder images from the UI
  - [x] UI reflects ordered images from backend responses after every mutation

  **QA Scenarios**:
  ```bash
  Scenario: Seller image manager happy path works end-to-end
    Tool: Bash
    Steps: cd frontend && npm run test:e2e
    Expected: e2e output includes successful image upload, primary toggle, reorder, and delete flow
    Evidence: .sisyphus/evidence/task-T5-image-manager.txt

  Scenario: Image constraint errors are surfaced clearly
    Tool: Bash
    Steps: cd frontend && npm run test:e2e
    Expected: e2e output includes at least one invalid-file or over-limit image case with visible error handling
    Evidence: .sisyphus/evidence/task-T5-image-errors.txt
  ```

  **Commit**: YES | Message: `feat(frontend): add seller listing image manager` | Files: image manager components/hooks/mutation helpers

- [x] T6. Add frontend unit/component and end-to-end verification coverage

  **What to do**: Expand the newly added frontend test stack with targeted automated coverage for session/API normalization, dashboard/listing form behavior, and image-management UI logic. Add one end-to-end seller flow that exercises the real route stack and one focused unit/component layer for deterministic logic. Make the tests reliable against the current repo environment and document any backend fixture assumptions they need.
  **Must NOT do**: Do not leave the project with only placeholder smoke tests. Do not introduce flaky network-dependent tests without deterministic setup.

  **Recommended Agent Profile**:
  - Category: `testing` - Reason: verification depth and test hardening
  - Skills: `[]` - no extra skill required
  - Omitted: `[]` - no omission needed

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: none | Blocked By: T1, T2, T3, T4, T5

  **References**:
  - Pattern: `frontend/package.json` - final test scripts to satisfy
  - Pattern: `backend/postman_collection.json` - deterministic API payload fixtures
  - Pattern: `backend/internal/router/router.go` - route assumptions for seller flows
  - Pattern: `frontend/AGENTS.md` - frontend scaffold and cookie-auth expectations

  **Acceptance Criteria**:
  - [x] Unit/component tests cover API normalization and at least one seller listing UI behavior
  - [x] End-to-end tests cover authenticated dashboard bootstrap and create/edit/image flow
  - [x] `npm test` and `npm run test:e2e` are reliable, documented, and pass in the project environment

  **QA Scenarios**:
  ```bash
  Scenario: Unit/component suite passes
    Tool: Bash
    Steps: cd frontend && npm test
    Expected: test suite exits 0 and includes seller-flow related coverage
    Evidence: .sisyphus/evidence/task-T6-unit-tests.txt

  Scenario: End-to-end seller flow suite passes
    Tool: Bash
    Steps: cd frontend && npm run test:e2e
    Expected: test suite exits 0 and covers dashboard + listing create/edit/image flow
    Evidence: .sisyphus/evidence/task-T6-e2e-tests.txt
  ```

  **Commit**: YES | Message: `test(frontend): cover seller listing flows` | Files: frontend test files and any minimal test helpers/config updates

- [x] T7. Sync docs and implementation guidance for the new frontend workflow

  **What to do**: Update the relevant AGENTS guidance and any frontend implementation notes that became inaccurate once the seller dashboard exists. Keep the docs bounded to the new seller flow, test stack, cookie-auth assumptions, and any new frontend conventions introduced by implementation. Include only files whose guidance genuinely changed.
  **Must NOT do**: Do not rewrite unrelated product docs or broaden this into marketing/public-site documentation.

  **Recommended Agent Profile**:
  - Category: `writing` - Reason: targeted implementation/documentation alignment
  - Skills: `[]` - no extra skill required
  - Omitted: `[]` - no omission needed

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: none | Blocked By: T4, T5, T6

  **References**:
  - Pattern: `frontend/AGENTS.md` - current frontend scaffold guidance to update
  - Pattern: `AGENTS.md` - root project guidance if frontend is no longer scaffold-only in the seller area
  - Pattern: `backend/AGENTS.md` - update only if frontend assumptions are referenced there

  **Acceptance Criteria**:
  - [x] Frontend guidance no longer describes seller-facing UI as scaffold-only if the flow now exists
  - [x] Docs mention the chosen frontend test stack and cookie-auth route assumptions
  - [x] Any changed AGENTS/docs remain accurate to the implemented frontend structure

  **QA Scenarios**:
  ```bash
  Scenario: Docs mention the real frontend workflow and test stack
    Tool: Bash
    Steps: python3 - <<'PY'
from pathlib import Path
paths = ['frontend/AGENTS.md', 'AGENTS.md']
for path in paths:
    if Path(path).exists():
        text = Path(path).read_text()
        assert 'seller' in text.lower() or 'dashboard' in text.lower() or 'test' in text.lower()
print('ok')
PY
    Expected: script prints `ok`
    Evidence: .sisyphus/evidence/task-T7-doc-sync.txt

  Scenario: Final frontend commands still pass after docs sync
    Tool: Bash
    Steps: cd frontend && npm run lint && npm run build
    Expected: both commands exit 0
    Evidence: .sisyphus/evidence/task-T7-final-frontend-build.txt
  ```

  **Commit**: YES | Message: `docs(frontend): document seller listing workflow` | Files: changed AGENTS/docs files only

## Final Verification Wave (4 parallel agents, ALL must APPROVE)
- [x] F1. Plan Compliance Audit - oracle
- [x] F2. Code Quality Review - reviewer
- [x] F3. Real Frontend QA - tester
- [x] F4. Scope Fidelity Check - deep

## Commit Strategy
- Commit 1: frontend test/app foundation
- Commit 2: auth/session + API client + dashboard shell
- Commit 3: listing form + category selection + image manager
- Commit 4: tests, docs, and verification alignment

## Success Criteria
- Seller-only frontend flow exists and is no longer scaffold content.
- Auth/session handling is cookie-based and route-protected.
- Form and image UX match backend contract behavior and constraints.
- Frontend tests and E2E checks run successfully without manual QA.

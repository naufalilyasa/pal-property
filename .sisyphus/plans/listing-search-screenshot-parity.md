# Listing Search Screenshot Parity

## TL;DR
> **Summary**: Push `frontend/app/(public)/listings/page.tsx` much closer to the provided brokerage-style screenshot by tightening the top search chrome, hardening the map-left/results-right shell, and making the results rhythm flatter and denser, all without changing the backend search contract.
> **Deliverables**:
> - compact top-of-page search shell with explicit page-local chrome
> - stable desktop map-left / results-right layout with defined mobile fallback
> - denser listing cards and result toolbar closer to the screenshot rhythm
> - shell-only `Map/List` and `Sort` controls with no backend contract changes
> - public-route automated coverage for query/default behavior and responsive rendering
> **Effort**: Medium
> **Parallel**: YES - 3 waves
> **Critical Path**: 1 -> 2 -> 4 -> 6 -> 7 -> F1-F4

## Context
### Original Request
The user wants the public property listings page to look substantially closer to the provided screenshot / `https://findrealestate.com/search`, not pixel-identical but recognizably similar in structure and visual rhythm.

### Interview Summary
- Scope confirmed: visual parity + interaction shell only.
- Preserve current server-rendered route ownership and backend query params.
- Do not expand backend/API contracts in this iteration.
- Favor a flatter brokerage-style search workspace over a premium catalog/hero layout.

### Metis Review (gaps addressed)
- Resolved unanswered tradeoffs by defaulting top-site chrome to `/listings` only, not a new shared public shell.
- Resolved `Map/List` and `Sort` as presentational shell controls for this iteration; they must not imply unsupported backend features.
- Added explicit guardrails for query-state fidelity, especially status default drift and reset behavior.
- Added required route-level Vitest plus Playwright coverage because no dedicated public `/listings` tests were surfaced.
- Added explicit mobile behavior instead of vague “responsive parity”.

## Work Objectives
### Core Objective
Deliver a decision-complete redesign of the public `/listings` route so it reads visually like the provided brokerage screenshot: compact search band, persistent map-left shell, dense result grid, lightweight result controls, and a flatter listing-card presentation.

### Deliverables
- Updated `frontend/app/(public)/listings/page.tsx` layout and page-local chrome.
- Updated `frontend/features/listings/components/listing-filters.tsx` toolbar/filter shell.
- Updated `frontend/features/listings/components/listing-card.tsx` result-card rhythm.
- New public-route tests for query/default behavior.
- New Playwright spec for desktop/mobile `/listings` rendering.

### Definition of Done (verifiable conditions with commands)
- `cd frontend && npx vitest run "app/(public)/listings/page.test.tsx"` exits `0`.
- `cd frontend && npx playwright test "e2e/public-listings.spec.ts" --project=chromium` exits `0`.
- `cd frontend && npm test` exits `0`.
- `cd frontend && npm run build` exits `0`.
- `/listings` preserves only the existing backend query params: `page`, `limit`, `city`, `category_id`, `price_min`, `price_max`, `status`.

### Must Have
- Route stays server-rendered and continues to call `getListings()`.
- Desktop layout keeps a visible left map panel and a right result grid.
- Mobile layout keeps the map below the toolbar and above results; do not hide it completely.
- `Map/List` and `Sort` are clearly shell controls unless fully implemented without backend changes.
- Filter defaults reflect actual resolved route params exactly.
- Clear/reset behavior removes stale query state instead of merely resetting form fields in-place.

### Must NOT Have (guardrails, AI slop patterns, scope boundaries)
- Must NOT add backend search params like `sort`, `view`, map bounds, polygons, or saved search IDs.
- Must NOT turn `/listings` into a client-owned data page or move primary data fetching into TanStack Query.
- Must NOT introduce fake MLS metadata not present in `ListingRecord`.
- Must NOT create a shared global public-shell refactor beyond the `/listings` page.
- Must NOT leave visually interactive controls in a misleading state; unsupported controls must be explicitly shell-only.

## Verification Strategy
> ZERO HUMAN INTERVENTION - all verification is agent-executed.
- Test decision: tests-after with Vitest + Playwright + frontend build.
- QA policy: every task includes agent-executed scenarios.
- Evidence: `.sisyphus/evidence/task-{N}-{slug}.{ext}`

## Execution Strategy
### Parallel Execution Waves
> Target: 5-8 tasks per wave. Shared dependencies are extracted first.

Wave 1: query-shell foundations and layout contract
- Task 1: lock query-state defaults/reset behavior
- Task 2: refactor page-level shell and responsive map/list composition
- Task 3: define route-level test harness for `/listings`

Wave 2: visual parity components
- Task 4: densify top toolbar/filter band
- Task 5: flatten result toolbar, pagination, and footer block
- Task 6: tighten listing-card rhythm to screenshot style

Wave 3: QA and polish
- Task 7: add Playwright responsive coverage
- Task 8: run final frontend verification and fix regressions

### Dependency Matrix (full, all tasks)
- 1 blocks 2, 4, 5, 7, 8
- 2 blocks 5, 7, 8
- 3 blocks 8
- 4 blocks 7, 8
- 5 blocks 7, 8
- 6 blocks 7, 8
- 7 blocks 8
- 8 unblocks final verification wave

### Agent Dispatch Summary
- Wave 1 -> 3 tasks -> implementation/testing
- Wave 2 -> 3 tasks -> visual-engineering/implementation
- Wave 3 -> 2 tasks -> testing/review

## TODOs
> Implementation + Test = ONE task. Never separate.

- [x] 1. Fix query-state fidelity for the `/listings` shell

  **What to do**: Update the `/listings` route and filter form so every displayed default comes from resolved `searchParams`. Replace the current DOM-only reset behavior with a route-safe clear flow that returns the user to `/listings` without stale query params. Preserve `limit` handling and keep the backend query set unchanged.
  **Must NOT do**: Do not add new backend params, and do not leave any visual default that differs from the actual server fetch inputs.

  **Recommended Agent Profile**:
  - Category: `implementation` - Reason: query-flow correctness across SSR route + client filter form
  - Skills: `[]` - no special skill required
  - Omitted: `['playwright']` - UI automation belongs in later QA tasks

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: 2, 4, 5, 7, 8 | Blocked By: none

  **References**:
  - Pattern: `frontend/app/(public)/listings/page.tsx` - route-owned query parsing and SSR fetch
  - Pattern: `frontend/features/listings/server/get-listings.ts` - canonical backend query contract
  - Pattern: `frontend/features/listings/components/listing-filters.tsx` - current toolbar shell and default/reset logic
  - API/Type: `frontend/lib/api/seller-listings.ts` - page payload shape used by `getListings()` return value

  **Acceptance Criteria**:
  - [ ] `frontend/app/(public)/listings/page.tsx` continues to pass only `page`, `limit`, `city`, `category_id`, `price_min`, `price_max`, and `status` to `getListings()`.
  - [ ] Filter defaults mirror resolved route params exactly on first render.
  - [ ] Clear/reset behavior routes to `/listings` without stale search params.

  **QA Scenarios**:
  ```
  Scenario: Query defaults render correctly
    Tool: Vitest
    Steps: Add/extend `frontend/app/(public)/listings/page.test.tsx`; mock `getListings()`; render the page with `city=Jakarta&status=active&limit=12`; assert toolbar defaults reflect those exact values.
    Expected: Test passes and confirms SSR query fidelity.
    Evidence: .sisyphus/evidence/task-1-query-state.txt

  Scenario: Clear flow removes stale params
    Tool: Playwright
    Steps: Open `/listings?city=Jakarta&status=active&price_min=500000000`; activate the clear action.
    Expected: URL becomes `/listings` (or `/listings?limit=12` only if the plan deliberately preserves limit) and no stale filter values remain visible.
    Evidence: .sisyphus/evidence/task-1-query-state-clear.png
  ```

  **Commit**: YES | Message: `fix(frontend): align listings query state with search shell` | Files: `frontend/app/(public)/listings/page.tsx`, `frontend/features/listings/components/listing-filters.tsx`, `frontend/app/(public)/listings/page.test.tsx`

- [x] 2. Rebuild the `/listings` page shell around screenshot parity

  **What to do**: Recompose `frontend/app/(public)/listings/page.tsx` into a flatter search workspace matching the screenshot’s structure: compact top page chrome, toolbar immediately under it, persistent left map panel, right-side result column, narrow spacing, and reduced “hero” emphasis. Keep the map visible on desktop and stacked below the toolbar on mobile.
  **Must NOT do**: Do not extract a new site-wide header/layout, and do not hide the map entirely on mobile.

  **Recommended Agent Profile**:
  - Category: `visual-engineering` - Reason: layout composition and responsive fidelity
  - Skills: `[]` - existing Tailwind patterns are sufficient
  - Omitted: `['frontend-ui-ux']` - not required unless the executor gets stuck

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: 5, 7, 8 | Blocked By: 1

  **References**:
  - Pattern: `frontend/app/(public)/listings/page.tsx` - current shell owner
  - Pattern: `frontend/app/(public)/listings/[slug]/page.tsx` - current public-route visual language to stay adjacent to
  - Pattern: `frontend/app/globals.css` - palette and atmosphere variables already used across the public app
  - External: `https://findrealestate.com/search` - target structure (compact top band, map-left, dense right results)

  **Acceptance Criteria**:
  - [ ] Desktop `/listings` renders a left map column and a right result column without horizontal overflow.
  - [ ] Mobile `/listings` stacks toolbar -> map -> results in a deterministic order.
  - [ ] The page no longer reads like a hero/catalog page; it reads like a search workspace.

  **QA Scenarios**:
  ```
  Scenario: Desktop shell matches map-left/list-right structure
    Tool: Playwright
    Steps: Open `/listings` at 1440x1200; assert a visible map region, a visible results header, and multiple `[data-testid="listing-card"]` when mocked or seeded data exists.
    Expected: Map is visibly left of results and the page has no horizontal overflow.
    Evidence: .sisyphus/evidence/task-2-desktop-shell.png

  Scenario: Mobile shell remains deterministic
    Tool: Playwright
    Steps: Open `/listings` at 390x844.
    Expected: Toolbar appears first, map appears before result cards, and the viewport can scroll vertically without horizontal clipping.
    Evidence: .sisyphus/evidence/task-2-mobile-shell.png
  ```

  **Commit**: YES | Message: `feat(frontend): reshape listings page into search workspace shell` | Files: `frontend/app/(public)/listings/page.tsx`

- [x] 3. Add route-level test coverage for the public listings page

  **What to do**: Introduce a dedicated test file for the public `/listings` route that mocks `getListings()` and verifies SSR route composition, query parsing, and data-testid coverage for filter/map/pagination regions.
  **Must NOT do**: Do not rely only on broad `npm test` smoke coverage.

  **Recommended Agent Profile**:
  - Category: `testing` - Reason: route-level SSR verification
  - Skills: `[]`
  - Omitted: `['playwright']` - route/component assertions belong in Vitest here

  **Parallelization**: Can Parallel: YES | Wave 1 | Blocks: 8 | Blocked By: none

  **References**:
  - Pattern: `frontend/app/page.test.tsx` - route-level render test structure
  - Pattern: `frontend/app/(dashboard)/dashboard/page.test.tsx` - mocked server helper route testing
  - Pattern: `frontend/app/(public)/listings/page.tsx` - selectors and shell regions to assert

  **Acceptance Criteria**:
  - [ ] `frontend/app/(public)/listings/page.test.tsx` exists and passes.
  - [ ] The test asserts route shell regions and mocked listings render.
  - [ ] The test locks query/default behavior introduced in Task 1.

  **QA Scenarios**:
  ```
  Scenario: Route renders listings shell with mocked data
    Tool: Vitest
    Steps: Mock `getListings()` to return 2 listings; render the page; assert `listing-filters`, `listing-pagination`, and 2 `listing-card` wrappers exist.
    Expected: Test passes and protects the SSR route shell.
    Evidence: .sisyphus/evidence/task-3-route-test.txt

  Scenario: Route handles empty state gracefully
    Tool: Vitest
    Steps: Mock `getListings()` to return `data: []`; render the page.
    Expected: Empty-state copy renders and no crash occurs.
    Evidence: .sisyphus/evidence/task-3-empty-state.txt
  ```

  **Commit**: YES | Message: `test(frontend): cover public listings route shell` | Files: `frontend/app/(public)/listings/page.test.tsx`

- [x] 4. Densify the top toolbar and filter band to MLS-style rhythm

  **What to do**: Compress `frontend/features/listings/components/listing-filters.tsx` into a flatter toolbar that feels closer to the screenshot: narrow controls, restrained labels, consistent heights, tighter gaps, and no oversized panel treatment. Keep it functional only for the existing filters.
  **Must NOT do**: Do not invent unsupported filters, and do not make the toolbar taller or more decorative than the screenshot.

  **Recommended Agent Profile**:
  - Category: `visual-engineering` - Reason: density, spacing, and visual parity work
  - Skills: `[]`
  - Omitted: `['playwright']`

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: 7, 8 | Blocked By: 1

  **References**:
  - Pattern: `frontend/features/listings/components/listing-filters.tsx` - current toolbar implementation
  - Pattern: `frontend/app/(public)/listings/page.tsx` - surrounding shell constraints and spacing
  - External: `https://findrealestate.com/search` - compact search/filter band rhythm

  **Acceptance Criteria**:
  - [ ] The toolbar fits in a flatter visual band on desktop without oversized stat cards.
  - [ ] All controls share a consistent compact height.
  - [ ] The toolbar remains keyboard-accessible and form-submittable.

  **QA Scenarios**:
  ```
  Scenario: Toolbar remains compact on desktop
    Tool: Playwright
    Steps: Open `/listings` at 1440x1200.
    Expected: Toolbar occupies one compact band above the map/results split and does not visually dominate the page.
    Evidence: .sisyphus/evidence/task-4-toolbar-desktop.png

  Scenario: Toolbar remains usable on mobile
    Tool: Playwright
    Steps: Open `/listings` at 390x844; tab or click through inputs/selects.
    Expected: Controls remain accessible, readable, and stack cleanly.
    Evidence: .sisyphus/evidence/task-4-toolbar-mobile.png
  ```

  **Commit**: YES | Message: `feat(frontend): tighten listings search toolbar density` | Files: `frontend/features/listings/components/listing-filters.tsx`

- [x] 5. Flatten the result toolbar, pagination, and footer block

  **What to do**: Tune the right-column result chrome in `frontend/app/(public)/listings/page.tsx` so it more closely matches the screenshot’s lightweight header (`results`, `sort`, view hint), minimal pagination, and dark footer/newsletter block. Keep the footer local to the page for this iteration.
  **Must NOT do**: Do not introduce global footer refactors or unsupported sorting behavior.

  **Recommended Agent Profile**:
  - Category: `visual-engineering` - Reason: right-column shell polish and parity
  - Skills: `[]`
  - Omitted: `['playwright']`

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: 7, 8 | Blocked By: 1, 2

  **References**:
  - Pattern: `frontend/app/(public)/listings/page.tsx` - result header, pagination, newsletter/footer block
  - External: `https://findrealestate.com/search` - result header density and footer structure

  **Acceptance Criteria**:
  - [ ] Result header reads as a lightweight search-results strip, not a feature panel.
  - [ ] Pagination remains visible and compact.
  - [ ] Footer block visually anchors the page without forcing global layout changes.

  **QA Scenarios**:
  ```
  Scenario: Result header stays lightweight
    Tool: Playwright
    Steps: Open `/listings` on desktop and inspect the result header region.
    Expected: Results count, sort shell, and pagination are visually lightweight and aligned with the screenshot direction.
    Evidence: .sisyphus/evidence/task-5-results-header.png

  Scenario: Footer remains page-local and visible
    Tool: Playwright
    Steps: Scroll to bottom of `/listings`.
    Expected: Dark footer/newsletter section appears only on this page route and remains readable on mobile and desktop.
    Evidence: .sisyphus/evidence/task-5-footer.png
  ```

  **Commit**: YES | Message: `feat(frontend): refine listings result chrome and footer` | Files: `frontend/app/(public)/listings/page.tsx`

- [x] 6. Tighten listing-card rhythm to the screenshot style

  **What to do**: Rework `frontend/features/listings/components/listing-card.tsx` so cards are flatter, denser, and more brokerage-listing-like: compact image block, restrained metadata, tight stats line, lightweight utility controls, and a less “editorial” presentation.
  **Must NOT do**: Do not add nonexistent data like agent names, coordinates, or days-on-market; do not make the cards taller than needed.

  **Recommended Agent Profile**:
  - Category: `visual-engineering` - Reason: card density and visual rhythm
  - Skills: `[]`
  - Omitted: `['playwright']`

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: 7, 8 | Blocked By: none

  **References**:
  - Pattern: `frontend/features/listings/components/listing-card.tsx` - current card component
  - API/Type: `frontend/lib/api/listing-form.ts` - `ListingRecord`, available fields and limitations
  - External: `https://findrealestate.com/search` - dense card rhythm and flatter property metadata

  **Acceptance Criteria**:
  - [ ] Cards render within a denser two-column desktop rhythm.
  - [ ] Card content uses only fields available in `ListingRecord`.
  - [ ] Utility actions remain clearly shell-level unless genuinely implemented.

  **QA Scenarios**:
  ```
  Scenario: Cards fit dense desktop rhythm
    Tool: Playwright
    Steps: Open `/listings` at 1440x1200 with multiple cards visible.
    Expected: Two columns remain visible without oversized card heights or excessive whitespace.
    Evidence: .sisyphus/evidence/task-6-card-density.png

  Scenario: Cards remain readable on mobile
    Tool: Playwright
    Steps: Open `/listings` at 390x844.
    Expected: Each card remains readable without clipped text or off-screen actions.
    Evidence: .sisyphus/evidence/task-6-card-mobile.png
  ```

  **Commit**: YES | Message: `feat(frontend): tighten public listing card rhythm` | Files: `frontend/features/listings/components/listing-card.tsx`

- [x] 7. Add Playwright coverage for desktop/mobile search-shell parity

  **What to do**: Create `frontend/e2e/public-listings.spec.ts` to validate the new `/listings` shell on desktop and mobile. The spec must verify shell regions, lack of horizontal overflow, deterministic map behavior, and visible listing-card rendering.
  **Must NOT do**: Do not make the test depend on exact pixels or live third-party map rendering internals.

  **Recommended Agent Profile**:
  - Category: `testing` - Reason: browser-level validation
  - Skills: `['playwright']` - needed for deterministic browser QA
  - Omitted: `[]`

  **Parallelization**: Can Parallel: NO | Wave 3 | Blocks: 8 | Blocked By: 1, 2, 4, 5, 6

  **References**:
  - Pattern: `frontend/playwright.config.ts` - existing E2E runner setup
  - Pattern: `frontend/e2e/` - existing browser coverage area
  - Pattern: `frontend/app/(public)/listings/page.tsx` - selectors and layout regions to assert

  **Acceptance Criteria**:
  - [ ] `frontend/e2e/public-listings.spec.ts` exists and passes under Chromium.
  - [ ] The test covers one desktop and one mobile viewport.
  - [ ] The test asserts visible `listing-filters`, `listing-pagination`, and at least one `listing-card` when data exists.

  **QA Scenarios**:
  ```
  Scenario: Desktop `/listings` shell
    Tool: Playwright
    Steps: Visit `/listings?city=Jakarta&status=active&limit=12` at desktop viewport; assert toolbar, map region, and result cards are visible.
    Expected: Search shell renders correctly with no horizontal overflow.
    Evidence: .sisyphus/evidence/task-7-desktop-playwright.png

  Scenario: Mobile `/listings` shell
    Tool: Playwright
    Steps: Visit `/listings` at mobile viewport.
    Expected: Toolbar, stacked map, and result cards appear in the defined order and remain scrollable.
    Evidence: .sisyphus/evidence/task-7-mobile-playwright.png
  ```

  **Commit**: YES | Message: `test(frontend): cover listings search shell in playwright` | Files: `frontend/e2e/public-listings.spec.ts`

- [x] 8. Run final frontend verification and fix regressions

  **What to do**: Run LSP diagnostics, the dedicated public-route Vitest file, the dedicated Playwright file, the full frontend test suite, and the production build. Fix any regressions introduced by the redesign while preserving the scope boundaries above.
  **Must NOT do**: Do not expand the scope into backend or unrelated dashboard changes.

  **Recommended Agent Profile**:
  - Category: `testing` - Reason: integrated verification and regression cleanup
  - Skills: `['playwright']` - needed for the E2E run
  - Omitted: `[]`

  **Parallelization**: Can Parallel: NO | Wave 3 | Blocks: none | Blocked By: 1, 2, 3, 4, 5, 6, 7

  **References**:
  - Pattern: `frontend/AGENTS.md` - command expectations and route conventions
  - Test: `frontend/app/(public)/listings/page.test.tsx` - route-level assertions
  - Test: `frontend/e2e/public-listings.spec.ts` - browser assertions

  **Acceptance Criteria**:
  - [ ] `lsp_diagnostics` reports no new errors in edited files.
  - [ ] `cd frontend && npx vitest run "app/(public)/listings/page.test.tsx"` exits `0`.
  - [ ] `cd frontend && npx playwright test "e2e/public-listings.spec.ts" --project=chromium` exits `0`.
  - [ ] `cd frontend && npm test` exits `0`.
  - [ ] `cd frontend && npm run build` exits `0`.

  **QA Scenarios**:
  ```
  Scenario: Focused verification stack
    Tool: Bash
    Steps: Run the dedicated Vitest, Playwright, full `npm test`, and `npm run build` commands in `frontend/`.
    Expected: All commands exit 0.
    Evidence: .sisyphus/evidence/task-8-verification.txt

  Scenario: File diagnostics stay clean
    Tool: LSP
    Steps: Check edited listing-page, card, filter, and test files.
    Expected: No new errors remain.
    Evidence: .sisyphus/evidence/task-8-lsp.txt
  ```

  **Commit**: NO | Message: `n/a` | Files: `n/a`

## Final Verification Wave (4 parallel agents, ALL must APPROVE)
- [x] F1. Plan Compliance Audit - oracle
- [x] F2. Code Quality Review - unspecified-high
- [x] F3. Real Manual QA - unspecified-high (+ playwright)
- [x] F4. Scope Fidelity Check - deep

## Commit Strategy
- Commit 1: query-state fidelity + route test
- Commit 2: page shell refactor
- Commit 3: toolbar density
- Commit 4: result chrome/footer refinement
- Commit 5: listing-card rhythm
- Commit 6: Playwright coverage
- Final step: verification only, no commit unless fixes are required

## Success Criteria
- `/listings` reads visually like the provided brokerage screenshot in structure and rhythm, not like a hero/catalog page.
- The route stays SSR and backend-query-safe.
- Query defaults/reset behavior are accurate and test-covered.
- Desktop and mobile shell behavior are explicitly defined and verified.
- No backend contract expansion is required to ship the iteration.

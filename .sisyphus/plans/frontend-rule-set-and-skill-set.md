# Frontend Rule Set And Skill Set

## TL;DR
> **Summary**: Establish the first real frontend architecture for `frontend/` using App Router, backend-owned cookie auth, native `fetch`, client-scoped TanStack Query, RHF + Zod + shadcn forms, and backend-mediated Cloudinary image flows.
> **Deliverables**:
> - Decision-complete frontend architecture and folder structure
> - Auth, fetch, env, query, form, image, and shadcn rules
> - Implementation task sequence with executable verification
> - Guardrails that block Axios, Zustand auth state, Auth.js authority, and token storage
> **Effort**: Medium
> **Parallel**: YES - 3 waves
> **Critical Path**: 1 -> 2 -> 5 -> 6 -> 8 -> 9 -> 10 -> 11 -> F1-F4

## Context
### Original Request
Create a decision-complete frontend rule set and skill set plan using mandatory stack choices: Next.js App Router, TypeScript, Tailwind, shadcn/ui, TanStack Query + native fetch only, RHF + Zod + shadcn Form, Cloudinary + next/image, and hybrid server/client rules.

### Interview Summary
- Frontend is still scaffold-level and has no established data, auth, form, query, or UI stack.
- Backend Go service already owns Google OAuth, httpOnly cookie auth, `/auth/me`, `/auth/refresh`, listing CRUD, categories, and listing image APIs.
- Backend remains the only auth authority in phase 1.
- Default architecture decisions applied because the user continued without overriding them:
  - No general Next BFF.
  - Protected routes are SSR-first and call backend `/auth/me` only.
  - No SSR refresh retry in phase 1; client-initiated protected requests own refresh+retry.
  - Direct browser-to-backend credentialed fetch for client code.
  - Server-side fetch to backend for Server Components and route gating.
  - Auth.js is forbidden in phase 1.
  - Admin UI is out of phase 1 scope; route conventions may reserve for it later.

### Metis Review (gaps addressed)
- Resolved SSR refresh ambiguity by explicitly excluding SSR refresh retry in phase 1.
- Locked unauthenticated protected-route behavior to redirect to `/login`.
- Locked production deployment assumption to same-site frontend/backend origins; cross-site auth is out of scope for this plan.
- Tightened route contract references to real backend endpoints and query parameters.
- Tightened banned-pattern rules for token storage, Axios, Zustand auth state, Auth.js, and overly broad Cloudinary config.

## Work Objectives
### Core Objective
Produce a decision-complete implementation plan for the first production-ready frontend architecture so executors can build a seller-facing App Router frontend that integrates with the existing Go backend without inventing new auth or transport layers.

### Deliverables
- Concrete frontend directory structure and route groups for public, auth, and seller dashboard flows.
- Exact auth/session rules for server and browser execution contexts.
- Exact fetch/query/hydration rules using native `fetch` and TanStack Query.
- Exact form, schema, and error-mapping rules using RHF + Zod + shadcn Form.
- Exact image upload/render rules using backend image APIs and `next/image` with Cloudinary delivery URLs.
- Exact env and config rules, including server-only validation and minimal `NEXT_PUBLIC_*` exposure.
- Exact guardrails and verification checks preventing forbidden libraries and patterns.

### Definition of Done (verifiable conditions with commands)
- `test -f frontend/package.json` exits `0`.
- `grep -n 'http://localhost:3000/dashboard' backend/internal/handler/http/auth.go` returns one redirect match used by the plan.
- `grep -n 'AllowCredentials: true' backend/internal/router/router.go` returns one CORS credentials match used by the plan.
- `grep -n 'type Response struct' backend/pkg/utils/response.go` returns the backend envelope definition used by the plan.
- `test -f .sisyphus/plans/frontend-rule-set-and-skill-set.md` exits `0` and the file contains `## TODOs` plus `## Final Verification Wave`.

### Must Have
- Backend cookie auth remains canonical.
- Server Components stay the default rendering mode.
- Native `fetch` remains the only HTTP client.
- TanStack Query remains client-only async state infrastructure.
- RHF + Zod + shadcn Form own non-trivial interactive form patterns.
- Cloudinary URLs are rendered via `next/image` with strict host/path config.
- Protected routes redirect to `/login` when `/auth/me` fails on SSR.

### Must NOT Have (guardrails, AI slop patterns, scope boundaries)
- No Auth.js or NextAuth in phase 1.
- No Axios.
- No Zustand for auth, current user, or backend-derived state.
- No `localStorage`, `sessionStorage`, or readable cookie token persistence.
- No general Next API proxy/BFF for listings, categories, or images.
- No SSR refresh retry in phase 1.
- No direct browser-to-Cloudinary uploads.
- No admin UI implementation in phase 1.
- No generic monolithic `lib/api.ts` that mixes server and browser behavior.

## Verification Strategy
> ZERO HUMAN INTERVENTION - all verification is agent-executed.
- Test decision: tests-after + Playwright for browser flows + static grep/node checks for architecture guardrails.
- QA policy: every task includes agent-executed happy-path and failure-path scenarios.
- Evidence: `.sisyphus/evidence/task-{N}-{slug}.{ext}`.

## Execution Strategy
### Parallel Execution Waves
> Target: 5-8 tasks per wave. Extract shared dependencies into Wave 1.

Wave 1: architecture foundation, config, contracts, provider scaffolding, rule enforcement.
Wave 2: route scaffolding, auth flow, public data reads, seller dashboard reads.
Wave 3: listing create/edit flows, image manager, verification hardening.

### Dependency Matrix (full, all tasks)
- 1 blocks 2, 3, 4, 5.
- 2 blocks 4, 6, 7, 8, 9, 10.
- 3 blocks 5, 7, 8, 9, 10.
- 4 blocks 6, 8, 9, 10.
- 5 blocks 8, 9, 10.
- 6 blocks 8, 9.
- 7 blocks 8.
- 8 blocks 10.
- 9 blocks 10.
- 10 blocks 11.
- 11 blocks F1-F4.

### Agent Dispatch Summary (wave -> task count -> categories)
- Wave 1 -> 4 tasks -> implementation, visual-engineering.
- Wave 2 -> 4 tasks -> implementation, visual-engineering.
- Wave 3 -> 3 tasks -> implementation, testing.
- Final Verification -> 4 tasks -> oracle, unspecified-high, unspecified-high, deep.

## TODOs
> Implementation + Test = ONE task. Never separate.
> EVERY task MUST have: Agent Profile + Parallelization + QA Scenarios.

- [x] 1. Replace the scaffold with the target frontend architecture baseline

  **What to do**: Install the mandated frontend dependencies, remove create-next-app placeholder UI, and establish the baseline directory structure for `app/`, `features/`, `components/ui`, `components/shared`, and `lib/`. Create route groups for `(public)`, `(auth)`, and `(dashboard)` plus a dedicated `/login` route and `/dashboard` entry route to satisfy the backend OAuth redirect. Keep `app/layout.tsx` as a Server Component and move any provider logic into a dedicated client wrapper file.
  **Must NOT do**: Do not add `pages/`, do not add Auth.js, Axios, Zustand, or a monolithic client root layout, and do not implement admin screens in this task.

  **Recommended Agent Profile**:
  - Category: `implementation` - Reason: establishes the baseline architecture and dependency graph.
  - Skills: `[]` - No extra skill is required for the baseline scaffold replacement.
  - Omitted: `[]` - No additional skills are intentionally required.

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: 2, 3, 4, 5 | Blocked By: none

  **References** (executor has NO interview context - be exhaustive):
  - Pattern: `frontend/AGENTS.md` - Existing frontend constraints: App Router only, server components by default, no auth tokens in localStorage.
  - Pattern: `frontend/tsconfig.json` - Existing strict TypeScript and `@/*` path alias.
  - Pattern: `frontend/app/layout.tsx` - Root layout must stay server-rendered.
  - Pattern: `frontend/app/page.tsx` - Placeholder content to remove.
  - API/Type: `backend/internal/handler/http/auth.go:95` - Backend OAuth callback redirects to `/dashboard`.
  - API/Type: `backend/internal/router/router.go:80` - Auth and API route grouping the frontend must align with.
  - External: `https://github.com/vercel/next.js/blob/adf8c612adddd103647c90ff0f511ea35c57076e/docs/01-app/01-getting-started/05-server-and-client-components.mdx#L411-L413` - Providers belong deep in the tree, not in a client root layout.

  **Acceptance Criteria** (agent-executable only):
  - [x] `cd frontend && node -e "const p=require('./package.json'); const d={...p.dependencies,...p.devDependencies}; const required=['@tanstack/react-query','react-hook-form','zod','@hookform/resolvers']; const missing=required.filter(x=>!d[x]); if(missing.length){console.error(missing.join(',')); process.exit(1)}"` exits `0`.
  - [x] `cd frontend && node -e "const p=require('./package.json'); const d={...p.dependencies,...p.devDependencies}; const banned=['axios','zustand','next-auth','@auth/core']; const found=banned.filter(x=>d[x]); if(found.length){console.error(found.join(',')); process.exit(1)}"` exits `0`.
  - [x] `test -d frontend/features && test -d frontend/components/ui && test -d frontend/lib` exits `0`.
  - [x] `grep -n 'Create Next App' frontend/app/layout.tsx` returns no matches.
  - [x] `test -f frontend/app/login/page.tsx && test -f "frontend/app/(dashboard)/dashboard/page.tsx"` exits `0`.

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```text
  Scenario: Baseline route shell renders
    Tool: Playwright
    Steps: Start the frontend dev server; open `/`; verify `[data-testid="home-shell"]` exists; open `/login`; verify `[data-testid="login-google-button"]` exists.
    Expected: Public home shell and login entry render without hydration errors.
    Evidence: .sisyphus/evidence/task-1-architecture-baseline.png

  Scenario: Forbidden dependency guard catches violations
    Tool: Bash
    Steps: Run the banned-dependency node check from Acceptance Criteria.
    Expected: Command exits `0`; any banned package would fail the task.
    Evidence: .sisyphus/evidence/task-1-architecture-baseline.txt
  ```

  **Commit**: YES | Message: `feat(frontend): establish app router architecture baseline` | Files: `frontend/package.json`, `frontend/app/layout.tsx`, `frontend/app/page.tsx`, `frontend/app/login/page.tsx`, `frontend/app/(dashboard)/dashboard/page.tsx`, `frontend/features/**`, `frontend/components/**`, `frontend/lib/**`

- [x] 2. Implement typed env parsing and split browser/server fetch layers

  **What to do**: Create a server-only env module validated with Zod, expose only a minimal `NEXT_PUBLIC_*` browser-safe subset, and create two fetch entry points: `serverFetch` for Server Components and server actions, and `browserFetch` for client code with `credentials: "include"`. Both layers must unwrap the backend response envelope, preserve `trace_id`, and normalize backend failures into one shared error shape. `serverFetch` must forward incoming cookies explicitly when calling the backend; `browserFetch` must own 401 -> `/auth/refresh` -> retry for client-initiated protected requests only.
  **Must NOT do**: Do not mix server and browser logic in one fetch helper, do not use Axios, and do not implement SSR refresh retry.

  **Recommended Agent Profile**:
  - Category: `implementation` - Reason: creates the transport and env contracts every feature depends on.
  - Skills: `[]` - No extra skill is required for the transport layer.
  - Omitted: `[]` - No additional skills are intentionally required.

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: 4, 6, 7, 8, 9, 10 | Blocked By: 1

  **References** (executor has NO interview context - be exhaustive):
  - Pattern: `backend/pkg/utils/response.go:9` - Canonical response envelope fields.
  - Pattern: `backend/internal/handler/http/auth.go:120` - Refresh endpoint rotates cookies and returns success.
  - Pattern: `backend/internal/router/router.go:49` - Backend CORS credentials are already enabled.
  - Pattern: `frontend/AGENTS.md` - Future frontend calls should use credentialed requests.
  - External: `https://github.com/vercel/next.js/blob/adf8c612adddd103647c90ff0f511ea35c57076e/docs/01-app/02-guides/environment-variables.mdx#L182-L226` - Server runtime env handling in App Router.
  - External: `https://github.com/vercel/next.js/blob/adf8c612adddd103647c90ff0f511ea35c57076e/docs/01-app/01-getting-started/05-server-and-client-components.mdx#L702-L776` - `server-only` guard for secret-bearing modules.
  - External: `https://github.com/TanStack/query/blob/09397ca84c5912060f312ab5fe5b15955ad5eac3/docs/framework/react/guides/query-functions.md#L30-L69` - Native `fetch` query functions should throw on non-OK responses.

  **Acceptance Criteria** (agent-executable only):
  - [x] `test -f frontend/lib/env/server.ts && test -f frontend/lib/env/public.ts && test -f frontend/lib/api/server-fetch.ts && test -f frontend/lib/api/browser-fetch.ts` exits `0`.
  - [x] `grep -R "server-only" frontend/lib/env/server.ts` returns one match.
  - [x] `grep -R 'credentials: "include"' frontend/lib/api/browser-fetch.ts` returns one match.
  - [x] `grep -R 'localStorage\|sessionStorage\|document.cookie' frontend/lib frontend/features frontend/app` returns no auth-storage matches.
  - [x] `grep -R 'trace_id' frontend/lib/api` returns at least one match documenting envelope preservation.

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```text
  Scenario: Browser fetch refreshes and retries client request
    Tool: Playwright
    Steps: Seed the browser with a valid refresh cookie and an expired access cookie; visit a client-driven protected page using `browserFetch`; trigger a query via `[data-testid="dashboard-refresh-button"]`.
    Expected: The first request 401s, browser fetch calls backend `/auth/refresh`, retries once, and the UI resolves to a signed-in state without redirecting.
    Evidence: .sisyphus/evidence/task-2-fetch-refresh.json

  Scenario: SSR protected fetch does not perform server-side refresh retry
    Tool: Playwright
    Steps: Seed the browser with a valid refresh cookie and an expired access cookie; request `/dashboard` as a fresh navigation.
    Expected: Server-side auth check fails and redirects to `/login`; there is no hidden SSR refresh loop.
    Evidence: .sisyphus/evidence/task-2-fetch-refresh-error.png
  ```

  **Commit**: YES | Message: `feat(frontend): add typed env and backend fetch layers` | Files: `frontend/lib/env/**`, `frontend/lib/api/**`, `frontend/lib/types/**`

- [x] 3. Establish UI primitives, theme foundation, and shadcn ownership rules

  **What to do**: Add the shadcn/ui setup, create the initial primitive inventory under `components/ui/`, add a shared visual theme in `app/globals.css`, and create composition rules that keep business logic out of UI primitives. Include app-shell components in `components/shared/` for header, sidebar, empty states, and status banners. Reserve `components/ui/` for generated or near-generated primitives only.
  **Must NOT do**: Do not place domain-specific listing/auth/category logic inside `components/ui/`, and do not leave the default scaffold typography/theme untouched.

  **Recommended Agent Profile**:
  - Category: `visual-engineering` - Reason: this task defines the reusable UI and visual baseline.
  - Skills: `[]` - No extra skill is required for the initial design system setup.
  - Omitted: `[]` - No additional skills are intentionally required.

  **Parallelization**: Can Parallel: YES | Wave 1 | Blocks: 5, 8, 9, 10 | Blocked By: 1

  **References** (executor has NO interview context - be exhaustive):
  - Pattern: `frontend/app/globals.css` - Current global theme entry point.
  - Pattern: `frontend/app/layout.tsx` - Root layout where shared shell classes and metadata will attach.
  - External: `https://github.com/shadcn-ui/ui/blob/31dbc6fc91950430b5d5647bc9a69d428495afb5/apps/v4/content/docs/forms/react-hook-form.mdx#L92-L176` - shadcn form composition style.
  - External: `https://github.com/vercel/next.js/blob/adf8c612adddd103647c90ff0f511ea35c57076e/docs/01-app/01-getting-started/05-server-and-client-components.mdx#L11-L31` - Server Components by default, client islands only when interactive.

  **Acceptance Criteria** (agent-executable only):
  - [x] `test -d frontend/components/ui && test -d frontend/components/shared` exits `0`.
  - [x] `grep -R 'use client' frontend/components/ui` returns no matches unless the primitive genuinely requires it.
  - [x] `grep -n 'font-sans' frontend/app/globals.css` returns a project-specific theme match instead of the starter theme only.
  - [x] `test -f frontend/components/shared/app-header.tsx && test -f frontend/components/shared/dashboard-sidebar.tsx` exits `0`.

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```text
  Scenario: Shared shell renders on desktop and mobile
    Tool: Playwright
    Steps: Open `/dashboard` in desktop and mobile viewports after authentication; verify `[data-testid="app-header"]` and `[data-testid="dashboard-sidebar"]` render; on mobile verify `[data-testid="dashboard-nav-toggle"]` opens the sidebar.
    Expected: Shared shell is responsive and no UI primitive contains domain-specific copy or network logic.
    Evidence: .sisyphus/evidence/task-3-ui-foundation.png

  Scenario: Primitive ownership rule remains clean
    Tool: Bash
    Steps: Run `grep -R 'fetch\|useQuery\|useMutation\|listing\|category\|auth/me' frontend/components/ui`.
    Expected: No business/domain/network logic exists under `components/ui`.
    Evidence: .sisyphus/evidence/task-3-ui-foundation.txt
  ```

  **Commit**: YES | Message: `feat(frontend): add shadcn ui foundation and shared shell` | Files: `frontend/app/globals.css`, `frontend/components/ui/**`, `frontend/components/shared/**`

- [x] 4. Add the client provider boundary and TanStack Query operating rules

  **What to do**: Create a single small client providers module mounted from the server root layout. It must host `QueryClientProvider` only, configure a non-zero default `staleTime`, and expose hydration helpers only for explicitly prefetched client subtrees. Document and enforce that Server Components own first-render reads and TanStack Query owns client-side async state, background refetch, and mutations only.
  **Must NOT do**: Do not convert `app/layout.tsx` into a client component, do not hydrate the entire app by default, and do not use TanStack Query as the initial auth bootstrap.

  **Recommended Agent Profile**:
  - Category: `implementation` - Reason: provider placement and query defaults are cross-cutting architecture decisions.
  - Skills: `[]` - No extra skill is required for provider setup.
  - Omitted: `[]` - No additional skills are intentionally required.

  **Parallelization**: Can Parallel: YES | Wave 1 | Blocks: 8, 9, 10 | Blocked By: 1, 2

  **References** (executor has NO interview context - be exhaustive):
  - Pattern: `frontend/app/layout.tsx` - Provider wrapper must be mounted from here while keeping layout server-rendered.
  - External: `https://github.com/TanStack/query/blob/09397ca84c5912060f312ab5fe5b15955ad5eac3/docs/framework/react/guides/advanced-ssr.md#L26-L77` - Small App Router provider wrapper and `staleTime` default.
  - External: `https://github.com/TanStack/query/blob/09397ca84c5912060f312ab5fe5b15955ad5eac3/docs/framework/react/guides/advanced-ssr.md#L314-L360` - Warning against dual ownership of the same data on server and client.

  **Acceptance Criteria** (agent-executable only):
  - [x] `test -f frontend/app/providers.tsx` exits `0`.
  - [x] `grep -n 'QueryClientProvider' frontend/app/providers.tsx` returns one match.
  - [x] `grep -n 'staleTime' frontend/app/providers.tsx` returns one match with a non-zero value.
  - [x] `grep -n 'use client' frontend/app/layout.tsx` returns no matches.

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```text
  Scenario: Query provider powers a client island only
    Tool: Playwright
    Steps: Visit a public page with a client-side filter island under `[data-testid="listing-filters"]`; interact with the filter controls.
    Expected: The filter island updates via TanStack Query without turning the whole route into a client-rendered page.
    Evidence: .sisyphus/evidence/task-4-query-provider.png

  Scenario: Server route remains server-rendered without client layout fallback
    Tool: Bash
    Steps: Run `grep -R 'use client' frontend/app/layout.tsx frontend/app/**/layout.tsx`.
    Expected: Root layout remains server-rendered; only explicit client provider or client islands use the directive.
    Evidence: .sisyphus/evidence/task-4-query-provider.txt
  ```

  **Commit**: YES | Message: `feat(frontend): add query provider boundary and rules` | Files: `frontend/app/providers.tsx`, `frontend/app/layout.tsx`, `frontend/lib/query/**`

- [x] 5. Implement auth entry routes and deterministic protected-route UX

  **What to do**: Build the `/login` route, auth CTA components, and route-group rules for `(public)`, `(auth)`, and `(dashboard)`. The login screen must start Google OAuth through the backend provider route, show signed-out and auth-error states, and never attempt to own session state itself. The protected dashboard group must redirect to `/login` whenever the SSR auth check fails. Reserve an auth status banner for expired-session messaging after client-side refresh failure.
  **Must NOT do**: Do not add a local credentials form, do not embed OAuth secrets, and do not keep a parallel `SessionProvider` that can disagree with `/auth/me`.

  **Recommended Agent Profile**:
  - Category: `visual-engineering` - Reason: route UX, login shell, and deterministic redirects are user-facing.
  - Skills: `[]` - No extra skill is required for the login route work.
  - Omitted: `[]` - No additional skills are intentionally required.

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: 6, 8, 9, 10 | Blocked By: 1, 3

  **References** (executor has NO interview context - be exhaustive):
  - Pattern: `backend/internal/router/router.go:87` - Google OAuth begins at `/auth/oauth/:provider`.
  - Pattern: `backend/internal/handler/http/auth.go:95` - Successful OAuth callback redirects to `/dashboard`.
  - Pattern: `backend/internal/router/router.go:96` - `/auth/me` is protected and must be respected by the frontend.
  - Pattern: `frontend/AGENTS.md` - Backend auth is cookie-based; tokens must not be stored in localStorage.
  - External: `https://github.com/vercel/next.js/blob/adf8c612adddd103647c90ff0f511ea35c57076e/docs/01-app/02-guides/authentication.mdx#L623-L653` - Cookie-auth route protection guidance.

  **Acceptance Criteria** (agent-executable only):
  - [x] `test -f frontend/app/login/page.tsx && test -f "frontend/app/(dashboard)/layout.tsx"` exits `0`.
  - [x] `grep -R '/auth/oauth/google' frontend/app frontend/features` returns at least one match for the Google login CTA.
  - [x] `grep -R 'SessionProvider\|useSession' frontend` returns no matches.
  - [x] `grep -R 'redirect("/login")\|redirect(\x27/login\x27)' "frontend/app/(dashboard)" frontend/features/auth` returns at least one SSR guard redirect match.

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```text
  Scenario: Anonymous user is redirected to login
    Tool: Playwright
    Steps: Clear cookies; visit `/dashboard` directly.
    Expected: Browser ends on `/login`; `[data-testid="dashboard-shell"]` is absent; `[data-testid="login-google-button"]` is visible.
    Evidence: .sisyphus/evidence/task-5-auth-routes.png

  Scenario: Login page shows recoverable auth error state
    Tool: Playwright
    Steps: Visit `/login?reason=session-expired`.
    Expected: `[data-testid="auth-status-banner"]` renders an expired-session message and the Google login button remains usable.
    Evidence: .sisyphus/evidence/task-5-auth-routes-error.png
  ```

  **Commit**: YES | Message: `feat(frontend): add login flow and protected route rules` | Files: `frontend/app/login/page.tsx`, `frontend/app/(dashboard)/**`, `frontend/features/auth/**`

- [x] 6. Implement server-side current-user resolution and dashboard auth guard helpers

  **What to do**: Create the server-only auth helper layer that reads incoming cookies, calls backend `/auth/me`, maps the response into a typed `CurrentUser` model, and exposes reusable helpers for `requireUser()` and `getOptionalUser()`. These helpers must be used by protected dashboard layouts/pages so SSR-first route gating remains consistent. If `/auth/me` fails on SSR, redirect to `/login`; do not attempt SSR refresh retry.
  **Must NOT do**: Do not duplicate current-user state in Zustand or React context, and do not call the refresh endpoint from SSR helpers.

  **Recommended Agent Profile**:
  - Category: `implementation` - Reason: this is the canonical auth boundary and server helper layer.
  - Skills: `[]` - No extra skill is required for the auth guard helper layer.
  - Omitted: `[]` - No additional skills are intentionally required.

  **Parallelization**: Can Parallel: NO | Wave 2 | Blocks: 8, 9, 10 | Blocked By: 2, 5

  **References** (executor has NO interview context - be exhaustive):
  - Pattern: `backend/internal/router/router.go:97` - `/auth/me` route exists and is protected.
  - Pattern: `backend/internal/handler/http/auth.go:99` - `GetMe` returns the authenticated user via the standard envelope.
  - Pattern: `backend/pkg/utils/response.go:9` - Envelope must be unwrapped consistently.
  - External: `https://github.com/vercel/next.js/blob/adf8c612adddd103647c90ff0f511ea35c57076e/docs/01-app/01-getting-started/05-server-and-client-components.mdx#L702-L776` - Server-only modules should be protected.
  - External: `https://github.com/vercel/next.js/blob/adf8c612adddd103647c90ff0f511ea35c57076e/docs/01-app/02-guides/authentication.mdx#L1129-L1189` - Auth verification should happen close to data access.

  **Acceptance Criteria** (agent-executable only):
  - [x] `test -f frontend/features/auth/server/current-user.ts && test -f frontend/features/auth/server/require-user.ts` exits `0`.
  - [x] `grep -R 'auth/me' frontend/features/auth/server "frontend/app/(dashboard)"` returns at least one match.
  - [x] `grep -R 'auth/refresh' frontend/features/auth/server` returns no matches.
  - [x] `grep -R 'server-only' frontend/features/auth/server` returns at least one match.

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```text
  Scenario: Authenticated server render resolves current user
    Tool: Playwright
    Steps: Seed valid access and refresh cookies; visit `/dashboard`.
    Expected: `[data-testid="dashboard-shell"]` renders and `[data-testid="current-user-email"]` shows the authenticated user email from `/auth/me`.
    Evidence: .sisyphus/evidence/task-6-current-user.png

  Scenario: Expired access cookie fails SSR guard without refresh retry
    Tool: Playwright
    Steps: Seed expired access cookie and valid refresh cookie; hard-load `/dashboard`.
    Expected: Browser lands on `/login`; no background dashboard content flashes before redirect.
    Evidence: .sisyphus/evidence/task-6-current-user-error.png
  ```

  **Commit**: YES | Message: `feat(frontend): add server auth guard helpers` | Files: `frontend/features/auth/server/**`, `frontend/app/(dashboard)/**`

- [x] 7. Implement public categories and listings read contracts with server-first rendering

  **What to do**: Create the public data layer for categories, listing browse, and listing detail. Public route pages must fetch on the server via `serverFetch`, map backend DTOs into frontend view models, and read the real query parameter contract (`page`, `limit`, `city`, `category_id`, `price_min`, `price_max`, `status`). Use TanStack Query only for optional client-side filter or pagination islands after the initial server render.
  **Must NOT do**: Do not invent new backend parameters, do not rename contract fields without a mapper, and do not make public browse pages client-only.

  **Recommended Agent Profile**:
  - Category: `implementation` - Reason: this task establishes the core public read models and server data patterns.
  - Skills: `[]` - No extra skill is required for public read flows.
  - Omitted: `[]` - No additional skills are intentionally required.

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: 8 | Blocked By: 2

  **References** (executor has NO interview context - be exhaustive):
  - Pattern: `backend/internal/router/router.go:103` - Public listing list route.
  - Pattern: `backend/internal/router/router.go:104` - Public listing slug route.
  - Pattern: `backend/internal/router/router.go:123` - Public categories route.
  - API/Type: `backend/internal/handler/http/listing.go:252` - Actual listing filter query parameters and limit cap.
  - API/Type: `backend/internal/dto/response/listing_response.go:10` - Listing response contract.
  - API/Type: `backend/internal/dto/response/category_response.go` - Category response contract.
  - Test: `backend/postman_collection.json` - Endpoint examples and payload expectations.

  **Acceptance Criteria** (agent-executable only):
  - [x] `test -f frontend/features/listings/server/get-listings.ts && test -f frontend/features/listings/server/get-listing-by-slug.ts && test -f frontend/features/categories/server/get-categories.ts` exits `0`.
  - [x] `grep -R 'price_min\|price_max\|category_id\|city\|page\|limit' frontend/features/listings` returns matches documenting the real filter contract.
  - [x] `grep -R 'useQuery' "frontend/app/(public)"` returns no matches for the top-level public pages themselves.
  - [x] `test -f "frontend/app/(public)/listings/page.tsx" && test -f "frontend/app/(public)/listings/[slug]/page.tsx"` exits `0`.

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```text
  Scenario: Public browse page renders server-fetched listing data
    Tool: Playwright
    Steps: Visit `/listings?page=1&limit=10`; verify `[data-testid="listing-card"]` renders at least one result and `[data-testid="listing-pagination"]` shows current page 1.
    Expected: Browse page renders with server data on first load and no client-only loading shell dominates the route.
    Evidence: .sisyphus/evidence/task-7-public-listings.png

  Scenario: Invalid filter input degrades safely
    Tool: Playwright
    Steps: Visit `/listings?limit=500&price_min=abc`.
    Expected: UI renders a safe fallback state using backend-normalized defaults and does not crash.
    Evidence: .sisyphus/evidence/task-7-public-listings-error.png
  ```

  **Commit**: YES | Message: `feat(frontend): add public listings and categories reads` | Files: `frontend/app/(public)/**`, `frontend/features/listings/server/**`, `frontend/features/categories/server/**`, `frontend/features/listings/components/**`

- [x] 8. Implement the seller dashboard listing index with SSR-first data and client revalidation islands

  **What to do**: Build the seller dashboard home and listing index routes using SSR-authenticated page renders plus optional client islands for refetch, sorting, and optimistic refresh. Read seller listings from backend `/auth/me/listings`, render them inside the dashboard shell, and keep any client query usage limited to local controls or explicit refresh buttons.
  **Must NOT do**: Do not fetch seller listings from a guessed route, do not gate access client-side only, and do not duplicate the dashboard list in both server ownership and uncontrolled client ownership.

  **Recommended Agent Profile**:
  - Category: `implementation` - Reason: combines the canonical protected route pattern with the seller’s primary read view.
  - Skills: `[]` - No extra skill is required for the dashboard index.
  - Omitted: `[]` - No additional skills are intentionally required.

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: 9, 10 | Blocked By: 2, 4, 5, 6, 7

  **References** (executor has NO interview context - be exhaustive):
  - Pattern: `backend/internal/router/router.go:119` - Seller listings route is `/auth/me/listings`.
  - Pattern: `backend/internal/handler/http/listing.go:228` - Seller listing filter flow.
  - API/Type: `backend/internal/dto/response/listing_response.go:45` - Paginated listing payload for seller dashboard.
  - External: `https://github.com/TanStack/query/blob/09397ca84c5912060f312ab5fe5b15955ad5eac3/docs/framework/react/guides/advanced-ssr.md#L161-L218` - Hydrate only explicit client subtrees when needed.

  **Acceptance Criteria** (agent-executable only):
  - [x] `test -f "frontend/app/(dashboard)/dashboard/page.tsx" && test -f "frontend/app/(dashboard)/dashboard/listings/page.tsx"` exits `0`.
  - [x] `grep -R '/auth/me/listings' frontend/features/listings "frontend/app/(dashboard)"` returns at least one match.
  - [x] `grep -R 'useQuery' "frontend/app/(dashboard)/dashboard/page.tsx" "frontend/app/(dashboard)/dashboard/listings/page.tsx"` returns no matches for the route files themselves.
  - [x] `grep -R 'dashboard-refresh-button' frontend` returns at least one match.

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```text
  Scenario: Seller dashboard listing index renders on first load
    Tool: Playwright
    Steps: Seed valid auth cookies; visit `/dashboard/listings`.
    Expected: `[data-testid="dashboard-listings-table"]` renders with listing rows and `[data-testid="dashboard-refresh-button"]` is visible.
    Evidence: .sisyphus/evidence/task-8-dashboard-listings.png

  Scenario: Anonymous access cannot see seller data
    Tool: Playwright
    Steps: Clear cookies; visit `/dashboard/listings`.
    Expected: Browser redirects to `/login`; `[data-testid="dashboard-listings-table"]` never appears.
    Evidence: .sisyphus/evidence/task-8-dashboard-listings-error.png
  ```

  **Commit**: YES | Message: `feat(frontend): add seller dashboard listing views` | Files: `frontend/app/(dashboard)/dashboard/**`, `frontend/features/listings/**`

- [x] 9. Implement listing create and edit forms with RHF, Zod, and backend error mapping

  **What to do**: Build the seller listing create/edit experience as client islands within protected dashboard routes. Use RHF + `zodResolver` + shadcn Form components for all non-trivial listing fields, map backend DTO contracts to form defaults and submission payloads, and normalize backend envelope errors into field-level or form-level messages. Submit mutations through `browserFetch` with credentialed requests and invalidate only the relevant client query keys after success.
  **Must NOT do**: Do not treat client Zod validation as the security boundary, do not use uncontrolled ad hoc forms, and do not duplicate listing schema definitions across multiple files without re-exporting a canonical schema.

  **Recommended Agent Profile**:
  - Category: `implementation` - Reason: combines the mandated form stack with authenticated listing mutations.
  - Skills: `[]` - No extra skill is required for the listing form stack.
  - Omitted: `[]` - No additional skills are intentionally required.

  **Parallelization**: Can Parallel: YES | Wave 3 | Blocks: 10, 11 | Blocked By: 2, 3, 4, 5, 6, 8

  **References** (executor has NO interview context - be exhaustive):
  - Pattern: `backend/internal/router/router.go:111` - Listing create route.
  - Pattern: `backend/internal/router/router.go:112` - Listing update route.
  - API/Type: `backend/internal/dto/request/listing_request.go` - Canonical listing create/update payloads.
  - API/Type: `backend/internal/dto/response/listing_response.go:10` - Canonical listing response payload.
  - External: `https://github.com/shadcn-ui/ui/blob/31dbc6fc91950430b5d5647bc9a69d428495afb5/apps/v4/content/docs/forms/react-hook-form.mdx#L92-L176` - shadcn Form + RHF structure.
  - External: `https://github.com/vercel/next.js/blob/adf8c612adddd103647c90ff0f511ea35c57076e/docs/01-app/02-guides/authentication.mdx#L99-L179` - Server-side validation still matters even when the client uses Zod.

  **Acceptance Criteria** (agent-executable only):
  - [x] `test -f frontend/features/listings/forms/listing-form.tsx && test -f frontend/features/listings/forms/listing-schema.ts` exits `0`.
  - [x] `grep -R 'zodResolver' frontend/features/listings/forms` returns at least one match.
  - [x] `grep -R 'FormField' frontend/features/listings/forms` returns at least one shadcn form-field match.
  - [x] `grep -R 'useMutation' frontend/features/listings` returns at least one match for create/update mutations.
  - [x] `grep -R 'price' frontend/features/listings/forms/listing-schema.ts` returns at least one price rule reflecting integer IDR handling.

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```text
  Scenario: Seller creates a valid listing
    Tool: Playwright
    Steps: Authenticate; visit `/dashboard/listings/new`; fill `[name="title"]`, `[name="price"]`, `[name="status"]`, and other required fields; submit via `[data-testid="listing-submit-button"]`.
    Expected: Success toast/banner appears, browser lands on the listing edit or detail route, and the new listing appears in `/dashboard/listings`.
    Evidence: .sisyphus/evidence/task-9-listing-form.png

  Scenario: Invalid listing payload surfaces field and form errors safely
    Tool: Playwright
    Steps: Authenticate; visit `/dashboard/listings/new`; submit with an empty title or invalid price.
    Expected: Zod-backed client validation prevents submission when invalid locally; backend validation failures render `[data-testid="listing-form-error"]` and field errors without crashing.
    Evidence: .sisyphus/evidence/task-9-listing-form-error.png
  ```

  **Commit**: YES | Message: `feat(frontend): add seller listing forms` | Files: `frontend/features/listings/forms/**`, `frontend/app/(dashboard)/dashboard/listings/new/**`, `frontend/app/(dashboard)/dashboard/listings/[id]/**`

- [x] 10. Implement the listing image manager with backend-mediated uploads and Cloudinary rendering

  **What to do**: Build the seller listing image manager UI as a client island attached to the listing edit screen. Upload images to backend multipart endpoint `/api/listings/:id/images` using form-data key `file`, render returned image URLs through `next/image`, and support delete, set-primary, and reorder flows through the real backend routes. Configure `next.config.ts` with strict Cloudinary `remotePatterns` and explicit `qualities` values. Enforce the backend max-image constraint in the UI and show deterministic failure messaging for invalid files, image limit, and storage-disabled responses.
  **Must NOT do**: Do not upload directly to Cloudinary from the browser, do not use broad wildcard `remotePatterns`, and do not hand-roll `<img>` for managed listing images.

  **Recommended Agent Profile**:
  - Category: `implementation` - Reason: image upload and ordering combine client interactivity with backend contract fidelity.
  - Skills: `[]` - No extra skill is required for the image manager.
  - Omitted: `[]` - No additional skills are intentionally required.

  **Parallelization**: Can Parallel: YES | Wave 3 | Blocks: 11 | Blocked By: 2, 3, 4, 8, 9

  **References** (executor has NO interview context - be exhaustive):
  - Pattern: `backend/internal/router/router.go:114` - Upload image route.
  - Pattern: `backend/internal/router/router.go:115` - Delete image route.
  - Pattern: `backend/internal/router/router.go:116` - Set primary image route.
  - Pattern: `backend/internal/router/router.go:117` - Reorder images route.
  - Pattern: `backend/internal/handler/http/listing.go:133` - Upload uses multipart key `file`.
  - API/Type: `backend/internal/dto/response/listing_response.go:32` - Listing image response metadata.
  - Pattern: `frontend/next.config.ts` - Current stock config to extend for Cloudinary delivery hosts.
  - External: `https://github.com/vercel/next.js/blob/adf8c612adddd103647c90ff0f511ea35c57076e/docs/01-app/03-api-reference/02-components/image.mdx#L533-L609` - Strict `remotePatterns` guidance.
  - External: `https://github.com/vercel/next.js/blob/adf8c612adddd103647c90ff0f511ea35c57076e/docs/01-app/03-api-reference/02-components/image.mdx#L696-L722` - `qualities` allowlist guidance.

  **Acceptance Criteria** (agent-executable only):
  - [x] `grep -n 'remotePatterns' frontend/next.config.ts` returns at least one Cloudinary host/path match.
  - [x] `grep -n 'qualities' frontend/next.config.ts` returns at least one match.
  - [x] `grep -R 'FormData' frontend/features/listings` returns at least one upload implementation match.
  - [x] `grep -R '/images/reorder\|/images/:imageId/primary\|/images/:imageId' frontend/features/listings "frontend/app/(dashboard)"` returns route usage matches.
  - [x] `grep -R '<Image' frontend/features/listings "frontend/app/(dashboard)"` returns at least one `next/image` usage match.

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```text
  Scenario: Seller uploads and reorders listing images
    Tool: Playwright
    Steps: Authenticate; open a listing edit page; upload two image fixtures through `[data-testid="listing-image-upload"]`; drag `[data-testid="listing-image-item"]` to reorder; click `[data-testid="listing-image-make-primary"]` on the second image.
    Expected: Uploaded previews render through `next/image`, reorder persists after refresh, and exactly one image is marked primary.
    Evidence: .sisyphus/evidence/task-10-image-manager.png

  Scenario: Invalid upload failure is handled gracefully
    Tool: Playwright
    Steps: Authenticate; attempt to upload an unsupported file or exceed the 10-image limit.
    Expected: `[data-testid="listing-image-error"]` renders a deterministic backend-derived error message; no broken preview cards remain in the UI.
    Evidence: .sisyphus/evidence/task-10-image-manager-error.png
  ```

  **Commit**: YES | Message: `feat(frontend): add listing image manager` | Files: `frontend/features/listings/images/**`, `frontend/next.config.ts`, `frontend/app/(dashboard)/dashboard/listings/[id]/**`

- [x] 11. Add architecture guardrail tests and automated verification fixtures

  **What to do**: Add the Playwright suite and static verification scripts that enforce the architecture rules from this plan. Cover route gating, login redirect behavior, forbidden dependency detection, forbidden token-storage patterns, and Cloudinary/image config checks. Ensure the CI-friendly verification entry points are documented in `frontend/package.json` scripts.
  **Must NOT do**: Do not make verification depend on manual clicking, and do not leave architecture guardrails as prose only.

  **Recommended Agent Profile**:
  - Category: `testing` - Reason: this task converts architecture rules into executable protection.
  - Skills: `[]` - No extra skill is required for the verification suite.
  - Omitted: `[]` - No additional skills are intentionally required.

  **Parallelization**: Can Parallel: NO | Wave 3 | Blocks: F1, F2, F3, F4 | Blocked By: 9, 10

  **References** (executor has NO interview context - be exhaustive):
  - Pattern: `frontend/package.json` - Verification scripts must be added here.
  - Pattern: `backend/internal/handler/http/auth.go:95` - `/dashboard` remains the post-login landing target.
  - Pattern: `backend/internal/router/router.go:49` - Credentials-based CORS must stay aligned with frontend checks.
  - Pattern: `backend/pkg/utils/response.go:9` - Error/success envelope used by auth and listing assertions.
  - External: `https://github.com/pmndrs/zustand/blob/206012dbd1ae046ea0aefb9cd7bf8bba913c6459/docs/learn/guides/nextjs.md#L15-L35` - Zustand should not own server/backend state in Next App Router.

  **Acceptance Criteria** (agent-executable only):
  - [x] `grep -n 'playwright' frontend/package.json` returns at least one test script match.
  - [x] `test -d frontend/e2e` exits `0`.
  - [x] `grep -R 'localStorage\|sessionStorage\|document.cookie' frontend/app frontend/features frontend/lib` returns no token-storage matches.
  - [x] `cd frontend && node -e "const p=require('./package.json'); const d={...p.dependencies,...p.devDependencies}; const banned=['axios','zustand','next-auth','@auth/core']; const found=banned.filter(x=>d[x]); if(found.length){console.error(found.join(',')); process.exit(1)}"` exits `0`.
  - [x] `cd frontend && npm run test:e2e -- --grep @auth` exits `0` against the implemented auth flows.

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```text
  Scenario: Architecture guardrail suite passes end-to-end
    Tool: Bash
    Steps: Run the banned dependency check, banned token-storage grep, and Playwright auth suite.
    Expected: All commands exit `0`, proving the rules are enforced by automation.
    Evidence: .sisyphus/evidence/task-11-guardrails.txt

  Scenario: Guardrail suite fails when a banned pattern is introduced
    Tool: Bash
    Steps: In a disposable temp copy of a guarded file, inject `localStorage.setItem('access_token', 'x')` or add `axios` to dependencies, then rerun the relevant static check without committing the mutation.
    Expected: The relevant static check exits non-zero, proving the guardrail is active.
    Evidence: .sisyphus/evidence/task-11-guardrails-error.txt
  ```

  **Commit**: YES | Message: `test(frontend): enforce architecture guardrails` | Files: `frontend/package.json`, `frontend/e2e/**`, `frontend/scripts/**`

## Final Verification Wave (4 parallel agents, ALL must APPROVE)
- [x] F1. Plan Compliance Audit - oracle
- [x] F2. Code Quality Review - unspecified-high
- [x] F3. Real Runtime QA - unspecified-high (+ playwright if UI)
- [x] F4. Scope Fidelity Check - deep

## Commit Strategy
- Use small commits at the end of each completed wave.
- Commit messages must reflect architecture intent, not just file churn.
- Do not commit placeholder-only routes without their corresponding verification.

## Success Criteria
- Frontend architecture matches the backend auth and DTO contracts without adding a parallel session authority.
- All protected route behavior is deterministic and verified.
- All fetch/query/form/image rules are implemented and enforced by code structure plus static checks.
- The frontend can render public listings/categories, gate seller dashboard routes, submit listing forms, and manage listing images against the existing backend APIs.

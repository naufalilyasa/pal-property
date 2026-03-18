# Draft: Root AGENTS.md Refresh

Use this draft to replace or heavily update `/mnt/data/Projects/pal-property/AGENTS.md`.

```md
# AGENTS.md — pal-property

**Project:** PAL Property  
**Primary Goal:** Property listing platform with Go backend and Next.js frontend  
**Current State:** Backend is mature; seller-facing frontend exists/planned depending on branch state.  

## Repo Overview

- `backend/` — Go 1.26 service using Fiber v3, GORM, PostgreSQL, Redis, Goth OAuth, and Cloudinary-backed listing images.
- `frontend/` — Next.js App Router app using TypeScript and Tailwind.
- `workers/`, `infra/`, `deploy/` — not primary implementation targets right now.

## Current Source-of-Truth Architecture

- **Auth authority:** backend Go is the sole auth/session authority.
- **Frontend auth model:** use backend-issued `httpOnly` cookies and `/auth/me`; do **not** introduce Auth.js or browser token storage in phase 1.
- **API envelope:** backend returns `{ success, message, data, trace_id }`; frontend must normalize this in one shared place.
- **Image flow:** listing image upload/delete/set-primary/reorder is backend-owned and Cloudinary-backed.

## Mandatory Frontend Stack

- Next.js App Router
- TypeScript
- Tailwind CSS
- shadcn/ui
- TanStack Query (React Query) for client-side async state and mutations
- native `fetch` only (**Axios is forbidden**)
- React Hook Form + Zod + shadcn Form for non-trivial forms
- Cloudinary-rendered URLs with `next/image`

## Server / Client Rules

### Server Components (default)
- Use for routing, layout composition, metadata, initial auth checks, and initial page data.
- Use for protected route gating based on backend cookie session.
- Use server-only helpers for `cookies()` access and credential forwarding.

### Client Components
- Use only for interactive islands:
  - RHF forms
  - mutation buttons
  - TanStack Query tables/lists that need client refresh
  - image-upload controls
  - reorder / small interactive state

### Explicit Prohibitions
- Do not make route roots `use client` unless absolutely required.
- Do not import server-only modules into client components.
- Do not use TanStack Query to replace page/layout server fetching.

## Recommended Folder Structure

```text
frontend/
├── src/
│   ├── app/                    # routes, layouts, loading/error boundaries
│   ├── features/
│   │   ├── auth/
│   │   ├── listings/
│   │   ├── categories/
│   │   └── images/
│   ├── components/
│   │   ├── ui/                 # shadcn/ui primitives only
│   │   └── shared/             # cross-feature presentational components
│   ├── lib/
│   │   ├── client/             # browser-safe fetch/query/mutation helpers
│   │   ├── server/             # cookies, auth guards, server fetch helpers
│   │   └── env/                # validated env access
│   └── types/                  # truly shared cross-feature types only
```

## Build / Lint / Test Commands

### Backend
- Full test suite: `cd backend && go test ./... -count=1`
- Single package: `cd backend && go test ./internal/service -count=1`
- Single test by name: `cd backend && go test ./... -count=1 -run TestListingService_UploadImage_Success -v`
- Handler suite: `cd backend && go test ./... -count=1 -run TestListingHandlerSuite -v`
- Build: `cd backend && go build ./...`
- Vet: `cd backend && go vet ./...`
- Run app: `cd backend && go run ./cmd/property-service`
- Hot reload: `cd backend && air`
- Migrate up: `cd backend && go run ./cmd/migrate/main.go`
- Migrate down: `cd backend && go run ./cmd/migrate/main.go down`

### Frontend
- Dev server: `cd frontend && npm run dev`
- Build: `cd frontend && npm run build`
- Lint: `cd frontend && npm run lint`
- Start built app: `cd frontend && npm run start`
- Unit/component tests: `cd frontend && npm test`
- E2E tests: `cd frontend && npm run test:e2e`

## Backend Code Style Rules

### Layering
- Keep the flow: `handler -> service -> repository -> domain`.
- Handlers stay thin transport only.
- Services own business rules and orchestration.
- Repositories own DB transaction details and persistence invariants.

### Context and Errors
- In handlers, pass `c.Context()` to services.
- Do not use `context.Background()` in request flow.
- Repositories must translate `gorm.ErrRecordNotFound` to `domain.ErrNotFound`.
- Return domain errors upward; let the global Fiber error handler map HTTP status.

### Types and Data
- Money is always `int64` in IDR.
- UUIDs use `uuid.UUID`; entities commonly generate UUID v7 in hooks.
- Do not leak provider internals (Cloudinary IDs, etc.) in public response DTOs unless explicitly intended.

### Testing
- Service tests: `package service_test`, flat `TestXxx`, `testify/mock`.
- Handler tests: `testify/suite` + Postgres testcontainers.
- Cloudinary-backed image tests must use fake storage, never live Cloudinary.

## Frontend Code Style Rules

### Data and Fetching
- Use native `fetch` only.
- Centralize backend envelope parsing in one shared helper.
- Protected requests must use `credentials: 'include'`.
- Do not scatter raw `fetch` calls across page components when a shared feature helper exists.

### TanStack Query
- Use for client-side async state and mutations only.
- Query keys belong to features, not a global dumping ground.
- Prefer server-side initial data load where page/layout SSR is enough.

### Forms
- Use React Hook Form + Zod + shadcn Form for non-trivial forms.
- Co-locate schemas with the feature that owns them.
- Keep backend DTO-to-form mapping in feature helpers/adapters, not inline in many components.

### UI / Components
- `components/ui/` is for shadcn primitives only.
- Put composed business widgets in `features/*` or `components/shared/`.
- Do not put business logic into shadcn/ui primitives.

### Images
- Render backend-returned Cloudinary URLs with `next/image`.
- Configure `next.config` image remote patterns explicitly.
- In phase 1, upload images through backend multipart endpoints, not direct browser-to-Cloudinary signing.

### State
- Avoid Zustand for now.
- Prefer server data + RHF state + local component state.

## Naming and Organization

### Backend
- Keep DTO names explicit: `CreateXRequest`, `UpdateXRequest`, `XResponse`.
- Repository/service interfaces should describe the business capability, not generic storage concerns.
- Wrap DB errors with operation context.

### Frontend
- Feature folders should own:
  - query/mutation helpers
  - schemas
  - mappers/adapters
  - tests
  - composed UI
- Keep server-only helpers under server-only folders.
- Avoid giant `utils.ts` files with mixed concerns.

## Environment Rules

- Use validated env access modules; do not spread raw `process.env` usage everywhere.
- Only expose browser-safe values through `NEXT_PUBLIC_*`.
- Backend remains authoritative for auth and cookie/session behavior.
- Cloudinary frontend usage should assume backend-mediated upload and backend-returned URLs.

## Agent Guidance

- Before editing, inspect the nearest local `AGENTS.md` if present.
- Prefer small, layer-appropriate changes over broad refactors.
- Preserve backend auth contracts and cookie flow.
- Preserve frontend server-first architecture.
- If you introduce a new convention, update the closest relevant `AGENTS.md`.

## Cursor / Copilot Rules

- No `.cursor/rules/`, `.cursorrules`, or `.github/copilot-instructions.md` files currently exist in this repo.
- Treat this `AGENTS.md` as the canonical in-repo instruction source until those files are explicitly added.
```

## Notes

- This draft intentionally aligns the root guidance with the approved frontend ruleset.
- Because Prometheus is planner-only, apply the actual root `AGENTS.md` edit in an implementation session.

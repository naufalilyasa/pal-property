# AGENTS.md — pal-property

**Generated:** 2026-04-04
**Commit:** 53a0ee6
**Branch:** main

## OVERVIEW

Property listing platform for Indonesia. The Go backend remains the system-of-record API, and the Next.js frontend now includes an admin-gated dashboard workspace, protected saved listings, public listing browse/detail routes, and a floating chat assistant.

**Stack:** Go 1.26 + Fiber v3 + GORM + PostgreSQL 17 + Redis + Goth OAuth | Next.js 16 App Router + React 19 + TypeScript + Tailwind v4 + TanStack Query + RHF + Zod | Cloudinary-backed listing images via backend APIs

**Implemented:** Google OAuth with httpOnly cookie auth, refresh rotation, `/auth/me`, listing CRUD, wilayah-backed location hierarchy, category APIs, listing image upload/delete/set-primary/reorder, Casbin-backed backend authorization with DB-fresh principals, admin-only dashboard/listing create-edit-image flows, saved listings, Elasticsearch-backed public listings browse/detail, and a Gemini-backed chat assistant with clickable recommendation cards
**Infrastructure:** Search indexing follows a modular-monolith path: PostgreSQL-backed outbox writes plus the `backend/cmd/listing-indexer` worker update Elasticsearch without Redpanda/Kafka in the active runtime path. Production support includes `docker-compose.prod.yml`, `backend/Dockerfile.prod`, and env-driven auth/search/chat wiring.

## SOURCE OF TRUTH ARCHITECTURE

- **Auth authority:** backend Go is the sole auth/session authority.
- **Authorization model:** persisted roles remain `user` and `admin`; Casbin handles route/resource authorization and seller capability remains ownership-aware.
- **Frontend auth model:** use backend-issued httpOnly cookies and `/auth/me`; Auth.js is forbidden in phase 1.
- **Transport:** native `fetch` only; Axios is forbidden.
- **API envelope:** backend returns `{ success, message, data, trace_id }`; frontend normalizes this centrally.
- **Images:** uploads go through backend multipart endpoints; frontend renders backend-returned URLs with `next/image`.
- **Client state:** do not introduce browser token storage or Zustand for auth/backend state.
- **Config parsing:** backend/pkg/config leverages `caarlos0/env/v11`, decodes the AES key, and keeps OAuth provider tokens encrypted at rest.
- **Prices:** listing and commerce values stay `int64` in IDR to protect precision across services.

## STRUCTURE

```text
pal-property/
├── .sisyphus/                # plans, notepads, local session workflow
├── backend/                  # active Go service
│   ├── cmd/                  # property-service, migrate, listing-indexer, seed-demo-listings
│   ├── internal/             # handlers, services, repositories, domain, DTOs, router
│   ├── pkg/                  # config, crypto, cloudinary, middleware, utils
│   ├── db/migrations/        # golang-migrate SQL files
│   └── postman_collection.json
├── frontend/                 # App Router frontend with admin, public, saved, and chat flows
│   ├── app/                  # public, dashboard, protected, login, seller-login routes
│   ├── components/           # ui primitives + shared shells
│   ├── features/             # auth, listings, chat, categories feature slices
│   ├── lib/                  # api, env, query, server helpers
│   └── e2e/                  # Playwright browser coverage
├── plan/                     # local OAuth client secret + setup artifacts
├── docker-compose.yml        # local infra + backend + listing-indexer
└── docker-compose.prod.yml   # VPS-oriented production stack
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Backend feature work | `backend/internal/` | flow = domain -> repository -> service -> handler -> router |
| Backend search read path | `backend/internal/handler/http/search.go` + `backend/internal/service/search_read_service.go` | public Elasticsearch-backed listing search |
| Auth routes + cookies | `backend/internal/handler/http/auth.go` | OAuth callback, `/auth/me`, refresh rotation |
| Backend authz foundation | `backend/pkg/authz/` | Casbin model, enforcer wiring, authz vocabulary |
| Listing transport | `backend/internal/handler/http/listing.go` | CRUD + multipart image endpoints |
| Listing business logic | `backend/internal/service/listing_service.go` | ownership, upload, delete, reorder, primary selection |
| Listing persistence | `backend/internal/repository/postgres/listing.go` | listing + listing_images queries/transactions |
| Backend config/env | `backend/pkg/config/config.go` | `LoadConfig()` + Cloudinary env validation |
| Backend command entrypoints | `backend/cmd/` | API server, migrate, indexer, seed binaries |
| Frontend protected routes | `frontend/app/(dashboard)/` | admin dashboard shell, overview, listings, edit/create routes |
| Frontend saved listings | `frontend/app/(protected)/saved-listings/` | protected SSR list + client toggle flows |
| Frontend public listings | `frontend/app/(public)/listings/` | server-rendered browse/detail routes |
| Frontend auth helpers | `frontend/features/auth/server/` | `/auth/me` server-side gating and redirects |
| Frontend forms/images | `frontend/features/listings/forms/` + `frontend/features/listings/images/` | RHF + Zod + image mutation workflow |
| Frontend chat UI | `frontend/features/chat/` + `frontend/app/layout.tsx` | floating assistant widget + API bridge |
| Planning workflow | `.sisyphus/` | plans, notepads, drafts, local session state |
| OAuth local secret | `plan/` | sensitive local Google OAuth client JSON |

## COMMANDS

```bash
# Backend
cd backend && air
cd backend && go run ./cmd/property-service
cd backend && go run ./cmd/migrate/main.go
cd backend && go run ./cmd/migrate/main.go down
cd backend && go run ./cmd/listing-indexer
cd backend && go run ./cmd/listing-indexer rebuild
cd backend && go run ./cmd/listing-indexer rebuild-chat
cd backend && go test ./... -count=1
cd backend && go test ./... -count=1 -run TestListingHandlerSuite -v
cd backend && go build ./...
cd backend && go vet ./...

# Frontend
cd frontend && npm run dev
cd frontend && npm run build
cd frontend && npm run start
cd frontend && npm run lint
cd frontend && npm test
cd frontend && npm run test:e2e
```

## BACKEND RULES

- Keep layering strict: `handler -> service -> repository -> domain`.
- In handlers, pass `c.Context()` to services; never use `context.Background()` in request flow.
- Repositories translate `gorm.ErrRecordNotFound` into domain errors.
- Authorization stays hybrid: middleware for coarse route checks, services for ownership-sensitive resource checks.
- Money is always `int64` in IDR.
- UUIDs use `uuid.UUID`; entities commonly generate UUID v7 in hooks.
- Listing-image tests must use fake storage, never live Cloudinary.
- Listing contract is currently in a compatibility phase: new typed property fields may coexist with legacy `specifications` JSON until cleanup is explicitly planned.

## FRONTEND RULES

- Server Components are the default for routes, layouts, auth checks, and initial page data.
- Keep the root layout server-rendered; mount the Query provider in `frontend/app/providers.tsx`.
- TanStack Query is for client-owned async state, refreshes, and mutations only.
- Use RHF + Zod + shadcn-style form primitives for non-trivial forms.
- Keep `components/ui/` free of business logic; compose feature widgets in `features/*` or `components/shared/`.
- Use browser-safe env via `NEXT_PUBLIC_*`; keep server-only env helpers under `frontend/lib/env/server.ts`.
- Do not import server-only modules into client components.
- Do not store auth tokens in `localStorage`, `sessionStorage`, or readable cookies.

## TESTING RULES

- `backend/internal/service/*_test.go`: `package service_test`, `testify/mock`, flat `TestXxx`.
- `backend/internal/handler/http/*_test.go`: `testify/suite` + testcontainers Postgres, with `logger.Log = zap.NewNop()` in setup.
- `config.Env.AppEnv = "testing"` disables rate limiting in integration-test paths.
- `frontend` uses Vitest + Testing Library for unit/component coverage.
- `frontend/e2e` uses Playwright with a local mock backend for browser-level seller/public flow checks.

## NOTES

- Host Postgres port is `5433`, not `5432`.
- `backend/postman_collection.json` covers Authentication, Categories, Listings, and listing-image routes.
- `frontend` currently implements admin dashboard flows, saved listings, public listing browse/detail, and a floating chat assistant; broader buyer/transaction flows beyond discovery are still future work.
- There is no `deploy/` directory in the repo right now; production assets live at the root (`docker-compose.prod.yml`) and under `backend/` (`Dockerfile.prod`, env examples).
- `backend/cmd/listing-indexer` is the active worker-style entrypoint; there is no separate `workers/` directory.
- `.sisyphus/boulder.json` is local session state; do not treat it as product code unless the task is specifically about planning workflow.
- `plan/` contains sensitive local OAuth material; avoid printing or copying secret values into commits, docs, or issues.
- **RBAC/Casbin:** Backend RBAC now flows through Casbin, so the documented authorization focus matches the implemented behavior.
- No .cursor/rules/, .cursorrules, or .github/copilot-instructions.md files currently exist in this repo.

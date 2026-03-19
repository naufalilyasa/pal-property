# AGENTS.md — frontend/

## OVERVIEW

Next.js 16 + React 19 + TypeScript App Router frontend. Current scope includes a seller workspace (`/dashboard`, `/dashboard/listings`, create/edit/image flows), a login entry route, and public listing browse/detail routes.

## STRUCTURE

```text
frontend/
├── app/
│   ├── (dashboard)/dashboard/            # protected seller routes
│   ├── (public)/listings/                # public browse + detail routes
│   ├── login/page.tsx                    # backend-owned Google OAuth entry
│   ├── layout.tsx                        # server root layout
│   ├── providers.tsx                     # QueryClientProvider boundary
│   ├── page.tsx                          # public home shell
│   └── globals.css                       # Tailwind theme vars + global styles
├── components/
│   ├── ui/                               # shadcn-style primitives only
│   └── shared/                           # app header, sidebar, shared shells
├── features/
│   ├── auth/server/                      # current-user + require-user helpers
│   ├── categories/server/                # public category reads
│   └── listings/                         # server reads, forms, images, widgets
├── lib/
│   ├── api/                              # envelope, browser/server fetch helpers
│   ├── env/                              # validated public/server env access
│   ├── query/                            # query keys/config
│   ├── server/                           # cookie forwarding helpers
│   └── session/                          # session bootstrap helpers
├── e2e/                                  # Playwright coverage
├── vitest.config.ts
├── playwright.config.ts
└── package.json
```

## CURRENT REALITY

- Backend cookie auth is authoritative; frontend never owns canonical session state.
- Seller routes are gated server-side via `/auth/me` and redirect to `/login` when unauthenticated.
- Public listing pages fetch server-side from backend APIs.
- Listing create/edit uses RHF + Zod and backend-aligned payload shaping.
- Listing image actions upload through backend endpoints and rehydrate from backend responses.

## CONVENTIONS

- App Router only; no `pages/` directory.
- Server Components by default; add `"use client"` only for interactive islands.
- Use native `fetch` only.
- Use `browserFetch` for client mutations/queries and `serverFetch` for server-side reads.
- Keep frontend env access centralized in `frontend/lib/env/`; when adding, renaming, or removing any frontend env var, update `frontend/.env-example` in the same change.
- Keep query keys in `frontend/lib/query/keys.ts`.
- Use `components/ui/` for primitives only, not business logic.
- Keep feature-owned schemas, mappers, and composed widgets under `features/*`.

## COMMANDS

```bash
cd frontend && npm run dev
cd frontend && npm run build
cd frontend && npm run start
cd frontend && npm run lint
cd frontend && npm test
cd frontend && npm run test:e2e
```

## ANTI-PATTERNS

- **NEVER** add Axios, Auth.js, or Zustand for auth/backend state.
- **NEVER** store access or refresh tokens in localStorage or sessionStorage.
- **NEVER** move page/layout data fetching wholesale into TanStack Query.
- **NEVER** import server-only env/auth helpers into client components.
- **NEVER** add a frontend env var without updating `frontend/.env-example` to match the validated schema.
- **NEVER** bypass the shared envelope/fetch helpers when adding new API modules.

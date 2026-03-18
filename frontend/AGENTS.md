# AGENTS.md — frontend/

## OVERVIEW

Next.js 16 + React 19 + Tailwind v4 + TypeScript App Router seller workspace. It now ships a real frontend baseline for seller listing operations, including dashboard listing views, create/edit flows, and listing image actions.

## STRUCTURE

```
frontend/
├── app/
│   ├── dashboard/                         # seller dashboard routes + shared shell
│   │   ├── _components/                   # dashboard shell + listing form UI
│   │   ├── listings/new/page.tsx          # create listing route
│   │   ├── listings/[listingId]/edit/     # edit listing route + image actions
│   │   └── page.tsx                       # seller listings dashboard
│   ├── layout.tsx                         # root metadata/fonts
│   ├── page.tsx                           # seller foundation landing page
│   └── globals.css                        # Tailwind v4 import + global theme vars
├── lib/
│   ├── api/                               # envelope-aware API client + seller endpoints
│   ├── server/cookies.ts                  # Next cookies() -> Cookie header helper
│   └── session/                           # seller session bootstrap helpers
├── e2e/                                   # Playwright seller-flow smoke coverage
├── public/             # static assets
├── next.config.ts      # stock Next config
├── tsconfig.json       # strict TS + @/* path alias
├── eslint.config.mjs   # Next core-web-vitals + TS rules
├── vitest.config.ts     # unit/component test config (jsdom)
├── playwright.config.ts # browser smoke config
└── package.json        # npm scripts and dependencies
```

## CURRENT REALITY

- Seller UI exists for `/dashboard`, `/dashboard/listings/new`, and `/dashboard/listings/[listingId]/edit`.
- API access uses a shared envelope-normalizing client (`lib/api/client.ts`) and seller-specific helpers (`lib/api/seller-listings.ts`, `lib/api/listing-form.ts`).
- Session bootstrap is implemented via `/auth/me` (`lib/session/bootstrap.ts`) and server-side cookie forwarding (`lib/server/cookies.ts`).
- Listing image workflows are wired in the edit route: upload, set primary, reorder, and delete, each rehydrated from backend responses.
- Test stack is active: Vitest + Testing Library for unit/component tests, Playwright for browser smoke coverage.

## CONVENTIONS

- App Router only; no `pages/` directory.
- Server Components by default; add `"use client"` only when needed.
- Tailwind v4 is configured via `@import "tailwindcss"` in CSS.
- TypeScript is strict, with path alias `@/*` mapped to the frontend root.

## COMMANDS

```bash
cd frontend && npm run dev
cd frontend && npm run build
cd frontend && npm run start
cd frontend && npm run lint
cd frontend && npm run test
cd frontend && npm run test:e2e
```

## ANTI-PATTERNS

- **NEVER** claim marketplace or buyer flows are implemented here, this frontend scope is seller-side.
- **NEVER** store auth tokens in localStorage, backend auth is httpOnly cookie-based.
- **NEVER** bypass the shared API envelope helpers when adding new frontend API modules.
- **NEVER** spread `Headers` into plain objects when forwarding cookies, keep `Headers` instances intact.

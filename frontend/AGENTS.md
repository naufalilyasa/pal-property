# AGENTS.md — frontend/

## OVERVIEW

Next.js 16 + React 19 + Tailwind v4 + TypeScript app-router project. It is still default scaffold content, not the real property product UI yet.

## STRUCTURE

```
frontend/
├── app/
│   ├── layout.tsx      # root layout + default metadata/fonts
│   ├── page.tsx        # create-next-app landing page
│   └── globals.css     # Tailwind v4 import + global theme vars
├── public/             # static assets
├── next.config.ts      # stock Next config
├── tsconfig.json       # strict TS + @/* path alias
├── eslint.config.mjs   # Next core-web-vitals + TS rules
└── package.json        # npm scripts and dependencies
```

## CURRENT REALITY

- `app/page.tsx` is still the default create-next-app screen.
- `app/layout.tsx` still uses default `Create Next App` metadata and Geist fonts.
- No data-fetching library, state library, API client layer, or tests are installed.
- Backend auth is cookie-based, so future frontend API calls should use credentialed requests.

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
```

## ANTI-PATTERNS

- **NEVER** describe the frontend as feature-complete; it is still scaffold-level.
- **NEVER** store auth tokens in localStorage; backend is designed for httpOnly cookies.
- **NEVER** invent a frontend state/query stack in docs unless it is actually installed.

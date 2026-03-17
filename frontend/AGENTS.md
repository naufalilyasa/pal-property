# AGENTS.md — frontend/

## OVERVIEW

Next.js 16 + React 19 + Tailwind v4, TypeScript 5. App Router. Bare scaffold — no features implemented yet.

## STRUCTURE

```
frontend/
├── app/              # Next.js App Router pages
│   ├── layout.tsx    # Root layout
│   ├── page.tsx      # Home page
│   └── globals.css   # Global styles (Tailwind v4 @import)
├── public/           # Static assets
├── next.config.ts    # Next.js config
├── tsconfig.json     # Strict TS
└── eslint.config.mjs # ESLint 9 flat config
```

## CONVENTIONS

- **Tailwind v4**: configured via `@import "tailwindcss"` in CSS (no `tailwind.config.js`)
- **App Router only**: no `pages/` directory
- **Server Components by default**: add `"use client"` only when needed
- **API communication**: backend at `http://localhost:8080`, auth via httpOnly cookies (no localStorage tokens)

## NOTES

- No state management library installed yet — use React 19 native patterns first
- No data fetching library yet — consider TanStack Query when implementing features
- `NEXT_PUBLIC_API_URL` env var pattern expected for backend URL
- Auth cookies are set by backend — frontend just needs to make credentialed requests: `fetch(url, { credentials: "include" })`

## COMMANDS

```bash
cd frontend && npm run dev    # dev server (port 3000)
cd frontend && npm run build  # production build
cd frontend && npm run lint   # ESLint
```

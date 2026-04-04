# AGENTS.md — frontend/features/auth/

## OVERVIEW

Frontend auth intent, route gating, and session-aware UI shells. Backend cookies and `/auth/me` stay authoritative.

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Current user read | `server/current-user.ts` | cookie-forwarded `/auth/me` read |
| Route gating | `server/require-user.ts` | redirect + intent-aware auth checks |
| Login intent helpers | `auth-intent.ts`, `auth-destination.ts` | public vs seller/admin login routing |
| Login shells | `components/auth-entry-shell.tsx` | Google OAuth CTA + shared chrome |
| Signed-in UI | `components/user-menu.tsx` | top-nav user state |

## CONVENTIONS

- Keep auth authority on the backend; frontend only forwards cookies and reacts to `/auth/me`.
- Put redirect logic in `server/require-user.ts` and intent helpers, not scattered across routes.
- Keep login entry pages as thin shells around `AuthEntryShell`; OAuth URLs come from shared helpers.
- Use seller/admin intent when protecting dashboard access.
- Swallow anonymous/expired-session reads via typed API helpers when the UX should degrade gracefully.

## ANTI-PATTERNS

- **NEVER** store auth tokens in browser storage or readable cookies.
- **NEVER** duplicate login-path or return-to logic inside individual pages.
- **NEVER** import server auth helpers into client components.
- **NEVER** treat role `user` as dashboard-capable; dashboard access is admin-only.

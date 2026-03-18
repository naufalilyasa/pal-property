# AGENTS.md — plan/

## OVERVIEW

Sensitive local setup material. Currently holds the Google OAuth client secret JSON used for local auth wiring.

## WHERE TO LOOK

| Need | Location | Notes |
|------|----------|-------|
| Local OAuth client config | `plan/client_secret_*.json` | contains live secret material |

## CONVENTIONS

- Keep this directory local-development only.
- Reference file paths, not raw secret values, in docs or notes.
- Prefer env/config wiring elsewhere; this folder is not a general config home.

## ANTI-PATTERNS

- **NEVER** print, paste, or summarize secret values in commits or chat.
- **NEVER** add more ad hoc secrets here without an explicit local-setup reason.

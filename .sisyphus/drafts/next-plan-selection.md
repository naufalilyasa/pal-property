# Draft: Next Plan Selection

## Requirements (confirmed)
- User wants to begin the next plan.

## Technical Decisions
- No decision yet on the next plan target; user choice is required because this materially changes scope.

## Research Findings
- Main repo still describes `frontend/` as scaffold-only, but the seller-dashboard worktree shows active seller frontend flows already exist.
- Public backend APIs already exist for listings (`/api/listings`, `/api/listings/slug/:slug`) and categories (`/api/categories`).
- Protected/admin category routes already exist and depend on role enforcement.
- Planned-but-unbuilt areas called out in AGENTS files: RBAC/Casbin, Redpanda producers/consumers, Elasticsearch indexing, expanded frontend flows.
- Existing plans already cover root AGENTS refresh, listing CRUD, images, seller dashboard, and frontend ruleset.

## Open Questions
- Which major area should the next plan target?

## Scope Boundaries
- INCLUDE: selecting the next planning target and then generating a decision-complete plan for it.
- EXCLUDE: implementation work.

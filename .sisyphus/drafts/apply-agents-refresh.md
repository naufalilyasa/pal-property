# Draft: Apply Root AGENTS Refresh

## Requirements (confirmed)
- [target]: update `/mnt/data/Projects/pal-property/AGENTS.md`
- [source]: use `.sisyphus/drafts/agents-md-refresh.md` as the base content
- [alignment]: keep it aligned with the approved frontend ruleset and current backend conventions
- [scope]: implementation plan only, not direct repo mutation in this session

## Technical Decisions
- [replace-strategy]: treat the existing root `AGENTS.md` as heavily stale and prefer a near-full rewrite over piecemeal edits
- [preserve]: retain repo-specific backend command/test/convention details that are still accurate
- [omit]: do not fabricate Cursor/Copilot sections beyond noting they do not exist

## Research Findings
- [canonical]: root `AGENTS.md` is the only in-repo agent instruction source right now
- [draft-ready]: `.sisyphus/drafts/agents-md-refresh.md` already contains a candidate refreshed version
- [frontend-rules]: `.sisyphus/plans/frontend-rule-set-and-skill-set.md` locked the frontend architecture direction that AGENTS must reflect

## Open Questions
- none

## Scope Boundaries
- INCLUDE: safe rollout plan for replacing root `AGENTS.md`, verification checks, and minimal sync points with child AGENTS
- EXCLUDE: editing child AGENTS files themselves, implementing frontend stack changes, or mutating code

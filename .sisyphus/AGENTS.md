# AGENTS.md — .sisyphus/

## OVERVIEW

Local planning workspace. Holds execution plans, drafts, notepads, and Boulder session state.

## STRUCTURE

```text
.sisyphus/
├── plans/       # execution plans; often checkbox-driven
├── notepads/    # learnings, decisions, issues, problems
├── drafts/      # intermediate plan drafts and selection notes
└── boulder.json # active-plan/session pointer; local workflow state
```

## WHERE TO LOOK

| Need | Location | Notes |
|------|----------|-------|
| Active execution plan | `.sisyphus/plans/` | checkboxes are the task tracker |
| Historical learnings | `.sisyphus/notepads/` | plan-specific notes only |
| Unfinalized planning text | `.sisyphus/drafts/` | not canonical unless promoted |
| Current session state | `.sisyphus/boulder.json` | local state, usually not product history |

## CONVENTIONS

- `plans/` are the authoritative execution docs once created.
- `notepads/` are append-oriented memory, not polished docs.
- `drafts/` can be messy; promote only finalized content into plans/docs.
- Keep filenames scoped to the feature or plan they support.

## ANTI-PATTERNS

- **NEVER** treat `boulder.json` as application config.
- **NEVER** mix product requirements and scratch notes in the same file.
- **NEVER** overwrite plan history casually when a note or new draft is enough.

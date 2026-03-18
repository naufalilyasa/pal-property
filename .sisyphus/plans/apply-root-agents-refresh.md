# Apply Root AGENTS Refresh

## TL;DR
> **Summary**: Refresh `/mnt/data/Projects/pal-property/AGENTS.md` so it stays truthful to the current repo while also encoding the approved frontend architecture direction as guidance for new frontend work.
> **Deliverables**:
> - updated root `AGENTS.md`
> - consistency checks against nested AGENTS files and manifests
> - verification that commands and claims are current-state accurate
> **Effort**: Short
> **Parallel**: NO
> **Critical Path**: T1 compare root/draft/nested AGENTS -> T2 rewrite root accurately -> T3 verify commands and contradictions -> F1-F4 final review

## Context
### Original Request
Prepare a plan to safely apply the drafted AGENTS refresh to the root `AGENTS.md`, aligned with the approved frontend ruleset and existing backend conventions.

### Interview Summary
- A candidate AGENTS refresh already exists at `.sisyphus/drafts/agents-md-refresh.md`.
- The approved frontend ruleset is documented in `.sisyphus/plans/frontend-rule-set-and-skill-set.md`.
- Backend Go remains the sole auth authority; Auth.js is out of scope for phase 1.
- No Cursor/Copilot rule files exist, so root `AGENTS.md` remains the main top-level instruction source.

### Metis Review (gaps addressed)
- Corrected the false assumption that root `AGENTS.md` is the only instruction source; nested `backend/*` and `frontend/AGENTS.md` files also exist and must not be contradicted.
- Locked the rollout stance: root file must remain **current-state accurate**, while any frontend ruleset language beyond current implementation must be clearly labeled as approved direction for new frontend work.
- Added manifest validation so nonexistent commands like `npm test` are not claimed unless package scripts actually exist.
- Preserved backend-specific operational knowledge (`5433`, fake storage, test patterns, `WHERE TO LOOK`) as non-negotiable content.

## Work Objectives
### Core Objective
Safely refresh the root `AGENTS.md` so agentic coding agents receive one top-level instruction file that is accurate, useful, and consistent with the repo’s nested AGENTS files and approved frontend architecture direction.

### Deliverables
- Root `AGENTS.md` rewritten or heavily updated using the prepared draft as a base.
- Current backend commands and conventions preserved.
- Frontend architecture direction included as labeled guidance for new frontend work, not as falsely installed current fact where untrue.
- Explicit note that no Cursor/Copilot rule files currently exist.

### Definition of Done (verifiable conditions with commands)
- `test -f /mnt/data/Projects/pal-property/AGENTS.md`
- `grep -n "backend Go is the sole auth/session authority" /mnt/data/Projects/pal-property/AGENTS.md`
- `grep -n "Axios is forbidden" /mnt/data/Projects/pal-property/AGENTS.md`
- `grep -n "5433" /mnt/data/Projects/pal-property/AGENTS.md`
- `grep -n "fake storage" /mnt/data/Projects/pal-property/AGENTS.md`
- `grep -n "No \.cursor/rules" /mnt/data/Projects/pal-property/AGENTS.md`
- `node -e "const p=require('./frontend/package.json'); console.log(Boolean(p.scripts.test), Boolean(p.scripts['test:e2e']))"`

### Must Have
- Root file must describe the repo accurately as it exists now.
- Root file must preserve backend operational/testing guidance that agents rely on.
- Root file must include the approved frontend ruleset as guidance for new frontend implementation work.
- Root file must explicitly forbid Auth.js in phase 1, Axios, and browser token storage.
- Root file must mention the absence of Cursor/Copilot rule files.

### Must NOT Have (guardrails, AI slop patterns, scope boundaries)
- No contradictions with nested AGENTS files.
- No claims that frontend tooling is installed if it is not present in the target branch.
- No invented commands not backed by manifests.
- No dropping of backend test patterns, `5433` note, or Cloudinary fake-storage guidance.

## Verification Strategy
> ZERO HUMAN INTERVENTION - all verification is agent-executed.
- File-content checks via `grep` for required guidance.
- Manifest checks via `frontend/package.json` and `backend/go.mod`.
- Consistency check against nested AGENTS files using read-only comparison.
- Evidence: `.sisyphus/evidence/task-{N}-{slug}.{ext}`

## Execution Strategy
### Parallel Execution Waves
> This is a small docs rollout; execute sequentially to avoid contradiction.

Wave 1: T1 source comparison, T2 root rewrite, T3 verification

### Dependency Matrix (full, all tasks)
- T1 blocks T2 and T3
- T2 blocks T3

### Agent Dispatch Summary (wave -> task count -> categories)
- Wave 1 -> 3 tasks -> `implementation`, `writing`, `review`

## TODOs

- [x] T1. Compare root draft against actual repo truth and nested AGENTS files

  **What to do**: Compare `/mnt/data/Projects/pal-property/AGENTS.md`, `.sisyphus/drafts/agents-md-refresh.md`, `frontend/AGENTS.md`, `backend/AGENTS.md`, `backend/internal/AGENTS.md`, `backend/internal/service/AGENTS.md`, `backend/internal/repository/AGENTS.md`, `backend/internal/handler/http/AGENTS.md`, `backend/internal/domain/AGENTS.md`, and `backend/pkg/AGENTS.md`. Identify which draft sections are safe to lift directly, which backend sections must be preserved verbatim or near-verbatim, and which frontend stack claims must be rewritten as “approved direction” rather than “already installed reality”.
  **Must NOT do**: Do not edit any files in this step.

  **Recommended Agent Profile**:
  - Category: `review` - Reason: high-signal comparison across instruction sources
  - Skills: `[]`
  - Omitted: `[]`

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: T2, T3 | Blocked By: none

  **References**:
  - Pattern: `/mnt/data/Projects/pal-property/AGENTS.md`
  - Pattern: `.sisyphus/drafts/agents-md-refresh.md`
  - Pattern: `frontend/AGENTS.md`
  - Pattern: `backend/AGENTS.md`
  - Pattern: `backend/internal/service/AGENTS.md`
  - Pattern: `backend/internal/handler/http/AGENTS.md`

  **Acceptance Criteria**:
  - [x] A comparison map exists for root draft vs nested AGENTS content
  - [x] Contradictory sections are explicitly identified before editing root `AGENTS.md`

  **QA Scenarios**:
  ```bash
  Scenario: Nested AGENTS sources are all enumerated
    Tool: Bash
    Steps: ls backend/AGENTS.md backend/internal/AGENTS.md backend/internal/service/AGENTS.md backend/internal/repository/AGENTS.md backend/internal/handler/http/AGENTS.md backend/internal/domain/AGENTS.md backend/pkg/AGENTS.md frontend/AGENTS.md
    Expected: all listed files exist
    Evidence: .sisyphus/evidence/task-T1-nested-agents.txt

  Scenario: Draft source exists
    Tool: Bash
    Steps: test -f .sisyphus/drafts/agents-md-refresh.md
    Expected: exits 0
    Evidence: .sisyphus/evidence/task-T1-draft-exists.txt
  ```

  **Commit**: NO | Message: `docs: compare root and nested agent guidance` | Files: none

- [x] T2. Rewrite root `AGENTS.md` with current-truth plus approved frontend direction

  **What to do**: Rewrite `/mnt/data/Projects/pal-property/AGENTS.md` using `.sisyphus/drafts/agents-md-refresh.md` as the base, but adjust every section so it is current-state accurate. Preserve backend commands, testing conventions, `WHERE TO LOOK`, Postgres port `5433`, Cloudinary fake-storage testing rule, and current backend auth ownership. Include the approved frontend ruleset under language such as “for new frontend implementation work” or “approved frontend direction”, not as falsely installed fact if the target branch has not yet adopted those tools.
  **Must NOT do**: Do not rewrite nested AGENTS files in this plan; this is root-only.

  **Recommended Agent Profile**:
  - Category: `writing` - Reason: precise documentation rewrite with architecture nuance
  - Skills: `[]`
  - Omitted: `[]`

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: T3 | Blocked By: T1

  **References**:
  - Pattern: `.sisyphus/drafts/agents-md-refresh.md`
  - Pattern: `/mnt/data/Projects/pal-property/AGENTS.md`
  - Pattern: `.sisyphus/plans/frontend-rule-set-and-skill-set.md`
  - Pattern: `frontend/package.json`
  - Pattern: `backend/go.mod`

  **Acceptance Criteria**:
  - [x] Root `AGENTS.md` is updated and remains about 150 lines
  - [x] Root `AGENTS.md` preserves repo-accurate backend commands and conventions
  - [x] Frontend rules are framed as approved implementation direction where necessary, not false installed fact

  **QA Scenarios**:
  ```bash
  Scenario: Root file includes mandatory repo guardrails
    Tool: Bash
    Steps: grep -n "backend Go is the sole auth/session authority" AGENTS.md && grep -n "Axios is forbidden" AGENTS.md && grep -n "5433" AGENTS.md
    Expected: all three grep commands match
    Evidence: .sisyphus/evidence/task-T2-root-guardrails.txt

  Scenario: Root file records missing Cursor/Copilot rule files accurately
    Tool: Bash
    Steps: grep -n "No \.cursor/rules" AGENTS.md
    Expected: grep matches a line noting absence of Cursor/Copilot rules
    Evidence: .sisyphus/evidence/task-T2-editor-rule-note.txt
  ```

  **Commit**: YES | Message: `docs: refresh root agent guidance` | Files: `AGENTS.md`

- [x] T3. Verify root guidance against manifests and nested instruction sources

  **What to do**: Verify that every command and stack claim in the refreshed root file is true for the target branch. Specifically check frontend scripts before mentioning `npm test` or `npm run test:e2e`; if they do not exist, phrase the frontend rules as target direction instead of present command. Check that root guidance does not contradict nested backend/frontend AGENTS files on auth authority, handler context (`c.Context()`), or test conventions.
  **Must NOT do**: Do not accept “close enough” wording; false commands or contradictions fail the task.

  **Recommended Agent Profile**:
  - Category: `review` - Reason: consistency and truthfulness gate
  - Skills: `[]`
  - Omitted: `[]`

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: none | Blocked By: T1, T2

  **References**:
  - Pattern: `frontend/package.json`
  - Pattern: `backend/AGENTS.md`
  - Pattern: `frontend/AGENTS.md`
  - Pattern: `backend/internal/handler/http/AGENTS.md`
  - Pattern: `/mnt/data/Projects/pal-property/AGENTS.md`

  **Acceptance Criteria**:
  - [x] No root command examples are invalid for the target branch
  - [x] No root statements contradict nested AGENTS guidance on current repo truth
  - [x] Root file clearly distinguishes current state from approved frontend direction where needed

  **QA Scenarios**:
  ```bash
  Scenario: Frontend command claims match package scripts
    Tool: Bash
    Steps: node -e "const p=require('./frontend/package.json'); console.log(Object.keys(p.scripts).sort().join(','))"
    Expected: output matches the commands claimed as current in root AGENTS.md
    Evidence: .sisyphus/evidence/task-T3-frontend-scripts.txt

  Scenario: Root file still preserves backend testing guidance
    Tool: Bash
    Steps: grep -n "package service_test" AGENTS.md && grep -n "testify/suite" AGENTS.md && grep -n "fake storage" AGENTS.md
    Expected: all three grep commands match
    Evidence: .sisyphus/evidence/task-T3-backend-test-guidance.txt
  ```

  **Commit**: NO | Message: `docs: verify root agent guidance accuracy` | Files: none

## Final Verification Wave (4 parallel agents, ALL must APPROVE)
- [x] F1. Plan Compliance Audit - oracle
- [x] F2. Docs Quality Review - reviewer
- [x] F3. Practicality Review - tester
- [x] F4. Scope Fidelity Check - deep

## Commit Strategy
- One doc commit only: root `AGENTS.md` refresh. Do not bundle nested AGENTS edits into the same change unless a follow-up plan explicitly calls for them.

## Success Criteria
- Root `AGENTS.md` becomes a reliable top-level guide for agentic coding agents in this repo.
- The file reflects both current backend truth and the approved frontend architecture direction without claiming nonexistent tooling as already installed.
- Root guidance no longer drifts from the backend/frontend conventions that actually exist in the repository.

# Plan: Update AGENTS.md with Missing Rules from AI_CONTEXT.md Review

**Generated:** 2026-03-04
**Scope:** Root `AGENTS.md` only — no code changes

## Context

`AI_CONTEXT.md` was reviewed against the live `AGENTS.md`. Several rules present in `AI_CONTEXT.md` are
either outdated (Viper, kafka-go, Casbin, imaging/ffmpeg, dedicated test DB) and must NOT be carried
forward, or are genuinely valuable conventions already used in the codebase but not yet documented in
`AGENTS.md`. This plan captures only the latter.

No code changes. No new files. One file updated: `/mnt/data/Projects/pal-property/AGENTS.md`.

---

## Task 1 — Fix stale description line (line 8)

**File:** `AGENTS.md`
**What:** Line 8 says "only auth is implemented end-to-end". Auth + Listing CRUD are now both done.
**Change:** Update the description sentence.

```
Property listing platform (Indonesia). Go REST API backend, Next.js frontend, event-driven workers (Redpanda/Kafka). Auth + Listing CRUD are fully implemented end-to-end.
```

**QA:** Read line 8 after edit — must not contain "only auth".

---

## Task 2 — Update stack line + add Config/Implemented/Planned lines (after line 10)

**File:** `AGENTS.md`
**What:** Stack line missing Sonic. No mention of caarlos0/env. No summary of what's implemented vs planned.
**Change:** Replace the `**Stack:**` line and add 2 new lines beneath it.

```markdown
**Stack:** Go 1.26 + Fiber v3 (Sonic JSON) + GORM + PostgreSQL 17 | Next.js 16 + React 19 + Tailwind v4 | Redpanda (Kafka-compat) | Redis | Elasticsearch 8

**Config:** `caarlos0/env` v11 (struct tags, no Viper). All config lives in `backend/pkg/config/config.go` as a single `Config` struct.

**Implemented:** Auth (OAuth2/Google, JWT RS256, refresh rotation), Listing CRUD (create/read/update/delete/list/filter, soft delete, view count)
**Planned (not yet impl):** RBAC/Casbin, Redpanda producers/consumers, Elasticsearch indexing, Image upload (S3/R2)
```

**QA:** Read updated block — must contain "Sonic JSON", "caarlos0/env", "Implemented:", "Planned:".

---

## Task 3 — Add listing integration test run commands (COMMANDS section)

**File:** `AGENTS.md`
**What:** COMMANDS section only shows auth integration test run command. Listing suite command missing.
**Change:** After the auth integration test command line, append:

```bash
cd backend && go test ./... -count=1 -run TestListingHandlerSuite -v  # listing integration
```

**QA:** Read COMMANDS section — both `TestAuthHandlerSuite` and `TestListingHandlerSuite` must be present.

---

## Task 4 — Expand KEY CONVENTIONS section

**File:** `AGENTS.md`
**What:** Missing 5 important runtime conventions already followed in code but not documented.
**Change:** After the existing 5 bullet points in KEY CONVENTIONS, append 5 more:

```markdown
- **JSON encoder**: Fiber is initialized with `sonic.Marshal` / `sonic.Unmarshal` — do not swap to `encoding/json`.
- **Context propagation**: always pass `c.UserContext()` from handler to service — never `context.Background()`.
- **Goroutine safety**: always copy variables before passing to goroutines (zero-allocation, avoid closure capture bugs).
- **Zap logger**: dev → stdout + file (`tmp/logs/app.log`); prod → stdout only. Tests → `logger.Log = zap.NewNop()`.
- **Rate limiter**: skipped automatically when `config.Env.AppEnv == "testing" || "development"` (see `router.go`).
```

**QA:** Read KEY CONVENTIONS — all 10 bullets must be present.

---

## Task 5 — Expand ANTI-PATTERNS section

**File:** `AGENTS.md`
**What:** 3 critical anti-patterns missing: context.Background() in handlers, unsafe goroutine capture, using Kafka testcontainer, and Viper.
**Change:** After the existing 6 NEVER bullets, append 4 more:

```markdown
- **NEVER** use `context.Background()` in handlers — always `c.UserContext()` so request cancellation propagates.
- **NEVER** pass goroutine captures without copying first — loop vars and Fiber ctx are unsafe across goroutine boundaries.
- **NEVER** spin up Redpanda/Kafka testcontainers in integration tests — mock the producer/consumer interface instead.
- **NEVER** use `viper` — project uses `caarlos0/env` for all config (struct tags on `pkg/config/config.go`).
```

**QA:** Read ANTI-PATTERNS — must contain 10 NEVER bullets total.

---

## Task 6 — Add TESTING RULES section

**File:** `AGENTS.md`
**What:** No dedicated testing section. Rules scattered or missing entirely.
**Where to insert:** After ANTI-PATTERNS section, before NOTES section.
**Change:** Insert new section:

```markdown
## TESTING RULES

- **Pyramid**: unit tests (mocks via `testify/mock`) + integration tests (real DB via `testcontainers-go`).
- **Target coverage**: 70–80% meaningful coverage. No coverage-padding tests.
- **Unit tests**: flat `TestXxx` functions, package `*_test`, use `mocks.NewXxx(t)` constructors.
- **Integration tests**: `testify/suite`, spin up `postgres:17-alpine` (and `redis:8.2-alpine` if needed) via testcontainers.
- **Test setup required** in `SetupSuite`:
  ```go
  logger.Log = zap.NewNop()      // silence logs
  config.Env.AppEnv = "testing"  // bypass rate limiter
  ```
- **Kafka/Redpanda**: always mock in tests — do NOT use testcontainer for message broker.
- **Truncate between tests**: `SetupTest` must TRUNCATE all relevant tables + flush Redis to ensure test isolation.
- **Run integration tests**: `go test ./... -count=1 -run TestXxxSuite -v -timeout 120s`
```

**QA:** `TESTING RULES` section must exist and contain all 8 bullets.

---

## Task 7 — Add NEW FEATURE WORKFLOW section

**File:** `AGENTS.md`
**What:** No documented workflow for adding new features. Essential for AI agent onboarding.
**Where to insert:** After TESTING RULES, before NOTES.
**Change:** Insert new section:

```markdown
## NEW FEATURE WORKFLOW

For every new backend feature, follow this order:
1. **Entity** — add/update struct in `domain/entity/`
2. **Domain interface** — add methods to `domain/xxx_repository.go`
3. **Repository** — implement in `repository/postgres/xxx.go`
4. **DTO** — add request/response in `dto/request/` and `dto/response/`
5. **Service** — implement business logic in `service/xxx_service.go`
6. **Handler** — add HTTP handler in `handler/http/xxx.go`
7. **Router** — register routes in `router/router.go`
8. **Config** — if new env var needed, update `pkg/config/config.go` + `.env-example`
9. **Unit tests** — `service/xxx_service_test.go` (mocks)
10. **Integration tests** — `handler/http/xxx_test.go` (testcontainers)
```

**QA:** `NEW FEATURE WORKFLOW` section must exist with 10 numbered steps.

---

## Task 8 — Update NOTES section

**File:** `AGENTS.md`
**What:** Missing note about RBAC being planned but not implemented. Important for future agents.
**Change:** Append to the end of NOTES:

```markdown
- RBAC (Casbin) is planned but not yet implemented — current auth is simple role string on User entity (`role varchar default 'user'`)
```

**QA:** NOTES section must contain the RBAC note.

---

## Final Verification Wave

After all edits:
1. Run `cat AGENTS.md` — verify no duplicate sections, no broken markdown
2. Verify line count is greater than the original (was ~103 lines, should be ~150+)
3. Confirm no mention of "Viper" (except in ANTI-PATTERNS NEVER bullet), no mention of "kafka-go", "casbin", "imaging", "ffmpeg" as active libraries

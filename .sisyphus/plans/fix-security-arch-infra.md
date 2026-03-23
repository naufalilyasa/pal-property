# Fix: Residual Backend Security, Config, and Infra Drift

## TL;DR
> **Summary**: The original broad security/infra plan is stale because most target items are already implemented. This revised plan focuses only on the residual drift still present in the current repo: DB SSL env-key mismatch, auth-service error semantics cleanup, env/doc consistency, and verification that already-landed security/infra fixes remain correct.
> **Deliverables**:
> - config/env key alignment for DB SSL mode
> - backward-compatible OAuth token read path plus encrypted-token migration safety
> - auth-service error cleanup to use domain-aware semantics instead of ad hoc string errors where appropriate
> - env/docs consistency audit for current backend config reality
> - focused verification for token encryption, domain error translation, CORS, Redpanda, and price storage invariants
> **Effort**: Medium
> **Parallel**: YES - 2 waves
> **Critical Path**: 1 -> 2 -> 3 -> 4 -> F1-F4

## Context
### Original Request
Plan `fix-security-arch-infra`.

### Current Repo Reality
- OAuth provider tokens are already AES-256-GCM encrypted in `backend/internal/repository/postgres/auth.go` using `backend/pkg/crypto/aes.go`.
- Domain sentinel errors already exist in `backend/internal/domain/errors.go`, and repositories already translate `gorm.ErrRecordNotFound` in the main postgres repositories.
- CORS already includes `AllowMethods` and `AllowCredentials` in `backend/internal/router/router.go`.
- Docker already uses Redpanda in `docker-compose.yml`.
- Listing price is already `int64` in `backend/internal/domain/entity/listing.go` and the listing DTOs.
- Config already uses `github.com/caarlos0/env/v11` in `backend/pkg/config/config.go`.

### Why the Old Plan Is Stale
- It treated six major migrations as greenfield work even though most of them already landed.
- It would risk unnecessary rewrites of working code (`pkg/crypto`, `domain/errors`, `docker-compose`, listing money fields, config parser).
- It did not account for post-RBAC repo state or the current AGENTS/docs baseline.

## Work Objectives
### Core Objective
Bring the backend’s remaining security/config/infrastructure drift to a fully consistent state without redoing already-completed migrations.

### Deliverables
- One consistent DB SSL env key across config code, examples, and local env usage.
- One safe read path for legacy plaintext OAuth rows while keeping encrypt-on-write behavior for new/updated rows.
- Auth-service error returns aligned with domain-aware handling and current repository semantics.
- AGENTS/env/docs updated to reflect current backend reality.
- Focused verification proving the already-landed security/infra fixes still hold.

### Definition of Done (verifiable conditions with commands)
- `grep -R 'DB_SSLMODE' backend pkg . -n` returns no stale config key usage in tracked repo files.
- `grep -R 'DB_SSL_MODE' backend/.env-example backend/pkg/config/config.go -n` returns aligned config/example usage.
- `cd backend && go build ./...` exits `0`.
- `cd backend && go test ./internal/repository/postgres -count=1 -run 'TestAuthRepository_' -v` exits `0` if repository auth tests are added; otherwise equivalent focused auth repository/service coverage passes.
- `cd backend && go test ./internal/service -count=1 -run 'TestGetMe_|TestRefreshToken_|TestCompleteAuth_' -v` exits `0`.

### Must Have
- Preserve current AES-GCM token encryption behavior.
- Preserve encrypt-on-write while safely handling legacy plaintext rows until they are migrated.
- Preserve current `int64` money storage.
- Preserve current CORS credential/method behavior.
- Preserve current Redpanda docker topology.
- Preserve current `caarlos0/env` config parser.
- Remove config drift and stale error-semantics mismatches only.

### Must NOT Have
- No reintroduction of Viper.
- No rollback from Redpanda to Kafka/Zookeeper.
- No price type regression to float/decimal-in-float handling.
- No plaintext OAuth token storage path.
- No broad rewrite of already-correct repo layers just to satisfy the original stale plan wording.

## Verification Strategy
> ZERO HUMAN INTERVENTION — all verification is agent-executed.
- Tests-after policy with focused build/grep/test checks.
- Prefer existing auth/config/repository/service tests over inventing broad new suites unless coverage is actually missing.

## Execution Strategy
### Parallel Execution Waves
Wave 1: config drift + token-read compatibility.
Wave 2: auth-service semantics, then docs/env sync, then focused verification.

### Dependency Matrix
- 1 blocks 3, 4, and 5.
- 2 blocks 4 and 5.
- 3 blocks 4 and 5.
- 4 blocks F1-F4.
- 5 blocks F1-F4.

## TODOs

- [x] 1. Fix DB SSL env-key drift and config-source consistency

  **What to do**: Align the DB SSL env key used in `backend/pkg/config/config.go` with the documented/example key in `backend/.env-example` and any tracked env/config references. Prefer the explicit `DB_SSL_MODE` spelling already documented in `.env-example` unless repo-wide evidence strongly supports a different canonical name. Keep config loading behavior via `caarlos0/env` and `godotenv.Load()` unchanged apart from the key alignment.
  **Must NOT do**: Do not reintroduce Viper, and do not invent parallel aliases unless there is a strong backward-compatibility reason explicitly implemented and documented.

  **Acceptance Criteria**:
  - [x] `grep -R 'DB_SSLMODE' backend pkg . -n` returns no stale tracked usage.
  - [x] `grep -R 'DB_SSL_MODE' backend/.env-example backend/pkg/config/config.go -n` returns aligned usage.
  - [x] `cd backend && go build ./...` exits `0`.

- [x] 2. Add backward-compatible OAuth token read handling for legacy plaintext rows

  **What to do**: Update the OAuth account read/write path so new writes stay encrypted but reads do not hard-fail if older database rows still contain plaintext provider tokens. Prefer a deliberate compatibility strategy in `backend/internal/repository/postgres/auth.go`: try decrypt, detect likely plaintext safely, and preserve a migration/backfill path rather than crashing live auth. Add focused tests around encrypted and plaintext reads plus invalid-ciphertext behavior.
  **Must NOT do**: Do not remove encryption-on-write, do not silently swallow malformed ciphertext as valid plaintext, and do not require a blocking one-shot data migration before deploy.

  **Acceptance Criteria**:
  - [x] `cd backend && go test ./internal/repository/postgres -count=1 -run 'TestAuthRepository_' -v` exits `0` if repo tests are added; otherwise equivalent focused auth repository/service coverage passes.
  - [x] `grep -n 'crypto.Decrypt' backend/internal/repository/postgres/auth.go` still shows decrypt-on-read behavior, but with guarded compatibility for legacy plaintext rows.
  - [x] `cd backend && go build ./...` exits `0`.

- [x] 3. Clean up auth-service error semantics around already-domain-aware repositories

  **What to do**: Review `backend/internal/service/auth_service.go` and replace ad hoc string-only error shaping where it obscures domain behavior, especially around `FindUserByID`, OAuth-account lookup consistency, refresh-token validation/cache failures, and token-generation/cache errors. Keep user-facing semantics clear, but avoid losing the ability to reason about `domain.ErrNotFound`, `domain.ErrUnauthorized`, and related paths in tests/handlers.
  **Must NOT do**: Do not leak raw repository/driver errors directly to handlers, and do not widen the auth API surface.

  **Acceptance Criteria**:
  - [x] `grep -n 'errors.New("user not found"' backend/internal/service/auth_service.go` returns no stale ad hoc not-found mapping if replaced by a stronger domain-aware path.
  - [x] `cd backend && go test ./internal/service -count=1 -run 'TestGetMe_|TestRefreshToken_|TestCompleteAuth_' -v` exits `0`.
  - [x] `cd backend && go build ./...` exits `0`.

- [x] 4. Sync docs and env guidance with current backend security/infra reality

  **What to do**: Update `AGENTS.md`, `backend/AGENTS.md`, `backend/pkg/AGENTS.md`, and tracked env guidance files that actually exist in the repo so they describe the current repo truth: Redpanda is already the broker, prices are `int64`, tokens are encrypted, config uses `caarlos0/env`, and only the residual config/auth-service cleanup remains.
  **Must NOT do**: Do not leave docs claiming planned work that is already implemented.

  **Acceptance Criteria**:
  - [x] `grep -n 'Redpanda' AGENTS.md backend/AGENTS.md` returns current-state wording, not future-state wording.
  - [x] `grep -n 'caarlos0/env\|encrypted' backend/pkg/AGENTS.md AGENTS.md` returns current-state wording.
  - [x] `grep -n 'RBAC/Casbin' AGENTS.md` does not regress current implemented/planned status while editing adjacent docs.
  - [x] `grep -R 'DB_SSL_MODE\|OAUTH_TOKEN_ENCRYPTION_KEY' backend/.env-example backend/pkg/config/config.go -n` returns current aligned guidance.

- [x] 5. Add focused verification and residual regression checks for the already-landed fixes

  **What to do**: Add or tighten only the missing focused tests/assertions needed to prove the current fixes remain true: encrypted OAuth token persistence path, repo not-found translation, auth-service semantics, DB SSL config load, and docker compose broker assumptions if needed via static checks.
  **Must NOT do**: Do not create a giant new integration matrix when focused static/service/repository checks are enough.

  **Acceptance Criteria**:
  - [x] `cd backend && go build ./...` exits `0`.
  - [x] `cd backend && go test ./internal/service -count=1 -run 'TestGetMe_|TestRefreshToken_|TestCompleteAuth_' -v` exits `0`.
  - [x] `cd backend && go test ./internal/repository/postgres -count=1 -run 'TestAuthRepository_' -v` exits `0` if repository auth tests are added.
  - [x] `grep -n 'redpanda:' docker-compose.yml` returns current broker config.
  - [x] `grep -n 'Price    int64' backend/internal/domain/entity/listing.go` returns the integer-money invariant.

## Final Verification Wave (4 parallel agents, ALL must APPROVE)
- [x] F1. Plan Compliance Audit — oracle
- [x] F2. Code Quality Review — reviewer
- [x] F3. Real Runtime QA — tester
- [x] F4. Scope Fidelity Check — deep

## Success Criteria
- The plan reflects current repo reality instead of stale migrations that are already done.
- Backend config/docs/error paths are internally consistent.
- No already-completed security/infra improvement is accidentally reverted during cleanup.

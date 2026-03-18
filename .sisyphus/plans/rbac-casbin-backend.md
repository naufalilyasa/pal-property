# RBAC Casbin Backend

## TL;DR
> **Summary**: Introduce Casbin as the backend authorization engine without changing the current auth/session authority or expanding persisted roles beyond `user` and `admin`. Keep route-level checks coarse in middleware, keep ownership-sensitive checks in services, and migrate the current `RequireRole("admin")` behavior behind a shared authz layer.
> **Deliverables**:
> - Casbin model + Postgres-backed policy storage integrated into the Go backend
> - Shared principal/authz service used by middleware and services
> - Category route auth migrated from exact role checks to policy checks
> - Listing owner/admin authorization migrated to shared authz helpers
> - Parity-preserving tests for role freshness, route checks, and ownership checks
> - Docs/config updates that keep rollout and local setup explicit
> **Effort**: Large
> **Parallel**: YES - 3 waves
> **Critical Path**: 1 -> 2 -> 3 -> 4 -> 5 -> 7 -> 8 -> F1-F4

## Context
### Original Request
Create the next decision-complete plan for `RBAC/Casbin backend`.

### Interview Summary
- Backend Go remains the only auth/session authority.
- Persisted roles stay `user` and `admin` for phase 1.
- Seller capability remains ownership-aware instead of becoming a stored `seller` role.
- Casbin should not replace authentication; it only replaces/coheres authorization.
- Current route-level admin gating and listing ownership checks must be preserved during rollout.

### Metis Review (gaps addressed)
- Preserve current role freshness semantics by continuing to read the current DB role during protected requests.
- Do not use adapter `AutoMigrate`; create Casbin policy storage through `golang-migrate` SQL.
- Treat enforcer/model/adapter boot failures as startup-fatal, never fail-open.
- Keep runtime policy management, watcher sync, UI policy editors, and multi-tenant/domain RBAC out of phase 1.
- Remove silent handler fallbacks like implicit `"user"` when role/principal wiring is missing.

## Work Objectives
### Core Objective
Add a production-safe, decision-complete authorization layer to the backend using Casbin, while preserving existing cookie-auth behavior, immediate role freshness, admin-only category writes, and owner/admin listing mutation semantics.

### Deliverables
- Casbin model source of truth in the repo.
- Postgres-backed Casbin policy table and seeded phase-1 policy rows.
- Shared principal/authz layer consumed by middleware and services.
- Route-level permission middleware for category admin routes.
- Service-level owner/admin authorization helper for listing mutations.
- Updated tests proving unchanged behavior plus new authz coverage.
- Updated backend/root guidance describing the new authz boundaries.

### Definition of Done (verifiable conditions with commands)
- `test -f backend/pkg/authz/model.conf` exits `0`.
- `test -f backend/db/migrations/000006_casbin_policy.up.sql` exits `0`.
- `grep -n 'RequireRole("admin")' backend/internal/router/router.go` returns no matches for category write routes after migration.
- `cd backend && go build ./...` exits `0`.
- `cd backend && go test ./internal/handler/http -count=1 -run TestCategoryHandlerSuite -v` exits `0`.
- `cd backend && go test ./internal/handler/http -count=1 -run TestListingHandlerSuite -v` exits `0`.

### Must Have
- Keep persisted roles to `user` and `admin` only.
- Keep backend DB role lookup authoritative during protected requests.
- Keep category create/update/delete admin-only.
- Keep listing update/delete/image mutations owner-or-admin.
- Seed phase-1 policy rows without requiring a runtime policy management API.
- Use one shared authz action/object vocabulary everywhere.

### Must NOT Have (guardrails, AI slop patterns, scope boundaries)
- No persisted `seller` role in phase 1.
- No Casbin replacement for authentication or JWT issuance.
- No startup `AutoMigrate` for Casbin policy storage.
- No watcher/pubsub synchronization in phase 1.
- No runtime policy CRUD endpoints or admin UI.
- No duplicated authorization logic between middleware and services without a shared authz service.
- No fail-open behavior on enforcer/adapter/model errors.

## Verification Strategy
> ZERO HUMAN INTERVENTION — all verification is agent-executed.
- Test decision: tests-after using existing Go service tests and handler integration suites.
- QA policy: every task includes explicit allow + deny/edge coverage.
- Evidence: `.sisyphus/evidence/task-{N}-{slug}.{ext}`.

## Execution Strategy
### Parallel Execution Waves
> Target: 5-8 tasks per wave. Extract shared dependencies into Wave 1.

Wave 1: foundation/authz package/migration/boot wiring.
Wave 2: route + service integration and principal refactor.
Wave 3: verification, parity tests, and docs cleanup.

### Dependency Matrix (full, all tasks)
- 1 blocks 2, 3, 4, 7, 9.
- 2 blocks 3, 4, 5, 6, 7, 8.
- 3 blocks 7, 8.
- 4 blocks 5, 7, 8.
- 5 blocks 7, 8.
- 6 blocks 7.
- 7 blocks 8, 9.
- 8 blocks 9.
- 9 blocks F1-F4.

### Agent Dispatch Summary (wave -> task count -> categories)
- Wave 1 -> 3 tasks -> implementation, implementation, testing.
- Wave 2 -> 3 tasks -> implementation, implementation, refactorer.
- Wave 3 -> 3 tasks -> testing, testing, doc-writer.
- Final Verification -> 4 tasks -> oracle, unspecified-high, unspecified-high, deep.

## TODOs
> Implementation + Test = ONE task. Never separate.
> EVERY task MUST have: Agent Profile + Parallelization + QA Scenarios.

- [x] 1. Add the Casbin foundation package, model file, adapter dependency, and policy-table migration

  **What to do**: Create a dedicated authorization package at `backend/pkg/authz/` with the Casbin enforcer wiring, subject/object/action vocabulary, and a repo-versioned `model.conf` source of truth. Add the chosen Go Casbin dependency plus a GORM/Postgres adapter in `backend/go.mod`. Create a new SQL migration pair at `backend/db/migrations/000006_casbin_policy.{up,down}.sql` that creates the Casbin rule table and seeds only the minimum phase-1 policy rows required for admin category writes. The migration must be the only place that creates policy storage; do not rely on adapter auto-creation.
  **Must NOT do**: Do not add runtime policy management APIs, do not add watcher infrastructure, and do not add a stored `seller` role.

  **Recommended Agent Profile**:
  - Category: `implementation` — Reason: foundation package + migration + dependency wiring.
  - Skills: `[]` — No extra skill is required.
  - Omitted: `[]` — No additional skills are intentionally required.

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: 2, 3, 4, 7, 9 | Blocked By: none

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `backend/go.mod` — existing dependency management style.
  - Pattern: `backend/db/migrations/000004_category_fk_set_null_and_seed.up.sql` — migration naming + seed style.
  - Pattern: `backend/db/migrations/000005_listing_images_cloudinary_metadata.up.sql` — latest migration sequencing.
  - Pattern: `backend/pkg/AGENTS.md` — `pkg/` is reusable support code, not `internal/` code.
  - API/Type: `backend/internal/domain/entity/user.go` — persisted role remains a single string field.
  - External: `https://github.com/casbin/casbin/blob/098508551fe30a208aafc7732b19c779fbc4059d/README.md#L63-L119` — PERM model structure.
  - External: `https://github.com/casbin/casbin/blob/098508551fe30a208aafc7732b19c779fbc4059d/persist/adapter.go#L24-L79` — adapter persistence contract.

  **Acceptance Criteria** (agent-executable only):
  - [x] `test -f backend/pkg/authz/model.conf && test -f backend/db/migrations/000006_casbin_policy.up.sql && test -f backend/db/migrations/000006_casbin_policy.down.sql` exits `0`.
  - [x] `grep -n 'casbin' backend/go.mod` returns at least one dependency match.
  - [x] `grep -n 'AutoMigrate' backend/pkg/authz backend/cmd/property-service backend/db/migrations` returns no Casbin-related automigration usage.
  - [x] `grep -n 'p, admin' backend/db/migrations/000006_casbin_policy.up.sql` returns at least one seeded admin policy row.
  - [x] `cd backend && go build ./...` exits `0`.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```
  Scenario: Enforcer foundation compiles and policy schema exists
    Tool: Bash
    Steps: Run `cd backend && go build ./...`; then assert the model and migration files exist.
    Expected: Build succeeds and the model/migration files are present.
    Evidence: .sisyphus/evidence/task-1-casbin-foundation.txt

  Scenario: Policy storage is migration-managed, not auto-created
    Tool: Bash
    Steps: Search for `AutoMigrate` in Casbin/authz wiring paths and inspect the new migration file for table creation.
    Expected: No runtime auto-migration path exists; policy schema is defined in SQL migration only.
    Evidence: .sisyphus/evidence/task-1-casbin-foundation-error.txt
  ```

  **Commit**: YES | Message: `feat(backend): add casbin authz foundation` | Files: `backend/go.mod`, `backend/go.sum`, `backend/pkg/authz/**`, `backend/db/migrations/000006_casbin_policy.*`

- [x] 2. Wire principal loading and enforcer lifecycle into startup and protected middleware

  **What to do**: Extend backend startup wiring in `backend/cmd/property-service/main.go` so the authz package, adapter, and model load at boot. Boot must fail fast if the model or adapter cannot initialize. Evolve `backend/pkg/middleware/auth.go` from “JWT + role loader” into a principal loader that still reads the current DB role on every protected request, keeps role freshness semantics intact, and stores a structured principal in Fiber locals alongside temporary compatibility values during rollout.
  **Must NOT do**: Do not move role authority into JWT claims or caches, and do not leave a code path that silently invents a default role when principal loading fails.

  **Recommended Agent Profile**:
  - Category: `implementation` — Reason: DI, middleware, and principal lifecycle are cross-cutting.
  - Skills: `[]` — No extra skill is required.
  - Omitted: `[]` — No additional skills are intentionally required.

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: 3, 4, 5, 6, 7, 8 | Blocked By: 1

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `backend/cmd/property-service/main.go` — current DI wiring style.
  - Pattern: `backend/pkg/middleware/auth.go` — current access-token validation and DB role lookup.
  - Pattern: `backend/pkg/utils/jwt/jwt.go` — JWT carries identity/session info, not role authority.
  - API/Type: `backend/internal/domain/entity/user.go` — canonical role source in DB.
  - External: `https://github.com/casbin/casbin/blob/098508551fe30a208aafc7732b19c779fbc4059d/README.md#L125-L135` — Casbin does not replace authentication/user management.

  **Acceptance Criteria** (agent-executable only):
  - [x] `grep -n 'user_role' backend/pkg/middleware/auth.go` returns at least one DB-backed role-loading path or compatibility write.
  - [x] `grep -n 'authz' backend/cmd/property-service/main.go` returns enforcer/authz wiring matches.
  - [x] `grep -n 'user_role = "user"' backend/internal/handler/http backend/pkg/middleware` returns no silent defaulting logic.
  - [x] `cd backend && go build ./...` exits `0`.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```
  Scenario: Protected middleware preserves current-role freshness
    Tool: Bash
    Steps: Run `cd backend && go test ./internal/handler/http -count=1 -run TestProtected_UsesCurrentDatabaseRole -v`.
    Expected: Exit code 0; the same JWT reflects a DB role change without reminting tokens.
    Evidence: .sisyphus/evidence/task-2-principal-loader.txt

  Scenario: Missing or broken principal loading denies access safely
    Tool: Bash
    Steps: Run the same suite’s deny-path coverage or a new focused test proving missing locals/invalid role lookup does not continue to handlers.
    Expected: The request fails with unauthorized/forbidden behavior and never falls back to `user`.
    Evidence: .sisyphus/evidence/task-2-principal-loader-error.txt
  ```

  **Commit**: YES | Message: `feat(backend): add authz principal loading` | Files: `backend/cmd/property-service/main.go`, `backend/pkg/middleware/auth.go`, `backend/pkg/authz/**`

- [x] 3. Replace exact admin-role route checks with Casbin-backed permission middleware for category writes

  **What to do**: Introduce a permission middleware that consumes the shared principal and authz service, then migrate the category admin routes in `backend/internal/router/router.go` from `RequireRole("admin")` to a permission-based Casbin check. Keep the route-level object/action vocabulary explicit and stable so tests and services use the same names.
  **Must NOT do**: Do not duplicate category authorization in both middleware and category service for the same operation, and do not leave `RequireRole("admin")` in active category routes after migration.

  **Recommended Agent Profile**:
  - Category: `implementation` — Reason: route-level policy enforcement migration.
  - Skills: `[]` — No extra skill is required.
  - Omitted: `[]` — No additional skills are intentionally required.

  **Parallelization**: Can Parallel: YES | Wave 1 | Blocks: 7, 8 | Blocked By: 1, 2

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `backend/pkg/middleware/role.go` — current exact-role middleware to replace.
  - Pattern: `backend/internal/router/router.go` — current category admin route group.
  - Test: `backend/internal/handler/http/category_test.go` — parity anchor for category create/update/delete permissions.
  - External: `https://github.com/casbin/casbin/blob/098508551fe30a208aafc7732b19c779fbc4059d/enforcer_test.go#L28-L55` — RESTful `(sub,obj,act)` enforcement pattern.

  **Acceptance Criteria** (agent-executable only):
  - [x] `grep -n 'RequireRole("admin")' backend/internal/router/router.go` returns no matches for category write routes.
  - [x] `grep -n 'RequirePermission' backend/internal/router/router.go backend/pkg/middleware` returns permission middleware usage.
  - [x] `cd backend && go test ./internal/handler/http -count=1 -run TestCategoryHandlerSuite -v` exits `0`.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```
  Scenario: Admin category writes still succeed
    Tool: Bash
    Steps: Run `cd backend && go test ./internal/handler/http -count=1 -run TestCategoryHandlerSuite -v`.
    Expected: Exit code 0; admin create/update/delete category paths pass.
    Evidence: .sisyphus/evidence/task-3-category-route-authz.txt

  Scenario: Non-admin category writes are denied
    Tool: Bash
    Steps: Run the deny-path tests inside `TestCategoryHandlerSuite` or targeted new tests covering user attempts on create/update/delete.
    Expected: Requests fail with 403, and no category mutation happens.
    Evidence: .sisyphus/evidence/task-3-category-route-authz-error.txt
  ```

  **Commit**: YES | Message: `feat(backend): migrate category auth to casbin middleware` | Files: `backend/internal/router/router.go`, `backend/pkg/middleware/**`, `backend/internal/handler/http/category_test.go`

- [x] 4. Add a shared service-level authorization helper for owner/admin listing mutations

  **What to do**: Create one shared authz helper used by listing service mutations to evaluate owner/admin rules with Casbin-aware inputs. Keep resource ownership finalization in services, because listing ownership requires a loaded listing record. Replace scattered ownership/admin authorization decisions with one reusable helper that accepts the current principal, resource metadata, and requested action.
  **Must NOT do**: Do not move listing ownership checks into route middleware, and do not duplicate owner/admin decisions across multiple listing service methods after this helper exists.

  **Recommended Agent Profile**:
  - Category: `implementation` — Reason: shared service authz logic for ownership-sensitive resources.
  - Skills: `[]` — No extra skill is required.
  - Omitted: `[]` — No additional skills are intentionally required.

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: 5, 7, 8 | Blocked By: 1, 2

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `backend/internal/service/listing_service.go` — current owner/admin checks for update/delete/image mutations.
  - Pattern: `backend/internal/domain/entity/listing.go` — listing owner metadata source (`UserID`).
  - Test: `backend/internal/service/listing_service_test.go` — existing owner/admin parity tests.
  - External: `https://github.com/casbin/casbin-website-v2/blob/master/docs/ABAC.mdx#L9-L52` — ABAC matcher on resource owner.
  - External: `https://github.com/casbin/casbin-website-v2/blob/master/docs/RBACWithConditions.mdx#L9-L69` — conditional role guidance when extra context is needed.

  **Acceptance Criteria** (agent-executable only):
  - [x] `grep -n 'checkOwnership' backend/internal/service/listing_service.go` returns either no legacy helper or only a thin wrapper over the new shared authz helper.
  - [x] `grep -n 'Enforce' backend/internal/service backend/pkg/authz` returns service-layer authz helper usage.
  - [x] `cd backend && go test ./internal/service -count=1 -run 'TestAuthzService_EnforceListingUpdate_OwnerAllowed|TestAuthzService_EnforceListingUpdate_AdminAllowed|TestAuthzService_EnforceListingUpdate_NonOwnerForbidden' -v` exits `0`.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```
  Scenario: Owner and admin listing mutations are allowed
    Tool: Bash
    Steps: Run `cd backend && go test ./internal/service -count=1 -run 'TestAuthzService_EnforceListingUpdate_OwnerAllowed|TestAuthzService_EnforceListingUpdate_AdminAllowed' -v`.
    Expected: Exit code 0 and both tests pass.
    Evidence: .sisyphus/evidence/task-4-service-authz.txt

  Scenario: Non-owner user listing mutation is denied
    Tool: Bash
    Steps: Run `cd backend && go test ./internal/service -count=1 -run 'TestAuthzService_EnforceListingUpdate_NonOwnerForbidden' -v`.
    Expected: Exit code 0 and the deny-path test passes.
    Evidence: .sisyphus/evidence/task-4-service-authz-error.txt
  ```

  **Commit**: YES | Message: `feat(backend): add listing authz helper` | Files: `backend/internal/service/**`, `backend/pkg/authz/**`, `backend/internal/service/listing_service_test.go`

- [x] 5. Refactor protected listing handlers and services to pass a principal instead of primitive role strings

  **What to do**: Remove the current pattern where handlers pull `user_id` and `user_role` separately and sometimes fall back to `"user"`. Introduce a structured principal/auth context and thread it through the listing handler -> service path so authorization decisions receive one canonical subject shape. Temporary compatibility locals are allowed only while older code paths still need them.
  **Must NOT do**: Do not leave any mutation handler path that defaults to `user` when principal data is missing, and do not make handlers responsible for policy decisions.

  **Recommended Agent Profile**:
  - Category: `refactorer` — Reason: this is interface and call-site cleanup around authz boundaries.
  - Skills: `[]` — No extra skill is required.
  - Omitted: `[]` — No additional skills are intentionally required.

  **Parallelization**: Can Parallel: NO | Wave 2 | Blocks: 7, 8 | Blocked By: 2, 4

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `backend/internal/handler/http/listing.go` — current primitive user/role extraction and fallback behavior.
  - Pattern: `backend/pkg/middleware/auth.go` — source of authenticated principal data.
  - Pattern: `backend/internal/service/listing_service.go` — service signatures currently accept `userID` and `userRole` separately.
  - Test: `backend/internal/handler/http/listing_test.go` — regression suite for listing protected behavior.

  **Acceptance Criteria** (agent-executable only):
  - [x] `grep -R 'userRole = "user"' backend/internal/handler/http backend/internal/service` returns no matches.
  - [x] `grep -R 'Locals("user_role")' backend/internal/handler/http/listing.go` returns no direct handler fallback logic.
  - [x] `cd backend && go test ./internal/handler/http -count=1 -run TestListingHandlerSuite -v` exits `0`.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```
  Scenario: Listing handlers still pass authenticated requests correctly
    Tool: Bash
    Steps: Run `cd backend && go test ./internal/handler/http -count=1 -run TestListingHandlerSuite -v`.
    Expected: Exit code 0; protected listing CRUD and image tests stay green.
    Evidence: .sisyphus/evidence/task-5-principal-refactor.txt

  Scenario: Missing principal does not silently become a normal user
    Tool: Bash
    Steps: Run a focused handler test that invokes a protected mutation path without valid principal locals after middleware.
    Expected: The request fails, and no mutation is performed.
    Evidence: .sisyphus/evidence/task-5-principal-refactor-error.txt
  ```

  **Commit**: YES | Message: `refactor(backend): use principal across listing handlers` | Files: `backend/internal/handler/http/listing.go`, `backend/internal/service/listing_service.go`, related tests

- [x] 6. Add rollout parity safeguards for immediate role freshness and deny-safe failures

  **What to do**: Make role freshness an explicit supported behavior in the new authz path. Add targeted regression coverage proving that a DB role change takes effect without forcing re-login, and define consistent failure behavior when policy loading or enforcement infrastructure is unavailable. Startup failures must stop boot; request-time dependency failures must deny with 5xx/forbidden semantics, never allow.
  **Must NOT do**: Do not move role authority into JWT claims in phase 1, and do not fail open on enforcer/adaptor errors.

  **Recommended Agent Profile**:
  - Category: `testing` — Reason: parity and deny-safe rollout behavior need explicit proof.
  - Skills: `[]` — No extra skill is required.
  - Omitted: `[]` — No additional skills are intentionally required.

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: 7 | Blocked By: 2

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `backend/pkg/middleware/auth.go` — current DB role reload on every protected request.
  - Pattern: `backend/internal/handler/http/category_test.go` — existing admin-path handler suite structure.
  - External: `https://github.com/casbin/casbin/blob/098508551fe30a208aafc7732b19c779fbc4059d/persist/watcher.go#L17-L29` — watcher is optional and out of scope for phase 1.

  **Acceptance Criteria** (agent-executable only):
  - [x] `cd backend && go test ./internal/handler/http -count=1 -run TestProtected_UsesCurrentDatabaseRole -v` exits `0`.
  - [x] `grep -R 'fail open\|allow on error' backend/pkg/authz backend/pkg/middleware backend/internal/service` returns no permissive-on-error logic.
  - [x] `cd backend && go build ./...` exits `0`.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```
  Scenario: Role changes take effect without re-login
    Tool: Bash
    Steps: Run `cd backend && go test ./internal/handler/http -count=1 -run TestProtected_UsesCurrentDatabaseRole -v`.
    Expected: Exit code 0; the same JWT observes updated DB role state.
    Evidence: .sisyphus/evidence/task-6-role-freshness.txt

  Scenario: Authz dependency failure denies instead of allowing
    Tool: Bash
    Steps: Run a focused middleware/authz test that injects a broken enforcer or adapter initialization path.
    Expected: Boot fails or request returns a deny/5xx path; no allow result is produced.
    Evidence: .sisyphus/evidence/task-6-role-freshness-error.txt
  ```

  **Commit**: YES | Message: `test(backend): preserve role freshness semantics` | Files: `backend/pkg/middleware/**`, `backend/internal/handler/http/**`, `backend/pkg/authz/**`

- [x] 7. Add dedicated authorization unit tests for model, policy, and service enforcement vocabulary

  **What to do**: Add a focused test suite around the new authz package and service helper(s) that covers model loading, seeded policy behavior, admin allow, normal-user deny, and owner/admin listing mutation cases. These tests should be the first place that proves the action/object vocabulary and deny-default semantics are correct before handler integration suites run.
  **Must NOT do**: Do not rely only on broad handler tests to validate policy logic, and do not hardcode policy semantics in multiple independent test helpers.

  **Recommended Agent Profile**:
  - Category: `testing` — Reason: focused unit coverage for policy semantics.
  - Skills: `[]` — No extra skill is required.
  - Omitted: `[]` — No additional skills are intentionally required.

  **Parallelization**: Can Parallel: YES | Wave 3 | Blocks: 8, 9 | Blocked By: 1, 2, 4, 5, 6

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `backend/internal/service/listing_service_test.go` — existing service mock-based testing style.
  - Pattern: `backend/internal/domain/errors.go` — deny paths should map to shared domain errors where appropriate.
  - External: `https://github.com/casbin/casbin/blob/098508551fe30a208aafc7732b19c779fbc4059d/enforcer_test.go#L87-L104` — allow/deny/domain-style enforcement testing patterns.

  **Acceptance Criteria** (agent-executable only):
  - [x] `cd backend && go test ./internal/service -count=1 -run 'TestAuthzService_EnforceCategoryCreate_AdminAllowed|TestAuthzService_EnforceCategoryCreate_UserForbidden|TestAuthzService_EnforceListingUpdate_OwnerAllowed|TestAuthzService_EnforceListingUpdate_AdminAllowed|TestAuthzService_EnforceListingUpdate_NonOwnerForbidden' -v` exits `0`.
  - [x] `cd backend && go test ./pkg/authz/... -count=1` exits `0`.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```
  Scenario: Core authz semantics pass at unit level
    Tool: Bash
    Steps: Run `cd backend && go test ./internal/service -count=1 -run 'TestAuthzService_EnforceCategoryCreate_AdminAllowed|TestAuthzService_EnforceCategoryCreate_UserForbidden|TestAuthzService_EnforceListingUpdate_OwnerAllowed|TestAuthzService_EnforceListingUpdate_AdminAllowed|TestAuthzService_EnforceListingUpdate_NonOwnerForbidden' -v`.
    Expected: Exit code 0 and each named test passes.
    Evidence: .sisyphus/evidence/task-7-authz-unit-tests.txt

  Scenario: Default deny stays intact when no matching allow policy exists
    Tool: Bash
    Steps: Run the deny-focused subset above and any pkg/authz tests covering absent-policy behavior.
    Expected: Deny-path tests pass and no unauthorized action is accidentally allowed.
    Evidence: .sisyphus/evidence/task-7-authz-unit-tests-error.txt
  ```

  **Commit**: YES | Message: `test(backend): cover casbin authorization semantics` | Files: `backend/pkg/authz/**`, `backend/internal/service/**`

- [x] 8. Add handler integration coverage for category permissions and listing parity under the new authz layer

  **What to do**: Extend the existing handler suites so they prove the Casbin-backed middleware/service combination preserves current HTTP behavior for category writes, listing mutations, image flows, and protected access. Reuse the current `testify/suite` + testcontainers structure instead of inventing a second integration style.
  **Must NOT do**: Do not replace the existing parity suites with only narrow unit tests, and do not leave category/listing authorization unverified at the HTTP layer.

  **Recommended Agent Profile**:
  - Category: `testing` — Reason: integration parity is the rollout safety net.
  - Skills: `[]` — No extra skill is required.
  - Omitted: `[]` — No additional skills are intentionally required.

  **Parallelization**: Can Parallel: NO | Wave 3 | Blocks: 9 | Blocked By: 3, 4, 5, 7

  **References** (executor has NO interview context — be exhaustive):
  - Test: `backend/internal/handler/http/category_test.go` — category suite structure and existing admin route registration.
  - Test: `backend/internal/handler/http/listing_test.go` — listing suite and protected route structure.
  - Test: `backend/internal/handler/http/auth_test.go` — auth-protected route suite pattern.
  - Pattern: `backend/internal/router/router.go` — canonical route registration targets.

  **Acceptance Criteria** (agent-executable only):
  - [x] `cd backend && go test ./internal/handler/http -count=1 -run TestCategoryHandlerSuite -v` exits `0`.
  - [x] `cd backend && go test ./internal/handler/http -count=1 -run TestListingHandlerSuite -v` exits `0`.
  - [x] `cd backend && go test ./internal/handler/http -count=1 -run TestProtected_UsesCurrentDatabaseRole -v` exits `0`.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```
  Scenario: Handler suites preserve current category and listing behavior
    Tool: Bash
    Steps: Run the category and listing handler suite commands above.
    Expected: Exit code 0 and final output contains `ok` for `./internal/handler/http`.
    Evidence: .sisyphus/evidence/task-8-handler-parity.txt

  Scenario: Unauthorized or forbidden paths still fail at HTTP level
    Tool: Bash
    Steps: Run the targeted protected-role freshness suite and deny-path cases inside the handler suites.
    Expected: Denied requests remain denied and suites still pass.
    Evidence: .sisyphus/evidence/task-8-handler-parity-error.txt
  ```

  **Commit**: YES | Message: `test(backend): verify authz handler parity` | Files: `backend/internal/handler/http/**`

- [x] 9. Update backend guidance, configuration notes, and rollout documentation for the new authz model

  **What to do**: Refresh the relevant AGENTS/docs so the backend authorization model is accurately described after Casbin is introduced. Document that backend DB role loading remains authoritative, persisted roles stay `user`/`admin`, phase-1 policy storage is migration-seeded, and runtime policy management/watchers are intentionally deferred. Update config notes only if a real new setting is added; avoid adding fake env toggles.
  **Must NOT do**: Do not claim watcher support, dynamic policy admin APIs, or stored `seller` roles if they are not implemented.

  **Recommended Agent Profile**:
  - Category: `doc-writer` — Reason: final documentation and rollout guidance cleanup.
  - Skills: `[]` — No extra skill is required.
  - Omitted: `[]` — No additional skills are intentionally required.

  **Parallelization**: Can Parallel: YES | Wave 3 | Blocks: F1-F4 | Blocked By: 1, 2, 7, 8

  **References** (executor has NO interview context — be exhaustive):
  - Pattern: `AGENTS.md` — root planned/implemented status and architecture rules.
  - Pattern: `backend/AGENTS.md` — backend feature areas and layer guidance.
  - Pattern: `backend/pkg/AGENTS.md` — config and middleware package notes.
  - Pattern: `backend/.env-example` and `backend/.env.docker` — only update if new env keys truly exist.

  **Acceptance Criteria** (agent-executable only):
  - [x] `grep -n 'RBAC/Casbin' AGENTS.md backend/AGENTS.md backend/pkg/AGENTS.md` returns updated guidance that matches the implemented phase-1 model.
  - [x] `grep -R 'seller role' AGENTS.md backend frontend` returns no misleading documentation that says seller is a stored role, unless explicitly documented as deferred.
  - [x] `cd backend && go build ./...` exits `0` after doc/config updates.

  **QA Scenarios** (MANDATORY — task incomplete without these):
  ```
  Scenario: Guidance matches implemented authz behavior
    Tool: Bash
    Steps: Run the grep checks above against AGENTS/config files.
    Expected: Docs reflect `user`/`admin` persisted roles, Casbin usage, and deferred watcher/runtime policy management.
    Evidence: .sisyphus/evidence/task-9-docs-rollout.txt

  Scenario: No phantom config or role claims were introduced
    Tool: Bash
    Steps: Search docs/env files for undeclared new env toggles or claims about a stored `seller` role.
    Expected: No misleading claims are present.
    Evidence: .sisyphus/evidence/task-9-docs-rollout-error.txt
  ```

  **Commit**: YES | Message: `docs(backend): document casbin rollout` | Files: `AGENTS.md`, `backend/AGENTS.md`, `backend/pkg/AGENTS.md`, optional env docs if needed

## Final Verification Wave (4 parallel agents, ALL must APPROVE)
- [x] F1. Plan Compliance Audit — oracle
- [x] F2. Code Quality Review — unspecified-high
- [x] F3. Real Runtime QA — unspecified-high
- [x] F4. Scope Fidelity Check — deep

## Commit Strategy
- Commit wave 1 foundation separately from route/service integration.
- Keep route-level migration and service-level migration in separate commits for rollback clarity.
- Keep tests with the code they validate.

## Success Criteria
- All category admin checks and listing owner/admin checks flow through one shared authz design.
- No route silently allows access when principal, policy, or enforcer state is missing.
- Current role freshness semantics are preserved.
- Existing category/listing behavior remains green under integration tests.

# Cloudinary Listing Images

## TL;DR
> **Summary**: Implement Cloudinary-backed listing image management in the Go backend using backend-streamed multipart uploads, Postgres-owned ordering/primary invariants, and app-level image lifecycle rules.
> **Deliverables**:
> - listing-image schema and domain contract upgrades
> - Cloudinary storage adapter and config wiring
> - upload/delete/set-primary/reorder endpoints with tests
> - listing response projection, Postman updates, and AGENTS.md updates
> **Effort**: Large
> **Parallel**: YES - 2 waves
> **Critical Path**: T1 schema/domain -> T2 storage/config -> T3 repository -> T4 service -> T5 handler/routes -> T7 verification/docs

## Context
### Original Request
Plan option 3: Cloudinary-based image management for listing photos, and include AGENTS.md alignment plus sub-agent execution guidance and context-token hygiene.

### Interview Summary
- Provider is Cloudinary, not S3/R2.
- Scope is backend only: upload, delete, set primary image, reorder images, API exposure, tests, Postman, env/docs/AGENTS updates.
- Implementation should follow current repo authority: `c.Context()` in handlers, flat mock-based service tests, and testcontainers-based handler suites.
- Execution plan must include sub-agent dispatch guidance and explicit `distill`/`prune` checkpoints after each implementation wave.

### Metis Review (gaps addressed)
- Resolved delete-path ambiguity by choosing app image UUID externally and stored Cloudinary `public_id` + `resource_type` + `type` internally for operational destroy; keep immutable `asset_id` for audit/future-proofing.
- Resolved upload-scope ambiguity by choosing single-file backend-streamed multipart upload for v1.
- Added guardrails for file limits, MIME validation, no soft-delete purge, and no frontend/direct-browser upload scope.
- Added explicit acceptance criteria for max-images, ownership/admin behavior, listing projection ordering, and Cloudinary-free test execution.

## Work Objectives
### Core Objective
Add a production-ready listing-image feature to the backend so authenticated owners/admins can upload one image at a time to Cloudinary, delete images, set a primary image, reorder images, and receive ordered image data from listing read endpoints.

### Deliverables
- Additive SQL migration that upgrades `listing_images` for Cloudinary metadata and soft-delete-safe constraints.
- Domain/storage contracts for listing image persistence plus Cloudinary upload/delete abstraction.
- Cloudinary config and adapter wired into `cmd/property-service/main.go`.
- Listing service methods for upload, delete, set-primary, reorder, and projection.
- Protected listing-image HTTP endpoints and DTOs.
- Service unit tests and handler integration tests using mocks/fakes only.
- `backend/postman_collection.json`, `backend/.env-example`, and AGENTS docs aligned with the new feature.

### Definition of Done (verifiable conditions with commands)
- `cd backend && go test ./... -count=1`
- `cd backend && go test ./... -count=1 -run TestListingHandlerSuite -v`
- `cd backend && go test ./... -count=1 -run TestListingService -v`
- `cd backend && go build ./...`
- `cd backend && go vet ./...`
- Listing read endpoints return ordered `images` arrays after upload/reorder/delete operations.
- No automated test path requires real Cloudinary credentials.

### Must Have
- Single-file multipart upload endpoint using form field `file`.
- App image UUID as the external image identifier in routes.
- Service-owned invariants: max 10 images per listing, exactly one primary image when images exist, sort order starts at 0 and remains contiguous, first image auto-becomes primary.
- Cloudinary metadata persisted locally: `asset_id`, `public_id`, `version`, `resource_type`, `type`, `format`, `bytes`, `width`, `height`, `url`, optional `original_filename`, plus app-owned `is_primary`, `sort_order`, and `deleted_at`.
- Best-effort Cloudinary compensation when DB persistence fails after upload.
- Best-effort Cloudinary destroy after DB delete transaction commits.
- AGENTS and env docs updated in the same implementation.

### Must NOT Have (guardrails, AI slop patterns, scope boundaries)
- No direct browser upload flow.
- No frontend screens or API client work.
- No background cleanup worker or outbox in v1.
- No hard purge of Cloudinary assets on listing soft-delete.
- No captions/alt text, no moderation pipeline, no transformation presets beyond storing the returned delivery URL.
- No broad refactor of unrelated listing CRUD code.
- No live-Cloudinary dependency in unit or handler integration tests.

## Verification Strategy
> ZERO HUMAN INTERVENTION - all verification is agent-executed.
- Test decision: tests-after + Go `testify/mock` service tests and `testify/suite` + testcontainers handler tests
- QA policy: every task includes executable API or test scenarios
- Evidence: `.sisyphus/evidence/task-{N}-{slug}.{ext}`
- Context hygiene: after each execution wave, the lead executor must run `distill` on exploration/test outputs that contain reusable signal and `prune` superseded outputs before starting the next wave

## Execution Strategy
### Parallel Execution Waves
> Target: 5-8 tasks per wave. Shared foundations land in Wave 1; API/docs/verification fan out in Wave 2.

Wave 1: T1 schema/domain, T2 storage/config, T3 repository, T4 service
Wave 2: T5 handler/router/DTOs, T6 service tests, T7 handler tests + Postman + AGENTS/docs

Executor rule: dispatch one sub-agent per task inside each wave, merge/verify Wave 1 before starting Wave 2 mutations, then run `distill` on retained findings and `prune` superseded outputs before the next wave.

### Dependency Matrix (full, all tasks)
- T1 blocks T3, T4, T5, T6, T7
- T2 blocks T4, T6, T7
- T3 blocks T4, T6, T7
- T4 blocks T5, T6, T7
- T5 blocks T7
- T6 and T7 depend on T4; T7 also depends on T5

### Agent Dispatch Summary (wave -> task count -> categories)
- Wave 1 -> 4 tasks -> `implementation`, `refactorer`
- Wave 2 -> 3 tasks -> `implementation`, `testing`, `writing`
- Final Verification -> 4 tasks -> `reviewer`, `review`, `tester`, `deep`

## TODOs
> Implementation + Test = ONE task. Never separate.
> Every task includes explicit QA scenarios.

<!-- TASKS -->

- [x] T1. Upgrade listing image schema and domain model for Cloudinary metadata

  **What to do**: Add a new additive migration `backend/db/migrations/000005_listing_images_cloudinary_metadata.{up,down}.sql` that upgrades `listing_images` without editing old migrations. Keep existing `url` column as the secure delivery URL. Add nullable-or-safe columns for `asset_id`, `public_id`, `version`, `resource_type`, `type`, `format`, `bytes`, `width`, `height`, `original_filename`, and `deleted_at`. Add indexes/constraints that backstop invariants without assuming a full contiguous-order DB constraint: unique active `(listing_id, sort_order)` and at-most-one active primary image per listing. Update `backend/internal/domain/entity/listing.go` to carry the new metadata fields, keeping app-facing image identity as the local UUID.
  **Must NOT do**: Do not rewrite `000001_init_schema.up.sql`; do not replace local image UUIDs with Cloudinary IDs in API routes; do not hard-delete existing image rows in migration logic.

  **Recommended Agent Profile**:
  - Category: `implementation` - Reason: schema + entity changes are foundational and precise
  - Skills: `[]` - no additional skill required
  - Omitted: `[]` - no omission needed

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: T3, T4, T5, T6, T7 | Blocked By: none

  **References**:
  - Pattern: `backend/db/migrations/000003_listing_crud_prep.up.sql` - additive migration style and explicit SQL comments
  - Pattern: `backend/db/migrations/000004_category_fk_set_null_and_seed.up.sql` - multi-step migration structure
  - API/Type: `backend/internal/domain/entity/listing.go` - existing `ListingImage` and `Listing.Images` relationship
  - API/Type: `backend/db/migrations/000001_init_schema.up.sql` - current `listing_images` table shape
  - Docs: `AGENTS.md` - project conventions for money, UUIDs, and testing

  **Acceptance Criteria**:
  - [ ] New migration file exists with `up` and `down` variants and does not modify old migration files
  - [ ] `ListingImage` entity includes Cloudinary metadata fields plus `deleted_at`
  - [ ] DB constraints prevent duplicate active `sort_order` values per listing and duplicate active primary images per listing

  **QA Scenarios**:
  ```bash
  Scenario: Migration files and entity shape exist
    Tool: Bash
    Steps: python3 - <<'PY'
import pathlib
up = pathlib.Path('backend/db/migrations/000005_listing_images_cloudinary_metadata.up.sql')
down = pathlib.Path('backend/db/migrations/000005_listing_images_cloudinary_metadata.down.sql')
entity = pathlib.Path('backend/internal/domain/entity/listing.go').read_text()
assert up.exists() and down.exists()
for needle in ['asset_id', 'public_id', 'deleted_at', 'width', 'height']:
    assert needle in up.read_text() or needle in entity
print('ok')
PY
    Expected: script prints `ok`
    Evidence: .sisyphus/evidence/task-T1-schema.txt

  Scenario: Old migration files remain untouched
    Tool: Bash
    Steps: git diff --name-only -- backend/db/migrations | python3 - <<'PY'
import sys
changed=[line.strip() for line in sys.stdin if line.strip()]
assert all(name.endswith('000005_listing_images_cloudinary_metadata.up.sql') or name.endswith('000005_listing_images_cloudinary_metadata.down.sql') for name in changed), changed
print('ok')
PY
    Expected: script prints `ok`
    Evidence: .sisyphus/evidence/task-T1-migration-guard.txt
  ```

  **Commit**: YES | Message: `feat(listing-images): add cloudinary image schema metadata` | Files: `backend/db/migrations/*`, `backend/internal/domain/entity/listing.go`

- [x] T2. Add Cloudinary config and mockable storage adapter

  **What to do**: Add Cloudinary runtime config to `backend/pkg/config/config.go` and sync `backend/.env-example` plus any local env file used by runtime startup. Introduce a small domain storage port dedicated to listing-image asset lifecycle, then implement a Cloudinary adapter under `backend/pkg/` using the Go SDK v2. Use backend-owned uploads from Go, configure via backend env only, and persist returned metadata for later DB storage. Decide deletion implementation explicitly: operational deletes use stored `public_id` + `resource_type` + `type`, with `asset_id` retained as immutable metadata but not required for the v1 destroy call.
  **Must NOT do**: Do not expose `api_secret` to clients; do not create direct-browser upload signing flow; do not make tests depend on real Cloudinary credentials.

  **Recommended Agent Profile**:
  - Category: `implementation` - Reason: config plus provider adapter work
  - Skills: `[]` - no additional skill required
  - Omitted: `[]` - no omission needed

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: T4, T6, T7 | Blocked By: none

  **References**:
  - Pattern: `backend/pkg/config/config.go` - runtime env authority and validation flow
  - Pattern: `backend/internal/repository/postgres/auth.go` - repository-side sensitive integration pattern for external token data
  - Pattern: `backend/pkg/AGENTS.md` - shared package/documentation expectations
  - External: `https://cloudinary.com/documentation/go_image_and_video_upload` - Go SDK upload response shape
  - External: `https://cloudinary.com/documentation/go_integration` - Go SDK v2 integration guidance
  - External: `https://cloudinary.com/documentation/image_upload_api_reference` - destroy/upload parameter behavior

  **Acceptance Criteria**:
  - [ ] Config struct and env loading support Cloudinary credentials without breaking existing config loading
  - [ ] Cloudinary adapter exposes upload and destroy operations through a mockable interface
  - [ ] No test path requires live Cloudinary credentials; a fake/mock implementation is available for service and handler tests

  **QA Scenarios**:
  ```bash
  Scenario: Config and adapter compile without credentials in tests
    Tool: Bash
    Steps: cd backend && go test ./... -count=1 -run TestListingService -v
    Expected: targeted service tests pass without needing Cloudinary env vars
    Evidence: .sisyphus/evidence/task-T2-config-storage.txt

  Scenario: Cloudinary secrets stay server-side
    Tool: Bash
    Steps: python3 - <<'PY'
from pathlib import Path
import subprocess
result = subprocess.run(['grep','-R','-n','CLOUDINARY_URL\|api_secret','backend','frontend'], capture_output=True, text=True)
lines = [line for line in result.stdout.splitlines() if line.strip()]
assert all(not line.startswith('frontend/') for line in lines), lines
print('ok')
PY
    Expected: script prints `ok`
    Evidence: .sisyphus/evidence/task-T2-secret-scan.txt
  ```

  **Commit**: YES | Message: `feat(listing-images): add cloudinary storage integration` | Files: `backend/pkg/config/*`, `backend/pkg/**`, `backend/.env-example`, `backend/.env.docker`

- [x] T3. Extend repository contracts and Postgres image operations

  **What to do**: Extend domain contracts so image persistence is explicit instead of hidden inside generic listing updates. Add repository methods for creating image rows, listing active images for a listing, loading a single image by local UUID, soft-deleting an image, promoting a primary image, normalizing primary after delete, and reordering a full image set. Implement these in `backend/internal/repository/postgres/listing.go` with transaction boundaries and row locking where multiple image rows are mutated. Keep listing read queries preloading active images and ordering them by ascending `sort_order`.
  **Must NOT do**: Do not put Cloudinary SDK calls in repository code; do not leave primary/sort mutations as multi-query non-transactional best effort.

  **Recommended Agent Profile**:
  - Category: `implementation` - Reason: contract and transaction-heavy persistence work
  - Skills: `[]` - no additional skill required
  - Omitted: `[]` - no omission needed

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: T4, T6, T7 | Blocked By: T1

  **References**:
  - API/Type: `backend/internal/domain/listing_repository.go` - current listing persistence contract
  - Pattern: `backend/internal/repository/postgres/listing.go` - preload/filter/order conventions
  - Pattern: `backend/internal/repository/AGENTS.md` - transaction and not-found translation rules
  - API/Type: `backend/internal/domain/entity/listing.go` - image relation shape

  **Acceptance Criteria**:
  - [ ] Domain contract includes explicit image persistence methods or a dedicated image repository interface
  - [ ] Postgres implementation mutates image ordering/primary state inside transactions
  - [ ] Listing reads preload active images sorted ascending by `sort_order`

  **QA Scenarios**:
  ```bash
  Scenario: Repository-backed listing reads project ordered images
    Tool: Bash
    Steps: cd backend && go test ./... -count=1 -run TestListingHandlerSuite -v
    Expected: handler integration suite passes with ordered image projection assertions
    Evidence: .sisyphus/evidence/task-T3-repository-ordering.txt

  Scenario: Repository contract exposes image operations
    Tool: Bash
    Steps: python3 - <<'PY'
from pathlib import Path
text = Path('backend/internal/domain/listing_repository.go').read_text()
for needle in ['CreateImage', 'FindImage', 'DeleteImage', 'ReorderImages', 'SetPrimaryImage']:
    assert needle in text, needle
print('ok')
PY
    Expected: script prints `ok`
    Evidence: .sisyphus/evidence/task-T3-contracts.txt
  ```

  **Commit**: YES | Message: `feat(listing-images): add image persistence workflows` | Files: `backend/internal/domain/*`, `backend/internal/repository/postgres/listing.go`, `backend/internal/domain/mocks/*`

- [x] T4. Add listing-image service workflows and response projection

  **What to do**: Extend `ListingService` so it owns all listing-image business rules. Add methods for upload, delete, set-primary, and reorder that enforce owner/admin access, max 10 images, accepted image-only uploads, first-image auto-primary, contiguous `sort_order`, and deterministic primary fallback after delete. Keep Cloudinary interaction behind the injected storage port. Update listing response mapping so `GetByID`, `GetBySlug`, `List`, and `ListByUserID` all include ordered image projections while hiding provider internals (`asset_id`, `public_id`).
  **Must NOT do**: Do not call Cloudinary directly from handlers; do not expose `public_id` or `asset_id` in public response DTOs; do not leave image ordering implicit in service logic.

  **Recommended Agent Profile**:
  - Category: `implementation` - Reason: service-layer orchestration and invariants
  - Skills: `[]` - no additional skill required
  - Omitted: `[]` - no omission needed

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: T5, T6, T7 | Blocked By: T1, T2, T3

  **References**:
  - Pattern: `backend/internal/service/listing_service.go` - ownership checks and listing response mapping
  - Pattern: `backend/internal/service/category_service.go` - domain-error-driven service logic
  - Pattern: `backend/internal/service/AGENTS.md` - service-layer dependency and test rules
  - API/Type: `backend/internal/dto/response/listing_response.go` - listing projection surface to extend
  - External: `https://cloudinary.com/documentation/go_image_and_video_upload` - upload response metadata to persist

  **Acceptance Criteria**:
  - [ ] Service exposes upload/delete/set-primary/reorder workflows with explicit domain errors for invalid file type, over-limit image count, not-found image, and forbidden ownership
  - [ ] Listing responses include `images` sorted by ascending `sort_order`
  - [ ] Deleting a primary image promotes the next lowest `sort_order` image to primary when one exists

  **QA Scenarios**:
  ```bash
  Scenario: Service tests cover image invariants without live Cloudinary
    Tool: Bash
    Steps: cd backend && go test ./... -count=1 -run TestListingService -v
    Expected: service tests pass and include upload/reorder/set-primary/delete cases
    Evidence: .sisyphus/evidence/task-T4-service-tests.txt

  Scenario: Listing responses expose ordered images only
    Tool: Bash
    Steps: python3 - <<'PY'
from pathlib import Path
text = Path('backend/internal/dto/response/listing_response.go').read_text()
assert 'Images' in text
assert 'PublicID' not in text and 'AssetID' not in text
print('ok')
PY
    Expected: script prints `ok`
    Evidence: .sisyphus/evidence/task-T4-response-shape.txt
  ```

  **Commit**: YES | Message: `feat(listing-images): add image service workflows` | Files: `backend/internal/service/listing_service.go`, `backend/internal/dto/response/listing_response.go`, `backend/internal/domain/errors.go`

- [x] T5. Add HTTP endpoints, multipart parsing, DTOs, and route registration

  **What to do**: Add protected listing-image endpoints under the existing listing route namespace using local listing UUID plus local image UUID. Use backend-streamed multipart upload with field name `file` for `POST /api/listings/:id/images`. Add JSON DTOs for `PATCH /api/listings/:id/images/reorder` with ordered `image_ids`, `PATCH /api/listings/:id/images/:imageId/primary`, and `DELETE /api/listings/:id/images/:imageId`. Keep handlers thin: extract auth locals, validate UUID params/body, pass `c.Context()`, and return the updated `ListingResponse`.
  **Must NOT do**: Do not add direct-upload signing endpoints; do not create ad hoc response envelopes; do not move ownership rules into handlers.

  **Recommended Agent Profile**:
  - Category: `implementation` - Reason: transport layer and route surface work
  - Skills: `[]` - no additional skill required
  - Omitted: `[]` - no omission needed

  **Parallelization**: Can Parallel: NO | Wave 2 | Blocks: T7 | Blocked By: T1, T4

  **References**:
  - Pattern: `backend/internal/handler/http/listing.go` - auth locals, UUID parsing, `c.Context()`, `utils.SendResponse`
  - Pattern: `backend/internal/router/router.go` - protected listing route registration
  - API/Type: `backend/internal/dto/request/listing_request.go` - DTO naming and validation tags
  - Pattern: `backend/internal/handler/http/AGENTS.md` - handler conventions

  **Acceptance Criteria**:
  - [ ] New routes exist for upload, delete, set-primary, and reorder under protected listing routes
  - [ ] Upload endpoint accepts exactly one multipart file field named `file`
  - [ ] All endpoints return the standard success envelope with updated listing image data

  **QA Scenarios**:
  ```bash
  Scenario: Upload endpoint validates multipart field and auth
    Tool: Bash
    Steps: cd backend && go test ./... -count=1 -run TestListingHandlerSuite -v
    Expected: handler suite passes upload success, unauthorized, invalid-file, and over-limit cases
    Evidence: .sisyphus/evidence/task-T5-upload-handler.txt

  Scenario: Route registration exposes listing image endpoints
    Tool: Bash
    Steps: python3 - <<'PY'
from pathlib import Path
text = Path('backend/internal/router/router.go').read_text()
for needle in ['/api/listings', 'images/reorder', '/primary']:
    assert needle in text, needle
print('ok')
PY
    Expected: script prints `ok`
    Evidence: .sisyphus/evidence/task-T5-routes.txt
  ```

  **Commit**: YES | Message: `feat(listing-images): add listing image endpoints` | Files: `backend/internal/handler/http/listing.go`, `backend/internal/dto/request/listing_request.go`, `backend/internal/router/router.go`

- [x] T6. Add service-unit coverage with storage and repository mocks

  **What to do**: Expand `backend/internal/service/listing_service_test.go` to cover all new image workflows using mock repository/storage collaborators. Include happy paths and failure paths for upload success, Cloudinary upload failure, DB persistence failure with compensation, max-images rejection, owner/admin authorization, set-primary transitions, reorder validation, delete fallback-primary behavior, and read projection mapping. Regenerate or hand-update mocks only where the repo already treats them as generated artifacts.
  **Must NOT do**: Do not hit real Cloudinary; do not move handler/integration concerns into service tests.

  **Recommended Agent Profile**:
  - Category: `testing` - Reason: focused unit-test expansion with mocks
  - Skills: `[]` - no additional skill required
  - Omitted: `[]` - no omission needed

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: none | Blocked By: T2, T3, T4

  **References**:
  - Pattern: `backend/internal/service/listing_service_test.go` - flat test style and mock expectations
  - Pattern: `backend/internal/service/auth_service_test.go` - collaborator mocking style
  - Pattern: `backend/internal/domain/mocks/` - generated mock location
  - Pattern: `backend/internal/service/AGENTS.md` - service test conventions

  **Acceptance Criteria**:
  - [ ] Listing service unit tests cover every new image workflow and major failure branch
  - [ ] Tests verify compensation path when DB persistence fails after upload
  - [ ] Tests verify response mapping includes images sorted by `sort_order`

  **QA Scenarios**:
  ```bash
  Scenario: Listing service unit tests pass
    Tool: Bash
    Steps: cd backend && go test ./... -count=1 -run TestListingService -v
    Expected: targeted service tests pass
    Evidence: .sisyphus/evidence/task-T6-service-suite.txt

  Scenario: No real Cloudinary use exists in unit tests
    Tool: Bash
    Steps: grep -R "cloudinary.New\|CLOUDINARY_URL\|api_secret" -n backend/internal/service/*_test.go backend/internal/handler/http/*_test.go || true
    Expected: no live Cloudinary client construction in tests
    Evidence: .sisyphus/evidence/task-T6-no-live-cloudinary.txt
  ```

  **Commit**: YES | Message: `test(listing-images): cover image service workflows` | Files: `backend/internal/service/listing_service_test.go`, `backend/internal/domain/mocks/*`

- [x] T7. Add handler integration coverage, update Postman, and align AGENTS/docs

  **What to do**: Extend `backend/internal/handler/http/listing_test.go` with multipart upload tests and JSON mutation tests using a fake storage implementation injected through app wiring, not real Cloudinary. Assert read endpoints project ordered images after mutations. Update `backend/postman_collection.json` with listing-image requests and variables/examples. Update `AGENTS.md`, `backend/AGENTS.md`, `backend/internal/AGENTS.md`, `backend/internal/domain/AGENTS.md`, `backend/internal/service/AGENTS.md`, `backend/internal/repository/AGENTS.md`, and `backend/pkg/AGENTS.md`; update `backend/internal/handler/http/AGENTS.md` only if multipart upload conventions become a documented standard. Sync `backend/.env-example` and `backend/.env.docker` with runtime config.
  **Must NOT do**: Do not leave docs/env drift for follow-up; do not rely on manual Postman editing without checking the collection JSON structure already used in repo.

  **Recommended Agent Profile**:
  - Category: `writing` - Reason: docs/postman alignment plus integration verification
  - Skills: `[]` - no additional skill required
  - Omitted: `[]` - no omission needed

  **Parallelization**: Can Parallel: YES | Wave 2 | Blocks: none | Blocked By: T1, T2, T3, T4, T5

  **References**:
  - Pattern: `backend/internal/handler/http/listing_test.go` - integration suite setup and helpers
  - Pattern: `backend/postman_collection.json` - feature-folder request organization
  - Pattern: `AGENTS.md` - root project guidance structure
  - Pattern: `backend/AGENTS.md` - backend-specific summary and command sections
  - Pattern: `backend/internal/domain/AGENTS.md` - domain interface/entity guidance to align with new image/storage contracts
  - Pattern: `backend/pkg/AGENTS.md` - shared config/helper guidance
  - Pattern: `backend/.env-example` - env documentation file to sync with runtime names

  **Acceptance Criteria**:
  - [ ] Handler integration suite covers upload, delete, set-primary, reorder, and read projection with fake storage only
  - [ ] Postman collection includes listing-image requests under the listings feature area
  - [ ] AGENTS and env docs mention Cloudinary accurately and no longer describe image upload only as future work

  **QA Scenarios**:
  ```bash
  Scenario: Handler integration suite passes image flows
    Tool: Bash
    Steps: cd backend && go test ./... -count=1 -run TestListingHandlerSuite -v
    Expected: listing handler suite passes image upload/delete/primary/reorder/read-projection cases
    Evidence: .sisyphus/evidence/task-T7-handler-suite.txt

  Scenario: Docs and Postman mention Cloudinary feature accurately
    Tool: Bash
    Steps: python3 - <<'PY'
from pathlib import Path
docs = [
  'AGENTS.md',
  'backend/AGENTS.md',
  'backend/internal/AGENTS.md',
  'backend/internal/domain/AGENTS.md',
  'backend/internal/service/AGENTS.md',
  'backend/internal/repository/AGENTS.md',
  'backend/pkg/AGENTS.md',
  'backend/.env-example',
  'backend/postman_collection.json',
]
for path in docs:
    text = Path(path).read_text()
    assert 'Cloudinary' in text or 'listing image' in text.lower(), path
print('ok')
PY
    Expected: script prints `ok`
    Evidence: .sisyphus/evidence/task-T7-docs-postman.txt
  ```

  **Commit**: YES | Message: `docs(listing-images): add tests postman and cloudinary guidance` | Files: `backend/internal/handler/http/listing_test.go`, `backend/postman_collection.json`, `AGENTS.md`, `backend/**/*.md`, `backend/.env-example`, `backend/.env.docker`

## Final Verification Wave (4 parallel agents, ALL must APPROVE)
- [x] F1. Plan Compliance Audit - oracle
- [x] F2. Code Quality Review - reviewer
- [x] F3. Real API QA - tester
- [x] F4. Scope Fidelity Check - deep

## Commit Strategy
- Commit 1: schema/domain/storage foundation
- Commit 2: service + handler/API surface
- Commit 3: tests + Postman + AGENTS/docs sync
- If the executor chooses fewer commits, keep boundaries logical and do not mix docs-only cleanup before the feature works end-to-end.

## Success Criteria
- Owners/admins can upload, reorder, set primary, and delete listing images through authenticated backend endpoints.
- Listing read endpoints include ordered image projections without exposing Cloudinary internals.
- Cloudinary lifecycle operations are abstracted behind a mockable storage port.
- Tests and docs pass without manual verification or live third-party credentials.

# Monolith Search Indexing Refactor

## TL;DR
> **Summary**: Replace the current Kafka/Redpanda event transport with a modular monolith indexing flow using the same backend service, an internal DB outbox, and the existing Elasticsearch search read model.
> **Goal**: Keep search indexing reliable while removing broker complexity.

## Why This Direction
- Current repo is still operationally small enough for a modular monolith.
- Search indexing is the only real event consumer path right now.
- Redpanda/Kafka adds deployment and debugging cost without enough product leverage yet.
- The current code already has good seams: projector, indexer command, search transport, search-read service.

## Target Architecture

### Keep
- Single Go backend service as system-of-record
- Elasticsearch as the read model for buyer/public search
- `search_projector.go` logic
- `search_read_service.go`
- `listing-indexer rebuild` command/path

### Replace
- Kafka publisher/consumer flow with DB-backed outbox

### New Runtime Flow
1. Request succeeds in backend service
2. Same DB transaction writes domain row(s) + outbox row(s)
3. Internal worker scans pending outbox rows
4. Worker projects changes to Elasticsearch
5. Worker marks outbox rows processed (or failed for retry)

## Rollout Plan

- [x] T1. Freeze monolith indexing rules
- [x] T2. Add outbox schema and domain contracts
- [x] T3. Write outbox records from listing/category success paths
- [x] T4. Add internal outbox processor / worker entrypoint
- [x] T5. Reuse search projector for outbox processing
- [x] T6. Remove Kafka/Redpanda runtime dependency from app flow
- [x] T7. Update config, compose, docs, and verification
- [x] F1. Plan Compliance Audit
- [x] F2. Code Quality Review
- [x] F3. Runtime/Test Review
- [x] F4. Scope Fidelity Review

## Rules
- No Auth.js/session changes
- No microservice split
- No synchronous indexing in request path as the default
- Search rebuild command remains mandatory
- Outbox processing must be idempotent

## Verification Targets
- `cd backend && go test ./... -count=1`
- `cd backend && go build ./...`
- `cd backend && go vet ./...`
- replay/rebuild path still works after Kafka removal

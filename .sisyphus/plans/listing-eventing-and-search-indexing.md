# Plan: Listing Eventing and Search Indexing Foundation

## TL;DR
> **Summary**: The security/config cleanup is done. The next highest-value backend step is to make listing mutations publish durable events and feed an Elasticsearch index so buyer-facing discovery can grow without coupling search reads to the OLTP database.
> **Deliverables**:
> - backend config/env contract for Redpanda + Elasticsearch runtime usage
> - listing/category mutation event producer hooks from the system-of-record backend
> - worker/consumer path that projects events into a search document model
> - initial Elasticsearch index mapping plus backfill/replay strategy
> - focused verification for publish, consume, replay, and search-read correctness
> **Effort**: High
> **Parallel**: YES - 2 waves after contracts settle
> **Critical Path**: 1 -> 2 -> 3 -> 4 -> F1-F4

## Context

### Why This Next
- Redpanda is already present in `docker-compose.yml`, but application code still does not produce or consume domain events.
- Elasticsearch is already present in local infra, but buyer-facing search still depends on direct backend reads rather than a dedicated search projection.
- Listings and categories now have enough CRUD stability that mutation events can become the integration seam for search, notifications, and future commerce workflows.

### Current Repo Reality
- Backend remains the system of record for auth, listings, categories, and listing images.
- Listing prices already use `int64`, which is safe for downstream event payloads and search projections.
- There is no `workers/` directory yet, so the first indexing consumer should either live under `backend/cmd/` or a newly introduced, clearly scoped worker package.

## Work Objectives

### Core Objective
Publish stable listing-domain events from backend writes and consume them into an Elasticsearch-backed listing search index without weakening the existing handler -> service -> repository layering.

### Deliverables
- One canonical config contract for Kafka-compatible broker and Elasticsearch endpoints.
- Event schemas for listing/category lifecycle actions that preserve IDs, ownership, price, status, category, and image summary fields.
- Producer calls wired to successful write paths only.
- Consumer/indexer flow that is idempotent and supports replay/backfill.
- Search-read API contract proposal for future buyer-facing listing discovery enhancements.

## Architecture Direction

### Event Source
- Emit events only after successful DB commits for:
  - listing created
  - listing updated
  - listing deleted/deactivated
  - listing images changed
  - category updated when search-visible category data changes

### Transport
- Use Redpanda via Kafka-compatible clients already implied by local infra.
- Keep producer/consumer abstractions behind backend interfaces so handlers and domain contracts stay transport-agnostic.

### Search Projection
- Maintain one Elasticsearch document per listing.
- Denormalize only search-critical fields: title, slug, description excerpt, category metadata, seller ID, location, price, active image URLs, status, timestamps.
- Treat the DB as authoritative; search index is rebuildable.

### Replay Strategy
- Support a replay/backfill command that scans authoritative listing rows and republishes or directly reindexes documents.
- Design consumer handlers to be idempotent so at-least-once delivery remains safe.

## Execution Strategy

### 1. Config and Interface Contract
- Add broker/search config fields in backend config and env guidance.
- Introduce producer and indexer interfaces in the correct backend layer.
- Decide event envelope shape before wiring any publishers.

### 2. Producer Wiring
- Hook listing/category services after successful writes.
- Emit domain events with stable event names, entity IDs, timestamps, and versioned payloads.
- Keep request success semantics unchanged if async publishing is intentionally best-effort; otherwise fail loudly and document it.

### 3. Consumer and Indexer
- Add a worker entrypoint for consuming events.
- Translate events into Elasticsearch upsert/delete operations.
- Handle missing/deleted source records deterministically.

### 4. Verification and Backfill
- Add focused tests for event payload construction and consumer projection logic.
- Add a replay/backfill command for rebuilding the index from DB state.
- Verify docker/local env docs match the runtime contract.

## Definition of Done
- Listing/category mutations publish events from backend application code.
- A consumer can build and update Elasticsearch listing documents from those events.
- Replay/backfill exists for index recovery.
- Env/docs describe the actual broker/search configuration.
- Focused tests and build checks pass.

## TODOs
- [x] 1. Define broker/search config contract and update env guidance
- [x] 2. Add event envelope + listing/category event payload definitions
- [x] 3. Wire producers into listing/category success paths
- [x] 4. Add consumer/indexer worker entrypoint
- [x] 5. Implement Elasticsearch document mapping and upsert/delete logic
- [x] 6. Add replay/backfill path for index rebuilds
- [x] 7. Add focused tests and runtime verification

## Final Verification Wave
- [x] F1. Producer contract review
- [x] F2. Consumer idempotency review
- [x] F3. Search document quality review
- [x] F4. Env/docs accuracy review

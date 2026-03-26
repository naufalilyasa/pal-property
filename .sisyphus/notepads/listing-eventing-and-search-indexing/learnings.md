# Learnings

- Eventing/search config contract is already present in `backend/pkg/config/config.go` and `.env-example`, including Kafka brokers, topic names, and Elasticsearch index settings.
- Base event envelope and listing/category event payload types already exist in `backend/internal/domain/eventing.go`, so the plan can treat producer/consumer wiring as the next real implementation step rather than redoing contracts.
- Producer hooks can be wired and proven with fake publishers in service tests without requiring Redpanda or testcontainers; full integration verification still depends on Docker-enabled test environments.
- A compile-safe consumer entrypoint can be shipped before real Elasticsearch projection exists if transport is isolated behind `domain.SearchProjector` and the worker boots with a noop projector temporarily.
- A thin raw-HTTP Elasticsearch client is enough for first-pass upsert/delete and easier to test with `httptest` than a full vendor client.
- Rebuild/backfill can be implemented safely as a paginated repository walk from the worker command without waiting for full Kafka replay infrastructure.
- The most cost-effective proof for eventing/indexing in this repo is a targeted slice: `./internal/service`, `./cmd/property-service`, `./cmd/listing-indexer`, and `./pkg/...`; Docker-backed suites can be rerun later in a fuller environment, but producer/consumer/indexer behavior is already directly covered without them.
- Once Docker is available, the full backend suite (`go test ./...`) is green as well, so the focused eventing/indexing slice and the broader handler/migrate suites now both support final sign-off.

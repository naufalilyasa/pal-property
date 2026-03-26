# Decisions

- T1 is satisfied by the existing config contract in `backend/pkg/config/config.go`, `.env-example`, and related config tests.
- T2 is satisfied by the existing event envelope and typed listing/category event payloads in `backend/internal/domain/eventing.go`.
- The next meaningful implementation slice is producer wiring in listing/category success paths.
- T3 is considered complete once producer wiring, bootstrap injection, and fake-publisher service proof are in place; full handler/migrate suites are environment-blocked here and will be rerun later where Docker is available.
- T4 is considered complete with a compile-safe Kafka consumer loop, worker entrypoint, and projector abstraction. Real Elasticsearch projection logic belongs to T5.
- T5 is implemented with a raw HTTP Elasticsearch transport plus a service-layer projector that maps listing events directly and reindexes listings on category changes.
- T6 is implemented as a `rebuild` mode on `cmd/listing-indexer` that paginates through repository listings and upserts them into the configured index.

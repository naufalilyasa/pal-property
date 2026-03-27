# Decisions

- Backend remains a single modular monolith; no microservice split.
- Elasticsearch stays as the read model for buyer/public search.
- Kafka/Redpanda transport will be replaced by a DB-backed outbox inside the same service.
- Default indexing path should be asynchronous via outbox processor, not synchronous request-path indexing.
- Existing projector logic and rebuild command are retained and adapted rather than replaced.
- T2 introduces a dedicated `search_index_jobs` table plus domain contracts before changing any existing listing/category write path behavior.
- T3 uses additive dual-write: listing/category success paths keep current publisher hooks temporarily but now also enqueue `search_index_jobs` through a dedicated repository.
- T4 replaces the Kafka consumer loop in `cmd/listing-indexer` with an internal outbox processor.
- T5 is implemented by reusing the current `search_projector` as the single projection engine for both outbox processing and rebuilds.

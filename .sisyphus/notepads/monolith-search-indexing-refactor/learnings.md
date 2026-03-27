# Learnings

- The current codebase already has strong seams for a monolith refactor: search projector, read service, rebuild flow, and event payloads can be reused while only the transport changes.
- The safest migration is schema-first outbox with dual-write/cutover, not immediate synchronous indexing or immediate Kafka removal.
- T2 compiles cleanly with additive outbox schema/contracts, so the next risky step is not schema design but wiring outbox writes into current success-path transactions.
- T3 can start as a dual-write transitional step: keep existing publisher hooks for now, but enqueue search index jobs from successful listing/category mutations so Kafka removal can happen later without losing durability.
- T4/T5 can share one implementation slice: an internal outbox processor loop plus `listing-indexer` worker can directly reuse the existing `search_projector` logic, so no second projector path is needed.

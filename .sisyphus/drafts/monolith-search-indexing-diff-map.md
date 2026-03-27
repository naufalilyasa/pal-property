# File-by-File Diff Map: Kafka -> Monolith Outbox

## KEEP (mostly unchanged)
- `backend/pkg/searchindex/elasticsearch.go`
- `backend/internal/service/search_projector.go`
- `backend/internal/service/search_read_service.go`
- `backend/search-read-contract.md`
- buyer-facing search frontend files

## REPLACE / REWORK
- `backend/internal/domain/eventing.go`
  - keep payload structs if still useful
  - replace transport-centric interfaces with outbox/job-centric interfaces

- `backend/internal/service/listing_service.go`
  - replace direct publisher hook with outbox-write hook

- `backend/internal/service/category_service.go`
  - replace direct publisher hook with outbox-write hook

- `backend/cmd/listing-indexer/main.go`
  - replace Kafka consumer loop with outbox processor loop

## REMOVE (after migration)
- `backend/pkg/eventing/kafka_publisher.go`
- `backend/pkg/eventing/kafka_consumer.go`
- Kafka-specific tests under `backend/pkg/eventing/*`
- Kafka bootstrap from `backend/cmd/property-service/main.go`
- Redpanda-specific runtime assumptions in docs/compose (if no longer needed)

## UPDATE DOCS / CONFIG
- `backend/pkg/config/config.go`
- `backend/.env-example`
- `backend/AGENTS.md`
- `backend/pkg/AGENTS.md`
- `docker-compose.yml`

## POSSIBLY ADD
- `backend/internal/domain/indexing_job.go`
- `backend/internal/repository/postgres/indexing_job.go`
- `backend/internal/service/indexing_job_processor.go`
- `backend/db/migrations/<new>_search_index_jobs.up.sql`
- `backend/db/migrations/<new>_search_index_jobs.down.sql`

# Target Outbox Schema and Worker Flow

## Table: `search_index_jobs`

Recommended columns:

```sql
id UUID PRIMARY KEY
aggregate_type VARCHAR(32) NOT NULL   -- listing | category
aggregate_id UUID NOT NULL
event_type VARCHAR(64) NOT NULL       -- listing.created, category.updated, etc.
payload JSONB NOT NULL
status VARCHAR(20) NOT NULL DEFAULT 'pending'   -- pending | processing | done | failed
attempt_count INT NOT NULL DEFAULT 0
available_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
last_error TEXT NULL
created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
processed_at TIMESTAMP NULL
```

Recommended indexes:
- `(status, available_at)`
- `(aggregate_type, aggregate_id)`
- `(created_at)`

## Write Path
- listing/category service writes outbox row in same DB transaction as domain change
- payload can reuse existing event payload structs if desired

## Worker Flow
1. fetch a batch of `pending` jobs ordered by `created_at`
2. mark selected rows `processing`
3. for each job:
   - decode payload
   - call projector
   - on success => `done`, set `processed_at`
   - on failure => increment `attempt_count`, set `failed` or requeue with `available_at`

## Retry Policy
- simple capped retry in phase 1
- exponential backoff optional, not mandatory for first rollout

## Rebuild Path
- keep `listing-indexer rebuild`
- rebuild should recreate index and upsert all active listings from DB
- outbox does not replace rebuild; it complements it

## Why This Is Better Than Kafka Here
- fewer moving parts
- no broker operations burden
- same consistency boundary as the DB transaction
- easier local/dev debugging
- still leaves room to move to external broker later if scale demands it

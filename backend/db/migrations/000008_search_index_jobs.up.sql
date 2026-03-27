CREATE TABLE search_index_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(32) NOT NULL,
    aggregate_id UUID NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempt_count INT NOT NULL DEFAULT 0,
    available_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP,
    CONSTRAINT chk_search_index_jobs_status CHECK (status IN ('pending', 'processing', 'done', 'failed'))
);

CREATE INDEX idx_search_index_jobs_status_available ON search_index_jobs (status, available_at);
CREATE INDEX idx_search_index_jobs_aggregate ON search_index_jobs (aggregate_type, aggregate_id);
CREATE INDEX idx_search_index_jobs_created_at ON search_index_jobs (created_at);

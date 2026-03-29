CREATE TABLE listing_videos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id UUID NOT NULL UNIQUE,
    url TEXT NOT NULL,
    asset_id VARCHAR(255),
    public_id VARCHAR(255),
    version BIGINT,
    resource_type VARCHAR(50),
    delivery_type VARCHAR(50),
    format VARCHAR(50),
    bytes BIGINT,
    width INT,
    height INT,
    duration_seconds INT,
    original_filename VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE listing_videos
    ADD CONSTRAINT fk_listing_videos_listing FOREIGN KEY (listing_id)
        REFERENCES listings(id)
        ON DELETE CASCADE;

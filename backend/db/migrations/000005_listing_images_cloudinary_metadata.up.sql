ALTER TABLE listing_images
    ADD COLUMN asset_id VARCHAR(255),
    ADD COLUMN public_id VARCHAR(255),
    ADD COLUMN version BIGINT,
    ADD COLUMN resource_type VARCHAR(50),
    ADD COLUMN type VARCHAR(50),
    ADD COLUMN format VARCHAR(50),
    ADD COLUMN bytes BIGINT,
    ADD COLUMN width INT,
    ADD COLUMN height INT,
    ADD COLUMN original_filename VARCHAR(255),
    ADD COLUMN deleted_at TIMESTAMP;

CREATE INDEX idx_listing_images_deleted_at ON listing_images(deleted_at);

CREATE UNIQUE INDEX idx_listing_images_listing_sort_order_active
    ON listing_images (listing_id, sort_order)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX idx_listing_images_primary_active
    ON listing_images (listing_id)
    WHERE is_primary IS TRUE AND deleted_at IS NULL;

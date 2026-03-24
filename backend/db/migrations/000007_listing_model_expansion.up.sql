-- 1. Add richer listing columns without removing compatibility fields.
ALTER TABLE listings
    ADD COLUMN transaction_type VARCHAR(20) NOT NULL DEFAULT 'sale',
    ADD COLUMN is_negotiable BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN special_offers JSONB NOT NULL DEFAULT '[]',
    ADD COLUMN location_province VARCHAR(100),
    ADD COLUMN latitude DECIMAL(10,8),
    ADD COLUMN longitude DECIMAL(11,8),
    ADD COLUMN bedroom_count INT,
    ADD COLUMN bathroom_count INT,
    ADD COLUMN floor_count INT,
    ADD COLUMN carport_capacity INT,
    ADD COLUMN land_area_sqm INT,
    ADD COLUMN building_area_sqm INT,
    ADD COLUMN certificate_type VARCHAR(50),
    ADD COLUMN condition VARCHAR(50),
    ADD COLUMN furnishing VARCHAR(50),
    ADD COLUMN electrical_power_va INT,
    ADD COLUMN facing_direction VARCHAR(50),
    ADD COLUMN year_built INT,
    ADD COLUMN facilities JSONB NOT NULL DEFAULT '[]';

-- 2. Add indexes for agreed phase-1 filters.
CREATE INDEX idx_listings_transaction_type ON listings (transaction_type);
CREATE INDEX idx_listings_location_province ON listings (location_province);
CREATE INDEX idx_listings_bedroom_count ON listings (bedroom_count);
CREATE INDEX idx_listings_bathroom_count ON listings (bathroom_count);
CREATE INDEX idx_listings_land_area_sqm ON listings (land_area_sqm);
CREATE INDEX idx_listings_building_area_sqm ON listings (building_area_sqm);
CREATE INDEX idx_listings_certificate_type ON listings (certificate_type);
CREATE INDEX idx_listings_condition ON listings (condition);
CREATE INDEX idx_listings_furnishing ON listings (furnishing);

-- 3. Expand allowed status values.
ALTER TABLE listings DROP CONSTRAINT IF EXISTS chk_listings_status;
ALTER TABLE listings
    ADD CONSTRAINT chk_listings_status
    CHECK (status IN ('active', 'inactive', 'sold', 'draft', 'archived'));

ALTER TABLE listings DROP CONSTRAINT IF EXISTS chk_listings_transaction_type;
ALTER TABLE listings
    ADD CONSTRAINT chk_listings_transaction_type
    CHECK (transaction_type IN ('sale', 'rent'));

-- 4. Backfill typed compatibility columns from legacy specifications JSON when present.
UPDATE listings
SET
    bedroom_count = CASE
        WHEN bedroom_count IS NULL AND specifications ? 'bedrooms' THEN (specifications->>'bedrooms')::INT
        ELSE bedroom_count
    END,
    bathroom_count = CASE
        WHEN bathroom_count IS NULL AND specifications ? 'bathrooms' THEN (specifications->>'bathrooms')::INT
        ELSE bathroom_count
    END,
    land_area_sqm = CASE
        WHEN land_area_sqm IS NULL AND specifications ? 'land_area_sqm' THEN (specifications->>'land_area_sqm')::INT
        ELSE land_area_sqm
    END,
    building_area_sqm = CASE
        WHEN building_area_sqm IS NULL AND specifications ? 'building_area_sqm' THEN (specifications->>'building_area_sqm')::INT
        ELSE building_area_sqm
    END;

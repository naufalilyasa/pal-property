CREATE TABLE IF NOT EXISTS indonesia_regions (
    code VARCHAR(13) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    level SMALLINT NOT NULL CHECK (level BETWEEN 1 AND 4),
    parent_code VARCHAR(13),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_indonesia_regions_level ON indonesia_regions(level);
CREATE INDEX IF NOT EXISTS idx_indonesia_regions_parent_code ON indonesia_regions(parent_code);
CREATE INDEX IF NOT EXISTS idx_indonesia_regions_level_parent_code ON indonesia_regions(level, parent_code);

ALTER TABLE listings
    ADD COLUMN IF NOT EXISTS location_province_code VARCHAR(13),
    ADD COLUMN IF NOT EXISTS location_city_code VARCHAR(13),
    ADD COLUMN IF NOT EXISTS location_district_code VARCHAR(13),
    ADD COLUMN IF NOT EXISTS location_village_code VARCHAR(13),
    ADD COLUMN IF NOT EXISTS location_village VARCHAR(100);

CREATE INDEX IF NOT EXISTS idx_listings_location_province_code ON listings(location_province_code);
CREATE INDEX IF NOT EXISTS idx_listings_location_city_code ON listings(location_city_code);
CREATE INDEX IF NOT EXISTS idx_listings_location_district_code ON listings(location_district_code);
CREATE INDEX IF NOT EXISTS idx_listings_location_village_code ON listings(location_village_code);

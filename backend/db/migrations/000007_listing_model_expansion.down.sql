ALTER TABLE listings DROP CONSTRAINT IF EXISTS chk_listings_transaction_type;

UPDATE listings
SET status = 'inactive'
WHERE status IN ('draft', 'archived');

ALTER TABLE listings DROP CONSTRAINT IF EXISTS chk_listings_status;
ALTER TABLE listings
    ADD CONSTRAINT chk_listings_status
    CHECK (status IN ('active', 'inactive', 'sold'));

DROP INDEX IF EXISTS idx_listings_furnishing;
DROP INDEX IF EXISTS idx_listings_condition;
DROP INDEX IF EXISTS idx_listings_certificate_type;
DROP INDEX IF EXISTS idx_listings_building_area_sqm;
DROP INDEX IF EXISTS idx_listings_land_area_sqm;
DROP INDEX IF EXISTS idx_listings_bathroom_count;
DROP INDEX IF EXISTS idx_listings_bedroom_count;
DROP INDEX IF EXISTS idx_listings_location_province;
DROP INDEX IF EXISTS idx_listings_transaction_type;

ALTER TABLE listings
    DROP COLUMN IF EXISTS facilities,
    DROP COLUMN IF EXISTS year_built,
    DROP COLUMN IF EXISTS facing_direction,
    DROP COLUMN IF EXISTS electrical_power_va,
    DROP COLUMN IF EXISTS furnishing,
    DROP COLUMN IF EXISTS condition,
    DROP COLUMN IF EXISTS certificate_type,
    DROP COLUMN IF EXISTS building_area_sqm,
    DROP COLUMN IF EXISTS land_area_sqm,
    DROP COLUMN IF EXISTS carport_capacity,
    DROP COLUMN IF EXISTS floor_count,
    DROP COLUMN IF EXISTS bathroom_count,
    DROP COLUMN IF EXISTS bedroom_count,
    DROP COLUMN IF EXISTS longitude,
    DROP COLUMN IF EXISTS latitude,
    DROP COLUMN IF EXISTS location_province,
    DROP COLUMN IF EXISTS special_offers,
    DROP COLUMN IF EXISTS is_negotiable,
    DROP COLUMN IF EXISTS transaction_type;

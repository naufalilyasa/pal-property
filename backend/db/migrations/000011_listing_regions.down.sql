DROP INDEX IF EXISTS idx_listings_location_village_code;
DROP INDEX IF EXISTS idx_listings_location_district_code;
DROP INDEX IF EXISTS idx_listings_location_city_code;
DROP INDEX IF EXISTS idx_listings_location_province_code;

ALTER TABLE listings
    DROP COLUMN IF EXISTS location_village,
    DROP COLUMN IF EXISTS location_village_code,
    DROP COLUMN IF EXISTS location_district_code,
    DROP COLUMN IF EXISTS location_city_code,
    DROP COLUMN IF EXISTS location_province_code;

DROP INDEX IF EXISTS idx_indonesia_regions_level_parent_code;
DROP INDEX IF EXISTS idx_indonesia_regions_parent_code;
DROP INDEX IF EXISTS idx_indonesia_regions_level;

DROP TABLE IF EXISTS indonesia_regions;

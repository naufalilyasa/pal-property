ALTER TABLE listings DROP CONSTRAINT IF EXISTS chk_listings_status;
DROP INDEX IF EXISTS idx_listings_slug_active;
ALTER TABLE listings ADD CONSTRAINT uni_listings_slug UNIQUE (slug);
ALTER TABLE listings ALTER COLUMN price TYPE DECIMAL(18,2) USING price::DECIMAL(18,2);

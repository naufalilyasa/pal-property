-- 1. Fix price column: DECIMAL(18,2) → BIGINT (match int64 entity convention)
ALTER TABLE listings ALTER COLUMN price TYPE BIGINT USING price::BIGINT;

-- 2. Drop plain unique index on slug (blocks soft-delete reuse)
DROP INDEX IF EXISTS idx_listings_slug; -- check exact name from 000001
ALTER TABLE listings DROP CONSTRAINT IF EXISTS uni_listings_slug;

-- 3. Partial unique index: only enforce uniqueness for non-deleted rows
CREATE UNIQUE INDEX idx_listings_slug_active
    ON listings (slug) WHERE deleted_at IS NULL;

-- 4. Add CHECK constraint for status values
ALTER TABLE listings
    ADD CONSTRAINT chk_listings_status
    CHECK (status IN ('active', 'inactive', 'sold'));

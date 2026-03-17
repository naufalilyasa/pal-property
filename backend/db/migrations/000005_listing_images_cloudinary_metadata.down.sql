DROP INDEX IF EXISTS idx_listing_images_primary_active;
DROP INDEX IF EXISTS idx_listing_images_listing_sort_order_active;
DROP INDEX IF EXISTS idx_listing_images_deleted_at;

ALTER TABLE listing_images
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS original_filename,
    DROP COLUMN IF EXISTS height,
    DROP COLUMN IF EXISTS width,
    DROP COLUMN IF EXISTS bytes,
    DROP COLUMN IF EXISTS format,
    DROP COLUMN IF EXISTS type,
    DROP COLUMN IF EXISTS resource_type,
    DROP COLUMN IF EXISTS version,
    DROP COLUMN IF EXISTS public_id,
    DROP COLUMN IF EXISTS asset_id;

-- Remove seed categories (children first due to FK self-reference, then parents)
DELETE FROM categories WHERE id LIKE '11111111-0000-7000-8000-%';

-- Revert FK to default (RESTRICT — Postgres default with no ON DELETE clause)
ALTER TABLE listings
    DROP CONSTRAINT IF EXISTS listings_category_id_fkey,
    ADD CONSTRAINT listings_category_id_fkey
        FOREIGN KEY (category_id) REFERENCES categories(id);

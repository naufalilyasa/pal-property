-- Drop existing FK constraint and re-add with ON DELETE SET NULL
ALTER TABLE listings
    DROP CONSTRAINT IF EXISTS listings_category_id_fkey,
    ADD CONSTRAINT listings_category_id_fkey
        FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;

-- Seed Root categories (3)
INSERT INTO categories (id, name, slug, parent_id, icon_url, created_at) VALUES
    ('11111111-0000-7000-8000-000000000001', 'Residensial', 'residensial', NULL, NULL, NOW()),
    ('11111111-0000-7000-8000-000000000002', 'Komersial',   'komersial',   NULL, NULL, NOW()),
    ('11111111-0000-7000-8000-000000000003', 'Lainnya',     'lainnya',     NULL, NULL, NOW())
ON CONFLICT (slug) DO NOTHING;

-- Seed Children of Residensial (5)
INSERT INTO categories (id, name, slug, parent_id, icon_url, created_at) VALUES
    ('11111111-0000-7000-8000-000000000011', 'Rumah',     'rumah',     '11111111-0000-7000-8000-000000000001', NULL, NOW()),
    ('11111111-0000-7000-8000-000000000012', 'Apartemen', 'apartemen', '11111111-0000-7000-8000-000000000001', NULL, NOW()),
    ('11111111-0000-7000-8000-000000000013', 'Kos',       'kos',       '11111111-0000-7000-8000-000000000001', NULL, NOW()),
    ('11111111-0000-7000-8000-000000000014', 'Villa',     'villa',     '11111111-0000-7000-8000-000000000001', NULL, NOW()),
    ('11111111-0000-7000-8000-000000000015', 'Townhouse', 'townhouse', '11111111-0000-7000-8000-000000000001', NULL, NOW())
ON CONFLICT (slug) DO NOTHING;

-- Seed Children of Komersial (4)
INSERT INTO categories (id, name, slug, parent_id, icon_url, created_at) VALUES
    ('11111111-0000-7000-8000-000000000021', 'Ruko',   'ruko',   '11111111-0000-7000-8000-000000000002', NULL, NOW()),
    ('11111111-0000-7000-8000-000000000022', 'Kantor', 'kantor', '11111111-0000-7000-8000-000000000002', NULL, NOW()),
    ('11111111-0000-7000-8000-000000000023', 'Gudang', 'gudang', '11111111-0000-7000-8000-000000000002', NULL, NOW()),
    ('11111111-0000-7000-8000-000000000024', 'Toko',   'toko',   '11111111-0000-7000-8000-000000000002', NULL, NOW())
ON CONFLICT (slug) DO NOTHING;

-- Seed Children of Lainnya (4)
INSERT INTO categories (id, name, slug, parent_id, icon_url, created_at) VALUES
    ('11111111-0000-7000-8000-000000000031', 'Tanah',  'tanah',  '11111111-0000-7000-8000-000000000003', NULL, NOW()),
    ('11111111-0000-7000-8000-000000000032', 'Kavling','kavling','11111111-0000-7000-8000-000000000003', NULL, NOW()),
    ('11111111-0000-7000-8000-000000000033', 'Kebun',  'kebun',  '11111111-0000-7000-8000-000000000003', NULL, NOW()),
    ('11111111-0000-7000-8000-000000000034', 'Ruko Perumahan', 'ruko-perumahan', '11111111-0000-7000-8000-000000000003', NULL, NOW())
ON CONFLICT (slug) DO NOTHING;

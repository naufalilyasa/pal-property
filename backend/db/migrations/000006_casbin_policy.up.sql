CREATE TABLE IF NOT EXISTS casbin_rule (
    id BIGSERIAL PRIMARY KEY,
    ptype VARCHAR(100) NOT NULL,
    v0 VARCHAR(255) NOT NULL DEFAULT '',
    v1 VARCHAR(255) NOT NULL DEFAULT '',
    v2 VARCHAR(255) NOT NULL DEFAULT '',
    v3 VARCHAR(255) NOT NULL DEFAULT '',
    v4 VARCHAR(255) NOT NULL DEFAULT '',
    v5 VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_casbin_rule_tuple
    ON casbin_rule (ptype, v0, v1, v2, v3, v4, v5);

CREATE INDEX IF NOT EXISTS idx_casbin_rule_lookup
    ON casbin_rule (ptype, v0, v1, v2);

INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', 'admin', 'category', 'create'),
    ('p', 'admin', 'category', 'update'),
    ('p', 'admin', 'category', 'delete'),
    ('p', 'admin', 'listing', 'update'),
    ('p', 'admin', 'listing', 'delete'),
    ('p', 'admin', 'listing', 'upload_image'),
    ('p', 'admin', 'listing', 'delete_image'),
    ('p', 'admin', 'listing', 'set_primary_image'),
    ('p', 'admin', 'listing', 'reorder_images'),
    ('p', 'admin', 'listing', 'upload_video'),
    ('p', 'admin', 'listing', 'delete_video'),
    ('p', 'owner', 'listing', 'update'),
    ('p', 'owner', 'listing', 'delete'),
    ('p', 'owner', 'listing', 'upload_image'),
    ('p', 'owner', 'listing', 'delete_image'),
    ('p', 'owner', 'listing', 'set_primary_image'),
    ('p', 'owner', 'listing', 'reorder_images'),
    ('p', 'owner', 'listing', 'upload_video'),
    ('p', 'owner', 'listing', 'delete_video')
ON CONFLICT DO NOTHING;

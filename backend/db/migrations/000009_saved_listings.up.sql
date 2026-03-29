CREATE TABLE saved_listings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    listing_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE saved_listings
    ADD CONSTRAINT fk_saved_listings_user FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,
    ADD CONSTRAINT fk_saved_listings_listing FOREIGN KEY (listing_id)
        REFERENCES listings(id)
        ON DELETE CASCADE,
    ADD CONSTRAINT uq_saved_listings_user_listing UNIQUE (user_id, listing_id);

CREATE INDEX idx_saved_listings_user_id ON saved_listings (user_id);
CREATE INDEX idx_saved_listings_listing_id ON saved_listings (listing_id);

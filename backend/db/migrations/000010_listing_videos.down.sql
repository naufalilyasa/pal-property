ALTER TABLE listing_videos
    DROP CONSTRAINT IF EXISTS fk_listing_videos_listing;

DROP TABLE IF EXISTS listing_videos;

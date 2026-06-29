ALTER TABLE listingservice.listings
    ADD COLUMN photo_urls TEXT[] NOT NULL DEFAULT '{}';

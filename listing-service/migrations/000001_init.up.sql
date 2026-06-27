CREATE SCHEMA listingservice;

CREATE TYPE listingservice.listing_status AS ENUM ('active', 'inactive', 'sold');
CREATE TYPE listingservice.body_type AS ENUM ('sedan', 'suv', 'hatchback', 'coupe', 'wagon', 'minivan', 'pickup');
CREATE TYPE listingservice.fuel_type AS ENUM ('gasoline', 'diesel', 'electric', 'hybrid', 'lpg');
CREATE TYPE listingservice.transmission_type AS ENUM ('automatic', 'manual', 'robot', 'variator');

CREATE TABLE IF NOT EXISTS listingservice.listings (
    id            UUID PRIMARY KEY,
    user_id       UUID NOT NULL,
    title         VARCHAR(200) NOT NULL CHECK (char_length(title) BETWEEN 3 AND 200),
    description   TEXT NOT NULL DEFAULT '',
    price         BIGINT NOT NULL CHECK (price > 0),
    status        listingservice.listing_status NOT NULL DEFAULT 'active',

    make          VARCHAR(100) NOT NULL,
    model         VARCHAR(100) NOT NULL,
    year          SMALLINT NOT NULL CHECK (year BETWEEN 1900 AND 2100),
    mileage       INT NOT NULL DEFAULT 0 CHECK (mileage >= 0),
    color         VARCHAR(50) NOT NULL DEFAULT '',
    body_type     listingservice.body_type NOT NULL,
    fuel_type     listingservice.fuel_type NOT NULL,
    transmission  listingservice.transmission_type NOT NULL,
    engine_volume NUMERIC(4, 1) NOT NULL DEFAULT 0 CHECK (engine_volume >= 0),

    city          VARCHAR(100) NOT NULL,
    region        VARCHAR(100) NOT NULL DEFAULT '',

    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_listings_user_id    ON listingservice.listings(user_id);
CREATE INDEX IF NOT EXISTS idx_listings_status     ON listingservice.listings(status);
CREATE INDEX IF NOT EXISTS idx_listings_make_model ON listingservice.listings(make, model);
CREATE INDEX IF NOT EXISTS idx_listings_price      ON listingservice.listings(price);
CREATE INDEX IF NOT EXISTS idx_listings_city       ON listingservice.listings(city);
CREATE INDEX IF NOT EXISTS idx_listings_created_at ON listingservice.listings(created_at DESC);

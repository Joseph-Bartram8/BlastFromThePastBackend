-- postgis used for geometry data types
CREATE EXTENSION IF NOT EXISTS postgis;

-- Enable UUID Extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users Table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

-- User Bios Table
CREATE TABLE user_bios (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    display_name VARCHAR(100) NOT NULL,
    store_name VARCHAR(100),
    bio_description TEXT,
    profile_image TEXT,
    show_real_name BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- User Markers Table
CREATE TABLE user_markers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    region TEXT NOT NULL CHECK (region IN (
        'North East', 'North West', 'Yorkshire and the Humber', 'West Midlands', 
        'East Midlands', 'South West', 'South East', 'London', 'East of England'
    )),
    marker_type TEXT NOT NULL CHECK (marker_type IN ('Shop', 'Collector', 'Event', 'Trade Meetup')),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Trigger to Auto-Update updated_at in user_bios
CREATE OR REPLACE FUNCTION update_user_bio_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_user_bio
BEFORE UPDATE ON user_bios
FOR EACH ROW
EXECUTE FUNCTION update_user_bio_timestamp();

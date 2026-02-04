-- Pahlawan Pangan Database Schema
-- Designed for 100TB+ scale with geo-sharding and time-based partitioning

-- Extension for PostGIS (Geo-spatial queries)
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS pg_partman;

-- Geo-region lookup table (for sharding)
CREATE TABLE geo_regions (
    id SERIAL PRIMARY KEY,
    region_name VARCHAR(100) NOT NULL,
    s2_cell_id BIGINT NOT NULL UNIQUE,
    geometry GEOGRAPHY(POLYGON, 4326) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_geo_regions_s2 ON geo_regions(s2_cell_id);
CREATE INDEX idx_geo_regions_geom ON geo_regions USING GIST(geometry);

-- Providers (Restaurants, Hotels, Groceries)
CREATE TABLE providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- 'restaurant', 'hotel', 'grocery'
    location GEOGRAPHY(POINT, 4326) NOT NULL,
    geo_region_id INT REFERENCES geo_regions(id),
    contact_phone VARCHAR(20),
    contact_email VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_providers_location ON providers USING GIST(location);
CREATE INDEX idx_providers_geo_region ON providers(geo_region_id);

-- NGOs and Food Banks
CREATE TABLE ngos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    location GEOGRAPHY(POINT, 4326) NOT NULL,
    geo_region_id INT REFERENCES geo_regions(id),
    capacity_kgs_per_day DECIMAL(10, 2),
    contact_phone VARCHAR(20),
    contact_email VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_ngos_location ON ngos USING GIST(location);
CREATE INDEX idx_ngos_geo_region ON ngos(geo_region_id);

-- Surplus (Partitioned by created_at and geo_region_id)
CREATE TABLE surplus (
    id UUID DEFAULT gen_random_uuid(),
    provider_id UUID NOT NULL REFERENCES providers(id),
    geo_region_id INT NOT NULL REFERENCES geo_regions(id),
    location GEOGRAPHY(POINT, 4326) NOT NULL,
    quantity_kgs DECIMAL(10, 2) NOT NULL,
    food_type VARCHAR(100),
    expiry_time TIMESTAMP NOT NULL,
    status VARCHAR(20) DEFAULT 'available', -- 'available', 'claimed', 'expired', 'completed'
    claimed_by_ngo_id UUID REFERENCES ngos(id),
    claimed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    version INT DEFAULT 1, -- Optimistic locking
    PRIMARY KEY (id, created_at, geo_region_id)
) PARTITION BY RANGE (created_at);

-- Create monthly partitions (managed by pg_partman)
SELECT partman.create_parent(
    p_parent_table := 'public.surplus',
    p_control := 'created_at',
    p_type := 'native',
    p_interval := '1 month',
    p_premake := 3
);

CREATE INDEX idx_surplus_location ON surplus USING GIST(location);
CREATE INDEX idx_surplus_status ON surplus(status, expiry_time);
CREATE INDEX idx_surplus_geo_region ON surplus(geo_region_id, created_at);
CREATE INDEX idx_surplus_provider ON surplus(provider_id, created_at);

-- Outbox Events (Transactional Outbox Pattern)
CREATE TABLE outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    published BOOLEAN DEFAULT FALSE,
    published_at TIMESTAMP,
    trace_id VARCHAR(64),
    retry_count INT DEFAULT 0
);

CREATE INDEX idx_outbox_unpublished ON outbox_events(published, created_at) WHERE published = false;
CREATE INDEX idx_outbox_trace ON outbox_events(trace_id);

-- Matching History (for analytics and ML)
CREATE TABLE matching_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    surplus_id UUID NOT NULL,
    ngo_id UUID NOT NULL,
    distance_km DECIMAL(10, 2),
    travel_time_seconds INT,
    matched_at TIMESTAMP DEFAULT NOW(),
    claim_latency_seconds DECIMAL(10, 3),
    successful BOOLEAN DEFAULT TRUE
) PARTITION BY RANGE (matched_at);

SELECT partman.create_parent(
    p_parent_table := 'public.matching_history',
    p_control := 'matched_at',
    p_type := 'native',
    p_interval := '1 month',
    p_premake := 3
);

CREATE INDEX idx_matching_surplus ON matching_history(surplus_id, matched_at);
CREATE INDEX idx_matching_ngo ON matching_history(ngo_id, matched_at);

-- Dead Letter Queue for failed matches
CREATE TABLE dlq_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    surplus_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    error_message TEXT,
    retry_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    resolved BOOLEAN DEFAULT FALSE,
    resolved_at TIMESTAMP
);

CREATE INDEX idx_dlq_unresolved ON dlq_events(resolved, created_at) WHERE resolved = false;

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for updated_at
CREATE TRIGGER update_providers_updated_at BEFORE UPDATE ON providers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ngos_updated_at BEFORE UPDATE ON ngos
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_surplus_updated_at BEFORE UPDATE ON surplus
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

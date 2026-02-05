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
    original_price DECIMAL(10, 2), -- For B2C market
    discount_price DECIMAL(10, 2), -- Calculated dynamically
    is_donation BOOLEAN DEFAULT TRUE,
    impact_points INT DEFAULT 0, -- Gamification for providers
    temperature_category VARCHAR(20) DEFAULT 'ambient', -- 'ambient', 'chilled', 'frozen', 'hot'
    health_certificate_url TEXT,
    safety_window_minutes INT DEFAULT 120, -- Default 2 hours safety window
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

-- Gamification: User Profiles (Consumers/People)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255),
    email VARCHAR(255) UNIQUE,
    total_impact_points INT DEFAULT 0,
    food_saved_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Leaderboard cache
CREATE TABLE leaderboards (
    region_id INT REFERENCES geo_regions(id),
    entity_id UUID, -- Can be Provider or User
    entity_name VARCHAR(255),
    entity_type VARCHAR(20), -- 'provider', 'user'
    rank INT,
    points INT,
    updated_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (region_id, entity_id)
);

-- Social Impact Feed (Unicorn Feature)
CREATE TABLE social_feed (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    provider_id UUID REFERENCES providers(id),
    surplus_id UUID,
    content_type VARCHAR(20), -- 'rescue_success', 'milestone', 'donation'
    media_url TEXT, -- Shared photo of the food rescue
    caption TEXT,
    likes_count INT DEFAULT 0,
    cheers_count INT DEFAULT 0, -- Like Strava 'Kudos'
    created_at TIMESTAMP DEFAULT NOW()
);

-- B2B: Subscription & Zero Waste Certification
CREATE TABLE business_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id UUID REFERENCES providers(id),
    plan_type VARCHAR(50), -- 'standard', 'premium', 'zero_waste_pro'
    status VARCHAR(20),
    next_billing_at TIMESTAMP,
    features JSONB -- e.g., {"priority_matching": true, "tax_report": true}
);

-- AI/Analytics: Waste Prediction Models
CREATE TABLE waste_predictions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id UUID REFERENCES providers(id),
    predicted_date DATE,
    predicted_quantity_kgs DECIMAL(10, 2),
    confidence_score DECIMAL(5, 2),
    context JSONB, -- e.g., {"event": "hujan_deras", "holiday": false}
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_social_user ON social_feed(user_id, created_at);
CREATE INDEX idx_predictions_provider ON waste_predictions(provider_id, predicted_date);

-- Pahlawan-Express: Logistics & Delivery
CREATE TABLE deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    surplus_id UUID NOT NULL,
    courier_id UUID, -- NULL if self-pickup
    status VARCHAR(20), -- 'searching', 'assigned', 'picked_up', 'delivered', 'failed', 'ready_for_pickup'
    fee DECIMAL(10, 2),
    courier_points INT,
    requires_cold_chain BOOLEAN DEFAULT FALSE,
    thermal_bag_verified BOOLEAN DEFAULT FALSE,
    fulfillment_method VARCHAR(20) DEFAULT 'courier', -- 'courier', 'self_pickup'
    pickup_verification_code VARCHAR(10), -- For self-pickup: QR/OTP
    is_verified_pickup BOOLEAN DEFAULT FALSE,
    external_tracking_id VARCHAR(255), -- Gojek/Grab Booking ID
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Self-Pickup Verification (Trust & Security)
CREATE TABLE pickup_checkins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    delivery_id UUID REFERENCES deliveries(id),
    provider_id UUID REFERENCES providers(id),
    checked_at TIMESTAMP DEFAULT NOW(),
    location GEOGRAPHY(POINT, 4326) -- Validate user is actually at the store
);


-- Pahlawan-Connect: POS Integration
CREATE TABLE pos_integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id UUID REFERENCES providers(id),
    integration_type VARCHAR(50), -- 'moka', 'majoo', 'esensi'
    api_key_encrypted BYTEA,
    last_sync_at TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);

-- Pahlawan-Carbon: ESG & Carbon Credits
CREATE TABLE carbon_credits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id UUID REFERENCES providers(id),
    total_co2_saved_kg DECIMAL(15, 2),
    credit_tokens BIGINT, -- 1 token per 100kg CO2 saved
    last_issued_at TIMESTAMP,
    is_certified BOOLEAN DEFAULT FALSE
);

-- Pahlawan-Comm: Community Group Buy
CREATE TABLE community_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    region_id INT REFERENCES geo_regions(id),
    group_name VARCHAR(100),
    coordinator_user_id UUID REFERENCES users(id),
    total_members INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE group_buy_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID REFERENCES community_groups(id),
    surplus_id UUID,
    total_quantity_kgs DECIMAL(10, 2),
    status VARCHAR(20) DEFAULT 'forming', -- 'forming', 'locked', 'completed'
    expires_at TIMESTAMP
);

CREATE INDEX idx_deliveries_surplus ON deliveries(surplus_id);
CREATE INDEX idx_carbon_provider ON carbon_credits(provider_id);
CREATE INDEX idx_comm_region ON community_groups(region_id);

-- Pahlawan-Comm: Community Drop Points (RT/RW Hubs)
CREATE TABLE community_drop_points (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    region_id INT REFERENCES geo_regions(id),
    name VARCHAR(100), -- e.g., 'Pos Satpam Cluster A', 'Rumah Ketua RT 05'
    address TEXT,
    geometry GEOGRAPHY(POINT, 4326),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Analytics: Provider ROI & Sustainability Impact
CREATE TABLE provider_analytics (
    provider_id UUID REFERENCES providers(id),
    total_revenue_saved DECIMAL(15, 2),
    total_waste_prevented_kgs DECIMAL(10, 2),
    carbon_offset_kg DECIMAL(10, 2),
    last_updated TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (provider_id)
);

-- Rating System: Trust & Safety (Standard Uber/Gojek)
CREATE TABLE ratings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_id UUID NOT NULL, -- ProviderID or CourierID
    user_id UUID NOT NULL REFERENCES users(id),
    score INT CHECK (score BETWEEN 1 AND 5),
    comment TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Voucher & Promo System (Standard Tokopedia)
CREATE TABLE vouchers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) UNIQUE NOT NULL,
    discount_type VARCHAR(20), -- 'percentage', 'fixed_amount'
    value DECIMAL(10, 2),
    min_order_idr DECIMAL(10, 2),
    max_discount_idr DECIMAL(10, 2),
    starts_at TIMESTAMP,
    expires_at TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);

-- Chat & Communication (Meta-data for Threads)
CREATE TABLE chat_threads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    surplus_id UUID NOT NULL,
    participant_a_id UUID NOT NULL, -- User/NGO
    participant_b_id UUID NOT NULL, -- Provider/Courier
    last_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- User/Provider Settings (Account Management)
CREATE TABLE account_settings (
    entity_id UUID PRIMARY KEY, -- User or Provider
    preferences JSONB DEFAULT '{}', -- e.g., {"notif_enabled": true, "theme": "dark"}
    security_verified BOOLEAN DEFAULT FALSE,
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_ratings_entity ON ratings(entity_id);
CREATE INDEX idx_vouchers_code ON vouchers(code) WHERE is_active = true;
CREATE INDEX idx_chat_surplus ON chat_threads(surplus_id);


CREATE INDEX idx_drop_points_geo ON community_drop_points USING GIST(geometry);

-- Safety & Liability: Digital Agreements
CREATE TABLE safety_liability_agreements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_id UUID NOT NULL, -- UserID or NGOID
    agreement_version VARCHAR(20),
    accepted_at TIMESTAMP DEFAULT NOW(),
    ip_address VARCHAR(45)
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

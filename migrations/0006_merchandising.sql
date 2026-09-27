-- Plane 9 merchandising: advertisers pay Nibble for labelled placements (never read by compare, D28).
-- Also fills plane 6 gaps: promotion fulfillment_mode / description, legacy restaurant targets (MVP shortcut).

ALTER TABLE promotions ADD COLUMN IF NOT EXISTS fulfillment_mode TEXT NOT NULL DEFAULT '';
ALTER TABLE promotions ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';

ALTER TABLE promotion_targets
    ADD COLUMN IF NOT EXISTS legacy_restaurant_id TEXT REFERENCES restaurants (id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_promotions_window ON promotions (starts_at, ends_at);
CREATE INDEX IF NOT EXISTS idx_promotion_targets_promotion ON promotion_targets (promotion_id);

CREATE TABLE IF NOT EXISTS advertisers (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    brand_id TEXT REFERENCES brands (id) ON DELETE SET NULL,
    contact_email TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sponsored_campaigns (
    id TEXT PRIMARY KEY,
    advertiser_id TEXT NOT NULL REFERENCES advertisers (id) ON DELETE CASCADE,
    market_id TEXT REFERENCES markets (id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    pricing_model TEXT NOT NULL,
    bid_cents BIGINT NOT NULL DEFAULT 0,
    daily_budget_cents BIGINT NOT NULL DEFAULT 0,
    total_budget_cents BIGINT NOT NULL DEFAULT 0,
    currency TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (ends_at > starts_at)
);

CREATE TABLE IF NOT EXISTS sponsored_placements (
    id TEXT PRIMARY KEY,
    campaign_id TEXT NOT NULL REFERENCES sponsored_campaigns (id) ON DELETE CASCADE,
    slot TEXT NOT NULL,
    priority INTEGER NOT NULL DEFAULT 0,
    legacy_restaurant_id TEXT REFERENCES restaurants (id) ON DELETE CASCADE,
    place_id TEXT REFERENCES places (id) ON DELETE CASCADE,
    brand_id TEXT REFERENCES brands (id) ON DELETE CASCADE,
    promotion_id TEXT REFERENCES promotions (id) ON DELETE SET NULL,
    category_id TEXT REFERENCES categories (id) ON DELETE SET NULL,
    cuisine_id TEXT REFERENCES cuisines (id) ON DELETE SET NULL,
    headline TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL DEFAULT '',
    image_url TEXT NOT NULL DEFAULT '',
    call_to_action TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (legacy_restaurant_id IS NOT NULL OR place_id IS NOT NULL OR brand_id IS NOT NULL)
);

CREATE TABLE IF NOT EXISTS sponsored_events (
    id TEXT PRIMARY KEY,
    placement_id TEXT NOT NULL REFERENCES sponsored_placements (id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    user_id TEXT REFERENCES users (id) ON DELETE SET NULL,
    surface TEXT NOT NULL DEFAULT '',
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_sponsored_campaigns_live ON sponsored_campaigns (status, starts_at, ends_at);
CREATE INDEX IF NOT EXISTS idx_sponsored_placements_slot ON sponsored_placements (slot, campaign_id);
CREATE INDEX IF NOT EXISTS idx_sponsored_events_placement_time ON sponsored_events (placement_id, occurred_at);

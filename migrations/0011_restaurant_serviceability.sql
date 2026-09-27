-- Bridge legacy storefront paths to target source stores so coverage can be
-- evaluated by provider and source-store service areas.
CREATE TABLE IF NOT EXISTS restaurant_source_store_paths (
    restaurant_id TEXT NOT NULL REFERENCES restaurants (id) ON DELETE CASCADE,
    provider_id TEXT NOT NULL REFERENCES providers (id) ON DELETE CASCADE,
    source_store_id TEXT NOT NULL REFERENCES source_stores (id) ON DELETE CASCADE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (restaurant_id, provider_id)
);

CREATE INDEX IF NOT EXISTS idx_restaurant_source_store_paths_source_store
    ON restaurant_source_store_paths (source_store_id);


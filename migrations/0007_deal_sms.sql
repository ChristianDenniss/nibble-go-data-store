CREATE TABLE IF NOT EXISTS deal_sms_sources (
    phone_number TEXT PRIMARY KEY,
    provider_id TEXT NOT NULL,
    label TEXT NOT NULL DEFAULT '',
    expiry_policy TEXT NOT NULL DEFAULT 'parsed',
    fixed_expiry_hours INTEGER NOT NULL DEFAULT 24,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE deal_sms_sources ADD COLUMN IF NOT EXISTS expiry_policy TEXT NOT NULL DEFAULT 'parsed';
ALTER TABLE deal_sms_sources ADD COLUMN IF NOT EXISTS fixed_expiry_hours INTEGER NOT NULL DEFAULT 24;

CREATE TABLE IF NOT EXISTS deal_sms_messages (
    id TEXT PRIMARY KEY,
    external_id TEXT NOT NULL UNIQUE,
    from_number TEXT NOT NULL,
    to_number TEXT NOT NULL,
    provider_id TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL,
    received_at TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL DEFAULT 'received',
    promotion_id TEXT NOT NULL DEFAULT '',
    parse_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_deal_sms_messages_received ON deal_sms_messages (received_at DESC);

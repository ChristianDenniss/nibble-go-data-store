CREATE TABLE IF NOT EXISTS gmail_connections (
    account_id TEXT PRIMARY KEY REFERENCES accounts (id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    last_sync_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS gmail_oauth_states (
    state TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS deal_email_messages (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    external_id TEXT NOT NULL,
    sender TEXT NOT NULL DEFAULT '',
    subject TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL DEFAULT '',
    received_at TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL DEFAULT 'received',
    promotion_id TEXT NOT NULL DEFAULT '',
    parse_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (account_id, external_id)
);

CREATE INDEX IF NOT EXISTS idx_deal_email_messages_received ON deal_email_messages (account_id, received_at DESC);

-- Direct ordering paths on legacy storefront restaurants. Empty string means "not offered".
-- phone: call-in number. app_url: the restaurant's own ordering app (store listing or deep link).

ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS phone TEXT NOT NULL DEFAULT '';
ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS app_url TEXT NOT NULL DEFAULT '';

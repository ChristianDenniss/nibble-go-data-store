-- Menu item images as external URLs (provider CDN or hand-picked). Empty string means "no image".
-- Moving to a Nibble-owned bucket later only changes the URLs stored here, not the schema.

ALTER TABLE menu_items ADD COLUMN IF NOT EXISTS image_url TEXT NOT NULL DEFAULT '';
ALTER TABLE source_items ADD COLUMN IF NOT EXISTS image_url TEXT NOT NULL DEFAULT '';

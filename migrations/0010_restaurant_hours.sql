-- Weekly schedules for legacy storefront restaurants. `store` is when the doors are open,
-- `delivery` is when delivery orders are taken; the two can differ.
-- opens/closes are local HH:MM; closes <= opens means the interval runs past midnight.

CREATE TABLE IF NOT EXISTS restaurant_hours (
    restaurant_id TEXT NOT NULL REFERENCES restaurants (id) ON DELETE CASCADE,
    service TEXT NOT NULL CHECK (service IN ('store', 'delivery')),
    day_of_week SMALLINT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    opens TEXT NOT NULL CHECK (opens ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$'),
    closes TEXT NOT NULL CHECK (closes ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$'),
    PRIMARY KEY (restaurant_id, service, day_of_week, opens)
);

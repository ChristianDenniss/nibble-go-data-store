-- Keep the browse taxonomy available independently of the restaurants and
-- menu items currently present in a deployment. This lets the storefront
-- show useful category/cuisine entry points even when they have no results yet.

INSERT INTO categories (id, slug, name, description) VALUES
    ('cat_food', 'food', 'Food', 'Restaurants near you'),
    ('cat_grocery', 'grocery', 'Grocery', 'Same-day grocery'),
    ('cat_convenience', 'convenience', 'Convenience', 'Snacks and essentials'),
    ('cat_alcohol', 'alcohol', 'Alcohol', 'Beer, wine, and more'),
    ('cat_pickup', 'pickup', 'Pickup', 'Skip the delivery fee'),
    ('cat_retail', 'retail', 'Retail', 'Stores and extras'),
    ('cat_pets', 'pets', 'Pets', 'Food and supplies'),
    ('cat_pharmacy', 'pharmacy', 'Pharmacy', 'Health and wellness'),
    ('cat_flowers', 'flowers', 'Flowers', 'Bouquets and plants'),
    ('cat_baby', 'baby', 'Baby', 'Diapers, formula, and more'),
    ('cat_beauty', 'beauty', 'Beauty', 'Skincare and cosmetics'),
    ('cat_bakery', 'bakery', 'Bakery', 'Fresh bread and pastries'),
    ('cat_gifts', 'gifts', 'Gifts', 'Last-minute presents')
ON CONFLICT DO NOTHING;

INSERT INTO cuisines (id, slug, name) VALUES
    ('cui_sushi', 'sushi', 'Sushi'),
    ('cui_pizza', 'pizza', 'Pizza'),
    ('cui_burgers', 'burgers', 'Burgers'),
    ('cui_mexican', 'mexican', 'Mexican'),
    ('cui_indian', 'indian', 'Indian'),
    ('cui_coffee', 'coffee', 'Coffee'),
    ('cui_healthy', 'healthy', 'Healthy'),
    ('cui_dessert', 'dessert', 'Dessert'),
    ('cui_chinese', 'chinese', 'Chinese'),
    ('cui_thai', 'thai', 'Thai'),
    ('cui_seafood', 'seafood', 'Seafood'),
    ('cui_sandwiches', 'sandwiches', 'Sandwiches'),
    ('cui_wings', 'wings', 'Wings'),
    ('cui_breakfast', 'breakfast', 'Breakfast'),
    ('cui_vegan', 'vegan', 'Vegan'),
    ('cui_donuts', 'donuts', 'Donuts')
ON CONFLICT DO NOTHING;

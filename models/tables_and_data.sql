-- Manufacturers
CREATE TABLE manufacturers (
                               id   VARCHAR(36) PRIMARY KEY,
                               name TEXT NOT NULL,
                               slug TEXT NOT NULL UNIQUE,
                               logo TEXT
);

-- Categories (supports hierarchy)
CREATE TABLE categories (
                            id         VARCHAR(36) PRIMARY KEY,
                            name       TEXT NOT NULL,
                            slug       TEXT NOT NULL UNIQUE,
                            parent_id  VARCHAR(36),
                            image      TEXT,
                            FOREIGN KEY (parent_id) REFERENCES categories(id) ON DELETE SET NULL
);

-- Products
CREATE TABLE products (
                          id              VARCHAR(36) PRIMARY KEY,
                          name            TEXT NOT NULL,
                          slug            TEXT NOT NULL UNIQUE,
                          manufacturer_id VARCHAR(36) NOT NULL,
                          category_id     VARCHAR(36) NOT NULL,
                          price           INTEGER NOT NULL,
                          old_price       INTEGER,
                          description     TEXT NOT NULL DEFAULT '',
                          features        JSON NOT NULL DEFAULT '[]',
                          image           JSON NOT NULL DEFAULT '[]',
                          stock           INTEGER NOT NULL DEFAULT 0,
                          rating          REAL NOT NULL DEFAULT 0,
                          reviews_count   INTEGER NOT NULL DEFAULT 0,
                          sku             TEXT NOT NULL UNIQUE,
                          availability    TEXT NOT NULL DEFAULT 'in_stock',
                          FOREIGN KEY (manufacturer_id) REFERENCES manufacturers(id) ON DELETE CASCADE,
                          FOREIGN KEY (category_id)     REFERENCES categories(id)     ON DELETE CASCADE
);

-- Sample data
INSERT INTO manufacturers (id, name, slug, logo) VALUES
                                                     ('m1', 'Apple',      'apple',      'https://example.com/apple.png'),
                                                     ('m2', 'Samsung',    'samsung',    'https://example.com/samsung.png'),
                                                     ('m3', 'Xiaomi',     'xiaomi',     'https://example.com/xiaomi.png');

INSERT INTO categories (id, name, slug, parent_id, image) VALUES
                                                              ('c1', 'Электроника',        'elektronika', NULL,     'https://example.com/cat_electronics.jpg'),
                                                              ('c2', 'Смартфоны',          'smartfony',   'c1',     'https://example.com/cat_smartphones.jpg'),
                                                              ('c3', 'Ноутбуки',           'noutbuki',    'c1',     'https://example.com/cat_laptops.jpg'),
                                                              ('c4', 'Аксессуары',         'aksessuary',  'c1',     NULL);

INSERT INTO products (id, name, slug, manufacturer_id, category_id, price, old_price, description, features, image, stock, rating, reviews_count, sku, availability) VALUES
                                                                                                                                                                         ('p1', 'iPhone 15 Pro', 'iphone-15-pro', 'm1', 'c2', 129900, 139900,
                                                                                                                                                                          'Новый флагман Apple', '["Titan","A17 Pro","USB-C"]', '["https://example.com/iphone1.jpg","https://example.com/iphone2.jpg"]', 15, 4.8, 124, 'IPH15PRO256', 'in_stock'),

                                                                                                                                                                         ('p2', 'Galaxy S24 Ultra', 'galaxy-s24-ultra', 'm2', 'c2', 139900, NULL,
                                                                                                                                                                          'Флагман Samsung 2024', '["S-Pen","200MP","Titanium"]', '["https://example.com/s24_1.jpg"]', 8, 4.9, 89, 'SMS24U512', 'in_stock'),

                                                                                                                                                                         ('p3', 'MacBook Air M3', 'macbook-air-m3', 'm1', 'c3', 149900, 169900,
                                                                                                                                                                          'Лёгкий и мощный ноутбук', '["M3 chip","13.6\" Liquid Retina","24h battery"]', '["https://example.com/mba_m3.jpg"]', 5, 5.0, 42, 'MBA13M3', 'in_stock'),

                                                                                                                                                                         ('p4', 'Redmi Note 13 Pro', 'redmi-note-13-pro', 'm3', 'c2', 34900, NULL,
                                                                                                                                                                          'Лучший в среднем сегменте', '["120Hz AMOLED","108MP","67W зарядка"]', '["https://example.com/redmi13.jpg"]', 32, 4.6, 201, 'RN13P8256', 'in_stock');
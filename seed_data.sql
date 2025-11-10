-- PostgreSQL Seed Data Script
-- This script creates a sample database with tables, data, and views for testing dessertfrog

-- Drop database if exists and create new one
DROP DATABASE IF EXISTS dessertfrog_demo;
CREATE DATABASE dessertfrog_demo;

-- Connect to the new database
\c dessertfrog_demo

-- Create schema
CREATE SCHEMA IF NOT EXISTS demo;

-- Set search path
SET search_path TO demo, public;

-- ============================================================================
-- TABLES
-- ============================================================================

-- Users table
CREATE TABLE demo.users (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    full_name VARCHAR(100),
    date_of_birth DATE,
    is_active BOOLEAN DEFAULT true,
    preferences JSONB,
    profile_data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE demo.users IS 'User accounts and profile information';
COMMENT ON COLUMN demo.users.user_id IS 'Unique user identifier';
COMMENT ON COLUMN demo.users.username IS 'Unique username for login';
COMMENT ON COLUMN demo.users.preferences IS 'User preferences and settings in JSON format';
COMMENT ON COLUMN demo.users.profile_data IS 'Additional user profile data in JSON format';

-- Products table
CREATE TABLE demo.products (
    product_id SERIAL PRIMARY KEY,
    product_name VARCHAR(200) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    price NUMERIC(10, 2) NOT NULL,
    stock_quantity INTEGER DEFAULT 0,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE demo.products IS 'Product catalog with pricing and inventory';

-- Orders table
CREATE TABLE demo.orders (
    order_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES demo.users(user_id),
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    total_amount NUMERIC(12, 2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    shipping_address TEXT,
    notes TEXT,
    shipping_details JSONB,
    payment_info JSONB
);

COMMENT ON TABLE demo.orders IS 'Customer orders';
COMMENT ON COLUMN demo.orders.shipping_details IS 'Shipping method, tracking info, and carrier details in JSON';
COMMENT ON COLUMN demo.orders.payment_info IS 'Payment method and transaction details in JSON';

-- Order items table
CREATE TABLE demo.order_items (
    order_item_id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES demo.orders(order_id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES demo.products(product_id),
    quantity INTEGER NOT NULL,
    unit_price NUMERIC(10, 2) NOT NULL,
    subtotal NUMERIC(12, 2) NOT NULL
);

COMMENT ON TABLE demo.order_items IS 'Line items for each order';

-- Categories table
CREATE TABLE demo.categories (
    category_id SERIAL PRIMARY KEY,
    category_name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    parent_category_id INTEGER REFERENCES demo.categories(category_id)
);

COMMENT ON TABLE demo.categories IS 'Product categories with hierarchical structure';

-- Reviews table
CREATE TABLE demo.reviews (
    review_id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES demo.products(product_id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES demo.users(user_id),
    rating INTEGER CHECK (rating BETWEEN 1 AND 5),
    review_text TEXT,
    review_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    review_metadata JSONB
);

COMMENT ON TABLE demo.reviews IS 'Product reviews and ratings';
COMMENT ON COLUMN demo.reviews.review_metadata IS 'Additional review data like images, helpfulness votes, verified purchase status';

-- ============================================================================
-- SEED DATA
-- ============================================================================

-- Insert users with JSON preferences and profile data
INSERT INTO demo.users (username, email, full_name, date_of_birth, is_active, preferences, profile_data) VALUES
    ('john_doe', 'john.doe@example.com', 'John Doe', '1990-05-15', true,
     '{"theme": "dark", "notifications": {"email": true, "sms": false, "push": true}, "language": "en", "currency": "USD"}'::jsonb,
     '{"avatar": "https://example.com/avatars/john.jpg", "bio": "Tech enthusiast and coffee lover", "social": {"twitter": "@johndoe", "linkedin": "johndoe"}, "interests": ["technology", "gaming", "coffee"]}'::jsonb),

    ('jane_smith', 'jane.smith@example.com', 'Jane Smith', '1985-08-22', true,
     '{"theme": "light", "notifications": {"email": true, "sms": true, "push": false}, "language": "en", "currency": "USD"}'::jsonb,
     '{"avatar": "https://example.com/avatars/jane.jpg", "bio": "Marketing professional and bookworm", "social": {"twitter": "@janesmith", "instagram": "janesmith"}, "interests": ["reading", "travel", "photography"]}'::jsonb),

    ('bob_wilson', 'bob.wilson@example.com', 'Bob Wilson', '1992-11-30', true,
     '{"theme": "dark", "notifications": {"email": false, "sms": false, "push": true}, "language": "en", "currency": "EUR"}'::jsonb,
     '{"avatar": "https://example.com/avatars/bob.jpg", "bio": "Fitness trainer and outdoor adventurer", "social": {"instagram": "bobwilson_fit"}, "interests": ["fitness", "hiking", "nutrition"]}'::jsonb),

    ('alice_brown', 'alice.brown@example.com', 'Alice Brown', '1988-03-10', true,
     '{"theme": "light", "notifications": {"email": true, "sms": false, "push": true}, "language": "en", "currency": "USD"}'::jsonb,
     '{"avatar": "https://example.com/avatars/alice.jpg", "bio": "Software engineer and open source contributor", "social": {"github": "alicebrown", "twitter": "@alice_codes"}, "interests": ["programming", "AI", "music"]}'::jsonb),

    ('charlie_davis', 'charlie.davis@example.com', 'Charlie Davis', '1995-07-25', false,
     '{"theme": "auto", "notifications": {"email": false, "sms": false, "push": false}, "language": "en", "currency": "GBP"}'::jsonb,
     '{"avatar": "https://example.com/avatars/charlie.jpg", "bio": "Student and aspiring entrepreneur", "social": {}, "interests": ["business", "startups", "podcasts"]}'::jsonb),

    ('diana_martinez', 'diana.martinez@example.com', 'Diana Martinez', '1991-12-05', true,
     '{"theme": "dark", "notifications": {"email": true, "sms": true, "push": true}, "language": "es", "currency": "EUR"}'::jsonb,
     '{"avatar": "https://example.com/avatars/diana.jpg", "bio": "Graphic designer and artist", "social": {"instagram": "diana_designs", "behance": "dianamart"}, "interests": ["design", "art", "fashion"]}'::jsonb),

    ('frank_garcia', 'frank.garcia@example.com', 'Frank Garcia', '1987-09-18', true,
     '{"theme": "light", "notifications": {"email": true, "sms": false, "push": false}, "language": "en", "currency": "USD"}'::jsonb,
     '{"avatar": "https://example.com/avatars/frank.jpg", "bio": "Data scientist and chess player", "social": {"linkedin": "frankgarcia"}, "interests": ["data science", "chess", "mathematics"]}'::jsonb),

    ('grace_lee', 'grace.lee@example.com', 'Grace Lee', '1993-04-12', true,
     '{"theme": "dark", "notifications": {"email": true, "sms": false, "push": true}, "language": "ko", "currency": "USD"}'::jsonb,
     '{"avatar": "https://example.com/avatars/grace.jpg", "bio": "Product manager and yoga instructor", "social": {"linkedin": "gracelee", "instagram": "grace_yoga"}, "interests": ["product management", "yoga", "meditation"]}'::jsonb),

    ('henry_white', 'henry.white@example.com', 'Henry White', '1989-06-28', true,
     '{"theme": "light", "notifications": {"email": false, "sms": true, "push": true}, "language": "en", "currency": "CAD"}'::jsonb,
     '{"avatar": "https://example.com/avatars/henry.jpg", "bio": "Chef and food blogger", "social": {"instagram": "henry_cooks", "youtube": "henrywhite"}, "interests": ["cooking", "food photography", "travel"]}'::jsonb),

    ('ivy_chen', 'ivy.chen@example.com', 'Ivy Chen', '1994-01-15', true,
     '{"theme": "auto", "notifications": {"email": true, "sms": true, "push": true}, "language": "zh", "currency": "USD"}'::jsonb,
     '{"avatar": "https://example.com/avatars/ivy.jpg", "bio": "UX designer and tea enthusiast", "social": {"dribbble": "ivychen", "twitter": "@ivy_ux"}, "interests": ["UX design", "tea", "calligraphy"]}'::jsonb);

-- Insert categories
INSERT INTO demo.categories (category_name, description, parent_category_id) VALUES
    ('Electronics', 'Electronic devices and gadgets', NULL),
    ('Computers', 'Desktop and laptop computers', 1),
    ('Smartphones', 'Mobile phones and accessories', 1),
    ('Clothing', 'Apparel and fashion items', NULL),
    ('Mens Wear', 'Clothing for men', 4),
    ('Womens Wear', 'Clothing for women', 4),
    ('Books', 'Physical and digital books', NULL),
    ('Fiction', 'Fiction novels and stories', 7),
    ('Non-Fiction', 'Educational and informational books', 7),
    ('Home & Garden', 'Home improvement and gardening', NULL);

-- Insert products with JSON metadata
INSERT INTO demo.products (product_name, description, category, price, stock_quantity, metadata) VALUES
    ('Laptop Pro 15"', 'High-performance laptop with 16GB RAM', 'Computers', 1299.99, 45, '{"brand": "TechCorp", "warranty": "2 years", "specs": {"cpu": "Intel i7", "ram": "16GB", "storage": "512GB SSD"}}'),
    ('Smartphone X', 'Latest smartphone with 5G support', 'Smartphones', 899.99, 120, '{"brand": "PhoneCo", "warranty": "1 year", "specs": {"screen": "6.5 inch", "camera": "48MP", "battery": "4500mAh"}}'),
    ('Wireless Headphones', 'Noise-canceling bluetooth headphones', 'Electronics', 249.99, 78, '{"brand": "AudioMax", "warranty": "1 year", "features": ["noise-canceling", "wireless", "40-hour battery"]}'),
    ('Running Shoes', 'Comfortable athletic shoes for running', 'Mens Wear', 89.99, 200, '{"brand": "SportFit", "sizes": ["8", "9", "10", "11", "12"], "colors": ["black", "blue", "red"]}'),
    ('Cotton T-Shirt', 'Soft cotton t-shirt in various colors', 'Mens Wear', 19.99, 350, '{"brand": "Casual Wear", "material": "100% cotton", "sizes": ["S", "M", "L", "XL"]}'),
    ('Summer Dress', 'Lightweight floral summer dress', 'Womens Wear', 49.99, 85, '{"brand": "Fashion Forward", "material": "cotton blend", "sizes": ["XS", "S", "M", "L"]}'),
    ('The Great Novel', 'Bestselling fiction book', 'Fiction', 24.99, 150, '{"author": "Famous Writer", "pages": 450, "isbn": "978-1234567890", "format": "hardcover"}'),
    ('Learn Python', 'Comprehensive Python programming guide', 'Non-Fiction', 39.99, 95, '{"author": "Tech Expert", "pages": 680, "isbn": "978-0987654321", "format": "paperback"}'),
    ('Coffee Maker', 'Programmable 12-cup coffee maker', 'Home & Garden', 79.99, 55, '{"brand": "BrewMaster", "warranty": "1 year", "features": ["programmable", "auto-shutoff", "thermal carafe"]}'),
    ('Garden Tools Set', 'Complete set of essential garden tools', 'Home & Garden', 129.99, 40, '{"brand": "GreenThumb", "items": ["shovel", "rake", "pruner", "gloves", "trowel"], "material": "stainless steel"}'),
    ('Office Chair', 'Ergonomic office chair with lumbar support', 'Home & Garden', 299.99, 30, '{"brand": "ComfortSeat", "warranty": "3 years", "features": ["adjustable height", "lumbar support", "breathable mesh"]}'),
    ('Gaming Mouse', 'RGB gaming mouse with programmable buttons', 'Computers', 69.99, 110, '{"brand": "GameGear", "warranty": "2 years", "specs": {"dpi": "16000", "buttons": "8", "rgb": true}}'),
    ('Yoga Mat', 'Non-slip yoga mat with carrying strap', 'Clothing', 34.99, 180, '{"brand": "FitLife", "material": "TPE", "dimensions": "72x24 inches", "colors": ["purple", "blue", "pink"]}'),
    ('Bluetooth Speaker', 'Portable waterproof bluetooth speaker', 'Electronics', 59.99, 95, '{"brand": "SoundWave", "warranty": "1 year", "features": ["waterproof", "12-hour battery", "360-degree sound"]}'),
    ('Desk Lamp', 'LED desk lamp with adjustable brightness', 'Home & Garden', 44.99, 65, '{"brand": "BrightLight", "warranty": "2 years", "features": ["dimmable", "USB port", "touch control"]}}');

-- Insert orders with JSON shipping and payment details
INSERT INTO demo.orders (user_id, order_date, total_amount, status, shipping_address, notes, shipping_details, payment_info) VALUES
    (1, '2024-01-15 10:30:00', 1349.98, 'delivered', '123 Main St, New York, NY 10001', 'Leave at front door',
     '{"method": "express", "carrier": "FedEx", "tracking": "FX123456789", "estimated_delivery": "2024-01-18", "signature_required": false}'::jsonb,
     '{"method": "credit_card", "last4": "4242", "brand": "Visa", "transaction_id": "txn_abc123"}'::jsonb),

    (2, '2024-01-20 14:15:00', 949.98, 'delivered', '456 Oak Ave, Los Angeles, CA 90001', NULL,
     '{"method": "standard", "carrier": "UPS", "tracking": "1Z999AA10123456784", "estimated_delivery": "2024-01-25", "signature_required": false}'::jsonb,
     '{"method": "paypal", "email": "jane.smith@example.com", "transaction_id": "PP-1234567890"}'::jsonb),

    (3, '2024-02-05 09:45:00', 169.97, 'shipped', '789 Pine Rd, Chicago, IL 60601', 'Call before delivery',
     '{"method": "express", "carrier": "DHL", "tracking": "DHL987654321", "estimated_delivery": "2024-02-08", "signature_required": true}'::jsonb,
     '{"method": "debit_card", "last4": "5678", "brand": "MasterCard", "transaction_id": "txn_def456"}'::jsonb),

    (4, '2024-02-10 16:20:00', 154.97, 'delivered', '321 Elm St, Houston, TX 77001', NULL,
     '{"method": "standard", "carrier": "USPS", "tracking": "9400111202555555555555", "estimated_delivery": "2024-02-15", "signature_required": false}'::jsonb,
     '{"method": "credit_card", "last4": "9012", "brand": "American Express", "transaction_id": "txn_ghi789"}'::jsonb),

    (1, '2024-02-15 11:00:00', 369.96, 'processing', '123 Main St, New York, NY 10001', NULL,
     '{"method": "express", "carrier": "FedEx", "tracking": null, "estimated_delivery": "2024-02-18", "signature_required": false}'::jsonb,
     '{"method": "credit_card", "last4": "4242", "brand": "Visa", "transaction_id": "txn_jkl012"}'::jsonb),

    (5, '2024-02-20 13:30:00', 89.99, 'pending', '654 Maple Dr, Phoenix, AZ 85001', NULL,
     '{"method": "standard", "carrier": null, "tracking": null, "estimated_delivery": null, "signature_required": false}'::jsonb,
     '{"method": "bank_transfer", "bank": "Chase", "transaction_id": "txn_mno345"}'::jsonb),

    (6, '2024-03-01 10:15:00', 229.97, 'shipped', '987 Cedar Ln, Philadelphia, PA 19101', 'Gift wrap requested',
     '{"method": "express", "carrier": "UPS", "tracking": "1Z999AA10987654321", "estimated_delivery": "2024-03-04", "signature_required": true, "gift_wrap": true}'::jsonb,
     '{"method": "paypal", "email": "diana.martinez@example.com", "transaction_id": "PP-9876543210"}'::jsonb),

    (7, '2024-03-05 15:45:00', 1529.96, 'delivered', '147 Birch Way, San Antonio, TX 78201', NULL,
     '{"method": "overnight", "carrier": "FedEx", "tracking": "FX987654321", "estimated_delivery": "2024-03-06", "signature_required": true}'::jsonb,
     '{"method": "credit_card", "last4": "3456", "brand": "Visa", "transaction_id": "txn_pqr678"}'::jsonb),

    (8, '2024-03-10 12:20:00', 104.98, 'delivered', '258 Spruce Ct, San Diego, CA 92101', NULL,
     '{"method": "standard", "carrier": "USPS", "tracking": "9400111202544444444444", "estimated_delivery": "2024-03-15", "signature_required": false}'::jsonb,
     '{"method": "apple_pay", "device": "iPhone 13", "transaction_id": "txn_stu901"}'::jsonb),

    (9, '2024-03-15 14:00:00', 299.99, 'processing', '369 Willow St, Dallas, TX 75201', 'Urgent delivery',
     '{"method": "overnight", "carrier": "FedEx", "tracking": null, "estimated_delivery": "2024-03-16", "signature_required": true}'::jsonb,
     '{"method": "google_pay", "email": "henry.white@example.com", "transaction_id": "txn_vwx234"}'::jsonb);

-- Insert order items
INSERT INTO demo.order_items (order_id, product_id, quantity, unit_price, subtotal) VALUES
    -- Order 1
    (1, 1, 1, 1299.99, 1299.99),
    (1, 12, 1, 69.99, 69.99),
    -- Order 2
    (2, 2, 1, 899.99, 899.99),
    (2, 3, 1, 249.99, 249.99),
    -- Order 3
    (3, 4, 1, 89.99, 89.99),
    (3, 5, 2, 19.99, 39.98),
    (3, 13, 1, 34.99, 34.99),
    -- Order 4
    (4, 6, 2, 49.99, 99.98),
    (4, 7, 1, 24.99, 24.99),
    (4, 8, 1, 39.99, 39.99),
    -- Order 5
    (5, 9, 1, 79.99, 79.99),
    (5, 10, 1, 129.99, 129.99),
    (5, 15, 2, 44.99, 89.98),
    -- Order 6
    (6, 4, 1, 89.99, 89.99),
    -- Order 7
    (7, 3, 1, 249.99, 249.99),
    (7, 14, 3, 59.99, 179.97),
    -- Order 8
    (8, 11, 1, 299.99, 299.99),
    (8, 1, 1, 1299.99, 1299.99),
    -- Order 9
    (9, 5, 3, 19.99, 59.97),
    (9, 13, 1, 34.99, 34.99),
    -- Order 10
    (10, 11, 1, 299.99, 299.99);

-- Insert reviews with JSON metadata
INSERT INTO demo.reviews (product_id, user_id, rating, review_text, review_date, review_metadata) VALUES
    (1, 2, 5, 'Excellent laptop! Fast performance and great build quality.', '2024-01-25 10:00:00',
     '{"verified_purchase": true, "helpful_votes": 45, "photos": ["https://example.com/reviews/1-1.jpg", "https://example.com/reviews/1-2.jpg"], "pros": ["fast", "sturdy", "great screen"], "cons": ["expensive"]}'::jsonb),

    (1, 4, 4, 'Good laptop but a bit pricey. Battery life could be better.', '2024-02-15 14:30:00',
     '{"verified_purchase": true, "helpful_votes": 23, "photos": [], "pros": ["performance", "design"], "cons": ["price", "battery life"]}'::jsonb),

    (2, 3, 5, 'Best smartphone I have ever owned. Camera is amazing!', '2024-02-08 16:45:00',
     '{"verified_purchase": true, "helpful_votes": 67, "photos": ["https://example.com/reviews/2-1.jpg"], "pros": ["camera", "display", "5G"], "cons": []}'::jsonb),

    (3, 1, 5, 'Noise canceling is superb. Very comfortable for long use.', '2024-01-28 09:15:00',
     '{"verified_purchase": true, "helpful_votes": 89, "photos": [], "pros": ["noise canceling", "comfort", "battery life"], "cons": []}'::jsonb),

    (3, 8, 4, 'Great sound quality, but a bit heavy for my taste.', '2024-03-12 11:20:00',
     '{"verified_purchase": false, "helpful_votes": 12, "photos": [], "pros": ["sound quality"], "cons": ["weight"]}'::jsonb),

    (4, 3, 5, 'Perfect running shoes. Very comfortable and durable.', '2024-02-18 13:00:00',
     '{"verified_purchase": true, "helpful_votes": 34, "photos": ["https://example.com/reviews/4-1.jpg"], "pros": ["comfort", "durability", "style"], "cons": []}'::jsonb),

    (5, 1, 4, 'Nice t-shirt, good quality cotton. Runs a bit small.', '2024-02-20 15:30:00',
     '{"verified_purchase": true, "helpful_votes": 18, "photos": [], "pros": ["quality", "soft"], "cons": ["sizing"]}'::jsonb),

    (6, 4, 5, 'Beautiful dress, fits perfectly. Love the floral pattern!', '2024-02-25 10:45:00',
     '{"verified_purchase": true, "helpful_votes": 56, "photos": ["https://example.com/reviews/6-1.jpg", "https://example.com/reviews/6-2.jpg"], "pros": ["design", "fit", "material"], "cons": []}'::jsonb),

    (7, 2, 5, 'Couldn not put it down! Great story and characters.', '2024-02-12 20:00:00',
     '{"verified_purchase": true, "helpful_votes": 102, "photos": [], "pros": ["plot", "characters", "writing"], "cons": []}'::jsonb),

    (8, 6, 5, 'Best Python book for beginners. Very well explained.', '2024-03-05 14:00:00',
     '{"verified_purchase": true, "helpful_votes": 78, "photos": [], "pros": ["clear explanations", "examples", "exercises"], "cons": []}'::jsonb),

    (9, 5, 4, 'Makes good coffee. Programming feature is convenient.', '2024-02-28 08:30:00',
     '{"verified_purchase": true, "helpful_votes": 41, "photos": ["https://example.com/reviews/9-1.jpg"], "pros": ["programmable", "quality"], "cons": ["noisy"]}'::jsonb),

    (10, 7, 5, 'High quality tools. Everything I need for gardening.', '2024-03-08 16:00:00',
     '{"verified_purchase": true, "helpful_votes": 29, "photos": ["https://example.com/reviews/10-1.jpg", "https://example.com/reviews/10-2.jpg"], "pros": ["quality", "complete set", "durable"], "cons": []}'::jsonb),

    (11, 9, 5, 'Most comfortable office chair I have used. Worth the price!', '2024-03-18 12:00:00',
     '{"verified_purchase": true, "helpful_votes": 94, "photos": ["https://example.com/reviews/11-1.jpg"], "pros": ["comfort", "lumbar support", "adjustable"], "cons": ["assembly required"]}'::jsonb),

    (12, 1, 4, 'Good gaming mouse with nice RGB lighting. Buttons are responsive.', '2024-02-01 19:30:00',
     '{"verified_purchase": true, "helpful_votes": 37, "photos": [], "pros": ["responsive", "RGB", "ergonomic"], "cons": ["software"]}'::jsonb),

    (13, 3, 5, 'Perfect yoga mat. Non-slip surface works great.', '2024-02-22 07:00:00',
     '{"verified_purchase": true, "helpful_votes": 51, "photos": ["https://example.com/reviews/13-1.jpg"], "pros": ["non-slip", "thickness", "portable"], "cons": []}'::jsonb);

-- ============================================================================
-- VIEWS
-- ============================================================================

-- View: Active users with their order count
CREATE VIEW demo.active_users_summary AS
SELECT
    u.user_id,
    u.username,
    u.email,
    u.full_name,
    u.created_at,
    COUNT(o.order_id) AS total_orders,
    COALESCE(SUM(o.total_amount), 0) AS lifetime_value
FROM demo.users u
LEFT JOIN demo.orders o ON u.user_id = o.user_id
WHERE u.is_active = true
GROUP BY u.user_id, u.username, u.email, u.full_name, u.created_at
ORDER BY lifetime_value DESC;

COMMENT ON VIEW demo.active_users_summary IS 'Summary of active users with their order statistics';

-- View: Product inventory status
CREATE VIEW demo.product_inventory_status AS
SELECT
    p.product_id,
    p.product_name,
    p.category,
    p.price,
    p.stock_quantity,
    CASE
        WHEN p.stock_quantity = 0 THEN 'Out of Stock'
        WHEN p.stock_quantity < 20 THEN 'Low Stock'
        WHEN p.stock_quantity < 50 THEN 'Medium Stock'
        ELSE 'In Stock'
    END AS stock_status,
    COALESCE(AVG(r.rating), 0) AS average_rating,
    COUNT(r.review_id) AS review_count
FROM demo.products p
LEFT JOIN demo.reviews r ON p.product_id = r.product_id
GROUP BY p.product_id, p.product_name, p.category, p.price, p.stock_quantity
ORDER BY p.product_name;

COMMENT ON VIEW demo.product_inventory_status IS 'Product inventory levels with stock status and ratings';

-- View: Order details with customer information
CREATE VIEW demo.order_details_full AS
SELECT
    o.order_id,
    o.order_date,
    o.status,
    o.total_amount,
    u.user_id,
    u.username,
    u.email,
    u.full_name,
    o.shipping_address,
    COUNT(oi.order_item_id) AS item_count
FROM demo.orders o
JOIN demo.users u ON o.user_id = u.user_id
LEFT JOIN demo.order_items oi ON o.order_id = oi.order_id
GROUP BY o.order_id, o.order_date, o.status, o.total_amount,
         u.user_id, u.username, u.email, u.full_name, o.shipping_address
ORDER BY o.order_date DESC;

COMMENT ON VIEW demo.order_details_full IS 'Complete order information with customer details';

-- View: Sales by category
CREATE VIEW demo.sales_by_category AS
SELECT
    p.category,
    COUNT(DISTINCT o.order_id) AS total_orders,
    SUM(oi.quantity) AS total_units_sold,
    SUM(oi.subtotal) AS total_revenue,
    AVG(oi.unit_price) AS average_price
FROM demo.order_items oi
JOIN demo.products p ON oi.product_id = p.product_id
JOIN demo.orders o ON oi.order_id = o.order_id
GROUP BY p.category
ORDER BY total_revenue DESC;

COMMENT ON VIEW demo.sales_by_category IS 'Sales statistics grouped by product category';

-- View: Top selling products
CREATE VIEW demo.top_selling_products AS
SELECT
    p.product_id,
    p.product_name,
    p.category,
    p.price,
    SUM(oi.quantity) AS total_sold,
    SUM(oi.subtotal) AS total_revenue,
    COUNT(DISTINCT oi.order_id) AS order_count,
    COALESCE(AVG(r.rating), 0) AS average_rating
FROM demo.products p
LEFT JOIN demo.order_items oi ON p.product_id = oi.product_id
LEFT JOIN demo.reviews r ON p.product_id = r.product_id
GROUP BY p.product_id, p.product_name, p.category, p.price
HAVING SUM(oi.quantity) IS NOT NULL
ORDER BY total_sold DESC
LIMIT 10;

COMMENT ON VIEW demo.top_selling_products IS 'Top 10 best-selling products by quantity';

-- View: Recent reviews with product and user info
CREATE VIEW demo.recent_reviews AS
SELECT
    r.review_id,
    r.review_date,
    r.rating,
    r.review_text,
    u.username,
    u.full_name,
    p.product_name,
    p.category
FROM demo.reviews r
JOIN demo.users u ON r.user_id = u.user_id
JOIN demo.products p ON r.product_id = p.product_id
ORDER BY r.review_date DESC
LIMIT 50;

COMMENT ON VIEW demo.recent_reviews IS 'Most recent 50 product reviews with context';

-- ============================================================================
-- INDEXES FOR PERFORMANCE
-- ============================================================================

CREATE INDEX idx_orders_user_id ON demo.orders(user_id);
CREATE INDEX idx_orders_status ON demo.orders(status);
CREATE INDEX idx_order_items_order_id ON demo.order_items(order_id);
CREATE INDEX idx_order_items_product_id ON demo.order_items(product_id);
CREATE INDEX idx_reviews_product_id ON demo.reviews(product_id);
CREATE INDEX idx_reviews_user_id ON demo.reviews(user_id);
CREATE INDEX idx_products_category ON demo.products(category);

-- GIN indexes for JSONB columns (enable fast JSON queries)
CREATE INDEX idx_products_metadata ON demo.products USING GIN(metadata);
CREATE INDEX idx_users_preferences ON demo.users USING GIN(preferences);
CREATE INDEX idx_users_profile_data ON demo.users USING GIN(profile_data);
CREATE INDEX idx_orders_shipping_details ON demo.orders USING GIN(shipping_details);
CREATE INDEX idx_orders_payment_info ON demo.orders USING GIN(payment_info);
CREATE INDEX idx_reviews_metadata ON demo.reviews USING GIN(review_metadata);

-- ============================================================================
-- SUMMARY
-- ============================================================================

-- Display summary of created objects
SELECT 'Database created: dessertfrog_demo' AS summary;
SELECT 'Schema created: demo' AS summary;
SELECT COUNT(*) || ' tables created' AS summary FROM information_schema.tables
WHERE table_schema = 'demo' AND table_type = 'BASE TABLE';
SELECT COUNT(*) || ' views created' AS summary FROM information_schema.views
WHERE table_schema = 'demo';
SELECT COUNT(*) || ' users inserted' AS summary FROM demo.users;
SELECT COUNT(*) || ' products inserted' AS summary FROM demo.products;
SELECT COUNT(*) || ' orders inserted' AS summary FROM demo.orders;
SELECT COUNT(*) || ' reviews inserted' AS summary FROM demo.reviews;

-- Display sample data
\echo '\n=== Sample Users ==='
SELECT user_id, username, email, full_name FROM demo.users LIMIT 5;

\echo '\n=== Sample Products ==='
SELECT product_id, product_name, category, price, stock_quantity FROM demo.products LIMIT 5;

\echo '\n=== Sample Orders ==='
SELECT order_id, user_id, order_date, total_amount, status FROM demo.orders LIMIT 5;

\echo '\n=== Active Users Summary View ==='
SELECT * FROM demo.active_users_summary LIMIT 5;

\echo '\n=== Product Inventory Status View ==='
SELECT * FROM demo.product_inventory_status LIMIT 5;

\echo '\n\nTo connect to this database, use:'
\echo 'psql -U postgres -d dessertfrog_demo'
\echo '\nOr with dessertfrog:'
\echo './dessertfrog --driver postgres --host localhost --port 5432 --username postgres --database dessertfrog_demo --schema demo'

-- ============================================================================
-- EXAMPLE JSON QUERIES
-- ============================================================================
-- Uncomment and run these queries to test JSON functionality

-- Query users by theme preference
-- SELECT username, preferences->>'theme' as theme FROM demo.users WHERE preferences->>'theme' = 'dark';

-- Query users with email notifications enabled
-- SELECT username, email FROM demo.users WHERE preferences->'notifications'->>'email' = 'true';

-- Find users interested in technology
-- SELECT username, profile_data->'interests' as interests FROM demo.users WHERE profile_data->'interests' ? 'technology';

-- Get orders shipped by FedEx
-- SELECT order_id, user_id, shipping_details->>'carrier' as carrier, shipping_details->>'tracking' as tracking FROM demo.orders WHERE shipping_details->>'carrier' = 'FedEx';

-- Find credit card payments
-- SELECT order_id, payment_info->>'method' as payment_method, payment_info->>'brand' as card_brand FROM demo.orders WHERE payment_info->>'method' = 'credit_card';

-- Get verified purchase reviews
-- SELECT product_id, rating, review_text FROM demo.reviews WHERE review_metadata->>'verified_purchase' = 'true';

-- Find reviews with photos
-- SELECT review_id, product_id, jsonb_array_length(review_metadata->'photos') as photo_count FROM demo.reviews WHERE jsonb_array_length(review_metadata->'photos') > 0;

-- Query products with specific warranty period
-- SELECT product_name, metadata->>'warranty' as warranty FROM demo.products WHERE metadata ? 'warranty' AND metadata->>'warranty' = '2 years';

-- Complex nested JSON query: users with push notifications enabled
-- SELECT username, preferences->'notifications'->>'push' as push_enabled FROM demo.users WHERE (preferences->'notifications'->>'push')::boolean = true;

-- creation of db tables
-- products
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    category VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- inventory
CREATE TABLE IF NOT EXISTS inventory (
    product_id INTEGER PRIMARY KEY REFERENCES products(id) ON DELETE CASCADE,
    stock INTEGER NOT NULL DEFAULT 0,
    reserved INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- orders
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    total_amount DECIMAL(10, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- order_items
CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL,
    price_at_order DECIMAL(10, 2) NOT NULL
);

-- populating with sample data
-- initial products
INSERT INTO products (id, name, description, price, category) VALUES
    (1, 'Laptop', 'High-performance laptop', 999.99, 'Electronics'),
    (2, 'Mouse', 'Wireless optical mouse', 29.99, 'Electronics'),
    (3, 'Keyboard', 'Mechanical keyboard', 79.99, 'Electronics'),
    (4, 'Monitor', '24-inch LCD monitor', 199.99, 'Electronics'),
    (5, 'Desk Chair', 'Ergonomic office chair', 149.99, 'Furniture')
ON CONFLICT (id) DO NOTHING;
-- initial inventory
INSERT INTO inventory (product_id, stock, reserved) VALUES
    (1, 50, 0),
    (2, 100, 0),
    (3, 25, 0),
    (4, 30, 0),
    (5, 15, 0)
ON CONFLICT (product_id) DO NOTHING;
-- ** no initial orders

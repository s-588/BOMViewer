CREATE TABLE materials (
    material_id SERIAL PRIMARY KEY,
    unit VARCHAR(50) NOT NULL,                   
    description TEXT                           
);

CREATE TABLE material_names (
    name_id SERIAL PRIMARY KEY,
    material_id INT NOT NULL REFERENCES materials(material_id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE(material_id, name)
);

CREATE TABLE products (
    product_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    price NUMERIC(12,2) NOT NULL CHECK (price >= 0)
);

CREATE TABLE product_materials (
    product_id INT REFERENCES products(product_id) ON DELETE CASCADE,
    material_id INT REFERENCES materials(material_id) ON DELETE RESTRICT,
    quantity NUMERIC(12,3) NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (product_id, material_id)
);

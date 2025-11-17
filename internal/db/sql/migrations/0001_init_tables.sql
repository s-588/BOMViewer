-- +goose Up
CREATE TABLE unit_types (
    unit_id INTEGER PRIMARY KEY AUTOINCREMENT,
    unit VARCHAR(50) NOT NULL
);

CREATE TABLE materials (
    material_id INTEGER PRIMARY KEY AUTOINCREMENT,
    unit_id INT NOT NULL,
    description TEXT,
    FOREIGN KEY (unit_id) REFERENCES unit_types (unit_id) ON DELETE SET NULL
);

CREATE TABLE material_names (
    name_id INTEGER PRIMARY KEY AUTOINCREMENT,
    material_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    FOREIGN KEY (material_id) REFERENCES materials (material_id) ON DELETE CASCADE
);

CREATE TABLE products (
    product_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    description TEXT
);

CREATE TABLE product_materials (
    product_id INT NOT NULL,
    material_id INT NOT NULL,
    quantity NUMERIC(12,3) CHECK (quantity > 0),
    quantity_text VARCHAR(50),
    PRIMARY KEY (product_id, material_id),
    FOREIGN KEY (product_id) REFERENCES products (product_id) ON DELETE CASCADE,
    FOREIGN KEY (material_id) REFERENCES materials (material_id) ON DELETE RESTRICT
);

CREATE TABLE files (
    file_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(50) NOT NULL,
    path VARCHAR(255)
);

CREATE TABLE files_materials (
    material_id INT NOT NULL,
    file_id INT NOT NULL,
    PRIMARY KEY (material_id, file_id),
    FOREIGN KEY (material_id) REFERENCES materials (material_id) ON DELETE CASCADE,
    FOREIGN KEY (file_id) REFERENCES files (file_id) ON DELETE CASCADE
);

CREATE TABLE files_products (
    product_id INT NOT NULL,
    file_id INT NOT NULL,
    PRIMARY KEY (product_id, file_id),
    FOREIGN KEY (product_id) REFERENCES products (product_id) ON DELETE CASCADE,
    FOREIGN KEY (file_id) REFERENCES files (file_id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE files_products;
DROP TABLE files_materials;
DROP TABLE files;
DROP TABLE product_materials;
DROP TABLE products;
DROP TABLE material_names;
DROP TABLE materials;
DROP TABLE unit_types;

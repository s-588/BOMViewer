-- THIS FILE ONLY FOR SQLC

CREATE TABLE materials (
  material_id INTEGER PRIMARY KEY AUTOINCREMENT,
  unit_id INT NOT NULL REFERENCES unit_types (unit_id) ON DELETE SET NULL,
  description TEXT
);

CREATE TABLE material_names (
  name_id INTEGER PRIMARY KEY AUTOINCREMENT,
  material_id INT NOT NULL REFERENCES materials (material_id) ON DELETE CASCADE,
  name VARCHAR(255) NOT NULL UNIQUE,
  is_primary BOOLEAN NOT NULL DEFAULT FALSE,
  UNIQUE (material_id, name)
);

-- allow only one true value of is_primary per group of same material_id's
CREATE UNIQUE INDEX unique_primary_name_per_material ON material_names (material_id)
WHERE
  is_primary = 1;

CREATE TABLE products (
  product_id INTEGER PRIMARY KEY AUTOINCREMENT,
  name VARCHAR(255) NOT NULL UNIQUE,
  description TEXT
);

CREATE TABLE product_materials (
  product_id INT REFERENCES products (product_id) ON DELETE CASCADE,
  material_id INT REFERENCES materials (material_id) ON DELETE RESTRICT,
  quantity NUMERIC(12, 3) CHECK (quantity > 0),
  quantity_text VARCHAR(50),
  PRIMARY KEY (product_id, material_id)
);

CREATE TABLE unit_types (
  unit_id INTEGER PRIMARY KEY AUTOINCREMENT,
  unit VARCHAR(50) NOT NULL
);

CREATE TABLE files (
  file_id INTEGER PRIMARY KEY AUTOINCREMENT,
  name varchar(50) NOT NULL,
  PATH varchar(255)
);

CREATE TABLE files_materials (
  material_id INTEGER REFERENCES materials (material_id) ON DELETE CASCADE,
  file_id INTEGER REFERENCES files (file_id) ON DELETE CASCADE,
  PRIMARY KEY (material_id, file_id)
);

CREATE TABLE files_products (
  product_id INTEGER REFERENCES products (product_id) ON DELETE CASCADE,
  file_id INTEGER REFERENCES files (file_id) ON DELETE CASCADE,
  PRIMARY KEY (product_id, file_id)
);

-- Optimized virtual table for text search
CREATE VIRTUAL TABLE fts_all USING fts5(
    type UNINDEXED,
    ref_id UNINDEXED,
    text,
    tokenize = 'unicode61'
);
-- 0002_indexes.sql

-- Only one primary name per material
CREATE UNIQUE INDEX idx_material_primary_name
ON material_names (material_id)
WHERE is_primary = 1;

-- Next indexes need for FTS5 optimization

-- Unique name per material
CREATE UNIQUE INDEX idx_material_names_unique
ON material_names (material_id, name);

-- Unique product name
CREATE UNIQUE INDEX idx_products_unique_name
ON products (name);

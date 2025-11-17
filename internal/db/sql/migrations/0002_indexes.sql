-- +goose Up
CREATE UNIQUE INDEX idx_material_primary_name
    ON material_names (material_id)
    WHERE is_primary = 1;

CREATE UNIQUE INDEX idx_material_names_unique
    ON material_names (material_id, name);

CREATE UNIQUE INDEX idx_products_unique_name
    ON products (name);

-- +goose Down
DROP INDEX IF EXISTS idx_material_primary_name;
DROP INDEX IF EXISTS idx_material_names_unique;
DROP INDEX IF EXISTS idx_products_unique_name;

-- name: GetAllProducts :many
SELECT
    p.product_id,
    p.name,
    p.description,
    pm.material_id,
    mn.name as material_name,
    m.unit_id,
    ut.unit as unit_name,
    pm.quantity,
    pm.quantity_text
FROM
    products p
    LEFT JOIN product_materials pm ON p.product_id = pm.product_id
    LEFT JOIN materials m ON pm.material_id = m.material_id
    LEFT JOIN material_names mn ON m.material_id = mn.material_id
    AND mn.is_primary = TRUE
    LEFT JOIN unit_types ut ON m.unit_id = ut.unit_id
ORDER BY
    p.product_id;

-- name: InsertProduct :one
INSERT INTO
    products (name, description)
VALUES
    (?, ?) RETURNING *;

-- name: GetProductByID :one
SELECT
    *
FROM
    products
WHERE
    product_id = ?;

-- name: GetProductMaterials :many
SELECT
    m.*,
    ut.unit AS unit,
    pm.quantity AS quantity,
    pm.quantity_text AS quantity_text,
    mn.name AS material_name
FROM
    product_materials pm
    INNER JOIN materials m ON m.material_id = pm.material_id
    INNER JOIN unit_types ut ON ut.unit_id = m.unit_id
    INNER JOIN material_names mn ON mn.material_id = m.material_id
WHERE
    pm.product_id = ?
    and mn.is_primary = TRUE;

-- name: DeleteProduct :exec
DELETE FROM products
WHERE
    product_id = ?;

-- name: AddProductMaterial :exec
insert into
    product_materials (product_id, material_id, quantity, quantity_text)
values
    (?, ?, ?, ?);

-- name: DeleteProductMaterial :exec
delete from product_materials
where
    product_id = ?
    and material_id = ?;

-- name: GetProductByName :one
select
    *
from
    products
where
    name = ?;

-- name: UpdateProductName :exec
UPDATE products
SET
    name = ?
WHERE
    product_id = ? RETURNING *;

-- name: UpdateProductMaterial :exec
UPDATE product_materials
SET
    quantity = ?,
    quantity_text = ?
WHERE
    product_id = ?
    AND material_id = ?;

-- name: UpdateProductDescription :exec
UPDATE products
SET
    description = ?
WHERE
    product_id = ? RETURNING *;

-- name: DeleteAllProductMaterials :exec
DELETE FROM product_materials
WHERE
    product_id = ?;
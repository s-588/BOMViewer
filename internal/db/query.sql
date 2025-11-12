-- name: GetAllMaterials :many
SELECT
  materials.*,
  unit_types.unit AS unit,
  material_names.name AS name,
  product_materials.quantity AS quantity,
  product_materials.quantity_text AS quantity_text
FROM
  materials
  inner join unit_types on materials.unit_id = unit_types.unit_id
  inner join material_names on material_names.material_id = materials.material_id
  inner join product_materials on product_materials.material_id = materials.material_id
where
  is_primary = TRUE ;

-- name: GetMaterialNames :many
select
  *
from
  material_names
where
  material_id = ?;

-- name: InsertMaterial :one
INSERT INTO
  materials (unit_id, description)
VALUES
  (
    (
      SELECT
        unit_id
      FROM
        unit_types
      WHERE
        unit_types.unit = ?
    ),
    ?
  ) RETURNING *;

-- name: AddMaterialName :one
INSERT INTO
  material_names (material_id, name, is_primary)
VALUES
  (?, ?, ?) RETURNING *;

-- name: GetMaterialByID :one
SELECT
  materials.*,
  unit_types.unit AS unit,
  product_materials.quantity AS quantity
FROM
  materials
  inner join unit_types on materials.unit_id = unit_types.unit_id
  inner join product_materials on product_materials.material_id = materials.material_id
WHERE
  materials.material_id = ?;

-- name: UpdateMaterialDescription :one
UPDATE materials
SET
  description = ?
WHERE
  material_id = ? RETURNING *;

-- name: UpdateMaterialUnit :one
UPDATE materials
SET
  unit = (
    SELECT
      unit_id
    FROM
      unit_types
    WHERE
      unit_types.unit = ?
  )
WHERE
  material_id = ? RETURNING *;

-- name: InsertMaterialName :one
INSERT INTO
  material_names (material_id, name, is_primary)
VALUES
  (?, ?, ?) RETURNING *;

-- name: DeleteAllMaterialNames :exec
DELETE FROM material_names
WHERE
  material_id = ?;

-- name: UpdateMaterialPrimaryName :one
UPDATE material_names
SET
  name = ?
WHERE
  material_id = ? RETURNING *;

-- name: SetMaterialPrimaryName :exec
UPDATE material_names
SET
  is_primary = TRUE
WHERE
  name = ?;

-- name: UnsetMaterialPrimaryName :exec
UPDATE material_names
SET
  is_primary = FALSE
WHERE
  is_primary = TRUE
  AND material_id = ?;

-- name: GetMaterialProducts :many
SELECT
  p.*,
  pm.quantity AS quantity,
  pm.quantity_text AS quantity_text,
  ut.unit AS unit
FROM
  products p
  INNER JOIN product_materials pm ON p.product_id = pm.product_id
  INNER JOIN materials m ON m.material_id = pm.material_id
  INNER JOIN unit_types ut ON ut.unit_id = m.unit_id
WHERE
  pm.material_id = ?;

-- name: GetMaterialFiles :many
SELECT
  f.*,
  fm.material_id AS material_id
FROM
  files f
  INNER JOIN files_materials fm ON f.file_id = fm.file_id
WHERE
  fm.material_id = ?;

-- name: InsertMaterialFile :one
INSERT INTO
  files_materials (material_id, file_id)
VALUES
  (?, ?) RETURNING *;

-- name: InsertFile :one
INSERT INTO
  files (name, path)
VALUES
  (?, ?) RETURNING *;

-- name: DeleteFile :exec
DELETE FROM files
WHERE
  file_id = ?;

-- name: DeleteMaterial :exec
DELETE FROM materials
WHERE
  material_id = ?;

-- name: GetAllProducts :many
SELECT
  *
FROM
  products;

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

-- name: GetProductFiles :many
SELECT
  f.*,
  fp.product_id AS product_id
FROM
  files f
  INNER JOIN files_products fp ON f.file_id = fp.file_id
WHERE
  fp.product_id = ?;

-- name: InsertProductFile :one
INSERT INTO
  files_products (product_id, file_id)
VALUES
  (?, ?) RETURNING *;

-- name: DeleteProductFile :exec
DELETE FROM files
WHERE
  file_id = ?;

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

-- name: GetMaterialByName :one
select
  m.*,
  pm.quantity AS quantity,
  pm.quantity_text AS quantity_text,
  mn.name AS material_name,
  ut.unit AS unit
from
  materials m
  inner join product_materials pm on m.material_id = pm.material_id
  inner join material_names mn on m.material_id = mn.material_id
  inner join unit_types ut on ut.unit_id = m.unit_id
where
  name = ?;

-- name: GetProductByName :one
select
  *
from
  products
where
  name = ?;

-- name: GetUnitByName :one
select
  *
from
  unit_types
where
  unit = ?;

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

-- name: GetAllUnits :many
SELECT
  *
FROM
  unit_types;
  
-- name: GetUnitByID :one
SELECT * from unit_types
where unit_id = ?;
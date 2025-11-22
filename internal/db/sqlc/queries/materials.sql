-- name: GetAllMaterials :many
SELECT 
    m.material_id,
    m.unit_id,
    m.description,
    ut.unit,
    mn.name AS primary_name,
    pm.quantity,
    pm.quantity_text,
    p.product_id,
    p.name AS product_name
FROM 
    materials m
    INNER JOIN unit_types ut ON m.unit_id = ut.unit_id
    INNER JOIN material_names mn ON m.material_id = mn.material_id AND mn.is_primary = TRUE
    LEFT JOIN product_materials pm ON m.material_id = pm.material_id
    LEFT JOIN products p ON pm.product_id = p.product_id
ORDER BY 
    m.material_id;

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

-- name: DeleteMaterial :exec
DELETE FROM materials
WHERE
    material_id = ?;

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
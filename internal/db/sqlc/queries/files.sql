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
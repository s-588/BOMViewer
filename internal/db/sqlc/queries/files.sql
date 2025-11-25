-- name: GetMaterialFiles :many
SELECT
    f.file_id,
    f.name,
    f.path,
    f.mime_type,
    f.file_type,
    fm.material_id
FROM
    files f
    INNER JOIN files_materials fm ON f.file_id = fm.file_id
WHERE
    fm.material_id = ?
    AND f.file_type <> 'profile-picture';

-- name: InsertMaterialFile :one
INSERT INTO
    files_materials (material_id, file_id)
VALUES
    (?, ?) RETURNING *;

-- name: InsertFile :one
INSERT INTO
    files (name, path, mime_type, file_type)
VALUES
    (?, ?, ?, ?) RETURNING *;

-- name: GetProductFiles :many
SELECT
    f.file_id,
    f.name,
    f.path,
    f.mime_type,
    f.file_type,
    fp.product_id
FROM
    files f
    INNER JOIN files_products fp ON f.file_id = fp.file_id
WHERE
    fp.product_id = ?
    AND f.file_type <> 'profile-picture';

-- name: InsertProductFile :one
INSERT INTO
    files_products (product_id, file_id)
VALUES
    (?, ?) RETURNING *;

-- name: DeleteProductFile :exec
DELETE FROM files_products
WHERE
    file_id = ?
    AND product_id = ?;

-- name: DeleteMaterialFile :exec
DELETE FROM files_materials
WHERE
    file_id = ?
    AND material_id = ?;

-- name: GetFileByID :one
SELECT
    *
FROM
    files
WHERE
    file_id = ?;

-- name: GetMaterialProfilePicture :one
SELECT
    f.*,
    fm.material_id
FROM
    files f
    INNER JOIN files_materials fm ON f.file_id = fm.file_id
WHERE
    f.file_type = 'profile-picture'
    AND fm.material_id = ?;

-- name: GetProductProfilePicture :one
SELECT
    f.*,
    fp.product_id
FROM
    files f
    INNER JOIN files_products fp ON f.file_id = fp.file_id
WHERE
    f.file_type = 'profile-picture'
    AND fp.product_id = ?;

-- name: UnsetMaterialProfilePicture :exec
UPDATE files
SET
    file_type = 'image'
WHERE
    file_type = 'profile-picture'
    AND file_id IN (
        SELECT
            file_id
        FROM
            files_materials
        WHERE
            material_id = ?
    );

-- name: UnsetProductProfilePicture :exec
UPDATE files
SET
    file_type = 'image'
WHERE
    file_type = 'profile-picture'
    AND file_id IN (
        SELECT
            file_id
        FROM
            files_products
        WHERE
            product_id = ?
    );

-- name: SetFileToProfilePicture :exec
UPDATE files
SET
    file_type = 'profile-picture'
WHERE
    file_id = ?;

-- name: DeleteFile :exec
DELETE FROM files
WHERE
    file_id = ?;

-- name: GetAllMaterialImages :many
SELECT
    f.file_id,
    f.name,
    f.path,
    f.mime_type,
    f.file_type
FROM
    files f
    INNER JOIN files_materials fm ON f.file_id = fm.file_id
WHERE
    fm.material_id = ?
    AND (
        f.file_type = 'image'
        OR f.file_type = 'profile-picture'
    );

-- name: GetAllProductImages :many
SELECT
    f.file_id,
    f.name,
    f.path,
    f.mime_type,
    f.file_type
FROM
    files f
    INNER JOIN files_products fp ON f.file_id = fp.file_id
WHERE
    fp.product_id = ?
    AND (
        f.file_type = 'image'
        OR f.file_type = 'profile-picture'
    );
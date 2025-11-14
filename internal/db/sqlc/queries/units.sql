-- name: GetAllUnits :many
SELECT
    *
FROM
    unit_types;

-- name: GetUnitByID :one
SELECT
    *
from
    unit_types
where
    unit_id = ?;

-- name: GetUnitByName :one
select
    *
from
    unit_types
where
    unit = ?;
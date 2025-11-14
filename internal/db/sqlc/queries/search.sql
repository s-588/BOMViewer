-- name: Search :many
SELECT
    *
FROM
    fts_all
WHERE
    text MATCH ?
LIMIT
    ?;
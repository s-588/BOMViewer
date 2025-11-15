-- name: SearchAll :many
SELECT
    fa.type,
    fa.ref_id,
    fa.text,
    COALESCE(
        case
            when fa.type = 'product' then p.name
            ELSE mn.name
        end,
        ''
    ) AS name
FROM
    fts_all fa
    left join products p on fa.ref_id = p.id
    and fa.type = 'product'
    left join material_names mn on fa.ref_id = mn.id
    and fa.type = 'material'
    and mn.is_primary = true
WHERE
    text MATCH ?
LIMIT
    ?;
-- name: SearchAll :many
SELECT
    f.type,
    f.ref_id,
    COALESCE(
        CASE
            WHEN f.type = 'product' THEN p.name
            WHEN f.type = 'material' THEN mn.name
        END,
        ''
    ) AS display_name,
    f.text,
    f.score
FROM
    fts f
    LEFT JOIN products p ON f.type = 'product'
    AND f.ref_id = p.product_id
    LEFT JOIN material_names mn ON f.type = 'material'
    AND f.ref_id = mn.material_id
    AND mn.is_primary = 1
WHERE
    f.text MATCH ?
ORDER BY
    f.score ASC
LIMIT
    ?;

-- name: SearchMaterials :many
SELECT
    f.ref_id,
    mn.name AS display_name,
    f.text,
    u.unit,
    COALESCE(pm.quantity, pm.quantity_text) AS quantity,
    score
FROM fts f
LEFT JOIN materials m
    ON f.ref_id = m.material_id
LEFT JOIN material_names mn
    ON f.ref_id = mn.material_id AND mn.is_primary = 1
LEFT JOIN product_materials pm
    ON m.material_id = pm.material_id
INNER JOIN unit_types u
    ON m.unit_id = u.unit_id
LEFT JOIN json_each(json(sqlc.arg(units))) je
    ON m.unit_id = je.value
WHERE
    f.text MATCH sqlc.arg(query)
    AND (
        json_array_length(json(sqlc.arg(units))) = 0
        OR je.value IS NOT NULL
    )
ORDER BY score ASC
LIMIT sqlc.arg(limit);


-- name: SearchProducts :many
SELECT
    f.ref_id,
    p.name AS display_name,
    f.text,
    f.score
FROM
    fts f
    LEFT JOIN products p ON f.ref_id = p.product_id
WHERE
    f.text MATCH ?
    AND f.type = 'product'
ORDER BY
    f.score ASC
LIMIT
    ?;
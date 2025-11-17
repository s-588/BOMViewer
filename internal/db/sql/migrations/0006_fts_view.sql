-- +goose Up
CREATE VIEW fts AS
SELECT
    type,
    ref_id,
    text,
    quantity,
    unit,
    bm25(fts_table) AS score
FROM fts_table;

-- +goose Down
DROP VIEW IF EXISTS fts;
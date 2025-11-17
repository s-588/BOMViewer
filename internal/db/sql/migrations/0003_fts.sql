-- +goose Up
CREATE VIRTUAL TABLE fts_table USING fts5(
    type UNINDEXED,
    ref_id UNINDEXED,
    text,
    tokenize='unicode61'
);

-- +goose Down
DROP TABLE IF EXISTS fts_table;

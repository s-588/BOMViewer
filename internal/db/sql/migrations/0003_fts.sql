-- 0003_fts.sql

CREATE VIRTUAL TABLE fts_all USING fts5(
    type UNINDEXED,
    ref_id UNINDEXED,
    text,
    tokenize='unicode61'
);

-- +goose Up
-- SQLite don't support ALTER TABLE ALTER COLUMN syntax so we need to create new table with changed columns.
-- And I decided to just add new columns to new table instead of using ALTER TABLE ADD COLUMN.
-- goose StatementBegin
CREATE TABLE
    files_new (
        file_id INTEGER PRIMARY KEY AUTOINCREMENT,
        name VARCHAR(50) NOT NULL,
        path TEXT NOT NULL,
        mime_type VARCHAR(100) DEFAULT 'application/octet-stream',
        file_type VARCHAR(20) DEFAULT 'document'
    );

-- Copy data 
INSERT INTO
    files_new
SELECT
    *,
    'application/octet-stream',
    'document'
FROM
    files;

-- Drop old table
DROP TABLE files;

-- Rename new table to original table name
ALTER TABLE files_new
RENAME TO files;

-- Fill file type since it null after copying data from old table
UPDATE files
SET
    file_type = 'document'
WHERE
    file_type IS NULL;
-- goose StatementEnd

-- +goose Down
-- goose StatementBegin
CREATE TABLE
    files_new (
        file_id INTEGER PRIMARY KEY AUTOINCREMENT,
        name VARCHAR(50) NOT NULL,
        PATH TEXT NOT NULL,
    );

-- Copy data 
INSERT INTO
    files_new
SELECT
    file_id, name, path
FROM
    files;

-- Drop old table
DROP TABLE files;

-- Rename new table to original table name
ALTER TABLE files_new
RENAME TO files;
-- goose StatementEnd
-- +goose Up
-- First, drop all triggers that reference material_names
DROP TRIGGER IF EXISTS material_ai;
DROP TRIGGER IF EXISTS material_au;
DROP TRIGGER IF EXISTS material_ad;
DROP TRIGGER IF EXISTS material_name_ai;
DROP TRIGGER IF EXISTS material_name_ad;
DROP TRIGGER IF EXISTS material_name_au;

-- Handle any existing duplicate names
DELETE FROM material_names
WHERE
    name_id NOT IN (
        SELECT
            MIN(name_id)
        FROM
            material_names
        GROUP BY
            name
    );

-- Create new table with UNIQUE constraint on name
CREATE TABLE
    material_names_new (
        name_id INTEGER PRIMARY KEY AUTOINCREMENT,
        material_id INT NOT NULL,
        name VARCHAR(255) NOT NULL UNIQUE,
        is_primary BOOLEAN NOT NULL DEFAULT FALSE,
        FOREIGN KEY (material_id) REFERENCES materials (material_id) ON DELETE CASCADE
    );

-- Copy data from old table
INSERT INTO
    material_names_new (name_id, material_id, name, is_primary)
SELECT
    name_id,
    material_id,
    name,
    is_primary
FROM
    material_names;

-- Drop old table
DROP TABLE material_names;

-- Rename new table
ALTER TABLE material_names_new
RENAME TO material_names;

-- Recreate necessary indexes
CREATE UNIQUE INDEX idx_material_primary_name ON material_names (material_id)
WHERE
    is_primary = 1;

-- Now recreate all the triggers from migration 0004
-- MATERIALS triggers
-- +goose StatementBegin
CREATE TRIGGER material_ai AFTER INSERT ON materials BEGIN
INSERT INTO
    fts_table (type, ref_id, text)
VALUES
    (
        'material',
        NEW.material_id,
        COALESCE((SELECT group_concat(name, ' ') FROM material_names WHERE material_id = NEW.material_id), '') || ' ' || COALESCE(NEW.description, '')
    );
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER material_au AFTER UPDATE ON materials BEGIN
DELETE FROM fts_table
WHERE
    type = 'material'
    AND ref_id = OLD.material_id;

INSERT INTO
    fts_table (type, ref_id, text)
VALUES
    (
        'material',
        NEW.material_id,
        COALESCE((SELECT group_concat(name, ' ') FROM material_names WHERE material_id = NEW.material_id), '') || ' ' || COALESCE(NEW.description, '')
    );
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER material_ad AFTER DELETE ON materials BEGIN
DELETE FROM fts_table
WHERE
    type = 'material'
    AND ref_id = OLD.material_id;
END;
-- +goose StatementEnd

-- MATERIAL NAMES triggers
-- +goose StatementBegin
CREATE TRIGGER material_name_ai AFTER INSERT ON material_names BEGIN
DELETE FROM fts_table
WHERE
    type = 'material'
    AND ref_id = NEW.material_id;

INSERT INTO
    fts_table (type, ref_id, text)
VALUES
    (
        'material',
        NEW.material_id,
        COALESCE((SELECT group_concat(name, ' ') FROM material_names WHERE material_id = NEW.material_id), '') || ' ' || COALESCE((SELECT description FROM materials WHERE material_id = NEW.material_id), '')
    );
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER material_name_ad AFTER DELETE ON material_names BEGIN
DELETE FROM fts_table
WHERE
    type = 'material'
    AND ref_id = OLD.material_id;

INSERT INTO
    fts_table (type, ref_id, text)
VALUES
    (
        'material',
        OLD.material_id,
        COALESCE((SELECT group_concat(name, ' ') FROM material_names WHERE material_id = OLD.material_id), '') || ' ' || COALESCE((SELECT description FROM materials WHERE material_id = OLD.material_id), '')
    );
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER material_name_au AFTER UPDATE ON material_names BEGIN
DELETE FROM fts_table
WHERE
    type = 'material'
    AND ref_id = NEW.material_id;

INSERT INTO
    fts_table (type, ref_id, text)
VALUES
    (
        'material',
        NEW.material_id,
        COALESCE((SELECT group_concat(name, ' ') FROM material_names WHERE material_id = NEW.material_id), '') || ' ' || COALESCE((SELECT description FROM materials WHERE material_id = NEW.material_id), '')
    );
END;
-- +goose StatementEnd

-- +goose Down
-- Drop all triggers
DROP TRIGGER IF EXISTS material_ai;
DROP TRIGGER IF EXISTS material_au;
DROP TRIGGER IF EXISTS material_ad;
DROP TRIGGER IF EXISTS material_name_ai;
DROP TRIGGER IF EXISTS material_name_ad;
DROP TRIGGER IF EXISTS material_name_au;

-- Revert to non-unique names
CREATE TABLE
    material_names_old (
        name_id INTEGER PRIMARY KEY AUTOINCREMENT,
        material_id INT NOT NULL,
        name VARCHAR(255) NOT NULL,
        is_primary BOOLEAN NOT NULL DEFAULT FALSE,
        FOREIGN KEY (material_id) REFERENCES materials (material_id) ON DELETE CASCADE
    );

-- Copy data back
INSERT INTO
    material_names_old (name_id, material_id, name, is_primary)
SELECT
    name_id,
    material_id,
    name,
    is_primary
FROM
    material_names;

-- Drop current table
DROP TABLE material_names;

-- Rename old table back
ALTER TABLE material_names_old
RENAME TO material_names;

-- Recreate original indexes
CREATE UNIQUE INDEX idx_material_primary_name ON material_names (material_id)
WHERE
    is_primary = 1;

CREATE UNIQUE INDEX idx_material_names_unique ON material_names (material_id, name);

-- Recreate triggers (same as in migration 0004)
-- MATERIALS triggers
-- +goose StatementBegin
CREATE TRIGGER material_ai AFTER INSERT ON materials BEGIN
INSERT INTO
    fts_table (type, ref_id, text)
VALUES
    (
        'material',
        NEW.material_id,
        COALESCE((SELECT group_concat(name, ' ') FROM material_names WHERE material_id = NEW.material_id), '') || ' ' || COALESCE(NEW.description, '')
    );
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER material_au AFTER UPDATE ON materials BEGIN
DELETE FROM fts_table
WHERE
    type = 'material'
    AND ref_id = OLD.material_id;

INSERT INTO
    fts_table (type, ref_id, text)
VALUES
    (
        'material',
        NEW.material_id,
        COALESCE((SELECT group_concat(name, ' ') FROM material_names WHERE material_id = NEW.material_id), '') || ' ' || COALESCE(NEW.description, '')
    );
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER material_ad AFTER DELETE ON materials BEGIN
DELETE FROM fts_table
WHERE
    type = 'material'
    AND ref_id = OLD.material_id;
END;
-- +goose StatementEnd

-- MATERIAL NAMES triggers
-- +goose StatementBegin
CREATE TRIGGER material_name_ai AFTER INSERT ON material_names BEGIN
DELETE FROM fts_table
WHERE
    type = 'material'
    AND ref_id = NEW.material_id;

INSERT INTO
    fts_table (type, ref_id, text)
VALUES
    (
        'material',
        NEW.material_id,
        COALESCE((SELECT group_concat(name, ' ') FROM material_names WHERE material_id = NEW.material_id), '') || ' ' || COALESCE((SELECT description FROM materials WHERE material_id = NEW.material_id), '')
    );
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER material_name_ad AFTER DELETE ON material_names BEGIN
DELETE FROM fts_table
WHERE
    type = 'material'
    AND ref_id = OLD.material_id;

INSERT INTO
    fts_table (type, ref_id, text)
VALUES
    (
        'material',
        OLD.material_id,
        COALESCE((SELECT group_concat(name, ' ') FROM material_names WHERE material_id = OLD.material_id), '') || ' ' || COALESCE((SELECT description FROM materials WHERE material_id = OLD.material_id), '')
    );
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER material_name_au AFTER UPDATE ON material_names BEGIN
DELETE FROM fts_table
WHERE
    type = 'material'
    AND ref_id = NEW.material_id;

INSERT INTO
    fts_table (type, ref_id, text)
VALUES
    (
        'material',
        NEW.material_id,
        COALESCE((SELECT group_concat(name, ' ') FROM material_names WHERE material_id = NEW.material_id), '') || ' ' || COALESCE((SELECT description FROM materials WHERE material_id = NEW.material_id), '')
    );
END;
-- +goose StatementEnd
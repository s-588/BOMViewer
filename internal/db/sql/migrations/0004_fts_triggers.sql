-- 0004_fts_triggers.sql

-- MATERIALS -----------------------------------------------------

CREATE TRIGGER material_ai AFTER INSERT ON materials BEGIN
    INSERT INTO fts_all (type, ref_id, text)
    VALUES (
        'material',
        NEW.material_id,
        (SELECT group_concat(name, ' ') FROM material_names WHERE material_id = NEW.material_id)
        || ' ' ||
        NEW.description
    );
END;

CREATE TRIGGER material_au AFTER UPDATE ON materials BEGIN
    DELETE FROM fts_all WHERE type='material' AND ref_id = OLD.material_id;

    INSERT INTO fts_all (type, ref_id, text)
    VALUES (
        'material',
        NEW.material_id,
        (SELECT group_concat(name, ' ') FROM material_names WHERE material_id = NEW.material_id)
        || ' ' ||
        NEW.description
    );
END;

CREATE TRIGGER material_ad AFTER DELETE ON materials BEGIN
    DELETE FROM fts_all WHERE type='material' AND ref_id = OLD.material_id;
END;

-- MATERIAL NAMES -------------------------------------------------

CREATE TRIGGER material_name_ai AFTER INSERT ON material_names BEGIN
    DELETE FROM fts_all WHERE type='material' AND ref_id = NEW.material_id;
    INSERT INTO fts_all (type, ref_id, text)
    VALUES (
        'material',
        NEW.material_id,
        (SELECT group_concat(name, ' ') FROM material_names WHERE material_id = NEW.material_id)
            || ' ' ||
        (SELECT description FROM materials WHERE material_id = NEW.material_id)
    );
END;

CREATE TRIGGER material_name_ad AFTER DELETE ON material_names BEGIN
    DELETE FROM fts_all WHERE type='material' AND ref_id = OLD.material_id;
    INSERT INTO fts_all (type, ref_id, text)
    VALUES (
        'material',
        OLD.material_id,
        (SELECT group_concat(name, ' ') FROM material_names WHERE material_id = OLD.material_id)
            || ' ' ||
        (SELECT description FROM materials WHERE material_id = OLD.material_id)
    );
END;

CREATE TRIGGER material_name_au AFTER UPDATE ON material_names BEGIN
    DELETE FROM fts_all WHERE type='material' AND ref_id = NEW.material_id;
    INSERT INTO fts_all (type, ref_id, text)
    VALUES (
        'material',
        NEW.material_id,
        (SELECT group_concat(name, ' ') FROM material_names WHERE material_id = NEW.material_id)
            || ' ' ||
        (SELECT description FROM materials WHERE material_id = NEW.material_id)
    );
END;

-- PRODUCTS -------------------------------------------------------

CREATE TRIGGER product_ai AFTER INSERT ON products BEGIN
    INSERT INTO fts_all (type, ref_id, text)
    VALUES ('product', NEW.product_id, NEW.name || ' ' || NEW.description);
END;

CREATE TRIGGER product_au AFTER UPDATE ON products BEGIN
    DELETE FROM fts_all WHERE type='product' AND ref_id = OLD.product_id;
    INSERT INTO fts_all (type, ref_id, text)
    VALUES ('product', NEW.product_id, NEW.name || ' ' || NEW.description);
END;

CREATE TRIGGER product_ad AFTER DELETE ON products BEGIN
    DELETE FROM fts_all WHERE type='product' AND ref_id = OLD.product_id;
END;

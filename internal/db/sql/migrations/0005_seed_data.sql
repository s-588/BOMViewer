-- 0005_seed_data.sql

INSERT INTO unit_types(unit) VALUES
    ('кг'),
    ('мм/кг'),
    ('м. пог.(кг.)'),
    ('мл'),
    ('л'),
    ('шт'),
    ('шт. (м)'),
    ('банка'),
    ('баллон'),
    ('см'),
    ('м'),
    ('м²'),
    ('м³');

INSERT INTO materials(unit_id) VALUES
    (1),(1),(1),(1),(1),(1),(1),(1),(1);

INSERT INTO material_names (material_id, name, is_primary) VALUES
    (1, 'Лист 25х1500х6000 10х СНД12', true),
    (2, 'Лист г/к 12х1500х6000', true),
    (3, 'Лист 3 г/к ст3пс/сп5', true),
    (4, 'Круг 10 ст20', true),
    (5, 'Круг 16', true),
    (6, 'Круг 20 (ст. 20, ст. 40, ст. 45)', true),
    (7, 'Труба х/к бесшовн.40х8 ст20', true),
    (8, 'Труба проф. 50х50х4 Ст3сп', true),
    (9, 'Труба проф. 50х50х2', true);

INSERT INTO products (name) VALUES
    ('Устройство остановки колесного транспорта'),
    ('Евроконтейнер ЕКМ');

INSERT INTO product_materials
(product_id, material_id, quantity, quantity_text) VALUES
    (1, 1, '61.75', null),
    (1, 2, '10.0', null),
    (1, 3, '2.0',  null),
    (1, 4, '8.5',  null),
    (1, 5, '1',    null),
    (1, 6, null,  '26,6/34,5'),
    (1, 7, null,  '38,1/51,06'),
    (1, 8, '3.55', null),
    (1, 9, '11.55', null);

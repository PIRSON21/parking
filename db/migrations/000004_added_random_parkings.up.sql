INSERT INTO parkings (parking_id, parking_name, parking_address, parking_width, parking_height)
VALUES
    (1, '1:Центр', 'ул. Ленина, 10', 5, 4),
    (2, '1:Центр', 'ул. Ленина, 20', 5, 6),
    (3, '1:Центр', 'ул. Советская, 5', 4, 5),
    (4, '1:ТЦ', 'пр. Победы, 25', 5, 5),
    (5, '1:ТЦ', 'пр. Победы, 30', 5, 6),
    (6, '1:ЖД', 'ул. Вокзальная, 3', 4, 4),
    (7, '1:ЖД', 'ул. Вокзальная, 7', 4, 6),
    (8, '1:Отель', 'ул. Горького, 12', 4, 6),
    (9, '1:Отель', 'ул. Горького, 15', 6, 5),
    (10, '1:БЦ', 'ул. Мира, 8', 5, 5);

SELECT setval('parkings_parking_id_seq', COALESCE((SELECT MAX(parking_id) FROM parkings), 1), false);

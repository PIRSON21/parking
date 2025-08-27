-- Вставляем менеджеров
INSERT INTO manager (manager_login, manager_password, manager_email) VALUES
    ('Александр Степанович', '$2y$10$rAzpI66iaFEF6E7E3tR3l.CzOaAa9Ecm/gG/H/hEkkr8PtRqB23Pm', 'admin1@ex.com'), -- pwd: aboba
    ('Иван Иванов', '$2y$10$DaxroQ.82gIk4MdDJ7mhhO6wYFPdQhSvxB4sGfqk1o7gPt5k0cJ9S', 'admin2@ex.com'), -- pwd: salam
    ('Ягон Дон', '$2y$10$m6W.50u1aVjWCE4PcEIF/.BNaFJi.C2izABx3nDvbymCWfhnJ3joi', 'admin3@ex.com'), -- pwd: egor_pomidor
    ('Яша Лава', '$2y$10$wEHjHgHlpn5jwfaGpBjMC.z248dRyceImPxgIHCFl0VxWP2vhkUPC', 'admin4@ex.com'); -- pwd: good_morning


SELECT setval('manager_manager_id_seq', COALESCE((SELECT MAX(manager_id) FROM manager), 1), false);

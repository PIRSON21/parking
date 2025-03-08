# Parking

Бэкенд сервис для моделирования работы платной парковки.

Создан на языке `Go` с использованием следующих библиотек:
- [cleanenv](https://github.com/ilyakaznacheev/cleanenv)
- [chi router](https://github.com/go-chi/chi)
- [chi render](https://github.com/go-chi/render)
- [pgx (используется только как драйвер, можно взять другой)](https://github.com/jackc/pgx)

База данных - `PostgresSQL`

## Установка
1. Перед установкой создайте ENV файла по [примеру](./config/template.env).
2. Установите миграции `migrate -database <url до вашей БД> -path db/migrations/ up -all`
3. При запуске укажите параметр `-path=<путь до env файла>`
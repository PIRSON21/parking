# Parking

Backend сервис для моделирования работы платной парковки.

Создан на языке `Go` с использованием следующих библиотек:
- [cleanenv](https://github.com/ilyakaznacheev/cleanenv)
- [chi router](https://github.com/go-chi/chi)
- [chi render](https://github.com/go-chi/render)
- [pgx (используется только как драйвер, можно взять другой)](https://github.com/jackc/pgx)
- [go validator](https://github.com/go-playground/validator)
- [gorilla websocket](https://github.com/gorilla/websocket)
- [ivahaev timer](https://github.com/ivahaev/timer)
- [testify](https://github.com/stretchr/testify)

База данных - `PostgresSQL`

## Запуск
Для полноценного запуска используется docker compose.

1. В папке `configs/` заполните файлы `app.env` и `db.env`
по примерам `app.env.template` и `db.env.template` соответственно.
2. Запустите команду
    ```bash
    docker compose --env-file=./configs/db.env up -d -y --wait
    ```
3. Для проверки работоспособности программы обратитесь к пути `<ваш адрес>/ping`
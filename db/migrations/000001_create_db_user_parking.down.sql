-- Завершаем все соединения с базой перед удалением
SELECT pg_terminate_backend(pg_stat_activity.pid)
FROM pg_stat_activity
WHERE pg_stat_activity.datname = 'parking'
AND pg_stat_activity.pid <> pg_backend_pid();

-- Удаление базы данных
DROP DATABASE IF EXISTS parking;


-- Удаление пользователя
DROP USER IF EXISTS parking;
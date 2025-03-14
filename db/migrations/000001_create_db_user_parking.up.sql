-- Создание базы данных
CREATE DATABASE parking;

-- Создание пользователя
CREATE USER parking WITH PASSWORD 'parking';

-- Назначение владельца базы данных
ALTER DATABASE parking OWNER TO parking;

-- Ограничение доступа к другим базам
REVOKE CONNECT ON DATABASE postgres FROM parking;

-- Назначение привелегий пользователю parking только для базы parking
GRANT CONNECT ON DATABASE parking TO parking;
GRANT USAGE ON SCHEMA public TO parking;
GRANT ALL PRIVILEGES ON SCHEMA public TO parking;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO parking;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO parking;
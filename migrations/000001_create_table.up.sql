-- migrations/000001_create_table.up.sql
-- Создание таблицы метрик
CREATE TABLE metrics (
                         id SERIAL PRIMARY KEY,
                         name VARCHAR(255) NOT NULL,
                         type_metrics VARCHAR(255) NOT NULL,
                         delta INTEGER,
                         value DOUBLE PRECISION
);

-- Базовый индекс для поиска по названию
CREATE UNIQUE INDEX idx_metrics_name ON metrics(name);

-- Базовый индекс для поиска по типу
CREATE INDEX idx_metrics_type ON metrics(type_metrics);
-- migrations/000001_create_table.down.sql
-- Откат создания таблицы
DROP INDEX IF EXISTS idx_metrics_name;
DROP INDEX IF EXISTS idx_metrics_type;
DROP TABLE IF EXISTS metrics; 
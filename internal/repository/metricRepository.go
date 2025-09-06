package repository

import (
	"context"
	"database/sql"
	"fmt"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"sort"
)

type MetricsRepository struct {
	db *sql.DB
}

func NewMetricsRepository(db *sql.DB) *MetricsRepository {
	return &MetricsRepository{
		db: db,
	}
}

func (r *MetricsRepository) UpdateMetric(ctx context.Context, value models.Metrics) error {
	switch value.MType {
	case models.Counter:
		_, err := r.db.ExecContext(ctx, "INSERT INTO metrics (name, type_metrics, delta) VALUES ($1, $2, $3) "+
			" ON CONFLICT (name) DO UPDATE"+
			" SET type_metrics = EXCLUDED.type_metrics,"+
			" delta = metrics.delta+EXCLUDED.delta",
			value.ID, value.MType, value.Delta)
		if err != nil {
			return fmt.Errorf("ошибка при обновлении Counter: %w", err)
		}
	case models.Gauge:
		_, err := r.db.ExecContext(ctx, "INSERT INTO metrics (name, type_metrics, \"value\") VALUES ($1, $2, $3) "+
			" ON CONFLICT (name) DO UPDATE"+
			" SET type_metrics = EXCLUDED.type_metrics,"+
			" value = EXCLUDED.value",
			value.ID, value.MType, value.Value)
		if err != nil {
			return fmt.Errorf("ошибка при обновлении Gauge: %w", err)
		}
	}
	return nil
}

func (r *MetricsRepository) GetAllMetrics(ctx context.Context) (map[string]*models.Metrics, error) {
	metrics := make(map[string]*models.Metrics)

	rows, err := r.db.QueryContext(ctx, "SELECT  \"name\", type_metrics, delta, \"value\" FROM metrics")
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении данных: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var v models.Metrics
		err = rows.Scan(&v.ID, &v.MType, &v.Delta, &v.Value)
		if err != nil {
			return nil, fmt.Errorf("ошибка при получении данных по строке: %w", err)
		}
		metrics[v.ID] = &v
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении данных: %w", err)
	}
	return metrics, nil
}

func (r *MetricsRepository) UpdateMetrics(ctx context.Context, values []models.Metrics) error {
	sort.Slice(values, func(i, j int) bool {
		return values[i].ID < values[j].ID
	})
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("не удалось создать транзакцию: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO metrics (name, type_metrics, delta, value) VALUES ($1, $2, $3, $4) "+
		" ON CONFLICT (name) DO UPDATE"+
		" SET type_metrics = COALESCE(EXCLUDED.type_metrics, metrics.type_metrics),"+
		" delta = CASE "+
		" WHEN $3 IS NOT NULL THEN metrics.delta + $3"+
		" ELSE metrics.delta "+
		" END, "+
		" value = COALESCE(EXCLUDED.value, metrics.value)")
	if err != nil {
		return fmt.Errorf("не удалось создать запрос: %w", err)
	}
	defer stmt.Close()

	for _, elem := range values {
		_, err := stmt.ExecContext(ctx, elem.ID, elem.MType, elem.Delta, elem.Value)
		if err != nil {
			return fmt.Errorf("не удалось вставить или обновить запись: %w", err)
		}
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("не удалось завершить запрос: %w", err)
	}
	return nil
}

func InitTables(ctx context.Context, db *sql.DB) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()
	query := `
			CREATE TABLE IF NOT EXISTS metrics (
            	id SERIAL PRIMARY KEY,
                name VARCHAR(255) NOT NULL,
                type_metrics VARCHAR(255) NOT NULL,
            	delta bigint,
                value DOUBLE PRECISION
)
`
	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		panic(err)
	}
	_, err = tx.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS idx_metrics_name ON metrics(name)`)
	if err != nil {
		panic(err)
	}

	_, err = tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_metrics_type ON metrics(type_metrics)`)
	if err != nil {
		panic(err)
	}
	err = tx.Commit()
	if err != nil {
		panic(err)
	}
}

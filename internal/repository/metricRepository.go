package repository

import (
	"context"
	"database/sql"
	"fmt"
	models "github.com/ValentinaKh/go-metrics/internal/model"
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

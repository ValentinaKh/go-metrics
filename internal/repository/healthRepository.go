package repository

import (
	"context"
	"database/sql"
)

type HealthRepository struct {
	db *sql.DB
}

func NewHealthRepository(db *sql.DB) *HealthRepository {
	return &HealthRepository{
		db: db,
	}
}

// Ping checks the database connection
func (h *HealthRepository) Ping(ctx context.Context) error {
	return h.db.PingContext(ctx)
}

package repository

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/ValentinaKh/go-metrics/internal/logger"
)

// MustConnectDB connect to db
func MustConnectDB(connectionString string) *sql.DB {
	if connectionString == "" {
		panic("connection string is empty")
	}
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		logger.Log.Panic("error open db", zap.Error(err))
		panic(err)
	}
	return db
}

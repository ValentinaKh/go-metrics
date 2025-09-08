package repository

import (
	"database/sql"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

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

package apperror

import (
	"context"
	"errors"
	"github.com/ValentinaKh/go-metrics/internal/config"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// ErrorClassification тип для классификации ошибок
type ErrorClassification int

const (
	// NonRetriable - операцию не следует повторять
	NonRetriable ErrorClassification = iota

	// Retriable - операцию можно повторить
	Retriable
)

type ErrorClassifier interface {
	Classify(err error) ErrorClassification
}

// PostgresErrorClassifier классификатор ошибок PostgreSQL
type PostgresErrorClassifier struct{}

func NewPostgresErrorClassifier() *PostgresErrorClassifier {
	return &PostgresErrorClassifier{}
}

// Classify классифицирует ошибку и возвращает PGErrorClassification
func (c *PostgresErrorClassifier) Classify(err error) ErrorClassification {
	if err == nil {
		return NonRetriable
	}

	// Проверяем и конвертируем в pgconn.PgError, если это возможно
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return ClassifyPgError(pgErr)
	}

	// По умолчанию считаем ошибку неповторяемой
	return NonRetriable
}

func ClassifyPgError(pgErr *pgconn.PgError) ErrorClassification {

	switch pgErr.Code {
	// Класс 08 - Ошибки соединения
	case pgerrcode.ConnectionException,
		pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure:
		return Retriable

	// Класс 40 - Откат транзакции
	case pgerrcode.TransactionRollback, // 40000
		pgerrcode.SerializationFailure, // 40001
		pgerrcode.DeadlockDetected:     // 40P01
		return Retriable

	// Класс 57 - Ошибка оператора
	case pgerrcode.CannotConnectNow: // 57P03
		return Retriable
	}

	// Можно добавить более конкретные проверки с использованием констант pgerrcode
	switch pgErr.Code {
	// Класс 22 - Ошибки данных
	case pgerrcode.DataException,
		pgerrcode.NullValueNotAllowedDataException:
		return NonRetriable

	// Класс 23 - Нарушение ограничений целостности
	case pgerrcode.IntegrityConstraintViolation,
		pgerrcode.RestrictViolation,
		pgerrcode.NotNullViolation,
		pgerrcode.ForeignKeyViolation,
		pgerrcode.UniqueViolation,
		pgerrcode.CheckViolation:
		return NonRetriable

	// Класс 42 - Синтаксические ошибки
	case pgerrcode.SyntaxErrorOrAccessRuleViolation,
		pgerrcode.SyntaxError,
		pgerrcode.UndefinedColumn,
		pgerrcode.UndefinedTable,
		pgerrcode.UndefinedFunction:
		return NonRetriable
	}

	// По умолчанию считаем ошибку неповторяемой
	return NonRetriable
}

// DoWithRetry выполняет повторный вызов функции в зависимости от типа ошибки.
// Конфиг повторов должен быть проверен на корректность до передачи в функцию
func DoWithRetry[T any](ctx context.Context, classifier ErrorClassifier, fn func() (T, error), cfg config.RetryConfig) (T, error) {
	var empty T
	attempts := cfg.MaxAttempts

	var lastErr error
	for i := 0; i <= attempts; i++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		if ctx.Err() != nil {
			return empty, ctx.Err()
		}
		lastErr = err

		classification := classifier.Classify(lastErr)
		if classification == NonRetriable {
			return empty, lastErr
		}

		if i < attempts {
			time.Sleep(cfg.Delays[i])
		}
	}
	return empty, lastErr
}

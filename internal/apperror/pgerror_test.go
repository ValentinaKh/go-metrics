package apperror

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/stretchr/testify/assert"
)

func TestPostgresErrorClassifier_Classify(t *testing.T) {
	classifier := NewPostgresErrorClassifier()

	t.Run("nil error should return NonRetriable", func(t *testing.T) {
		result := classifier.Classify(nil)
		assert.Equal(t, NonRetriable, result)
	})

	t.Run("non-pg error should return NonRetriable", func(t *testing.T) {
		regularErr := errors.New("application error")
		result := classifier.Classify(regularErr)
		assert.Equal(t, NonRetriable, result)
	})

	t.Run("wrapped non-pg error should return NonRetriable", func(t *testing.T) {
		innerErr := errors.New("failed to process data")
		wrappedErr := fmt.Errorf("failed to connect: %w", innerErr)
		result := classifier.Classify(wrappedErr)
		assert.Equal(t, NonRetriable, result)
	})
}

func TestPostgresErrorClassifier_RetriableErrors(t *testing.T) {
	classifier := NewPostgresErrorClassifier()

	retriableCases := []string{
		pgerrcode.ConnectionException,
		pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure,
		pgerrcode.TransactionRollback,
		pgerrcode.SerializationFailure,
		pgerrcode.DeadlockDetected,
		pgerrcode.CannotConnectNow,
	}

	for _, code := range retriableCases {
		t.Run("error with code "+code+" should be Retriable", func(t *testing.T) {
			pgErr := &pgconn.PgError{Code: code}
			result := classifier.Classify(pgErr)
			assert.Equal(t, Retriable, result)
		})
	}
}

func TestPostgresErrorClassifier_NonRetriableErrors(t *testing.T) {
	classifier := NewPostgresErrorClassifier()

	nonRetriableCases := []string{

		pgerrcode.DataException,
		pgerrcode.NullValueNotAllowedDataException,
		pgerrcode.IntegrityConstraintViolation,
		pgerrcode.RestrictViolation,
		pgerrcode.NotNullViolation,
		pgerrcode.ForeignKeyViolation,
		pgerrcode.UniqueViolation,
		pgerrcode.CheckViolation,
		pgerrcode.SyntaxErrorOrAccessRuleViolation,
		pgerrcode.SyntaxError,
		pgerrcode.UndefinedColumn,
		pgerrcode.UndefinedTable,
		pgerrcode.UndefinedFunction,
	}

	for _, code := range nonRetriableCases {
		t.Run("error with code "+code+" should be NonRetriable", func(t *testing.T) {
			pgErr := &pgconn.PgError{Code: code}
			result := classifier.Classify(pgErr)
			assert.Equal(t, NonRetriable, result)
		})
	}
}

func TestPostgresErrorClassifier_UnknownErrorCode(t *testing.T) {
	classifier := NewPostgresErrorClassifier()
	pgErr := &pgconn.PgError{Code: "XX000"}
	result := classifier.Classify(pgErr)

	assert.Equal(t, NonRetriable, result)
}

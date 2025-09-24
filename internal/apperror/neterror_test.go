package apperror

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestNetworkErrorClassifier_Classify(t *testing.T) {
	classifier := NewNetworkErrorClassifier()

	t.Run("nil error should return NonRetriable", func(t *testing.T) {
		result := classifier.Classify(nil)
		assert.Equal(t, NonRetriable, result)
	})

	t.Run("net.OpError should return Retriable", func(t *testing.T) {
		opErr := &net.OpError{
			Op:  "read",
			Net: "tcp",
			Err: errors.New("connection reset by peer"),
		}

		result := classifier.Classify(opErr)
		assert.Equal(t, Retriable, result)
	})

	t.Run("wrapped net.OpError should return Retriable", func(t *testing.T) {
		opErr := &net.OpError{
			Op:  "dial",
			Net: "tcp",
			Err: errors.New("connection refused"),
		}

		wrappedErr := fmt.Errorf("failed to connect: %w", opErr)

		result := classifier.Classify(wrappedErr)
		assert.Equal(t, Retriable, result)
	})

	t.Run("non-network error should return NonRetriable", func(t *testing.T) {
		regularErr := errors.New("some application error")

		result := classifier.Classify(regularErr)
		assert.Equal(t, NonRetriable, result)
	})
}

package apperror

import (
	"errors"
	"net"
)

// NetworkErrorClassifier классификатор ошибок при обращении к серверу
type NetworkErrorClassifier struct{}

func NewNetworkErrorClassifier() *NetworkErrorClassifier {
	return &NetworkErrorClassifier{}
}

// Classify классифицирует ошибку и возвращает ErrorClassification
func (c *NetworkErrorClassifier) Classify(err error) ErrorClassification {
	if err == nil {
		return NonRetriable
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return Retriable
	}

	return NonRetriable
}

package storage

import (
	"fmt"
	"github.com/ValentinaKh/go-metrics/internal/model"
	"sync"
)

type Storage interface {
	UpdateMetric(key string, value models.Metrics) error
}

type memStorage struct {
	mutex   sync.Mutex
	storage map[string]*models.Metrics
}

func NewMemStorage() Storage {
	return &memStorage{
		mutex:   sync.Mutex{},
		storage: map[string]*models.Metrics{},
	}
}

func (s *memStorage) UpdateMetric(key string, value models.Metrics) error {
	metric, ok := s.storage[key]
	if !ok {
		s.mutex.Lock()
		defer s.mutex.Unlock()
		s.storage[key] = &value
	} else {
		if metric.MType != value.MType {
			return fmt.Errorf("incorrect type")
		}

		s.mutex.Lock()
		defer s.mutex.Unlock()

		switch value.MType {
		case models.Counter:
			metric.Delta = addIntPtr(metric.Delta, value.Delta)
		case models.Gauge:
			metric.Value = value.Value
		}
	}

	return nil
}

func addIntPtr(a, b *int64) *int64 {
	va := int64(0)
	if a != nil {
		va = *a
	}
	res := va + *b
	return &res
}

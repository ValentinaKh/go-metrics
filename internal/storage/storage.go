package storage

import (
	"context"
	"fmt"
	"sync"

	models "github.com/ValentinaKh/go-metrics/internal/model"
)

type MemStorage struct {
	mutex   sync.Mutex
	storage map[string]*models.Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		mutex:   sync.Mutex{},
		storage: make(map[string]*models.Metrics),
	}
}

func (s *MemStorage) UpdateMetric(_ context.Context, value models.Metrics) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	key := value.ID
	metric, ok := s.storage[key]
	if !ok {
		s.storage[key] = &value
	} else {
		if metric.MType != value.MType {
			return fmt.Errorf("incorrect type")
		}

		switch value.MType {
		case models.Counter:
			metric.Delta = addIntPtr(metric.Delta, value.Delta)
		case models.Gauge:
			metric.Value = value.Value
		}
	}
	return nil
}

func (s *MemStorage) GetAndClear() map[string]*models.Metrics {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	copyMap := make(map[string]*models.Metrics)
	for k, v := range s.storage {
		copyMap[k] = v
	}
	clear(s.storage)
	return copyMap
}

func (s *MemStorage) GetAllMetrics(_ context.Context) (map[string]*models.Metrics, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	copyMap := make(map[string]*models.Metrics)
	for k, v := range s.storage {
		copyMap[k] = v
	}
	return copyMap, nil
}

func addIntPtr(a, b *int64) *int64 {
	va := int64(0)
	if a != nil {
		va = *a
	}
	res := va + *b
	return &res
}

func (s *MemStorage) UpdateMetrics(_ context.Context, values []models.Metrics) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, value := range values {
		key := value.ID
		metric, ok := s.storage[key]
		if !ok {
			s.storage[key] = &value
		} else {
			if metric.MType != value.MType {
				return fmt.Errorf("incorrect type")
			}

			switch value.MType {
			case models.Counter:
				metric.Delta = addIntPtr(metric.Delta, value.Delta)
			case models.Gauge:
				metric.Value = value.Value
			}
		}
	}
	return nil
}

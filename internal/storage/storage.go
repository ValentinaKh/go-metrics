package storage

import (
	"fmt"
	"github.com/ValentinaKh/go-metrics/internal/model"
	"sync"
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

func (s *MemStorage) UpdateMetric(key string, value models.Metrics) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

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

func (s *MemStorage) GetAllMetrics() map[string]*models.Metrics {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	copyMap := make(map[string]*models.Metrics)
	for k, v := range s.storage {
		copyMap[k] = v
	}
	return copyMap
}

func addIntPtr(a, b *int64) *int64 {
	va := int64(0)
	if a != nil {
		va = *a
	}
	res := va + *b
	return &res
}

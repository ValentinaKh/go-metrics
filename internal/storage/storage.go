package storage

import (
	"fmt"
	"github.com/ValentinaKh/go-metrics/internal/model"
	"sync"
)

type Storage interface {
	// UpdateMetric обновляем метрику в хранилище
	UpdateMetric(key string, value models.Metrics) error
	// GetAndClear овозвращаем то, что находится в хранилище и очищаем хранилище
	GetAndClear() map[string]*models.Metrics
}

type memStorage struct {
	mutex   sync.Mutex
	storage map[string]*models.Metrics
}

func NewMemStorage() Storage {
	return &memStorage{
		mutex:   sync.Mutex{},
		storage: make(map[string]*models.Metrics),
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

func (s *memStorage) GetAndClear() map[string]*models.Metrics {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	copyMap := make(map[string]*models.Metrics)
	for k, v := range s.storage {
		copyMap[k] = v
	}
	clear(s.storage)
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

package decorator

import (
	"encoding/json"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/ValentinaKh/go-metrics/internal/storage"
	"os"
)

type StoreWithSyncFile struct {
	*storeWithFile
}

func NewStoreWithSyncFile(storage *storage.MemStorage, fileName string) (*StoreWithSyncFile, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}
	s := &StoreWithSyncFile{
		storeWithFile: &storeWithFile{
			encoder:    json.NewEncoder(file),
			file:       file,
			MemStorage: storage,
		},
	}
	return s, nil
}

func (s *StoreWithSyncFile) UpdateMetric(value models.Metrics) error {
	err := s.MemStorage.UpdateMetric(value)
	if err != nil {
		return err
	}
	err = s.flushToFile()
	if err != nil {
		return err
	}
	return nil
}

func (s *StoreWithSyncFile) Close() error {
	return s.file.Close()
}

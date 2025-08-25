package decorator

import (
	"context"
	"encoding/json"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/ValentinaKh/go-metrics/internal/storage"
	"go.uber.org/zap"
	"os"
	"time"
)

type storeWithFile struct {
	*storage.MemStorage
	encoder *json.Encoder
	file    *os.File
}

type StoreWithAsyncFile struct {
	*storeWithFile
	interval time.Duration
}

const errorMsg = "Error when writing data on a file"

func NewStoreWithAsyncFile(notifyCtx context.Context, storage *storage.MemStorage,
	interval time.Duration, fileName string) (*StoreWithAsyncFile, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}

	s := &StoreWithAsyncFile{
		storeWithFile: &storeWithFile{
			encoder:    json.NewEncoder(file),
			file:       file,
			MemStorage: storage,
		},
		interval: interval,
	}
	go s.StartFlush(notifyCtx)
	return s, nil
}

func (s *StoreWithAsyncFile) StartFlush(notifyCtx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer func() {
		ticker.Stop()
		s.file.Close()
	}()

	for {
		select {
		case <-notifyCtx.Done():
			err := s.flushToFile()
			if err != nil {
				logger.Log.Error(errorMsg, zap.Error(err))
			} else {
				logger.Log.Info("AsyncFileStore stopped")
			}
			return
		case <-ticker.C:
			err := s.flushToFile()
			if err != nil {
				logger.Log.Error(errorMsg, zap.Error(err))
			}
		}
	}
}

func (s *storeWithFile) flushToFile() error {
	//удаляем все данные
	err := s.file.Truncate(0)
	if err != nil {
		return err
	}

	_, err = s.file.Seek(0, 0)
	if err != nil {
		return err
	}

	metrics := s.GetAllMetrics()
	tmp := make([]*models.Metrics, len(metrics))

	for k := range metrics {
		tmp = append(tmp, metrics[k])
	}
	if err := s.encoder.Encode(tmp); err != nil {
		return err
	}
	return nil
}

package decorator

import (
	"context"
	"github.com/ValentinaKh/go-metrics/internal/fileworker"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/ValentinaKh/go-metrics/internal/storage"
	"go.uber.org/zap"
	"time"
)

type StoreWithAsyncFile struct {
	*storage.MemStorage
	writer   fileworker.Writer
	interval time.Duration
}

const errorMsg = "Error when writing data on a file"

func NewStoreWithAsyncFile(notifyCtx context.Context, storage *storage.MemStorage,
	interval time.Duration, writer fileworker.Writer) (*StoreWithAsyncFile, error) {

	s := &StoreWithAsyncFile{
		MemStorage: storage,
		writer:     writer,
		interval:   interval,
	}
	go s.StartFlush(notifyCtx)
	return s, nil
}

func (s *StoreWithAsyncFile) StartFlush(notifyCtx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer func() {
		ticker.Stop()
		err := s.writer.Close()
		if err != nil {
			logger.Log.Error(err.Error())
		}
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

func (s *StoreWithAsyncFile) flushToFile() error {
	metrics, err := s.GetAllMetrics(context.TODO())
	if err != nil {
		return err
	}
	tmp := make([]*models.Metrics, 0)

	for k := range metrics {
		tmp = append(tmp, metrics[k])
	}
	if len(tmp) > 0 {
		if err := s.writer.Write(tmp); err != nil {
			return err
		}
	}
	return nil
}

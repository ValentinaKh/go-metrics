package agent

import (
	"context"
	"fmt"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/ValentinaKh/go-metrics/internal/storage"
	"strconv"
	"time"
)

type Agent interface {
	Push(ctx context.Context)
}

type metricAgent struct {
	s              storage.Storage
	h              Sender
	reportInterval time.Duration
}

func NewMetricAgent(s storage.Storage, h Sender, reportInterval time.Duration) Agent {
	return &metricAgent{s, h, reportInterval}
}

func (s *metricAgent) Push(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("close agent")
			return
		default:
			err := s.send()
			if err != nil {
				fmt.Println(err.Error())
			}
			time.Sleep(s.reportInterval)
		}
	}
}

func (s *metricAgent) send() error {
	metrics := s.s.GetAndClear()
	for key, metric := range metrics {
		var url string
		switch metric.MType {
		case models.Gauge:
			url = "http://localhost:8080/update/" + metric.MType + "/" + key + "/" + strconv.FormatFloat(*metric.Value, 'f', -1, 64)
		case models.Counter:
			url = "http://localhost:8080/update/" + metric.MType + "/" + key + "/" + strconv.FormatInt(*metric.Delta, 10)
		}
		if err := s.h.Send(url); err != nil {
			return err
		}
	}
	return nil
}

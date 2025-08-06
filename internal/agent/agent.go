package agent

import (
	"context"
	"fmt"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/ValentinaKh/go-metrics/internal/service"
	"log/slog"
	"net/url"
	"strconv"
	"time"
)

type Agent interface {
	Push(ctx context.Context)
}

type MetricAgent struct {
	s              service.Storage
	h              Sender
	reportInterval time.Duration
	host           string
}

func NewMetricAgent(s service.Storage, h Sender, reportInterval time.Duration, host string) *MetricAgent {
	return &MetricAgent{s, h, reportInterval, host}
}

func (s *MetricAgent) Push(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("close agent")
			return
		default:
			err := s.send()
			if err != nil {
				slog.Error(err.Error())
			}
			time.Sleep(s.reportInterval)
		}
	}
}

func (s *MetricAgent) send() error {
	metrics := s.s.GetAndClear()
	for key, metric := range metrics {
		var serverURL string
		switch metric.MType {
		case models.Gauge:
			serverURL = s.buildURL(metric.MType, key, strconv.FormatFloat(*metric.Value, 'f', -1, 64))
		case models.Counter:
			serverURL = s.buildURL(metric.MType, key, strconv.FormatInt(*metric.Delta, 10))
		}
		if err := s.h.Send(serverURL); err != nil {
			return err
		}
	}
	return nil
}

func (s *MetricAgent) buildURL(metricType, key, value string) string {
	u := &url.URL{
		Scheme: "http",
		Host:   s.host,
		Path:   fmt.Sprintf("/update/%s/%s/%s", metricType, key, value),
	}
	return u.String()
}

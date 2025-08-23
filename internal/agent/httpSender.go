package agent

import (
	"github.com/ValentinaKh/go-metrics/internal/logger"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type Sender interface {
	Send(url string) error
}

type HTTPSender struct {
	client *resty.Client
}

func NewPostSender() *HTTPSender {
	return &HTTPSender{client: resty.New()}
}

func (s *HTTPSender) Send(url string) error {
	resp, err := s.client.R().
		SetHeader("Content-Type", "text/plain").
		Post(url)
	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		logger.Log.Info("Status Code:", zap.Int("code", resp.StatusCode()))
	}

	return nil
}

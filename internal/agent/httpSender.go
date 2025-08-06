package agent

import (
	"github.com/go-resty/resty/v2"
	"log/slog"
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
		slog.Info("Status Code:", slog.Int("code", resp.StatusCode()))
	}

	return nil
}

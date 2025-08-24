package agent

import (
	"github.com/ValentinaKh/go-metrics/internal/logger"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"net/url"
)

type Sender interface {
	Send(data []byte) error
}

type HTTPSender struct {
	client *resty.Client
	url    string
}

func NewPostSender(host string) *HTTPSender {
	return &HTTPSender{client: resty.New(), url: buildURL(host)}
}

func (s *HTTPSender) Send(data []byte) error {
	resp, err := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(data).
		Post(s.url)
	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		logger.Log.Info("Status Code:", zap.Int("code", resp.StatusCode()))
	}

	return nil
}

func buildURL(host string) string {
	u := &url.URL{
		Scheme: "http",
		Host:   host,
		Path:   "/update/",
	}
	return u.String()
}

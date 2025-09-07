package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"github.com/ValentinaKh/go-metrics/internal/apperror"
	"github.com/ValentinaKh/go-metrics/internal/config"
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
	cfg    config.RetryConfig
}

func NewPostSender(host string, cfg config.RetryConfig) *HTTPSender {
	return &HTTPSender{client: resty.New(), url: buildURL(host), cfg: cfg}
}

func (s *HTTPSender) Send(data []byte) error {
	var compressedBody bytes.Buffer
	gz := gzip.NewWriter(&compressedBody)
	_, err := gz.Write(data)
	if err != nil {
		return err
	}
	err = gz.Close()
	if err != nil {
		return err
	}

	response, err := apperror.DoWithRetry(context.TODO(), apperror.NewNetworkErrorClassifier(), func() (*resty.Response, error) {
		return s.client.R().
			SetHeaders(map[string]string{"Content-Type": "application/json", "Content-Encoding": "gzip"}).
			SetBody(compressedBody.Bytes()).
			Post(s.url)
	}, s.cfg)
	if err != nil {
		return err
	}

	if response.StatusCode() != 200 {
		logger.Log.Info("Status Code:", zap.Int("code", response.StatusCode()))
	}

	return nil
}

func buildURL(host string) string {
	u := &url.URL{
		Scheme: "http",
		Host:   host,
		Path:   "/updates/",
	}
	return u.String()
}

package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"net/url"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"

	"github.com/ValentinaKh/go-metrics/internal/logger"
	"github.com/ValentinaKh/go-metrics/internal/retry"
	"github.com/ValentinaKh/go-metrics/internal/utils"
)

type HTTPSender struct {
	client    *resty.Client
	url       string
	retrier   *retry.Retrier
	secureKey string
}

func NewPostSender(host string, retrier *retry.Retrier, secureKey string) *HTTPSender {
	return &HTTPSender{client: resty.New(), url: buildURL(host), retrier: retrier, secureKey: secureKey}
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

	response, err := retry.DoWithRetry(context.TODO(), s.retrier, func() (*resty.Response, error) {
		body := compressedBody.Bytes()
		prep := s.client.R().
			SetHeaders(map[string]string{"Content-Type": "application/json", "Content-Encoding": "gzip"})

		if s.secureKey != "" {
			hash := utils.Hash(s.secureKey, body)
			prep.SetHeader("HashSHA256", fmt.Sprintf("%x", hash))
		}
		return prep.SetBody(body).Post(s.url)
	})
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

package agent

import (
	"bytes"
	"compress/gzip"
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

	resp, err := s.client.R().
		SetHeaders(map[string]string{"Content-Type": "application/json", "Content-Encoding": "gzip"}).
		SetBody(compressedBody.Bytes()).
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
		Path:   "/updates/",
	}
	return u.String()
}

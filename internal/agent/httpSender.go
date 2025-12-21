package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"github.com/ValentinaKh/go-metrics/internal/crypto"
	"net/url"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"

	"github.com/ValentinaKh/go-metrics/internal/logger"
	"github.com/ValentinaKh/go-metrics/internal/retry"
	"github.com/ValentinaKh/go-metrics/internal/utils"
)

// HTTPSender - позволяет отправлять данные по HTTP. Имеет возможность повторной отправки в случае неудачной попытки.
type HTTPSender struct {
	client    *resty.Client
	url       string
	retrier   *retry.Retrier
	secureKey string
	cs        *crypto.CryptoService[*x509.Certificate, *rsa.PublicKey]
}

func NewPostSender(host string, retrier *retry.Retrier, secureKey string,
	cs *crypto.CryptoService[*x509.Certificate, *rsa.PublicKey]) *HTTPSender {
	return &HTTPSender{client: resty.New(), url: buildURL(host), retrier: retrier, secureKey: secureKey, cs: cs}
}

// Send - Отправляет сжатые по gzip, а так же подписанные, если задан ключ, SHA256 данные на сервер.
// В случае неудачи повторяет попытку в соотвествии с настройками retrier.
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
		if s.cs != nil {
			body, err = s.cs.Transform(body)
			if err != nil {
				return nil, err
			}
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

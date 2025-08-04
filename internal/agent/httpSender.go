package agent

import (
	"fmt"
	"github.com/go-resty/resty/v2"
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
		fmt.Printf("Status Code: %d\r\n", resp.StatusCode())
	}

	return nil
}

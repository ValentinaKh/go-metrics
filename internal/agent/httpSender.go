package agent

import (
	"fmt"
	"io"
	"net/http"
)

type Sender interface {
	Send(url string) error
}

type HTTPSender struct {
	client *http.Client
}

func NewPostSender() Sender {
	return &HTTPSender{client: &http.Client{}}
}

func (s *HTTPSender) Send(url string) error {
	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "text/plain")
	response, err := s.client.Do(request)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(response.Body)

	if response.StatusCode != 200 {
		fmt.Printf("Status Code: %d\r\n", response.StatusCode)
	}

	return nil
}

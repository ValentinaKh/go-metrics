package rest

import (
	"encoding/json"
	"github.com/ValentinaKh/go-metrics/internal/audit"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/go-resty/resty/v2"
	"time"
)

// AuditHandler используется для записи аудита в rest api
type AuditHandler struct {
	client *resty.Client
	url    string
}

func NewAuditHandler(url string) *AuditHandler {
	return &AuditHandler{
		client: resty.New(),
		url:    url,
	}
}

func (s *AuditHandler) Update(request []models.Metrics, ip string) {
	rs, err := json.Marshal(audit.Dto{
		TS:        time.Now().Unix(),
		Metrics:   request,
		IPAddress: ip,
	})
	if err != nil {
		logger.Log.Error(err.Error())
		return
	}

	_, err = s.client.R().
		SetHeaders(map[string]string{"Content-Type": "application/json"}).SetBody(rs).Post(s.url)
	if err != nil {
		logger.Log.Error(err.Error())
	}

}

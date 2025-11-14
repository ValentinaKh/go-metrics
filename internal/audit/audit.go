package audit

import (
	"encoding/json"
	"github.com/ValentinaKh/go-metrics/internal/fileworker"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/go-resty/resty/v2"
	"time"
)

type Publisher interface {
	Register(observer)
	Notify(request []models.Metrics, ip string)
}

type observer interface {
	update(request []models.Metrics, ip string)
}

type Auditor struct {
	observers []observer
}

func (e *Auditor) Register(o observer) {
	e.observers = append(e.observers, o)
}

func (e *Auditor) Notify(request []models.Metrics, ip string) {
	for _, observer := range e.observers {
		observer.update(request, ip)
	}
}

type FileAuditHandler struct {
	writer fileworker.Writer
}

func NewFileAuditHandler(writer fileworker.Writer) *FileAuditHandler {
	return &FileAuditHandler{
		writer: writer,
	}
}

func (e *FileAuditHandler) update(request []models.Metrics, ip string) {
	err := e.writer.Write(auditDto{
		Ts:        time.Now().Unix(),
		Metrics:   request,
		IpAddress: ip,
	})
	if err != nil {
		logger.Log.Error(err.Error())
	}
}

type RestAuditHandler struct {
	client *resty.Client
	url    string
}

func NewRestAuditHandler(url string) *RestAuditHandler {
	return &RestAuditHandler{
		client: resty.New(),
		url:    url,
	}
}

func (s *RestAuditHandler) update(request []models.Metrics, ip string) {
	rs, err := json.Marshal(auditDto{
		Ts:        time.Now().Unix(),
		Metrics:   request,
		IpAddress: ip,
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

type auditDto struct {
	Ts        int64            `json:"ts"`
	Metrics   []models.Metrics `json:"metrics"`
	IpAddress string           `json:"ip_address"`
}

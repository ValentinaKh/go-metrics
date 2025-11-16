package file

import (
	"github.com/ValentinaKh/go-metrics/internal/audit"
	"github.com/ValentinaKh/go-metrics/internal/fileworker"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"time"
)

// AuditHandler используется для записи аудита в файл
type AuditHandler struct {
	writer fileworker.Writer
}

func NewFileAuditHandler(writer fileworker.Writer) *AuditHandler {
	return &AuditHandler{
		writer: writer,
	}
}

func (e *AuditHandler) Update(request []models.Metrics, ip string) {
	err := e.writer.Write(audit.Dto{
		TS:        time.Now().Unix(),
		Metrics:   request,
		IPAddress: ip,
	})
	if err != nil {
		logger.Log.Error(err.Error())
	}
}

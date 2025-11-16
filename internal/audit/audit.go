// Package audit используется для записи аудита в файл или в rest api
package audit

import (
	"context"
	models "github.com/ValentinaKh/go-metrics/internal/model"
)

type Publisher interface {
	Register(observer)
	Notify(request []models.Metrics, ip string)
}

type observer interface {
	Update(request []models.Metrics, ip string)
}

type Auditor struct {
	observers []observer
	tasks     chan task
}

func NewAuditor(ctx context.Context, queueSize uint64) *Auditor {
	a := &Auditor{
		tasks: make(chan task, queueSize),
	}
	a.startWorker(ctx)
	return a
}

// Register добавляет наблюдателя в список наблюдателей
func (e *Auditor) Register(o observer) {
	e.observers = append(e.observers, o)
}

// Notify вызывает метод update у всех наблюдателей, оповещает об изменении метрики
func (e *Auditor) Notify(request []models.Metrics, ip string) {
	for _, observer := range e.observers {
		observer.Update(request, ip)
	}
}
func (e *Auditor) startWorker(ctx context.Context) {
	go func() {
		for {
			select {
			case task, ok := <-e.tasks:
				if !ok {
					return
				}
				e.Notify(task.metrics, task.ip)
			case <-ctx.Done():
				return
			}
		}
	}()
}

type Dto struct {
	TS        int64            `json:"ts"`
	Metrics   []models.Metrics `json:"metrics"`
	IPAddress string           `json:"ip_address"`
}

type task struct {
	metrics []models.Metrics
	ip      string
}

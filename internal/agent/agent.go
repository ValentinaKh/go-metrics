// Package agent - содержит функции для запуска агента.
package agent

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"github.com/ValentinaKh/go-metrics/internal/crypto"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	"github.com/ValentinaKh/go-metrics/internal/proto/client"
	"github.com/ValentinaKh/go-metrics/internal/utils"
	"go.uber.org/zap"
	"sync"
	"time"

	"github.com/ValentinaKh/go-metrics/internal/apperror"
	"github.com/ValentinaKh/go-metrics/internal/config"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/ValentinaKh/go-metrics/internal/retry"
	"github.com/ValentinaKh/go-metrics/internal/service/collector"
	"github.com/ValentinaKh/go-metrics/internal/service/provider"
	"github.com/ValentinaKh/go-metrics/internal/service/writer"
	"github.com/ValentinaKh/go-metrics/internal/storage"
)

// ConfigureAgent - создает и запускает агента.
func ConfigureAgent(shutdownCtx context.Context, cfg *config.AgentArg, rCfg *config.RetryConfig, mChan chan []models.Metrics) (*sync.WaitGroup, error) {
	st := storage.NewMemStorage()
	metricPublisher, msgCh := NewMetricsPublisher(st, time.Duration(cfg.ReportInterval)*time.Second)

	var cs *crypto.CryptoService[*x509.Certificate, *rsa.PublicKey]
	var err error
	if cfg.CryptoKey != "" {
		cs, err = crypto.NewPublicKeyService(cfg.CryptoKey)
		if err != nil {
			return nil, err
		}
	}

	var wg sync.WaitGroup
	ip, err := utils.GetLocalIP()
	if err != nil {
		panic(err)
	}
	var grpcClient *client.GRPCClient
	if cfg.GRPCServerPort != "" {
		grpcClient, err = client.NewGRPCClient(cfg.GRPCServerPort, ip)
		if err != nil {
			logger.Log.Error("Не удалось запустить grpc client", zap.Error(err))
		}
	}
	for idx := 0; idx < int(cfg.RateLimit); idx++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			senders := make([]Sender, 0)
			senders = append(senders, NewPostSender(cfg.Host,
				retry.NewRetrier(
					retry.NewClassifierRetryPolicy(apperror.NewNetworkErrorClassifier(), rCfg.MaxAttempts),
					retry.NewStaticDelayStrategy(rCfg.Delays),
					&retry.SleepTimeProvider{}), cfg.Key, cs, ip))

			if grpcClient != nil {
				senders = append(senders, grpcClient)
			}
			NewMetricSender(senders, msgCh).Push(shutdownCtx)
		}()
	}

	duration := time.Duration(cfg.PollInterval)
	runtimeCollector := collector.NewMetricCollector(provider.NewRuntimeProvider(), duration*time.Second, mChan)
	systemCollector := collector.NewMetricCollector(provider.NewSystemProvider(), duration*time.Second, mChan)
	w := writer.NewMetricWriter(st, mChan)

	//по заданию надо добавить еще одну горутину с новыми метриками
	wg.Add(1)
	go func() {
		defer wg.Done()
		runtimeCollector.Collect(shutdownCtx)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		systemCollector.Collect(shutdownCtx)
	}()

	go metricPublisher.Publish(shutdownCtx)
	go w.Write(shutdownCtx)

	return &wg, nil
}

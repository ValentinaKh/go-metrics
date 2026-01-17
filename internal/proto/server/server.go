package server

import (
	"context"
	"github.com/ValentinaKh/go-metrics/internal/handler"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/ValentinaKh/go-metrics/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
	"time"
)

type MetricsGRPCService struct {
	proto.UnimplementedMetricsServer
	service handler.Service
}

func NewMetricsGRPCService(service handler.Service) *MetricsGRPCService {
	return &MetricsGRPCService{
		service: service,
	}
}

func (m *MetricsGRPCService) UpdateMetrics(ctx context.Context, r *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	timeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	errU := m.service.UpdateMetrics(timeout, convert(r.GetMetrics()))
	if errU != nil {
		return nil, status.Error(codes.Internal, errU.Error())
	}
	resp := proto.UpdateMetricsResponse_builder{}.Build()
	logger.Log.Info("обновление метрик по grpc")
	return resp, nil
}

func convert(m []*proto.Metric) []models.Metrics {
	result := make([]models.Metrics, len(m))
	for _, metric := range m {
		delta := metric.GetDelta()
		value := metric.GetValue()

		var mType string
		switch metric.GetType() {
		case proto.Metric_COUNTER:
			mType = "counter"
		case proto.Metric_GAUGE:
			mType = "gauge"
		}

		result = append(result, models.Metrics{
			ID:    metric.GetId(),
			MType: mType,
			Delta: &delta,
			Value: &value,
		})
	}
	return result
}

type GRPCServer struct {
	port             string
	service          proto.MetricsServer
	srv              *grpc.Server
	unaryInterceptor *UnaryInterceptor
}

func NewGRPCServer(port string, service proto.MetricsServer, unaryInterceptor *UnaryInterceptor) *GRPCServer {
	return &GRPCServer{
		port:             port,
		service:          service,
		unaryInterceptor: unaryInterceptor,
	}
}

func (s *GRPCServer) Run() error {
	listen, err := net.Listen("tcp", s.port)
	if err != nil {
		logger.Log.Error("ошибка при инициализации listener", zap.Error(err))
		return err
	}

	if s.unaryInterceptor != nil {
		s.srv = grpc.NewServer(grpc.UnaryInterceptor(s.unaryInterceptor.check))
	} else {
		// Создаем gRPC сервер без зарегистрированной службы
		s.srv = grpc.NewServer()
	}
	// Регистрируем сервис
	proto.RegisterMetricsServer(s.srv, s.service)

	logger.Log.Info("сервер gRPC начал работу")
	// Получение запроса gRpc
	go func() {
		if err := s.srv.Serve(listen); err != nil {
			logger.Log.Error("ошибка при работе сервера", zap.Error(err))
		}
	}()
	return nil
}

func (s *GRPCServer) Stop() {
	if s.srv != nil {
		s.srv.GracefulStop()
	}
}

package client

import (
	"context"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	models "github.com/ValentinaKh/go-metrics/internal/model"
	"github.com/ValentinaKh/go-metrics/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"net"
)

type GRPCClient struct {
	conn *grpc.ClientConn
	c    proto.MetricsClient
	ip   net.IP
}

func NewGRPCClient(port string, ip net.IP) (*GRPCClient, error) {
	// Устанавливаем соединение с сервером
	conn, err := grpc.NewClient(port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Error("ошибка при установлении соединения с сервером", zap.Error(err))
		return nil, err
	}
	c := proto.NewMetricsClient(conn)
	return &GRPCClient{conn: conn, c: c, ip: ip}, nil
}

func (g *GRPCClient) Close() {
	err := g.conn.Close()
	if err != nil {
		logger.Log.Error("ошибка при закрытии соединения с сервером", zap.Error(err))
	}
}

func (g *GRPCClient) Send(data []*models.Metrics) error {
	md := metadata.New(map[string]string{"X-Real-IP": g.ip.String()})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	metrics := proto.UpdateMetricsRequest_builder{Metrics: convertToProto(data)}.Build()
	_, err := g.c.UpdateMetrics(ctx, metrics)
	if err != nil {
		return err
	}
	return nil
}

func convertToProto(data []*models.Metrics) []*proto.Metric {
	res := make([]*proto.Metric, 0, len(data))
	for _, metric := range data {
		var t proto.Metric_MType
		if metric.MType == "counter" {
			t = proto.Metric_COUNTER
		} else {
			t = proto.Metric_GAUGE
		}

		m := proto.Metric_builder{
			Id:    metric.ID,
			Type:  t,
			Delta: getValue(metric.Delta, 0),
			Value: getValue(metric.Value, 0),
		}.Build()
		res = append(res, m)
	}
	return res
}

func getValue[T any](ptr *T, defaultValue T) T {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

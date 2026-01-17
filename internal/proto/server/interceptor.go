package server

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net"
)

type UnaryInterceptor struct {
	trustedSubnet *net.IPNet
}

func NewUnaryInterceptor(subnet string) (*UnaryInterceptor, error) {
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, err
	}
	return &UnaryInterceptor{trustedSubnet: ipnet}, nil
}

func (u *UnaryInterceptor) check(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var clientIP string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("X-Real-IP")
		if len(values) > 0 {
			clientIP = values[0]
		}
	}
	if clientIP == "" {
		return nil, status.Error(codes.PermissionDenied, "invalid ip")
	}

	ip := net.ParseIP(clientIP)
	if ip == nil || !u.trustedSubnet.Contains(ip) {
		return nil, status.Error(codes.PermissionDenied, "invalid ip")
	}
	return handler(ctx, req)
}

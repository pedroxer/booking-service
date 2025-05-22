package metric_interceptor

import (
	"context"
	"github.com/pedroxer/booking-service/internal/prometheus"
	"google.golang.org/grpc"
)

type MetricInterceptor struct {
}

func NewMetricInterceptor() *MetricInterceptor {
	return &MetricInterceptor{}
}

func (m *MetricInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		prometheus.RpcMetricCounterInc()
		return handler(ctx, req)
	}
}

func (m *MetricInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		prometheus.RpcMetricCounterInc()
		return handler(srv, stream)
	}
}

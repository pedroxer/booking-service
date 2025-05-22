package grpc_app

import (
	"fmt"
	my_grpc "github.com/pedroxer/booking-service/internal/grpc"
	"github.com/pedroxer/booking-service/internal/grpc/metric_interceptor"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
)

type App struct {
	logger     *log.Logger
	grpcServer *grpc.Server
	port       int
}
type Interceptor interface {
	Unary() grpc.UnaryServerInterceptor
	Stream() grpc.StreamServerInterceptor
}

func NewApp(log *log.Logger, port int, bookingService my_grpc.BookingInterface) *App {
	interceptor := metric_interceptor.NewMetricInterceptor()
	server := grpc.NewServer(grpc.UnaryInterceptor(interceptor.Unary()), grpc.StreamInterceptor(interceptor.Stream()))
	my_grpc.RegisterBookingServiceServer(server, log, bookingService)
	return &App{
		logger:     log,
		grpcServer: server,
		port:       port,
	}
}

func (a *App) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		log.Fatal(err)
	}
	a.logger.Infof("starting grpc server on port %d", a.port)
	return a.grpcServer.Serve(l)
}

func (a *App) Stop() {
	a.logger.Info("stopping grpc server")
	a.grpcServer.GracefulStop()
}

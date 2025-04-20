package app

import (
	grpc_app "github.com/pedroxer/booking-service/internal/app/grpc"
	proto_gen "github.com/pedroxer/booking-service/internal/proto_gen/protos"
	"github.com/pedroxer/booking-service/internal/services/booking"
	"github.com/pedroxer/booking-service/internal/storage"
	log "github.com/sirupsen/logrus"
)

type App struct {
	GRPCSrv *grpc_app.App
}

func NewApp(log *log.Logger, grpcPort int, store *storage.Storage, resourceClient proto_gen.ResourceServiceClient) *App {
	bookingService := booking.NewBookingService(log, resourceClient, store, store, store)
	grpcApp := grpc_app.NewApp(
		log,
		grpcPort,
		bookingService,
	)

	return &App{
		GRPCSrv: grpcApp,
	}
}

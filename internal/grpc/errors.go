package my_grpc

import (
	"errors"
	"github.com/pedroxer/booking-service/internal/utills"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func generateErrors(err error) error {
	switch true {
	case errors.Is(err, utills.ErrNoRows):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

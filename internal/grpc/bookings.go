package my_grpc

import (
	"context"
	"github.com/pedroxer/booking-service/internal/models"
	proto_gen "github.com/pedroxer/booking-service/internal/proto_gen/protos"
	"github.com/pedroxer/booking-service/internal/utills"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type BookingInterface interface {
	GetBookings(ctx context.Context, bookingType string, startTime, endTime time.Time, userId string, resourceID, page int64) ([]models.Booking, int64, error)
	GetBookingById(ctx context.Context, bookingType string, bookingId int64) (models.Booking, error)
	CreateBooking(ctx context.Context, bookingType, status string, startTime, endTime time.Time, userId string, resourceId int64) (models.Booking, error)
	UpdateBooking(ctx context.Context, bookingType, status string, bookingID int64, startTime, endTime time.Time) (models.Booking, error)
	CancelBooking(ctx context.Context, bookingType string, bookingId int64) (bool, error)
	ApproveBooking(ctx context.Context, uniqueTag string) (bool, error) // Только для workplace
	GetTimeSlotsForBooking(ctx context.Context, bookingType string, resourceId int64, date time.Time) ([]models.TimeSlot, error)
}

type bookingAPI struct {
	proto_gen.UnimplementedBookingServiceServer
	bookingService BookingInterface
	logger         *log.Logger
}

func RegisterBookingServiceServer(server *grpc.Server, log *log.Logger, bookingService BookingInterface) {
	proto_gen.RegisterBookingServiceServer(server, &bookingAPI{logger: log, bookingService: bookingService})
}

func (b *bookingAPI) CreateBooking(ctx context.Context, req *proto_gen.CreateBookingRequest) (*proto_gen.Booking, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}
	if req.BookingType == "" {
		return nil, status.Error(codes.InvalidArgument, "resource type is required. Available resource types: workplace, parking")
	}
	if req.ResourceId == 0 {
		return nil, status.Error(codes.InvalidArgument, "resource id is required")
	}
	if req.StartTime == nil {
		return nil, status.Error(codes.InvalidArgument, "start time is required")
	}
	if req.EndTime == nil {
		return nil, status.Error(codes.InvalidArgument, "end time is required")
	}
	b.logger.Infof("Creating booking with user id: %d and resource type: %s, id: %d ", req.UserId, req.BookingType, req.ResourceId)
	resp, err := b.bookingService.CreateBooking(ctx, req.BookingType, utills.StatusPending, protoTimestampToTime(req.StartTime), protoTimestampToTime(req.EndTime), req.UserId, req.ResourceId)
	if err != nil {
		b.logger.Errorf("Error creating booking: %v", err)
		return nil, generateErrors(err)
	}
	return bookingToGrpcBooking(&resp), nil
}

func (b *bookingAPI) GetBookingById(ctx context.Context, req *proto_gen.GetBookingByIdRequest) (*proto_gen.Booking, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "booking id is required")
	}
	if req.BookingType == "" {
		return nil, status.Error(codes.InvalidArgument, "resource type is required. Available resource types: workplace, parking")
	}
	resp, err := b.bookingService.GetBookingById(ctx, req.BookingType, req.Id)
	if err != nil {
		b.logger.Errorf("Error getting booking: %v", err)
		return nil, generateErrors(err)
	}
	return bookingToGrpcBooking(&resp), nil
}

func (b *bookingAPI) GetBookings(ctx context.Context, req *proto_gen.GetBookingsRequest) (*proto_gen.GetBookingsResponse, error) {
	if req.Page == 0 {
		req.Page = 1
	}

	if req.BookingType == "" {
		return nil, status.Error(codes.InvalidArgument, "resource type is required. Available resource types: workplace, parking")
	}
	//if req.UserId == 0 {
	//	return nil, status.Error(codes.InvalidArgument, "user id is required")
	//}
	resp, count, err := b.bookingService.GetBookings(ctx, req.BookingType, protoTimestampToTime(req.StartTime), protoTimestampToTime(req.EndTime), req.UserId, req.ResourceId, req.Page)
	if err != nil {
		b.logger.Errorf("Error getting bookings: %v", err)
		return nil, generateErrors(err)
	}
	grpcResp := &proto_gen.GetBookingsResponse{
		Page:       req.Page,
		PageSize:   utills.PageSize,
		TotalCount: count/utills.PageSize + 1,
	}
	for _, booking := range resp {
		grpcResp.Bookings = append(grpcResp.Bookings, bookingToGrpcBooking(&booking))
	}
	return grpcResp, nil
}

func (b *bookingAPI) UpdateBooking(ctx context.Context, req *proto_gen.UpdateBookingRequest) (*proto_gen.Booking, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "booking id is required")
	}
	if req.BookingType == "" {
		return nil, status.Error(codes.InvalidArgument, "resource type is required. Available resource types: workplace, parking")
	}
	b.logger.Infof("updating booking %s with id: %d", req.BookingType, req.Id)
	resp, err := b.bookingService.UpdateBooking(ctx, req.BookingType, req.Status, req.Id, protoTimestampToTime(req.StartTime), protoTimestampToTime(req.EndTime))
	if err != nil {
		b.logger.Errorf("Error updating booking: %v", err)
		return nil, generateErrors(err)
	}
	return bookingToGrpcBooking(&resp), nil
}

func (b *bookingAPI) CancelBooking(ctx context.Context, req *proto_gen.CancelBookingRequest) (*proto_gen.CancelBookingResponse, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "booking id is required")
	}
	if req.BookingType == "" {
		return nil, status.Error(codes.InvalidArgument, "resource type is required. Available resource types: workplace, parking")
	}
	b.logger.Infof("Canceling booking %s with id: %d", req.BookingType, req.Id)
	resp, err := b.bookingService.CancelBooking(ctx, req.BookingType, req.Id)
	if err != nil {
		b.logger.Errorf("Error canceling booking: %v", err)
		return &proto_gen.CancelBookingResponse{
			Success: resp,
		}, generateErrors(err)
	}
	return &proto_gen.CancelBookingResponse{
		Success: resp,
	}, nil
}

func (b *bookingAPI) ApproveByQRBooking(ctx context.Context, req *proto_gen.ApproveByQRBookingRequest) (*proto_gen.ApproveByQRBookingResponse, error) {
	if req.UniqueTag == "" {
		return nil, status.Error(codes.InvalidArgument, "unique tag is required")
	}
	success, err := b.bookingService.ApproveBooking(ctx, req.UniqueTag)
	if err != nil {
		b.logger.Errorf("Error approving booking: %v", err)
		return &proto_gen.ApproveByQRBookingResponse{Success: success}, generateErrors(err)
	}
	return &proto_gen.ApproveByQRBookingResponse{Success: success}, nil
}

func (b *bookingAPI) GetSlotsToBooking(ctx context.Context, req *proto_gen.GetSlotsToBookingRequest) (*proto_gen.GetSlotsToBookingResponse, error) {
	if req.BookingType == "" {
		return nil, status.Error(codes.InvalidArgument, "resource type is required. Available resource types: workplace, parking")
	}
	if req.ResourceId == 0 {
		return nil, status.Error(codes.InvalidArgument, "resource id is required")
	}
	if req.Date == nil {
		return nil, status.Error(codes.InvalidArgument, "date is required")
	}
	b.logger.Info("getting free slot for booking: ", req.BookingType, " with resourceId: ", req.ResourceId)
	timeSlots, err := b.bookingService.GetTimeSlotsForBooking(ctx, req.BookingType, req.ResourceId, protoTimestampToTime(req.Date))
	if err != nil {
		b.logger.Errorf("Error getting free slots for booking: %v", err)
		return nil, generateErrors(err)
	}
	grpcTimeSlots := make([]*proto_gen.TimeSlot, len(timeSlots))
	for i, timeSlot := range timeSlots {
		grpcTimeSlots[i] = &proto_gen.TimeSlot{
			StartTime: timestamppb.New(timeSlot.StartTime),
			EndTime:   timestamppb.New(timeSlot.EndTime),
		}
	}
	return &proto_gen.GetSlotsToBookingResponse{Slots: grpcTimeSlots}, nil
}

func bookingToGrpcBooking(model *models.Booking) *proto_gen.Booking {
	return &proto_gen.Booking{
		Id:         model.BookingId,
		UserId:     model.UserId,
		ResourceId: model.ResourceId,
		StartTime:  timestamppb.New(model.StartTime),
		EndTime:    timestamppb.New(model.EndTime),
		Status:     model.Status,
		CreatedAt:  timestamppb.New(model.CreatedAt),
		UpdatedAt:  timestamppb.New(model.UpdatedAt),
	}
}

func protoTimestampToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{} // Zero time
	}
	return ts.AsTime()
}

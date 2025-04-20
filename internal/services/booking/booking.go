package booking

import (
	"context"
	"errors"
	"github.com/pedroxer/booking-service/internal/models"
	proto_gen "github.com/pedroxer/booking-service/internal/proto_gen/protos"
	"github.com/pedroxer/booking-service/internal/storage"
	"github.com/pedroxer/booking-service/internal/utills"
	log "github.com/sirupsen/logrus"
	"time"
)

type BookingGetter interface {
	GetBookings(ctx context.Context, filters []storage.Field, bookingType string, page int64) ([]models.Booking, int64, error)
	GetBookingsById(ctx context.Context, bookingType string, bookingId int64) (models.Booking, error)
	GetTimeSlotsForResource(ctx context.Context, bookingType string, resourceId int64, date time.Time) ([]models.TimeSlot, error)
}

type BookingCreater interface {
	CreateBooking(ctx context.Context, bookingType, status string, startTime, endTime time.Time, userId, resourceId int64) (models.Booking, error)
}

type BookingUpdater interface {
	UpdateBooking(ctx context.Context, bookingID int64, updateFields []storage.Field, bookingType string) (models.Booking, error)
	DeleteBooking(ctx context.Context, bookingType string, bookingId int64) error
	ApproveBooking(ctx context.Context, workplaceId int64) (bool, error)
}
type BookingService struct {
	logger         *log.Logger
	resourceClient proto_gen.ResourceServiceClient
	bookingGetter  BookingGetter
	bookingUpdater BookingUpdater
	bookingCreater BookingCreater
}

func NewBookingService(logger *log.Logger, resourceClient proto_gen.ResourceServiceClient, bookingGetter BookingGetter, creater BookingCreater, updater BookingUpdater) *BookingService {

	return &BookingService{
		logger:         logger,
		resourceClient: resourceClient,
		bookingGetter:  bookingGetter,
		bookingCreater: creater,
		bookingUpdater: updater,
	}

}
func (b BookingService) GetBookings(ctx context.Context, bookingType string, startTime, endTime time.Time, userId, resourceID, page int64) ([]models.Booking, int64, error) {

	filters := make([]storage.Field, 0)
	if !startTime.IsZero() {
		filters = append(filters, storage.Field{
			Name:  "start_date",
			Value: startTime,
		})
	}
	if !endTime.IsZero() {
		filters = append(filters, storage.Field{
			Name:  "end_date",
			Value: endTime,
		})
	}
	if bookingType == utills.WorkplaceType {
		if resourceID != 0 {
			filters = append(filters, storage.Field{
				Name:  "workplace_id",
				Value: resourceID,
			})
		}
	} else if bookingType == utills.ParkingType {
		if resourceID != 0 {
			filters = append(filters, storage.Field{
				Name:  "parking_space_id",
				Value: resourceID,
			})
		}
	}
	if userId != 0 {
		filters = append(filters, storage.Field{
			Name:  "user_id",
			Value: userId,
		})
	}

	bookings, count, err := b.bookingGetter.GetBookings(ctx, filters, bookingType, page)
	if err != nil {
		b.logger.Warnf("Error getting bookings: %s", err.Error())
		return nil, 0, err
	}
	return bookings, count, err
}

func (b BookingService) GetBookingById(ctx context.Context, bookingType string, bookingId int64) (models.Booking, error) {
	booking, err := b.bookingGetter.GetBookingsById(ctx, bookingType, bookingId)
	if err != nil {
		b.logger.Warnf("Error getting booking: %s", err.Error())
		return models.Booking{}, err
	}
	return booking, nil
}

func (b BookingService) CreateBooking(ctx context.Context, bookingType, status string, startTime, endTime time.Time, userId, resourceId int64) (models.Booking, error) {
	var (
		resourceAvailable bool
		err               error
	)
	if bookingType == utills.WorkplaceType {
		workplace, err := b.resourceClient.GetWorkplaceById(ctx, &proto_gen.GetWorkplaceByIdRequest{
			Id: resourceId,
		})
		if err != nil {
			b.logger.Warn("Error getting resource ", err)
			return models.Booking{}, err
		}
		resourceAvailable = workplace.IsAvailable
	} else if bookingType == utills.ParkingType {
		parking, err := b.resourceClient.GetParkingSpaceById(ctx, &proto_gen.GetParkingSpaceByIdRequest{
			Id: resourceId,
		})
		if err != nil {
			b.logger.Warn("Error getting resource ", err)
			return models.Booking{}, err
		}
		resourceAvailable = parking.IsAvailable
	}

	if !resourceAvailable {
		b.logger.Warn("Resource is not available")
		return models.Booking{}, errors.New("resource is not available")
	}

	booking, err := b.bookingCreater.CreateBooking(ctx, bookingType, status, startTime, endTime, userId, resourceId)
	if err != nil {
		b.logger.Warnf("Error creating booking: %s", err.Error())
		return models.Booking{}, err
	}
	if bookingType == utills.WorkplaceType {
		_, err = b.resourceClient.UpdateWorkplace(ctx, &proto_gen.UpdateWorkplaceRequest{
			Id:          resourceId,
			IsAvailable: false,
		})
	} else if bookingType == utills.ParkingType {
		_, err = b.resourceClient.UpdateParkingSpace(ctx, &proto_gen.UpdateParkingSpaceRequest{
			Id:          resourceId,
			IsAvailable: false,
		})
	}
	if err != nil {
		b.logger.Warnf("Error updating resource: %s", err.Error())
		return models.Booking{}, err
	}
	return booking, nil
}

func (b BookingService) UpdateBooking(ctx context.Context, bookingType, status string, bookingID int64, startTime, endTime time.Time) (models.Booking, error) {
	updateFields := make([]storage.Field, 0)
	if startTime != time.Unix(0, 0) {
		updateFields = append(updateFields, storage.Field{
			Name:  "start_date",
			Value: startTime,
		})
	}
	if endTime != time.Unix(0, 0) {
		updateFields = append(updateFields, storage.Field{
			Name:  "end_date",
			Value: endTime,
		})
	}
	if status != "" {
		updateFields = append(updateFields, storage.Field{
			Name:  "status",
			Value: status,
		})
	}
	booking, err := b.bookingUpdater.UpdateBooking(ctx, bookingID, updateFields, bookingType)
	if err != nil {
		b.logger.Warnf("Error updating booking: %s", err.Error())
		return models.Booking{}, err
	}
	return booking, nil
}

func (b BookingService) CancelBooking(ctx context.Context, bookingType string, bookingId int64) (bool, error) {
	booking, err := b.bookingGetter.GetBookingsById(ctx, bookingType, bookingId)
	if err != nil {
		b.logger.Warnf("Error getting booking: %s", err.Error())
		return false, err
	}
	if bookingType == utills.WorkplaceType {
		_, err = b.resourceClient.UpdateWorkplace(ctx, &proto_gen.UpdateWorkplaceRequest{
			Id:          booking.ResourceId,
			IsAvailable: true,
		})
	} else if bookingType == utills.ParkingType {
		_, err = b.resourceClient.UpdateParkingSpace(ctx, &proto_gen.UpdateParkingSpaceRequest{
			Id:          booking.ResourceId,
			IsAvailable: true,
		})
	}
	if err != nil {
		b.logger.Warnf("Error updating resource: %s", err.Error())
		return false, err
	}

	err = b.bookingUpdater.DeleteBooking(ctx, bookingType, bookingId)
	if err != nil {
		b.logger.Warnf("Error deleting booking: %s", err.Error())
		return false, err
	}
	return true, nil

}

func (b BookingService) ApproveBooking(ctx context.Context, uniqueTag string) (bool, error) {
	workplace, err := b.resourceClient.GetWorkplaceByUniqueTag(ctx, &proto_gen.GetWorkplaceByUniqueTagRequest{UniqueTag: uniqueTag})
	if err != nil {
		b.logger.Warnf("Error getting resource: %s", err.Error())
		return false, err
	}

	success, err := b.bookingUpdater.ApproveBooking(ctx, workplace.Id)
	if err != nil {
		b.logger.Warnf("Error approving booking: %s", err.Error())
		return false, err
	}
	return success, nil
}

func (b BookingService) GetTimeSlotsForBooking(ctx context.Context, bookingType string, resourceId int64, date time.Time) ([]models.TimeSlot, error) {
	timeSlots, err := b.bookingGetter.GetTimeSlotsForResource(ctx, bookingType, resourceId, date)
	if err != nil {
		b.logger.Warnf("Error getting free slots: %s", err.Error())
		return nil, err
	}
	resultTimeSlots := make([]models.TimeSlot, 0)
	for i := 0; i < len(timeSlots)-1; i++ {
		var timeSlot models.TimeSlot
		diff := timeSlots[i].EndTime.Sub(timeSlots[i+1].StartTime)
		if diff.Hours() > 0 {
			timeSlot.StartTime = timeSlots[i].EndTime
			timeSlot.EndTime = timeSlots[i+1].StartTime
			timeSlot.Busy = false
		}
		resultTimeSlots = append(resultTimeSlots, timeSlot)
	}
	timeSlots = append(timeSlots, resultTimeSlots...)
	return timeSlots, nil
}

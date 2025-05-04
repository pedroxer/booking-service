package storage

import (
	"context"
	"fmt"
	"github.com/pedroxer/booking-service/internal/models"
	"github.com/pedroxer/booking-service/internal/utills"
	"strings"
	"time"
)

func (s *Storage) GetBookings(ctx context.Context, filters []Field, bookingType string, page int64) ([]models.Booking, int64, error) {

	var bookingColumnsFields = map[string]SearchField{
		"id":         {NameWhere: "id", NameOrder: "id"},
		"user_id":    {NameWhere: "user_id", NameOrder: "user_id"},
		"start_date": {NameWhere: "start_date", NameOrder: "start_date"},
		"end_date":   {NameWhere: "end_date", NameOrder: "end_date"},
		"status":     {NameWhere: "status", NameOrder: "status"},
		"created_at": {NameWhere: "created_at", NameOrder: "created_at"},
		"updated_at": {NameWhere: "updated_at", NameOrder: "updated_at"},
	}

	var bookingFileds = []string{
		"id",
		"user_id",
		"start_date",
		"end_date",
		"status",
		"created_at",
		"updated_at",
	}
	from := ` FROM `

	switch bookingType {
	case utills.WorkplaceType:
		bookingColumnsFields["workplace_id"] = SearchField{NameWhere: "workplace_id", NameOrder: "workplace_id"}
		from += "booking_service.booking"
		bookingFileds = append(bookingFileds, "workplace_id")
	case utills.ParkingType:
		bookingColumnsFields["parking_space_id"] = SearchField{NameWhere: "parking_space_id", NameOrder: "parking_space_id"}
		from += "booking_service.parking_bookings"
		bookingFileds = append(bookingFileds, "parking_space_id")
	default:
		s.logger.Warnf("unknown booking type: %s", bookingType)
		return nil, 0, fmt.Errorf("booking type %s not supported", bookingType)
	}

	selectQuery := "SELECT " + strings.Join(bookingFileds, ",") + from
	countQuery := `SELECT count(*) FROM (` + selectQuery

	where, err := GenerateSearch(bookingColumnsFields, filters)
	if err != nil {
		s.logger.Warn(err)
		return nil, 0, err
	}

	var conditions strings.Builder
	if len(filters) > 0 {
		conditions.WriteString(" WHERE")
		conditions.WriteString(where)
	}
	selectQuery += conditions.String() + GenerateLimits(page, utills.PageSize)

	rows, err := s.db.Query(ctx, selectQuery)
	if err != nil {
		s.logger.Warn(err)
		return nil, 0, err
	}
	defer rows.Close()
	var bookings []models.Booking
	var bookingCount int64
	for rows.Next() {
		var booking models.Booking
		if err := rows.Scan(
			&booking.BookingId,
			&booking.UserId,
			&booking.StartTime,
			&booking.EndTime,
			&booking.Status,
			&booking.CreatedAt,
			&booking.UpdatedAt,
			&booking.ResourceId); err != nil {
			s.logger.Warn(err)
			return nil, 0, err
		}
		bookings = append(bookings, booking)

	}

	countQuery += conditions.String() + ") as cnt"

	if err := s.db.QueryRow(ctx, countQuery).Scan(&bookingCount); err != nil {
		s.logger.Warn(err)
		return nil, 0, err
	}

	if bookingType == utills.WorkplaceType {
		bookingFileds = bookingFileds[0 : len(bookingColumnsFields)-1]
		delete(bookingColumnsFields, "workplace_id")
	} else if bookingType == utills.ParkingType {
		bookingFileds = bookingFileds[0 : len(bookingColumnsFields)-1]
		delete(bookingColumnsFields, "parking_space_id")
	}
	return bookings, bookingCount, nil
}

func (s *Storage) GetBookingsById(ctx context.Context, bookingType string, bookingId int64) (models.Booking, error) {
	var bookingColumnsFields = map[string]SearchField{
		"id":         {NameWhere: "id", NameOrder: "id"},
		"user_id":    {NameWhere: "user_id", NameOrder: "user_id"},
		"start_date": {NameWhere: "start_date", NameOrder: "start_date"},
		"end_date":   {NameWhere: "end_date", NameOrder: "end_date"},
		"status":     {NameWhere: "status", NameOrder: "status"},
		"created_at": {NameWhere: "created_at", NameOrder: "created_at"},
		"updated_at": {NameWhere: "updated_at", NameOrder: "updated_at"},
	}

	var bookingFileds = []string{
		"id",
		"user_id",
		"start_date",
		"end_date",
		"status",
		"created_at",
		"updated_at",
	}

	from := ` FROM `

	switch bookingType {
	case utills.WorkplaceType:
		from += "booking_service.booking"
		bookingFileds = append(bookingFileds, "workplace_id")
	case utills.ParkingType:
		from += "booking_service.parking_bookings"
		bookingFileds = append(bookingFileds, "parking_space_id")
	default:
		s.logger.Warnf("unknown booking type: %s", bookingType)
		return models.Booking{}, fmt.Errorf("booking type %s not supported", bookingType)
	}

	selectQuery := "SELECT " + strings.Join(bookingFileds, ",") + from + " WHERE id = $1"
	var booking models.Booking
	if err := s.db.QueryRow(ctx, selectQuery, bookingId).Scan(&booking.BookingId,
		&booking.UserId,
		&booking.StartTime,
		&booking.EndTime,
		&booking.Status,
		&booking.CreatedAt,
		&booking.UpdatedAt,
		&booking.ResourceId); err != nil {
		s.logger.Warn(err)
		return models.Booking{}, err
	}

	bookingFileds = bookingFileds[0 : len(bookingColumnsFields)-1]

	return booking, nil
}

func (s *Storage) CreateBooking(ctx context.Context, bookingType, status string, startTime, endTime time.Time, userId string, resourceId int64) (models.Booking, error) {
	query := `INSERT INTO `
	if bookingType == utills.WorkplaceType {
		query += "booking_service.booking (user_id, workplace_id, start_date, end_date, status, created_at, updated_at) "
	} else if bookingType == utills.ParkingType {
		query += "booking_service.parking_bookings (user_id, parking_space_id, start_date, end_date, status, created_at, updated_at) "
	}
	query += " VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *"

	var booking models.Booking
	if err := s.db.QueryRow(ctx, query, userId, resourceId, startTime, endTime, status, time.Now(), time.Now()).Scan(&booking.BookingId,
		&booking.UserId,
		&booking.ResourceId,
		&booking.StartTime,
		&booking.EndTime,
		&booking.Status,
		&booking.CreatedAt,
		&booking.UpdatedAt); err != nil {
		s.logger.Warn(err)
		return models.Booking{}, err
	}

	return booking, nil
}

func (s *Storage) ApproveBooking(ctx context.Context, workplaceId int64) (bool, error) {
	query := `UPDATE booking_service.booking SET status = $2 WHERE workplace_id = $1`

	if _, err := s.db.Exec(ctx, query, workplaceId, utills.StatusWorking); err != nil {
		s.logger.Warn(err)
		return false, err
	}
	return true, nil
}

func (s *Storage) UpdateBooking(ctx context.Context, bookingID int64, updateFields []Field, bookingType string) (models.Booking, error) {
	var bookingColumnsFields = map[string]SearchField{
		"id":         {NameWhere: "id", NameOrder: "id"},
		"user_id":    {NameWhere: "user_id", NameOrder: "user_id"},
		"start_date": {NameWhere: "start_date", NameOrder: "start_date"},
		"end_date":   {NameWhere: "end_date", NameOrder: "end_date"},
		"status":     {NameWhere: "status", NameOrder: "status"},
		"created_at": {NameWhere: "created_at", NameOrder: "created_at"},
		"updated_at": {NameWhere: "updated_at", NameOrder: "updated_at"},
	}

	var bookingFileds = []string{
		"id",
		"user_id",
		"start_date",
		"end_date",
		"status",
		"created_at",
		"updated_at",
	}

	updateQuery := ` UPDATE`
	switch bookingType {
	case utills.WorkplaceType:
		bookingColumnsFields["workplace_id"] = SearchField{NameWhere: "workplace_id", NameOrder: "workplace_id"}
		updateQuery += "booking"
		bookingFileds = append(bookingFileds, "workplace_id")
	case utills.ParkingType:
		bookingColumnsFields["parking_space_id"] = SearchField{NameWhere: "parking_space_id", NameOrder: "parking_space_id"}
		updateQuery += "parking_bookings"
		bookingFileds = append(bookingFileds, "parking_space_id")
	default:
		s.logger.Warnf("unknown booking type: %s", bookingType)
		return models.Booking{}, fmt.Errorf("booking type %s not supported", bookingType)
	}
	updateQuery += " SET "

	updates, err := GenerateUpdates(bookingColumnsFields, updateFields)
	if err != nil {
		s.logger.Warn(err)
		return models.Booking{}, err
	}
	updateQuery += updates + fmt.Sprintf(" WHERE id = %d RETURNING *", bookingID)
	var booking models.Booking
	if err := s.db.QueryRow(ctx, updateQuery, bookingID).Scan(&booking.BookingId,
		&booking.UserId,
		&booking.StartTime,
		&booking.EndTime,
		&booking.Status,
		&booking.CreatedAt,
		&booking.UpdatedAt,
		&booking.ResourceId); err != nil {
		s.logger.Warn(err)
		return models.Booking{}, err
	}
	return booking, nil
}

func (s *Storage) GetTimeSlotsForResource(ctx context.Context, bookingType string, resourceId int64, date time.Time) ([]models.TimeSlot, error) {
	query := `SELECT start_date, end_date FROM `
	if bookingType == utills.WorkplaceType {
		query += "booking_service.booking WHERE workplace_id = $1 "
	} else if bookingType == utills.ParkingType {
		query += "booking_service.parking_bookings WHERE parking_space_id = $1 "
	}
	query += "AND start_date >= $2 AND end_date <= $3 ORDER BY start_date"
	rows, err := s.db.Query(ctx, query, resourceId, date.Format(utills.TimeLayout), date.Add(time.Hour*24).Format(utills.TimeLayout))
	if err != nil {
		s.logger.Warn(err)
		return nil, err
	}
	var timeSlots []models.TimeSlot
	for rows.Next() {
		var timeSlot models.TimeSlot
		if err := rows.Scan(&timeSlot.StartTime, &timeSlot.EndTime); err != nil {
			s.logger.Warn(err)
			return nil, err
		}
		timeSlot.Busy = true
		timeSlots = append(timeSlots, timeSlot)
	}

	return timeSlots, nil

}
func (s *Storage) DeleteBooking(ctx context.Context, bookingType string, bookingId int64) error {
	query := `DELETE FROM `
	if bookingType == utills.WorkplaceType {
		query += "booking_service.booking WHERE id = $1"
	} else {
		query += "booking_service.parking_bookings WHERE id = $1"
	}
	if _, err := s.db.Exec(ctx, query, bookingId); err != nil {
		s.logger.Warn(err)
		return err
	}
	return nil

}

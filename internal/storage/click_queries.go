package storage

import (
	"context"
	"time"
)

func (s *Storage) AddToClickHouse(ctx context.Context,
	bookingId, resourceId int64,
	userId, bookingType, bookingStatus, address, zone string,
	floor, number int64,
	eventDate, eventTime, startBookingTime, endBookingTime time.Time,
	durationMinutes int64) error {
	query := `INSERT INTO booking_analytics (booking_id, 
                               resource_id, 
                               user_id, 
                               booking_type, 
                               booking_status,
                               address, 
                               zone, 
                               floor, 
                               number, 
                               event_date, 
                               event_time, 
                               start_booking_time, 
                               end_booking_time, 
							   duration_minutes) VALUES ($1,
							                             $2,
							                             $3,
							                             $4,
							                             $5,
							                             $6,
							                             $7,
							                             $8,
							                             $9,
							                             $10,
							                             $11,
							                             $12,
							                             $13,
							                             $14);`
	err := s.clickDb.Exec(ctx, query, bookingId,
		resourceId,
		userId,
		bookingType,
		bookingStatus,
		address,
		zone,
		floor,
		number,
		eventDate,
		eventTime,
		startBookingTime,
		endBookingTime,
		durationMinutes)
	if err != nil {
		return err
	}
	return nil
}

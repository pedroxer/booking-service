package models

import "time"

type Booking struct {
	BookingId  int64     `json:"id"`
	UserId     int64     `json:"user_id"`
	ResourceId int64     `json:"resource_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type TimeSlot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Busy      bool      `json:"busy"`
}

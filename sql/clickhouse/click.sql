CREATE DATABASE IF NOT EXISTS analytics;

CREATE TABLE analytics.booking_analytics(
    booking_id int,
    resource_id int,
    user_id varchar,
    booking_type Enum('workplace' = 1, 'parking' = 2),
    booking_status Enum('PENDING' = 1, 'CONFIRMED' = 2, 'DONE' = 3, 'CANCELED' = 4),

    address varchar,
    zone varchar,
    floor int,
    number int,

    event_date Date,
    event_time DateTime,
    start_booking_time DateTime,
    end_booking_time DateTime,
    duration_minutes int

)
ENGINE = MergeTree
ORDER BY (booking_id, user_id, event_date)
PRIMARY KEY (booking_id, user_id)

package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/pedroxer/booking-service/internal/config"
)

func ConnectToPg(cfg *config.Postgres) (*pgx.Conn, error) {
	dsn := fmt.Sprintf(`postgres://%s:%s@%s:%d/%s`,
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Db)
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	if err = conn.Ping(context.Background()); err != nil {
		return nil, err
	}
	return conn, nil
}

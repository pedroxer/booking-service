package storage

import (
	"github.com/jackc/pgx/v5"
	"github.com/pedroxer/booking-service/internal/config"
	"github.com/pedroxer/booking-service/internal/database"
	log "github.com/sirupsen/logrus"
)

type Storage struct {
	db     *pgx.Conn
	logger *log.Logger
}

func NewStorage(cfg *config.Postgres, logger *log.Logger) (*Storage, error) {
	pgConn, err := database.ConnectToPg(cfg)
	if err != nil {
		return nil, err
	}
	return &Storage{
		db:     pgConn,
		logger: logger,
	}, nil
}

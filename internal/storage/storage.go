package storage

import (
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/jackc/pgx/v5"
	"github.com/pedroxer/booking-service/internal/config"
	"github.com/pedroxer/booking-service/internal/database"
	log "github.com/sirupsen/logrus"
)

type Storage struct {
	pgDb    *pgx.Conn
	clickDb driver.Conn
	logger  *log.Logger
}

func NewStorage(pgCfg *config.Postgres, clickCfg *config.Clickhouse, logger *log.Logger) (*Storage, error) {
	pgConn, err := database.ConnectToPg(pgCfg)
	if err != nil {
		return nil, err
	}
	clickConn, err := database.ConnectToClick(clickCfg)
	if err != nil {
		return nil, err
	}
	return &Storage{
		pgDb:    pgConn,
		clickDb: clickConn,
		logger:  logger,
	}, nil
}

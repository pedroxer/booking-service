package main

import (
	"encoding/json"
	"github.com/caarlos0/env/v6"
	"github.com/pedroxer/booking-service/internal/app"
	"github.com/pedroxer/booking-service/internal/config"
	"github.com/pedroxer/booking-service/internal/prometheus"
	"github.com/pedroxer/booking-service/internal/storage"
	"github.com/pedroxer/booking-service/internal/utills"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	log := setupLogger()
	data, err := os.ReadFile("./config/config.json")
	if err != nil {
		log.Fatal(err)
	}
	cfg := new(config.Config)

	if err := json.Unmarshal(data, cfg); err != nil {
		log.Fatal(err)
	}
	if err := env.Parse(cfg); err != nil {
		log.Fatal(err)
	}

	go func() {
		err = prometheus.RunRestServer()
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Info("starting http server")
	if err := prometheus.MetricsInit(); err != nil {
		log.Fatal(err)
	}
	store, err := storage.NewStorage(&cfg.Postgres, &cfg.Clickhouse, log)
	if err != nil {
		log.Fatalf("failed connect to db %s", err)
	}
	log.Info("connected to db")
	resourceClient, err := utills.CreateResourceClient(cfg.ResourceService)
	if err != nil {
		log.Fatal("failed to create resource client ", err)
	}
	log.Info("connected to resource service")
	app := app.NewApp(log, cfg.Port, store, resourceClient)
	if err := app.GRPCSrv.Run(); err != nil {
		log.Fatal(err)
	}
}

func setupLogger() *log.Logger {
	log := log.New()
	log.ReportCaller = true
	return log
}

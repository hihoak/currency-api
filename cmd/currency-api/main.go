package main

import (
	"context"
	"flag"
	"github.com/hihoak/currency-api/internal/app/registrator"
	"github.com/hihoak/currency-api/internal/app/users"
	"github.com/hihoak/currency-api/internal/clients/storager"
	"github.com/hihoak/currency-api/internal/pkg/config"
	"github.com/hihoak/currency-api/internal/pkg/logger"
	"net/http"

	_ "github.com/lib/pq"
)

var (
	configFile = ".currency_api.yaml"
)

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/.calendar_config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()
	ctx := context.Background()

	cfg := config.New(configFile)
	logg := logger.New(cfg.Logger)
	logg.Info().Msg("Successfully initialize config...")

	store := storager.New(logg, cfg.Database)
	if err := store.Connect(ctx); err != nil {
		logg.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer func() {
		if err := store.Close(); err != nil {
			logg.Error().Err(err).Msg("failed to close connection to database")
		}
	}()

	reg := registrator.New(logg, store)
	usr := users.New(logg, store)

	http.HandleFunc("/user/register", reg.RegisterNewUser())
	http.HandleFunc("/user/approve", reg.ApproveUsersRequest())
	http.HandleFunc("/user/block", reg.BlockUser())
	http.HandleFunc("/user/list", usr.ListUsers())

	if err := http.ListenAndServe(cfg.Server.Address, nil); err != nil {
		logg.Error().Err(err).Msg("service is stopped")
		return
	}
	logg.Info().Msg("service is stopped")
}

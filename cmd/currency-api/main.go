package main

import (
	"context"
	"flag"
	"github.com/hihoak/currency-api/internal/app/registrator"
	"github.com/hihoak/currency-api/internal/app/users"
	"github.com/hihoak/currency-api/internal/app/walleter"
	"github.com/hihoak/currency-api/internal/clients/exchanger"
	"github.com/hihoak/currency-api/internal/clients/storager"
	"github.com/hihoak/currency-api/internal/pkg/config"
	"github.com/hihoak/currency-api/internal/pkg/logger"
	"github.com/hihoak/currency-api/internal/clients/quoter/mock_quoter"
	"net/http"
	"os/signal"
	"syscall"

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

	quoter := mock_quoter.New()
	// start mock quotes
	quoter.Start()

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	exch := exchanger.New(ctx, logg, quoter)
	// start inner exchanger with bigger time step
	exch.Start()

	reg := registrator.New(logg, store)
	usr := users.New(logg, store)
	wal := walleter.New(logg, store, exch)

	http.HandleFunc("/register", reg.RegisterNewUser())
	http.HandleFunc("/register/approve", reg.ApproveUsersRequest())
	http.HandleFunc("/login", reg.LoginUser())

	http.HandleFunc("/user/block", usr.BlockOrUnblockUser())
	http.HandleFunc("/user/list", usr.ListUsers())
	http.HandleFunc("/user/info", usr.GetUserFullInfo())

	http.HandleFunc("/wallet/get", wal.GetWallet())
	http.HandleFunc("/wallet/list", wal.ListUsersWallets())
	http.HandleFunc("/wallet/create", wal.CreateNewWallet())
	http.HandleFunc("/wallet/money/add", wal.AddMoneyToWallet())
	http.HandleFunc("/wallet/exchange", wal.ExchangeMoney())

	if err := http.ListenAndServe(cfg.Server.Address, nil); err != nil {
		logg.Error().Err(err).Msg("service is stopped")
		return
	}

	<-ctx.Done()
	logg.Info().Msg("service is stopped")
}

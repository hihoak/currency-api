package walleter

import (
	"context"
	"github.com/hihoak/currency-api/internal/pkg/logger"
	"github.com/hihoak/currency-api/internal/pkg/models"
	"net/http"
)

type Storager interface {
	AddMoneyToWallet(ctx context.Context, walletID int64, amount int64) error
	GetMoneyFromWallet(ctx context.Context, walletID int64, amount int64) (int64, error)
	SaveWallet(ctx context.Context, wallet *models.Wallet) (*models.Wallet, error)
}

type Walleter struct {
	logg *logger.Logger

	storage Storager
}

func (w *Walleter) AddMoneyToWallet() func(http.ResponseWriter, *http.Request) {
	w.logg.Info().Msg("registering AddMoneyToWallet handler...")
	return func(writer http.ResponseWriter, request *http.Request) {
		w.logg.Info().Msg("start AddMoneyToWallet handler...")

		w.logg.Info().Msg("end AddMoneyToWallet handler")
	}
}

func (w *Walleter) GetMoneyFromWallet() func(http.ResponseWriter, *http.Request) {
	w.logg.Info().Msg("registering GetMoneyFromWallet handler...")
	return func(writer http.ResponseWriter, request *http.Request) {
		w.logg.Info().Msg("start GetMoneyFromWallet handler...")

		w.logg.Info().Msg("end GetMoneyFromWallet handler")
	}
}

func (w *Walleter) CreateNewWallet() func(http.ResponseWriter, *http.Request) {
	w.logg.Info().Msg("registering CreateNewWallet handler...")
	return func(writer http.ResponseWriter, request *http.Request) {
		w.logg.Info().Msg("start CreateNewWallet handler...")

		w.logg.Info().Msg("end CreateNewWallet handler")
	}
}

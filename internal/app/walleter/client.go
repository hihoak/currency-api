package walleter

import (
	"context"
	"github.com/hihoak/currency-api/internal/pkg/logger"
	"github.com/hihoak/currency-api/internal/pkg/models"
	"net/http"
)

type Storager interface {
	AddMoneyToWallet(ctx context.Context, walletID int64, amount int64) (*models.Wallet, error)
	GetWallet(ctx context.Context, walletID int64) (*models.Wallet, error)
	SaveWallet(ctx context.Context, wallet *models.Wallet) (*models.Wallet, error)
}

type Walleter struct {
	logg *logger.Logger

	storage Storager
}

func (w *Walleter) CreateNewWallet() func(http.ResponseWriter, *http.Request) {
	w.logg.Info().Msg("registering CreateNewWallet handler...")
	return func(writer http.ResponseWriter, request *http.Request) {
		w.logg.Info().Msg("start CreateNewWallet handler...")

		w.logg.Info().Msg("end CreateNewWallet handler")
	}
}

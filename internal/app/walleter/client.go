package walleter

import (
	"context"
	"github.com/hihoak/currency-api/internal/pkg/logger"
	"github.com/hihoak/currency-api/internal/pkg/models"
)

type Storager interface {
	AddMoneyToWallet(ctx context.Context, walletID int64, amount int64) (*models.Wallet, error)
	GetWallet(ctx context.Context, walletID int64) (*models.Wallet, error)
	SaveWalletUnary(ctx context.Context, wallet *models.Wallet) (int64, error)
}

type Walleter struct {
	logg *logger.Logger

	storage Storager
}

func New(logg *logger.Logger, storage Storager) *Walleter {
	return &Walleter{
		logg: logg,
		storage: storage,
	}
}

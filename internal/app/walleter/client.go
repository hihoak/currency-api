package walleter

import (
	"context"
	"github.com/hihoak/currency-api/internal/clients/exchanger"
	"github.com/hihoak/currency-api/internal/pkg/logger"
	"github.com/hihoak/currency-api/internal/pkg/models"
)

type Storager interface {
	AddMoneyToWallet(ctx context.Context, walletID int64, amount int64) (*models.Wallet, error)
	GetWallet(ctx context.Context, walletID int64) (*models.Wallet, error)
	SaveWalletUnary(ctx context.Context, wallet *models.Wallet) (int64, error)
	GetUserWallets(ctx context.Context, userID int64) ([]*models.Wallet, error)
	MoneyExchange(ctx context.Context, userID, fromWalletID, toWalletID int64, amount int64, toAmount int64) (*models.Wallet, *models.Wallet, error)
}

type Exchanger interface {
	GetCourse(from, to models.Currencies) exchanger.CourseInfo
}

type Walleter struct {
	logg *logger.Logger

	storage Storager
	exchange Exchanger
}

func New(logg *logger.Logger, storage Storager, exchange Exchanger) *Walleter {
	return &Walleter{
		logg: logg,
		storage: storage,
		exchange: exchange,
	}
}

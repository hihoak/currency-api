package users

import (
	"context"
	"github.com/hihoak/currency-api/internal/pkg/logger"
	"github.com/hihoak/currency-api/internal/pkg/models"
)

type Storager interface {
	ListUsers(ctx context.Context, count, offset int64) ([]*models.User, error)
	GetUser(ctx context.Context, userID int64) (*models.User, error)
	GetUserWallets(ctx context.Context, userID int64) ([]*models.Wallet, error)
	BlockOrUnblockUser(ctx context.Context, userID int64, block bool) error
}

type Users struct {
	logg *logger.Logger

	storage Storager
}

func New(logg *logger.Logger, storage Storager) *Users {
	return &Users{
		logg: logg,
		storage: storage,
	}
}

package registrator

import (
	"context"
	"github.com/hihoak/currency-api/internal/pkg/logger"
	"github.com/hihoak/currency-api/internal/pkg/models"
)

type Storager interface {
	SaveNewUser(ctx context.Context, user *models.User, wallet *models.Wallet) (int64, error)
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
	ApproveUsersRequest(ctx context.Context, userID int64) error
}

type Registrator struct {
	logg *logger.Logger

	storage Storager
}

func New(logg *logger.Logger, storage Storager) *Registrator {
	return &Registrator{
		logg: logg,
		storage: storage,
	}
}

package registrator

import (
	"context"
	"fmt"
	"github.com/hihoak/currency-api/internal/pkg/logger"
	"github.com/hihoak/currency-api/internal/pkg/models"
	jsoniter "github.com/json-iterator/go"
	"net/http"
)

type Storager interface {
	SaveNewUser(ctx context.Context, user *models.User, wallet *models.Wallet) (int64, error)
	ApproveUsersRequest(ctx context.Context, userID int64) error
	BlockUser(ctx context.Context, userID int64) error
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

type BlockUserRequest struct {
	User *models.User
}

func (r *Registrator) BlockUser() func(http.ResponseWriter, *http.Request) {
	r.logg.Info().Msg("registering BlockUser handler...")
	return func(writer http.ResponseWriter, request *http.Request) {
		r.logg.Info().Msg("start BlockUser handler...")
		dec := jsoniter.NewDecoder(request.Body)
		dec.DisallowUnknownFields()

		requestJSON := &BlockUserRequest{}
		if err := dec.Decode(&requestJSON); err != nil {
			r.logg.Error().Err(err).Msgf("failed to parse json")
			http.Error(writer, fmt.Sprintf("failed to parse json: %v", err), http.StatusBadRequest)
			return
		}

		err := r.storage.BlockUser(context.Background(), requestJSON.User.ID)
		if err != nil {
			r.logg.Error().Err(err).Msgf("failed to block user")
			http.Error(writer, fmt.Sprintf("failed to block user: %v", err), http.StatusBadRequest)
			return
		}
		writer.WriteHeader(http.StatusOK)
		r.logg.Info().Msg("end BlockUser handler")
	}
}

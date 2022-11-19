package users

import (
	"context"
	"fmt"
	"github.com/hihoak/currency-api/internal/pkg/models"
	jsoniter "github.com/json-iterator/go"
	"net/http"
)

type BlockOrUnblockUserRequest struct {
	User *models.User
	Block bool
}

func (u *Users) BlockOrUnblockUser() func(http.ResponseWriter, *http.Request) {
	u.logg.Info().Msg("registering BlockUser handler...")
	return func(writer http.ResponseWriter, request *http.Request) {
		u.logg.Info().Msg("start BlockUser handler...")
		dec := jsoniter.NewDecoder(request.Body)
		dec.DisallowUnknownFields()

		requestJSON := &BlockOrUnblockUserRequest{}
		if err := dec.Decode(&requestJSON); err != nil {
			u.logg.Error().Err(err).Msgf("failed to parse json")
			http.Error(writer, fmt.Sprintf("failed to parse json: %v", err), http.StatusBadRequest)
			return
		}

		err := u.storage.BlockOrUnblockUser(context.Background(), requestJSON.User.ID, requestJSON.Block)
		if err != nil {
			u.logg.Error().Err(err).Msgf("failed to block user")
			http.Error(writer, fmt.Sprintf("failed to block user: %v", err), http.StatusBadRequest)
			return
		}
		writer.WriteHeader(http.StatusOK)
		u.logg.Info().Msg("end BlockUser handler")
	}
}

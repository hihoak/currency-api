package registrator

import (
	"context"
	"fmt"
	"github.com/hihoak/currency-api/internal/pkg/models"
	jsoniter "github.com/json-iterator/go"
	"net/http"
)

type ApproveUsersRequestRequest struct {
	User *models.User
}

func (r *Registrator) ApproveUsersRequest() func(http.ResponseWriter, *http.Request) {
	r.logg.Info().Msg("registering ApproveUsersRequest handler...")
	return func(writer http.ResponseWriter, request *http.Request) {
		r.logg.Info().Msg("start ApproveUsersRequest handler...")

		dec := jsoniter.NewDecoder(request.Body)
		dec.DisallowUnknownFields()
		requestJSON := &ApproveUsersRequestRequest{}
		if err := dec.Decode(&requestJSON); err != nil {
			r.logg.Error().Err(err).Msgf("failed to parse json")
			http.Error(writer, fmt.Sprintf("failed to parse json: %v", err), http.StatusBadRequest)
			return
		}
		r.logg.Debug().Msgf("successfully parse request: %v", requestJSON)
		if err := r.storage.ApproveUsersRequest(context.Background(), requestJSON.User.ID); err != nil {
			r.logg.Error().Err(err).Msgf("failed to approve user")
			http.Error(writer, fmt.Sprintf("failed to approver user: %v", err), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
		r.logg.Info().Msg("end ApproveUsersRequest handler")
	}
}

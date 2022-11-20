package walleter

import (
	"context"
	"errors"
	"fmt"
	"github.com/hihoak/currency-api/internal/pkg/errs"
	jsoniter "github.com/json-iterator/go"
	"net/http"
)

type ListUsersWalletsRequest struct {
	UserID int64 `json:"user_id"`
}

func (w *Walleter) ListUsersWallets() func(http.ResponseWriter, *http.Request) {
	w.logg.Info().Msg("registering ListUsersWallets handler...")
	return func(writer http.ResponseWriter, request *http.Request) {
		w.logg.Info().Msg("start ListUsersWallets handler...")
		dec := jsoniter.NewDecoder(request.Body)
		dec.DisallowUnknownFields()

		requestJSON := &ListUsersWalletsRequest{}
		if err := dec.Decode(&requestJSON); err != nil {
			w.logg.Error().Err(err).Msgf("failed to parse json")
			http.Error(writer, fmt.Sprintf("failed to parse json: %v", err), http.StatusBadRequest)
			return
		}

		wallets, err := w.storage.GetUserWallets(context.Background(), requestJSON.UserID)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				w.logg.Error().Err(err).Msgf("not found wallets by user_id: %d", requestJSON.UserID)
				http.Error(writer, fmt.Sprintf("not found wallets by user_id: %d: %v", requestJSON.UserID, err), http.StatusNotFound)
				return
			}
			w.logg.Error().Err(err).Msgf("failed to get wallets by user id: %d", requestJSON.UserID)
			http.Error(writer, fmt.Sprintf("failed to get wallets by user id: %d: %v", requestJSON.UserID, err), http.StatusInternalServerError)
			return
		}

		respJson, err := jsoniter.Marshal(wallets)
		if err != nil {
			w.logg.Error().Err(err).Msgf("failed to marshall request")
			http.Error(writer, fmt.Sprintf("failed to marshall request: %v", err), http.StatusInternalServerError)
			return
		}
		if _, err := writer.Write(respJson); err != nil {
			w.logg.Error().Err(err).Msgf("failed to write response")
			http.Error(writer, fmt.Sprintf("failed to write response: %v", err), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
		w.logg.Info().Msg("end ListUsersWallets handler")
	}
}


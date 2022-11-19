package users

import (
	"context"
	"errors"
	"fmt"
	"github.com/hihoak/currency-api/internal/pkg/errs"
	"github.com/hihoak/currency-api/internal/pkg/models"
	jsoniter "github.com/json-iterator/go"
	"net/http"
)

type GetUserFullInfoRequest struct {
	ID int64
}

type GetUserFullInfoResponse struct {
	User *models.User
	Wallets []*models.Wallet
}

func (u *Users) GetUserFullInfo() func(http.ResponseWriter, *http.Request) {
	u.logg.Info().Msg("registering GetUserFullInfo handler...")
	return func(writer http.ResponseWriter, request *http.Request) {
		u.logg.Info().Msg("start GetUserFullInfo handler...")
		dec := jsoniter.NewDecoder(request.Body)
		dec.DisallowUnknownFields()

		requestJSON := &GetUserFullInfoRequest{}
		if err := dec.Decode(&requestJSON); err != nil {
			u.logg.Error().Err(err).Msgf("failed to parse json")
			http.Error(writer, fmt.Sprintf("failed to parse json: %v", err), http.StatusBadRequest)
			return
		}

		user, err := u.storage.GetUser(context.Background(), requestJSON.ID)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				u.logg.Error().Err(err).Msgf("not found user by id: %d", requestJSON.ID)
				http.Error(writer, fmt.Sprintf("not found user by id: %d: %v", requestJSON.ID, err), http.StatusNotFound)
				return
			}
			u.logg.Error().Err(err).Msgf("failed to get user by id: %d", requestJSON.ID)
			http.Error(writer, fmt.Sprintf("failed to get user by id: %d: %v", requestJSON.ID, err), http.StatusInternalServerError)
			return
		}
		wallets, err := u.storage.GetUserWallets(context.Background(), requestJSON.ID)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				u.logg.Error().Err(err).Msgf("not found wallets by user_id: %d", requestJSON.ID)
				http.Error(writer, fmt.Sprintf("not found wallets by user_id: %d: %v", requestJSON.ID, err), http.StatusNotFound)
				return
			}
			u.logg.Error().Err(err).Msgf("failed to get wallets by user id: %d", requestJSON.ID)
			http.Error(writer, fmt.Sprintf("failed to get wallets by user id: %d: %v", requestJSON.ID, err), http.StatusInternalServerError)
			return
		}

		respJson, err := jsoniter.Marshal(&GetUserFullInfoResponse{User: user, Wallets: wallets})
		if err != nil {
			u.logg.Error().Err(err).Msgf("failed to marshall request")
			http.Error(writer, fmt.Sprintf("failed to marshall request: %v", err), http.StatusInternalServerError)
			return
		}
		if _, err := writer.Write(respJson); err != nil {
			u.logg.Error().Err(err).Msgf("failed to write response")
			http.Error(writer, fmt.Sprintf("failed to write response: %v", err), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
		u.logg.Info().Msg("end GetUserFullInfo handler")
	}
}


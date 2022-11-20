package walleter

import (
	"context"
	"errors"
	"fmt"
	"github.com/hihoak/currency-api/internal/clients/exchanger"
	"github.com/hihoak/currency-api/internal/pkg/errs"
	"github.com/hihoak/currency-api/internal/pkg/models"
	jsoniter "github.com/json-iterator/go"
	"net/http"
)

type ListUsersWalletsRequest struct {
	UserID int64 `json:"user_id"`
}

type UsersWalletsResponse struct {
	ID int64 `json:"id"`
	UserID int64 `json:"user_id"`
	Currency models.Currencies `json:"currency"`
	Value int64 `json:"value"`
	CourseInfo exchanger.CourseInfo `json:"course_info"`
	Inactive bool `json:"inactive"`
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

		res := make([]*UsersWalletsResponse, len(wallets))
		for idx, wallet := range wallets {
			courseInfo := w.exchange.GetCourse(wallet.Currency, models.RUB)
			res[idx] = &UsersWalletsResponse{
				ID: wallet.ID,
				UserID: wallet.UserID,
				Currency: wallet.Currency,
				Value: wallet.Value,
				CourseInfo: courseInfo,
			}
		}

	main:
		for _, currency := range models.AllSupportedCurrencies {
			for _, r := range res {
				if r.Currency == currency {
					continue main
				}
			}
			res = append(res, &UsersWalletsResponse{
				Currency: currency,
				Inactive: true,
				CourseInfo: w.exchange.GetCourse(currency, models.RUB),
			})
		}


		respJson, err := jsoniter.Marshal(res)
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

func currencyInUserWallets(wallets []*models.Wallet, currency models.Currencies) bool {
	for _, wallet := range wallets {
		if wallet.Currency == currency {
			return true
		}
	}
	return false
}

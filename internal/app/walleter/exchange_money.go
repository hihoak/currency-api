package walleter

import (
	"context"
	"errors"
	"fmt"
	"github.com/hihoak/currency-api/internal/pkg/errs"
	"github.com/hihoak/currency-api/internal/pkg/models"
	jsoniter "github.com/json-iterator/go"
	"math"
	"net/http"
)

type ExchangeMoneyRequest struct {
	UserID int64 `json:"user_id"`
	FromWalletID int64 `json:"from_wallet_id"`
	ToWalletID int64 `json:"to_wallet_id"`
	FromCurrency models.Currencies `json:"from_currency"`
	ToCurrency models.Currencies `json:"to_currency"`
	Amount int64 `json:"amount"`
}

type ExchangeMoneyResponse struct {
	FromWallet *models.Wallet `json:"from_wallet"`
	ToWallet *models.Wallet `json:"to_wallet"`
	Quote float64 `json:"quote"`
	FromAmount int64 `json:"from_amount"`
	ToAmount int64 `json:"to_amount"`
}

func (w *Walleter) ExchangeMoney() func(http.ResponseWriter, *http.Request) {
	w.logg.Info().Msg("registering ExchangeMoney handler...")
	return func(writer http.ResponseWriter, request *http.Request) {
		w.logg.Info().Msg("start ExchangeMoney handler...")
		dec := jsoniter.NewDecoder(request.Body)
		dec.DisallowUnknownFields()

		requestJSON := &ExchangeMoneyRequest{}
		if err := dec.Decode(&requestJSON); err != nil {
			w.logg.Error().Err(err).Msgf("failed to parse json")
			http.Error(writer, fmt.Sprintf("failed to parse json: %v", err), http.StatusBadRequest)
			return
		}

		if requestJSON.Amount <= 0 {
			w.logg.Warn().Msgf("amount can't be equal or less than zero")
			http.Error(writer, fmt.Sprintf("amount can't be equal or less than zero"), http.StatusBadRequest)
			return
		}

		realCourse := w.exchange.GetCourse(requestJSON.FromCurrency, requestJSON.ToCurrency)
		toAmount := int64(math.Floor(float64(requestJSON.Amount) * realCourse.Value))

		fromWallet, toWallet, err := w.storage.MoneyExchange(context.Background(),
			requestJSON.UserID, requestJSON.FromWalletID, requestJSON.ToWalletID, requestJSON.Amount, toAmount, requestJSON.FromCurrency, requestJSON.ToCurrency, realCourse.Value)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				w.logg.Error().Err(err).Msgf("not found wallets by user_id: %d", requestJSON.UserID)
				http.Error(writer, fmt.Sprintf("not found wallets by user_id: %d: %v", requestJSON.UserID, err), http.StatusNotFound)
				return
			}
			if errors.Is(err, errs.ErrNotEnoughMoney) {
				w.logg.Error().Err(err).Msgf("not enough money in wallets for user_id: %d", requestJSON.UserID)
				http.Error(writer, fmt.Sprintf("not enough money in wallets for user_id %d: %v", requestJSON.UserID, err), http.StatusConflict)
				return
			}
			w.logg.Error().Err(err).Msgf("failed to get wallets by user id: %d", requestJSON.UserID)
			http.Error(writer, fmt.Sprintf("failed to get wallets by user id: %d: %v", requestJSON.UserID, err), http.StatusInternalServerError)
			return
		}

		respJson, err := jsoniter.Marshal(&ExchangeMoneyResponse{
			FromWallet: fromWallet,
			ToWallet: toWallet,
			Quote: realCourse.Value,
			FromAmount: requestJSON.Amount,
			ToAmount: toAmount,
		})
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
		w.logg.Info().Msg("end ExchangeMoney handler")
	}
}


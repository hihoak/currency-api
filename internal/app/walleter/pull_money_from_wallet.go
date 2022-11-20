package walleter

import (
	"context"
	"errors"
	"fmt"
	"github.com/hihoak/currency-api/internal/pkg/errs"
	jsoniter "github.com/json-iterator/go"
	"net/http"
)

type PullMoneyFromWalletRequest struct {
	ID int64
	Amount int64
}

func (w *Walleter) PullMoneyFromWallet() func(http.ResponseWriter, *http.Request) {
	w.logg.Info().Msg("registering PullMoneyFromWallet handler...")
	return func(writer http.ResponseWriter, request *http.Request) {
		w.logg.Info().Msg("start PullMoneyFromWallet handler...")

		dec := jsoniter.NewDecoder(request.Body)
		dec.DisallowUnknownFields()

		requestJSON := &PullMoneyFromWalletRequest{}
		if err := dec.Decode(&requestJSON); err != nil {
			w.logg.Error().Err(err).Msgf("failed to parse json")
			http.Error(writer, fmt.Sprintf("failed to parse json: %v", err), http.StatusBadRequest)
			return
		}

		updatedWallet, err := w.storage.PullMoneyFromWallet(context.Background(), requestJSON.ID, requestJSON.Amount)
		if err != nil {
			if errors.Is(err, errs.ErrNotEnoughMoney) {
				w.logg.Warn().Err(err).Msgf("failed to pull money, not enough money")
				http.Error(writer, fmt.Sprintf("failed to pull money, not enough money: %v", err), http.StatusConflict)
				return
			}
			w.logg.Error().Err(err).Msgf("failed to pull money")
			http.Error(writer, fmt.Sprintf("failed to pull money: %v", err), http.StatusInternalServerError)
			return
		}

		responseJSON, err := jsoniter.Marshal(updatedWallet)
		if err != nil {
			w.logg.Error().Err(err).Msgf("failed to parse wallet")
			http.Error(writer, fmt.Sprintf("failed to parse wallet: %v", err), http.StatusInternalServerError)
			return
		}

		if _, err := writer.Write(responseJSON); err != nil {
			w.logg.Error().Err(err).Msgf("failed to write response")
			http.Error(writer, fmt.Sprintf("failed to write response: %v", err), http.StatusInternalServerError)
			return
		}

		writer.WriteHeader(http.StatusOK)
		w.logg.Info().Msg("end PullMoneyFromWallet handler")
	}
}

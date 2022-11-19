package walleter

import (
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"net/http"
)

type AddMoneyToWalletRequest struct {
	ID int64
	Value int64
}

func (w *Walleter) AddMoneyToWallet() func(http.ResponseWriter, *http.Request) {
	w.logg.Info().Msg("registering AddMoneyToWallet handler...")
	return func(writer http.ResponseWriter, request *http.Request) {
		w.logg.Info().Msg("start AddMoneyToWallet handler...")

		dec := jsoniter.NewDecoder(request.Body)
		dec.DisallowUnknownFields()

		requestJSON := &AddMoneyToWalletRequest{}
		if err := dec.Decode(&requestJSON); err != nil {
			w.logg.Error().Err(err).Msgf("failed to parse json")
			http.Error(writer, fmt.Sprintf("failed to parse json: %v", err), http.StatusBadRequest)
			return
		}

		updatedWallet, err := w.storage.AddMoneyToWallet(context.Background(), requestJSON.ID, requestJSON.Value)
		if err != nil {
			w.logg.Error().Err(err).Msgf("failed to add money")
			http.Error(writer, fmt.Sprintf("failed to add money: %v", err), http.StatusInternalServerError)
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
		w.logg.Info().Msg("end AddMoneyToWallet handler")
	}
}

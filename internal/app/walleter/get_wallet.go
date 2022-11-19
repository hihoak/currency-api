package walleter

import (
	"context"
	"errors"
	"fmt"
	"github.com/hihoak/currency-api/internal/pkg/errs"
	jsoniter "github.com/json-iterator/go"
	"net/http"
)

type GetWalletRequest struct {
	ID int64
}

func (w *Walleter) GetWallet() func(http.ResponseWriter, *http.Request) {
	w.logg.Info().Msg("registering GetWallet handler...")
	return func(writer http.ResponseWriter, request *http.Request) {
		w.logg.Info().Msg("start GetWallet handler...")
		dec := jsoniter.NewDecoder(request.Body)
		dec.DisallowUnknownFields()

		requestJSON := &GetWalletRequest{}
		if err := dec.Decode(&requestJSON); err != nil {
			w.logg.Error().Err(err).Msgf("failed to parse json")
			http.Error(writer, fmt.Sprintf("failed to parse json: %v", err), http.StatusBadRequest)
			return
		}

		wallet, err := w.storage.GetWallet(context.Background(), requestJSON.ID)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				w.logg.Error().Err(err).Msgf("not found wallet with id '%d'", requestJSON.ID)
				http.Error(writer, fmt.Sprintf("not found wallet with id '%d': %v", requestJSON.ID, err), http.StatusNotFound)
				return
			}
			w.logg.Error().Err(err).Msgf("failed to get wallet")
			http.Error(writer, fmt.Sprintf("failed to get wallet: %v", err), http.StatusInternalServerError)
			return
		}

		responseJSON, err := jsoniter.Marshal(wallet)
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
		w.logg.Info().Msg("end GetWallet handler")
	}
}

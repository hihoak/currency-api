package walleter

import (
	"context"
	"fmt"
	"github.com/hihoak/currency-api/internal/pkg/models"
	jsoniter "github.com/json-iterator/go"
	"net/http"
)

type CreateNewWalletRequest struct {
	Wallet *models.Wallet
}

type CreateNewWalletResponse struct {
	ID int64
}

func (w *Walleter) CreateNewWallet() func(http.ResponseWriter, *http.Request) {
	w.logg.Info().Msg("registering CreateNewWallet handler...")
	return func(writer http.ResponseWriter, request *http.Request) {
		w.logg.Info().Msg("start CreateNewWallet handler...")
		dec := jsoniter.NewDecoder(request.Body)
		dec.DisallowUnknownFields()

		requestJSON := &CreateNewWalletRequest{}
		if err := dec.Decode(&requestJSON); err != nil {
			w.logg.Error().Err(err).Msgf("failed to parse json")
			http.Error(writer, fmt.Sprintf("failed to parse json: %v", err), http.StatusBadRequest)
			return
		}

		id, err := w.storage.SaveWalletUnary(context.Background(), requestJSON.Wallet)
		if err != nil {
			w.logg.Error().Err(err).Msgf("failed to create wallet")
			http.Error(writer, fmt.Sprintf("failed to create wallet: %v", err), http.StatusInternalServerError)
			return
		}

		respJson, err := jsoniter.Marshal(&CreateNewWalletResponse{ID: id})
		if err != nil {
			w.logg.Error().Err(err).Msgf("failed to marshall user")
			http.Error(writer, fmt.Sprintf("failed to marshall user: %v", err), http.StatusInternalServerError)
			return
		}
		if _, err := writer.Write(respJson); err != nil {
			w.logg.Error().Err(err).Msgf("failed to write response")
			http.Error(writer, fmt.Sprintf("failed to write response: %v", err), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
		w.logg.Info().Msg("end CreateNewWallet handler")
	}
}

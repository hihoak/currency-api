package walleter

import (
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"net/http"
)

type ListTransactionsRequest struct {
	UserID int64 `json:"user_id"`
}

func (w *Walleter) ListTransactions() func(http.ResponseWriter, *http.Request) {
	w.logg.Info().Msg("registering ListTransactions handler...")
	return func(writer http.ResponseWriter, request *http.Request) {
		w.logg.Info().Msg("start ListTransactions handler...")
		dec := jsoniter.NewDecoder(request.Body)
		dec.DisallowUnknownFields()

		requestJSON := &ListTransactionsRequest{}
		if err := dec.Decode(&requestJSON); err != nil {
			w.logg.Error().Err(err).Msgf("failed to parse json")
			http.Error(writer, fmt.Sprintf("failed to parse json: %v", err), http.StatusBadRequest)
			return
		}

		transactions, err := w.storage.ListTransactions(context.Background(), requestJSON.UserID)
		if err != nil {
			w.logg.Error().Err(err).Msgf("failed to get transactions")
			http.Error(writer, fmt.Sprintf("failed to get transactions: %v", err), http.StatusInternalServerError)
			return
		}

		w.logg.Debug().Msgf("got %d transactions", len(transactions))

		responseJSON, err := jsoniter.Marshal(transactions)
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
		w.logg.Info().Msg("end ListTransactions handler")
	}
}

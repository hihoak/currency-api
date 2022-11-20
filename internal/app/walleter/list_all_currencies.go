package walleter

import (
	"fmt"
	"github.com/hihoak/currency-api/internal/pkg/models"
	jsoniter "github.com/json-iterator/go"
	"net/http"
)

func (w *Walleter) ListCurrencies() func(http.ResponseWriter, *http.Request) {
	w.logg.Info().Msg("registering ListCurrencies handler...")
	return func(writer http.ResponseWriter, request *http.Request) {
		w.logg.Info().Msg("start ListCurrencies handler...")
		dec := jsoniter.NewDecoder(request.Body)
		dec.DisallowUnknownFields()

		responseJSON, err := jsoniter.Marshal(models.AllSupportedCurrencies)
		if err != nil {
			w.logg.Error().Err(err).Msgf("failed to parse AllSupportedCurrencies")
			http.Error(writer, fmt.Sprintf("failed to parse wallet: %v", err), http.StatusInternalServerError)
			return
		}

		if _, err := writer.Write(responseJSON); err != nil {
			w.logg.Error().Err(err).Msgf("failed to write response")
			http.Error(writer, fmt.Sprintf("failed to write response: %v", err), http.StatusInternalServerError)
			return
		}

		writer.WriteHeader(http.StatusOK)
		w.logg.Info().Msg("end ListCurrencies handler")
	}
}

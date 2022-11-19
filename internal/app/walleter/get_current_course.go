package walleter

import (
	"fmt"
	"github.com/hihoak/currency-api/internal/pkg/models"
	jsoniter "github.com/json-iterator/go"
	"net/http"
)

type GetCourseRequest struct {
	From models.Currencies `json:"from"`
	To models.Currencies `json:"to"`
}

type GetCourseResponse struct {
	From models.Currencies `json:"from"`
	To models.Currencies `json:"to"`
	Course float64 `json:"course"`
}

func (w *Walleter) GetCourse() func(http.ResponseWriter, *http.Request) {
	w.logg.Info().Msg("registering GetCourse handler...")
	return func(writer http.ResponseWriter, request *http.Request) {
		w.logg.Info().Msg("start GetCourse handler...")
		dec := jsoniter.NewDecoder(request.Body)
		dec.DisallowUnknownFields()

		requestJSON := &GetCourseRequest{}
		if err := dec.Decode(&requestJSON); err != nil {
			w.logg.Error().Err(err).Msgf("failed to parse json")
			http.Error(writer, fmt.Sprintf("failed to parse json: %v", err), http.StatusBadRequest)
			return
		}

		response := &GetCourseResponse{
			From: requestJSON.From,
			To: requestJSON.To,
			Course: w.exchange.GetCourse(requestJSON.From, requestJSON.To),
		}
		responseJSON, err := jsoniter.Marshal(&response)
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
		w.logg.Info().Msg("end GetCourse handler")
	}
}

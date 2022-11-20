package timeliner

import (
	"context"
	"errors"
	"fmt"
	"github.com/hihoak/currency-api/internal/pkg/errs"
	"github.com/hihoak/currency-api/internal/pkg/logger"
	"github.com/hihoak/currency-api/internal/pkg/models"
	jsoniter "github.com/json-iterator/go"
	"net/http"
	"time"
)

type Storager interface {
	ListCourses(ctx context.Context, fromCurrency, toCurrency models.Currencies, fromTime int64, toTime int64) ([]*models.Course, error)
}

type Timeline struct {
	storage Storager
	logg *logger.Logger
}

func New(logg *logger.Logger, storage Storager) *Timeline {
	return &Timeline{
		logg: logg,
		storage: storage,
	}
}

type ListCoursesRequest struct {
	From models.Currencies `json:"from"`
	To models.Currencies `json:"to"`
	FromTime int64 `json:"from_time"`
	ToTime int64 `json:"to_time"`
}

func (t *Timeline) ListCourses() func(http.ResponseWriter, *http.Request) {
	t.logg.Info().Msg("registering ListCourses handler...")
	return func(writer http.ResponseWriter, request *http.Request) {
		t.logg.Info().Msg("start ListCourses handler...")
		dec := jsoniter.NewDecoder(request.Body)
		dec.DisallowUnknownFields()

		requestJSON := &ListCoursesRequest{}
		if err := dec.Decode(&requestJSON); err != nil {
			t.logg.Error().Err(err).Msgf("failed to parse json")
			http.Error(writer, fmt.Sprintf("failed to parse json: %v", err), http.StatusBadRequest)
			return
		}

		fromTime := time.Unix(requestJSON.FromTime, 0)
		toTime := time.Unix(requestJSON.ToTime, 0)

		courses, err := t.storage.ListCourses(context.Background(), requestJSON.From, requestJSON.To, requestJSON.FromTime, requestJSON.ToTime)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				t.logg.Error().Err(err).Msgf("not found courses %s to %s from %s to %s", requestJSON.From, requestJSON.To, fromTime, toTime)
				http.Error(writer, fmt.Sprintf("not found courses %s to %s from %s to %s: %w", requestJSON.From, requestJSON.To, fromTime, toTime, err), http.StatusNotFound)
				return
			}
			t.logg.Error().Err(err).Msgf("failed to list courses %s to %s from %s to %s", requestJSON.From, requestJSON.To, fromTime, toTime)
			http.Error(writer, fmt.Sprintf("failed to list courses %s to %s from %s to %s: %w", requestJSON.From, requestJSON.To, fromTime, toTime, err), http.StatusInternalServerError)
			return
		}

		respJson, err := jsoniter.Marshal(courses)
		if err != nil {
			t.logg.Error().Err(err).Msgf("failed to marshall request")
			http.Error(writer, fmt.Sprintf("failed to marshall request: %v", err), http.StatusInternalServerError)
			return
		}
		if _, err := writer.Write(respJson); err != nil {
			t.logg.Error().Err(err).Msgf("failed to write response")
			http.Error(writer, fmt.Sprintf("failed to write response: %v", err), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
		t.logg.Info().Msg("end ListCourses handler")
	}
}

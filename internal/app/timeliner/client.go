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
	ListCourses(ctx context.Context, fromCurrency, toCurrency models.Currencies, fromTime time.Time, toTime time.Time) ([]*models.Course, error)
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
	FromTime string `json:"from_time"`
	ToTime string `json:"to_time"`
}

type ListCoursesResponse struct {
	Courses []*models.Course
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

		fromTime, err := time.Parse("2006-01-02 15:04:05.999-07", requestJSON.FromTime)
		if err != nil {
			t.logg.Error().Err(err).Msgf("wrong format of date. Example: 2006-01-02 15:04:05.999-07")
			http.Error(writer, fmt.Sprintf("failed to parse time: example: 2006-01-02 15:04:05.999-07: %v", err), http.StatusBadRequest)
			return
		}
		toTime, err := time.Parse("2006-01-02 15:04:05.999-07", requestJSON.ToTime)
		if err != nil {
			t.logg.Error().Err(err).Msgf("wrong format of date. Example: 2006-01-02 15:04:05.999-07")
			http.Error(writer, fmt.Sprintf("failed to parse time: example: 2006-01-02 15:04:05.999-07: %v", err), http.StatusBadRequest)
			return
		}

		courses, err := t.storage.ListCourses(context.Background(), requestJSON.From, requestJSON.To, fromTime, toTime)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				t.logg.Error().Err(err).Msgf("not found courses %s to %s from %s to %s", requestJSON.From, requestJSON.To, requestJSON.FromTime, requestJSON.ToTime)
				http.Error(writer, fmt.Sprintf("not found courses %s to %s from %s to %s", requestJSON.From, requestJSON.To, requestJSON.FromTime, requestJSON.ToTime), http.StatusNotFound)
				return
			}
			t.logg.Error().Err(err).Msgf("failed to list courses %s to %s from %s to %s", requestJSON.From, requestJSON.To, requestJSON.FromTime, requestJSON.ToTime)
			http.Error(writer, fmt.Sprintf("failed to list courses %s to %s from %s to %s", requestJSON.From, requestJSON.To, requestJSON.FromTime, requestJSON.ToTime), http.StatusInternalServerError)
			return
		}

		respJson, err := jsoniter.Marshal(&ListCoursesResponse{Courses: courses})
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

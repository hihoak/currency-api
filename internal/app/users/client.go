package users

import (
	"context"
	"fmt"
	"github.com/hihoak/currency-api/internal/pkg/logger"
	"github.com/hihoak/currency-api/internal/pkg/models"
	jsoniter "github.com/json-iterator/go"
	"net/http"
)

type Storager interface {
	ListUsers(ctx context.Context, count, offset int64) ([]*models.User, error)
}

type Users struct {
	logg *logger.Logger

	storage Storager
}

func New(logg *logger.Logger, storage Storager) *Users {
	return &Users{
		logg: logg,
		storage: storage,
	}
}

type ListUserRequest struct {
	Offset int64
	Count int64
}

type ListUserResponse struct {
	Users []*models.User
}

func (u *Users) ListUsers() func(http.ResponseWriter, *http.Request) {
	u.logg.Info().Msg("registering ListUsers handler...")
	return func(writer http.ResponseWriter, request *http.Request) {
		u.logg.Info().Msg("start ListUsers handler...")
		dec := jsoniter.NewDecoder(request.Body)
		dec.DisallowUnknownFields()

		requestJSON := &ListUserRequest{}
		if err := dec.Decode(&requestJSON); err != nil {
			u.logg.Error().Err(err).Msgf("failed to parse json")
			http.Error(writer, fmt.Sprintf("failed to parse json: %v", err), http.StatusBadRequest)
			return
		}

		users, err := u.storage.ListUsers(context.Background(), requestJSON.Count, requestJSON.Offset)
		if err != nil {
			u.logg.Error().Err(err).Msgf("failed to ListUsers")
			http.Error(writer, fmt.Sprintf("failed to ListUsers: %v", err), http.StatusInternalServerError)
			return
		}

		respJson, err := jsoniter.Marshal(&ListUserResponse{Users: users})
		if err != nil {
			u.logg.Error().Err(err).Msgf("failed to marshall user")
			http.Error(writer, fmt.Sprintf("failed to marshall user: %v", err), http.StatusInternalServerError)
			return
		}
		if _, err := writer.Write(respJson); err != nil {
			u.logg.Error().Err(err).Msgf("failed to write response")
			http.Error(writer, fmt.Sprintf("failed to write response: %v", err), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
		u.logg.Info().Msg("end ListUsers handler")
	}
}

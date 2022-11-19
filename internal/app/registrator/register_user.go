package registrator

import (
	"context"
	"errors"
	"fmt"
	"github.com/hihoak/currency-api/internal/pkg/errs"
	"github.com/hihoak/currency-api/internal/pkg/models"
	jsoniter "github.com/json-iterator/go"
	"net/http"
)

type RegisterUserRequest struct {
	User *models.User

	Wallet *models.Wallet
}

type RegisterUserResponse struct {
	ID int64
}

func (r *Registrator) RegisterNewUser() func(http.ResponseWriter, *http.Request) {
	r.logg.Info().Msg("registering RegisterUser handler...")
	return func(writer http.ResponseWriter, request *http.Request) {
		r.logg.Info().Msg("start RegisterUser handler...")
		dec := jsoniter.NewDecoder(request.Body)
		dec.DisallowUnknownFields()

		requestJSON := &RegisterUserRequest{}
		if err := dec.Decode(&requestJSON); err != nil {
			r.logg.Error().Err(err).Msgf("failed to parse json")
			http.Error(writer, fmt.Sprintf("failed to parse json: %v", err), http.StatusBadRequest)
			return
		}
		r.logg.Debug().Msgf("successfully parse request: %v", requestJSON)
		user := requestJSON.User
		wallet := requestJSON.Wallet
		id, err := r.storage.SaveNewUser(context.TODO(), user, wallet)
		if err != nil {
			if errors.Is(err, errs.ErrUserAlreadyExists) {
				r.logg.Warn().Err(err).Msgf("failed to add user '%v'", *user)
				http.Error(writer, fmt.Sprintf("user already exists: %v", err), http.StatusConflict)
				return
			}
			r.logg.Error().Err(err).Msgf("failed to add user: %v", *user)
			http.Error(writer, fmt.Sprintf("failed to create user: %v", err), http.StatusInternalServerError)
			return
		}

		response := &RegisterUserResponse{
			ID: id,
		}
		responseJSON, errMarshall := jsoniter.Marshal(response)
		if errMarshall != nil {
			r.logg.Error().Err(errMarshall).Msgf("failed to marshall: %v", response)
			http.Error(writer, "failed to marshall response, but user was saved successfully", http.StatusNotExtended)
			return
		}
		_, errWrite := writer.Write(responseJSON)
		if errWrite != nil {
			r.logg.Error().Err(errWrite).Msgf("failed to send response: %v", responseJSON)
			http.Error(writer, "failed to send response, but user was saved successfully", http.StatusNotExtended)
			return
		}
		writer.WriteHeader(http.StatusOK)
		r.logg.Info().Msg("end RegisterUser handler")
	}
}

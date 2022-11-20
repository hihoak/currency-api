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

type LoginUserRequest struct {
	PhoneNumber string `json:"phone_number"`
	Email string `json:"email"`
	Password string `json:"password"`
}

type LoginUserResponse struct {
	ID          int64  `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	MiddleName  string `json:"middle_name" db:"middle_name"`
	Surname     string `json:"surname" db:"surname"`
	Mail        string `json:"mail" db:"mail"`
	PhoneNumber string `json:"phone_number" db:"phone_number"`
	Blocked     bool   `json:"blocked" db:"blocked"`
	Registered  bool   `json:"registered" db:"registered"`
	Admin       bool   `json:"admin" db:"admin"`
	Password    string `json:"password" db:"password"`
	MainWalletID int64 `json:"main_wallet_id"`
}

func (r *Registrator) LoginUser() func(http.ResponseWriter, *http.Request) {
	r.logg.Info().Msg("registering LoginUser handler...")
	return func(writer http.ResponseWriter, request *http.Request) {
		r.logg.Info().Msg("start LoginUser handler...")
		dec := jsoniter.NewDecoder(request.Body)
		dec.DisallowUnknownFields()

		requestJSON := &LoginUserRequest{}
		if err := dec.Decode(&requestJSON); err != nil {
			r.logg.Error().Err(err).Msgf("failed to parse json")
			http.Error(writer, fmt.Sprintf("failed to parse json: %v", err), http.StatusBadRequest)
			return
		}

		user, err := r.storage.GetUserByPhoneNumberOrEmail(context.Background(), requestJSON.PhoneNumber, requestJSON.Email)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				r.logg.Error().Err(err).Msgf("user not found")
				http.Error(writer, fmt.Sprintf("user not found: %v", err), http.StatusUnauthorized)
				return
			}
			r.logg.Error().Err(err).Msgf("failed to get user")
			http.Error(writer, fmt.Sprintf("failed to get user: %v", err), http.StatusInternalServerError)
			return
		}

		if user.Password != requestJSON.Password {
			r.logg.Error().Err(err).Msgf("failed user password doesn't match")
			http.Error(writer, fmt.Sprintf("failed user password doesn't match: %v", err), http.StatusUnauthorized)
			return
		}

		wallets, err := r.storage.GetUserWallets(context.Background(), user.ID)
		if err != nil {
			r.logg.Error().Err(err).Msgf("failed to get wallets user")
			http.Error(writer, fmt.Sprintf("failed to get wallets user: %v", err), http.StatusInternalServerError)
			return
		}

		resp := &LoginUserResponse{
			ID: user.ID,
			Name: user.Name,
			Surname: user.Surname,
			MiddleName: user.MiddleName,
			Mail: user.Mail,
			PhoneNumber: user.PhoneNumber,
			Blocked: user.Blocked,
			Registered: user.Registered,
			Admin: user.Admin,
			Password: user.Password,
		}
		for _, wallet := range wallets {
			if wallet.Currency == models.RUB {
				resp.MainWalletID = wallet.ID
				break
			}
		}

		responseJSON, err := jsoniter.Marshal(resp)
		if err != nil {
			r.logg.Error().Err(err).Msgf("failed to parse user")
			http.Error(writer, fmt.Sprintf("failed to parse user: %v", err), http.StatusInternalServerError)
			return
		}

		if _, err := writer.Write(responseJSON); err != nil {
			r.logg.Error().Err(err).Msgf("failed to write response")
			http.Error(writer, fmt.Sprintf("failed to write response: %v", err), http.StatusInternalServerError)
			return
		}

		writer.WriteHeader(http.StatusOK)
		r.logg.Info().Msg("end LoginUser handler")
	}
}


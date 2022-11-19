package storager

import (
	"github.com/hihoak/currency-api/internal/pkg/models"
	"github.com/jmoiron/sqlx"
)

func (s *Storage) fromSQLRowsToUsers(rows *sqlx.Rows) ([]*models.User, error) {
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			s.log.Error().Err(closeErr).Msg("Failed to close rows")
		}
	}()
	users := make([]*models.User, 0)
	for rows.Next() {
		var user models.User
		if scanErr := rows.StructScan(&user); scanErr != nil {
			return nil, scanErr
		}
		users = append(users, &user)
	}
	return users, nil
}

func (s *Storage) fromSQLRowsToWallets(rows *sqlx.Rows) ([]*models.Wallet, error) {
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			s.log.Error().Err(closeErr).Msg("Failed to close rows")
		}
	}()
	wallets := make([]*models.Wallet, 0)
	for rows.Next() {
		var wallet models.Wallet
		if scanErr := rows.StructScan(&wallet); scanErr != nil {
			return nil, scanErr
		}
		wallets = append(wallets, &wallet)
	}
	return wallets, nil
}

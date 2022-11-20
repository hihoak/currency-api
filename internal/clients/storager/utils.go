package storager

import (
	"github.com/hihoak/currency-api/internal/pkg/models"
	"github.com/jmoiron/sqlx"
	"time"
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

func (s *Storage) fromSQLRowsToTransactions(rows *sqlx.Rows) ([]*models.Transaction, error) {
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			s.log.Error().Err(closeErr).Msg("Failed to close rows")
		}
	}()
	transactions := make([]*models.Transaction, 0)
	for rows.Next() {
		var transaction models.Transaction
		if scanErr := rows.StructScan(&transaction); scanErr != nil {
			return nil, scanErr
		}
		transactions = append(transactions, &transaction)
	}
	return transactions, nil
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

func (s *Storage) fromSQLRowsToCourses(rows *sqlx.Rows) ([]*models.Course, error) {
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			s.log.Error().Err(closeErr).Msg("Failed to close rows")
		}
	}()
	courses := make([]*models.Course, 0)
	for rows.Next() {
		var course models.Course
		if scanErr := rows.StructScan(&course); scanErr != nil {
			return nil, scanErr
		}
		courses = append(courses, &course)
	}
	return courses, nil
}

func (s *Storage) timeToSQLTimeWithTimezone(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.999-07")
}

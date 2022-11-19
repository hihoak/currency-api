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
		var event models.User
		if scanErr := rows.StructScan(&event); scanErr != nil {
			return nil, scanErr
		}
		users = append(users, &event)
	}
	return users, nil
}

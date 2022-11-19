package storager

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/hihoak/currency-api/internal/pkg/config"
	"github.com/hihoak/currency-api/internal/pkg/errs"
	"github.com/hihoak/currency-api/internal/pkg/logger"
	"github.com/hihoak/currency-api/internal/pkg/models"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"time"
)

type Storage struct {
	host     string
	port     string
	user     string
	password string
	dbname   string

	connectionTimeout time.Duration
	operationTimeout  time.Duration

	log *logger.Logger

	db *sqlx.DB
}

func New(log *logger.Logger, storageSection config.DatabaseSection) *Storage {
	return &Storage{
		host:              storageSection.Host,
		port:              storageSection.Port,
		user:              storageSection.User,
		password:          storageSection.Password,
		dbname:            storageSection.DBName,
		connectionTimeout: storageSection.ConnectionTimeout,
		operationTimeout:  storageSection.OperationTimeout,

		log: log,
	}
}

func (s *Storage) Connect(ctx context.Context) error {
	s.log.Info().Msgf("Start connection to database %s:%s with timeout %v", s.host, s.port, s.connectionTimeout)
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	db, err := sqlx.ConnectContext(ctx, "postgres",
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable search_path=public",
			s.host, s.port, s.user, s.password, s.dbname))
	if err != nil {
		return fmt.Errorf("%s: %w", err.Error(), errs.ErrConnectionFailed)
	}
	err = db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", err.Error(), errs.ErrPingFailed)
	}
	s.db = db
	s.log.Info().Msg("Successfully connected to database")
	return nil
}

func (s *Storage) Close() error {
	s.log.Info().Msg("Start closing connection to database...")
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("%s: %w", err.Error(), errs.ErrCloseConnectionFailed)
	}
	s.log.Info().Msg("Successfully close connection to database")
	return s.db.Close()
}

func (s *Storage) SaveNewUser(ctx context.Context, user *models.User, wallet *models.Wallet) (int64, error) {
	s.log.Debug().Msgf("storage: start saving user: %s", user.Username)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin context: %w", err)
	}
	defer func() {
		if err != nil {
			err := tx.Rollback()
			s.log.Error().Err(err).Msg("failed to rollback transaction")
		}
	}()
	query := `
	INSERT INTO users (name, middle_name, surname, mail, phone_number, blocked, registered, admin, username, password)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING id`
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	res, err := tx.QueryContext(ctx, query, user.Name, user.MiddleName, user.Surname, user.Mail, user.PhoneNumber, user.Blocked, user.Registered, user.Admin, user.Username, user.Password)
	if err != nil {
		var dbErr *pq.Error
		if errors.As(err, &dbErr) {
			if dbErr.Code == "23505" {
				return 0, fmt.Errorf("%s: %w", dbErr, errs.ErrUserAlreadyExists)
			}
		}
		return 0, fmt.Errorf("failed to save user: %w", err)
	}

	var userID int64
	for res.Next() {
		err = res.Scan(&userID)
		if err != nil {
			s.log.Error().Err(err).Msg("failed to get returning id")
			return 0, err
		}
	}

	wallet.UserID = userID
	err = s.SaveWallet(ctx, tx, wallet)
	if err != nil {
		return 0, fmt.Errorf("failed")
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	s.log.Debug().Msgf("storage: user saved successfully with id: %d", userID)
	return userID, nil
}

func (s *Storage) SaveWallet(ctx context.Context, tx *sql.Tx, wallet *models.Wallet) error {
	s.log.Debug().Msgf("storage: start saving wallet")
	query := `
	INSERT INTO wallets (user_id, currency, value)
	VALUES ($1, $2, $3)
	RETURNING id`
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	_, err := tx.ExecContext(ctx, query, wallet.UserID, wallet.Currency, wallet.Value)
	if err != nil {
		return fmt.Errorf("failed to save wallet: %w", err)
	}
	s.log.Debug().Msgf("storage: wallet saved successfully")
	return nil
}

func (s *Storage) ApproveUsersRequest(ctx context.Context, userID int64) error {
	q := `
	UPDATE users
	SET registered = true
	WHERE id = $1;`
	_, err := s.db.QueryxContext(ctx, q, userID)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) BlockUser(ctx context.Context, userID int64) error {
	q := `
	UPDATE users
	SET blocked = true
	WHERE id = $1;`
	_, err := s.db.QueryxContext(ctx, q, true, userID)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) ListUsers(ctx context.Context, count, offset int64) ([]*models.User, error) {
	s.log.Debug().Msg("Start listing users")
	query := `
	SELECT *
	FROM users
	OFFSET $1
	LIMIT $2`
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	rows, err := s.db.QueryxContext(ctx, query, offset, count)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	events, scanErr := s.fromSQLRowsToUsers(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan users: %w", scanErr)
	}
	s.log.Debug().Msgf("Successfully list events")
	return events, nil
}


//func (s *Storage) AddEvent(ctx context.Context, title string) error {
//	query := `
//		INSERT INTO events (id, title)
//        VALUES (:id, :title)`
//	event := &storage.Event{
//		Title: title,
//	}
//	event.ID = xid.New().String()
//	s.log.Debug().Msgf("Start adding event with id %s", event.ID)
//	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
//	defer cancel()
//	_, err := s.db.NamedExecContext(ctx, query, event)
//	if err != nil {
//		return fmt.Errorf("%s: %w", err.Error(), errs.ErrAddEvent)
//	}
//	s.log.Debug().Msgf("Successfully add event with id %s", event.ID)
//	return err
//}
//
//func (s *Storage) ModifyEvent(ctx context.Context, event *storage.Event) error {
//	s.log.Debug().Msgf("Start editing event with id %s", event.ID)
//	query := `
//	UPDATE events
//	SET title = :title
//	WHERE id = :id;`
//	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
//	defer cancel()
//	_, err := s.db.NamedExecContext(ctx, query, event)
//	if err != nil {
//		return fmt.Errorf("%s: %w", err.Error(), errs.ErrUpdateEvent)
//	}
//	s.log.Debug().Msgf("Successfully update event with id %s", event.ID)
//	return nil
//}
//
//func (s *Storage) DeleteEvent(ctx context.Context, id string) error {
//	s.log.Debug().Msgf("Start deleting event with id %s", id)
//	query := `DELETE FROM events WHERE id=:id;`
//	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
//	defer cancel()
//	_, err := s.db.NamedExecContext(ctx, query, map[string]interface{}{"id": id})
//	if err != nil {
//		return fmt.Errorf("%s: %w", err.Error(), errs.ErrDeleteEvent)
//	}
//	s.log.Debug().Msgf("Successfully deleted event with id %s", id)
//	return nil
//}
//
//func (s *Storage) GetEvent(ctx context.Context, id string) (*storage.Event, error) {
//	s.log.Debug().Msgf("Start getting event with id %s", id)
//	query := `
//	SELECT id, title
//	FROM events
//	WHERE id=$1;
//	`
//	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
//	defer cancel()
//	row := s.db.QueryRowxContext(ctx, query, id)
//	var event storage.Event
//	if err := row.StructScan(&event); err != nil {
//		if errors.Is(err, sql.ErrNoRows) {
//			return nil, fmt.Errorf("%s: %w", err.Error(), errs.ErrNotFoundEvent)
//		}
//		return nil, fmt.Errorf("%s: %w", err.Error(), errs.ErrGetEvent)
//	}
//	s.log.Debug().Msgf("Successfully got event with id %s", id)
//	return &event, nil
//}
//
//func (s *Storage) ListEvents(ctx context.Context) ([]*storage.Event, error) {
//	s.log.Debug().Msg("Start listing events")
//	query := `
//	SELECT id, title
//	FROM events;
//	`
//	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
//	defer cancel()
//	rows, err := s.db.QueryxContext(ctx, query)
//	if err != nil {
//		return nil, fmt.Errorf("%s: %w", err.Error(), errs.ErrListEvents)
//	}
//	events, scanErr := s.fromSQLRowsToEvents(rows)
//	if err != nil {
//		return nil, fmt.Errorf("%s: %w", scanErr.Error(), errs.ErrListEvents)
//	}
//	s.log.Debug().Msgf("Successfully list events")
//	return events, nil
//}
//
//func (s *Storage) ListEventsToNotify(ctx context.Context,
//	fromTime time.Time, period time.Duration,
//) ([]*storage.Event, error) {
//	s.log.Debug().Msg("ListEventsToNotify - start method")
//	query := strings.Builder{}
//	query.WriteString(`
//	SELECT id, title, start_date, user_id
//	FROM events `)
//	sqlFromTimeStr := s.timeToSQLTimeWithTimezone(fromTime)
//	sqlToTimeStr := s.timeToSQLTimeWithTimezone(fromTime.Add(period))
//	query.WriteString(
//		fmt.Sprintf("WHERE notify_date >= '%s' AND notify_date <= '%s';",
//			sqlFromTimeStr, sqlToTimeStr))
//	s.log.Debug().Msgf("ListEventsToNotify - try to execute query: '%s'", query.String())
//
//	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
//	defer cancel()
//	rows, err := s.db.QueryxContext(ctx, query.String())
//	if err != nil {
//		return nil, fmt.Errorf("%s: %w", err.Error(), errs.ErrListEventsToNotify)
//	}
//	events, err := s.fromSQLRowsToEvents(rows)
//	if err != nil {
//		return nil, fmt.Errorf("ListEventsToNotify - failed to scan events from rows: %w", err)
//	}
//	s.log.Debug().Msgf("ListEventsToNotify: got '%d' events to notify", len(events))
//	return events, nil
//}
//
//func (s *Storage) DeleteOldEventsBeforeTime(
//	ctx context.Context,
//	fromTime time.Time,
//	maxLiveTime time.Duration,
//) ([]*storage.Event, error) {
//	s.log.Debug().Msg("DeleteOldEventsBeforeTime: start deleting old events")
//	query := strings.Builder{}
//	query.WriteString("DELETE FROM events ")
//	query.WriteString(fmt.Sprintf("WHERE '%s' - end_date > '%s' RETURNING *;",
//		s.timeToSQLTimeWithTimezone(fromTime),
//		s.durationToSQLInterval(maxLiveTime)))
//	s.log.Debug().Msgf("DeleteOldEventsBeforeTime: deleting with query: %s", query.String())
//	rows, err := s.db.QueryxContext(ctx, query.String())
//	if err != nil {
//		return nil, fmt.Errorf("failed to delete old events with query: %s: %w", query.String(), err)
//	}
//	events, err := s.fromSQLRowsToEvents(rows)
//	if err != nil {
//		return nil, fmt.Errorf("ListEventsToNotify - failed to scan events from rows: %w", err)
//	}
//	s.log.Debug().Msgf("DeleteOldEventsBeforeTime: delete '%d' old events", len(events))
//	return events, nil
//}

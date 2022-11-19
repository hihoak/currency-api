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
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
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

func (s *Storage) BlockOrUnblockUser(ctx context.Context, userID int64, block bool) error {
	q := `
	UPDATE users
	SET blocked = $2
	WHERE id = $1;`
	_, err := s.db.QueryxContext(ctx, q, true, userID, block)
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

func (s *Storage) GetUserByLogin(ctx context.Context, username string) (*models.User, error) {
	s.log.Debug().Msg("Start listing users")
	query := `
	SELECT *
	FROM users
	WHERE username = $1`
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	rows, err := s.db.QueryxContext(ctx, query, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	events, scanErr := s.fromSQLRowsToUsers(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan users: %w", scanErr)
	}
	if len(events) == 0 {
		return nil, fmt.Errorf("user with username %s not found: %w", username, errs.ErrNotFound)
	}
	s.log.Debug().Msgf("Successfully get user")
	return events[0], nil
}

func (s *Storage) GetWallet(ctx context.Context, walletID int64) (*models.Wallet, error) {
	s.log.Debug().Msg("Start getting wallet")
	query := `
	SELECT *
	FROM wallets
	WHERE id = $1`
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	rows, err := s.db.QueryxContext(ctx, query, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}
	wallets, err := s.fromSQLRowsToWallets(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan wallets: %w", err)
	}
	if len(wallets) == 0 {
		return nil, fmt.Errorf("wallet with id %d: %w", walletID, errs.ErrNotFound)
	}
	s.log.Debug().Msgf("Successfully get wallet")
	return wallets[0], nil
}

func (s *Storage) GetWalletTX(ctx context.Context, tx *sqlx.Tx, walletID int64) (*models.Wallet, error) {
	s.log.Debug().Msg("Start getting wallet")
	query := `
	SELECT *
	FROM wallets
	WHERE id = $1`
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	rows, err := tx.QueryxContext(ctx, query, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}
	wallets, err := s.fromSQLRowsToWallets(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan wallets: %w", err)
	}
	if len(wallets) == 0 {
		return nil, fmt.Errorf("wallet with id %d: %w", walletID, errs.ErrNotFound)
	}
	s.log.Debug().Msgf("Successfully get wallet")
	return wallets[0], nil
}

func (s *Storage) AddMoneyToWallet(ctx context.Context, walletID int64, amount int64) (*models.Wallet, error) {
	s.log.Debug().Msg("Start adding money to wallet")
	wallet, err := s.GetWallet(ctx, walletID)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				s.log.Error().Err(err).Msg("failed to rollback")
			}
		}
	}()
	newValue := wallet.Value + amount
	_, err = s.SetMoneyToWalletTX(ctx, tx, walletID, newValue)
	if err != nil {
		return nil, err
	}

	err = s.AddTransactionTX(ctx, tx, walletID, wallet.UserID, "ADD MONEY")
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	s.log.Debug().Msgf("Successfully add money to wallet")
	wallet.Value = newValue
	return wallet, nil
}

func (s *Storage) AddTransactionTX(ctx context.Context, tx *sqlx.Tx, walletID int64, userID int64, operationName string) error {
	query := `
	INSERT INTO transactions (wallet_id, user_id, date, operation_name)
	VALUES ($1, $2, now(), $3)
	RETURNING id`
	_, err := tx.QueryxContext(ctx, query, walletID, userID, operationName)
	return err
}

func (s *Storage) SetMoneyToWalletTX(ctx context.Context, tx *sqlx.Tx, walletID int64, value int64) (int64, error) {
	s.log.Info().Msg("SetMoneyToWalletTX start")
	q := `
	UPDATE wallets
	SET value = $2
	WHERE id = $1
	RETURNING value;`
	res, err := tx.QueryxContext(ctx, q, walletID, value)
	if err != nil {
		return 0, err
	}
	var id int64
	for res.Next() {
		if err = res.Scan(&id); err != nil {
			return 0, err
		}
	}

	return id, err
}

func (s *Storage) GetUser(ctx context.Context, userID int64) (*models.User, error) {
	s.log.Debug().Msg("Start listing users")
	query := `
	SELECT *
	FROM users
	WHERE id = '$1'`
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	rows, err := s.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	events, scanErr := s.fromSQLRowsToUsers(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan users: %w", scanErr)
	}
	if len(events) == 0 {
		return nil, fmt.Errorf("user with id %d not found: %w", userID, errs.ErrNotFound)
	}
	s.log.Debug().Msgf("Successfully get user")
	return events[0], nil
}

func (s *Storage) GetUserWallets(ctx context.Context, userID int64) ([]*models.Wallet, error) {
	s.log.Debug().Msgf("Start listing wallets for user '%d'", userID)
	query := `
	SELECT *
	FROM wallets
	WHERE user_id = '$1'`
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	rows, err := s.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallets: %w", err)
	}
	wallets, scanErr := s.fromSQLRowsToWallets(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan wallets: %w", scanErr)
	}
	if len(wallets) == 0 {
		return nil, fmt.Errorf("wallets with user_id %d not found: %w", userID, errs.ErrNotFound)
	}
	s.log.Debug().Msgf("Successfully get wallets")
	return wallets, nil
}

func (s *Storage) GetUserWalletsTX(ctx context.Context, tx *sqlx.Tx, userID int64) ([]*models.Wallet, error) {
	s.log.Debug().Msgf("Start listing wallets for user '%d'", userID)
	query := `
	SELECT *
	FROM wallets
	WHERE user_id = '$1'`
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	rows, err := tx.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallets: %w", err)
	}
	wallets, scanErr := s.fromSQLRowsToWallets(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan wallets: %w", scanErr)
	}
	if len(wallets) == 0 {
		return nil, fmt.Errorf("wallets with user_id %d not found: %w", userID, errs.ErrNotFound)
	}
	s.log.Debug().Msgf("Successfully get wallets")
	return wallets, nil
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

func (s *Storage) SaveWalletUnary(ctx context.Context, wallet *models.Wallet) (int64, error) {
	s.log.Debug().Msgf("storage: start saving wallet")
	query := `
	INSERT INTO wallets (user_id, currency, value)
	VALUES ($1, $2, $3)
	RETURNING id`
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	rows, err := s.db.QueryxContext(ctx, query, wallet.UserID, wallet.Currency, wallet.Value)
	if err != nil {
		return 0, fmt.Errorf("failed to save wallet: %w", err)
	}
	var id int64
	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
	}
	s.log.Debug().Msgf("storage: wallet saved successfully")
	return id, nil
}

func (s *Storage) MoneyExchange(
	ctx context.Context,
	userID int64,
	fromWalletID int64,
	toWalletID int64,
	fromAmount int64,
	toAmount int64,
) (*models.Wallet, *models.Wallet, error) {
	s.log.Info().Msgf("start MoneyExhange from %d to %d", fromWalletID, toWalletID)
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err != nil {
			if err = tx.Rollback(); err != nil {
				s.log.Error().Err(err).Msg("failed to rollback transaction")
			}
		}
	}()

	userWallets, err := s.GetUserWalletsTX(ctx, tx, fromWalletID)
	if err != nil {
		return nil, nil, err
	}

	var fromWallet, toWallet *models.Wallet
	for _, wallet := range userWallets {
		if wallet.ID == fromWalletID {
			fromWallet = wallet
			continue
		}
		if wallet.ID == toWalletID {
			toWallet = wallet
		}
	}

	if fromWallet == nil {
		s.log.Warn().Msgf("not found wallet with id %d for user with id %d", fromWalletID, userID)
		return nil, nil, fmt.Errorf("not found wallet with id %d for user with id %d: %w", fromWalletID, userID, errs.ErrNotFound)
	}
	if toWallet == nil {
		s.log.Warn().Msgf("not found wallet with id %d for user with id %d", toWalletID, userID)
		return nil, nil, fmt.Errorf("not found wallet with id %d for user with id %d: %w", toWalletID, userID, errs.ErrNotFound)
	}

	if fromWallet.Value < fromAmount {
		s.log.Warn().Msgf("not much money on the wallet id %d for user with id %d", fromWallet, userID)
		return nil, nil, fmt.Errorf("not much money on the wallet id %d for user with id %d: %w", fromWallet, userID, errs.ErrNotEnoughMoney)
	}

	newFromWalletValue := fromWallet.Value - fromAmount
	_, err = s.SetMoneyToWalletTX(ctx, tx, fromWalletID, newFromWalletValue)
	if err != nil {
		return nil, nil, err
	}
	newToWalletValue := toWallet.Value + toAmount
	_, err = s.SetMoneyToWalletTX(ctx, tx, toWalletID, toWallet.Value + newToWalletValue)
	if err != nil {
		return nil, nil, err
	}

	err = s.AddTransactionTX(ctx, tx, fromWalletID, userID, "EXCHANGE MONEY")
	if err != nil {
		return nil, nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, nil, err
	}
	s.log.Info().Msgf("start MoneyExhange from %d to %d", fromWalletID, toWalletID)

	fromWallet.Value = newFromWalletValue
	toWallet.Value = newToWalletValue
	return fromWallet, toWallet, nil
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

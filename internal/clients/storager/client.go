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
	s.log.Debug().Msgf("storage: start saving user: %s", user.PhoneNumber)

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
	INSERT INTO users (name, middle_name, surname, mail, phone_number, blocked, registered, admin, password)
	VALUES ($1, $2, $3, $4, $5, false, false, false, $6)
	RETURNING id`
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	res, err := tx.QueryContext(ctx, query, user.Name, user.MiddleName, user.Surname, user.Mail, user.PhoneNumber, user.Password)
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
	rows, err := s.db.QueryxContext(ctx, q, userID)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}

func (s *Storage) BlockOrUnblockUser(ctx context.Context, userID int64, block bool) error {
	q := `
	UPDATE users
	SET blocked = $2
	WHERE id = $1;`
	rows, err := s.db.QueryxContext(ctx, q, userID, block)
	if err != nil {
		return err
	}
	defer rows.Close()
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
	defer rows.Close()
	events, err := s.fromSQLRowsToUsers(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan users: %w", err)
	}
	s.log.Debug().Msgf("Successfully list events")
	return events, nil
}

func (s *Storage) GetUserByPhoneNumberOrEmail(ctx context.Context, phoneNumber, mail string) (*models.User, error) {
	s.log.Debug().Msg("Start listing users")
	query := `
	SELECT *
	FROM users
	WHERE phone_number = $1 or mail = $2`
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	rows, err := s.db.QueryxContext(ctx, query, phoneNumber, mail)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	defer rows.Close()
	users, err := s.fromSQLRowsToUsers(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan users: %w", err)
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("user with phoneNumber %s or email %s not found: %w", phoneNumber, mail, errs.ErrNotFound)
	}
	s.log.Debug().Msgf("Successfully get user")
	return users[0], nil
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
	defer rows.Close()
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
	defer rows.Close()
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

func (s *Storage) ListTransactions(ctx context.Context, userID int64) ([]*models.Transaction, error) {
	s.log.Debug().Msg("Start listing transactions")
	query := `
	SELECT *
	FROM transactions
	WHERE user_id = $1
	ORDER BY date;`
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	rows, err := s.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}
	defer rows.Close()
	transactions, err := s.fromSQLRowsToTransactions(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan transactions: %w", err)
	}
	s.log.Debug().Msgf("Successfully list transactions")
	return transactions, nil
}

func (s *Storage) PullMoneyFromWallet(ctx context.Context, walletID int64, amount int64) (*models.Wallet, error) {
	s.log.Debug().Msg("Start pulling money from wallet")
	wallet, err := s.GetWallet(ctx, walletID)
	if err != nil {
		return nil, err
	}

	if wallet.Value < amount {
		return nil, fmt.Errorf("can't pull money from walliet id %d: %w", walletID, errs.ErrNotEnoughMoney)
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
	newValue := wallet.Value - amount
	_, err = s.SetMoneyToWalletTX(ctx, tx, walletID, newValue)
	if err != nil {
		return nil, err
	}

	err = s.AddTransactionTX(
		ctx, tx, wallet.UserID,
		"PULL MONEY", 0,
		wallet.ID, 0, amount,
		"", wallet.Currency, 1)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	s.log.Debug().Msgf("Successfully pull money from wallet")
	wallet.Value = newValue
	return wallet, nil
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

	err = s.AddTransactionTX(
		ctx, tx, wallet.UserID,
		"ADD MONEY", wallet.ID,
		0, amount, 0,
		wallet.Currency, "", 1)
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

func (s *Storage) AddTransactionTX(
	ctx context.Context,
	tx *sqlx.Tx,
	userID int64,
	operationName string,
	incomeWalletID, outcomeWalletID, incomeAmount, outcomeAmount int64,
	incomeWalletCurrency models.Currencies,
	outcomeWalletCurrency models.Currencies,
	courseValue float64,
) error {
	query := `
	INSERT INTO transactions (date, user_id, operation_name, income_amount, outcome_amount, income_wallet_id, outcome_wallet_id, income_wallet_currency, outcome_wallet_currency, course_value)
	VALUES (now(), $1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id`
	rows, err := tx.QueryxContext(ctx, query, userID, operationName, incomeAmount, outcomeAmount, incomeWalletID, outcomeWalletID, incomeWalletCurrency, outcomeWalletCurrency, courseValue)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
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
	defer res.Close()
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
	WHERE id = $1`
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	rows, err := s.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	defer rows.Close()
	events, err := s.fromSQLRowsToUsers(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan users: %w", err)
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
	WHERE user_id = $1`
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	rows, err := s.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallets: %w", err)
	}
	defer rows.Close()
	wallets, err := s.fromSQLRowsToWallets(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan wallets: %w", err)
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
	WHERE user_id = $1`
	ctx, cancel := context.WithTimeout(ctx, s.connectionTimeout)
	defer cancel()
	rows, err := tx.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallets: %w", err)
	}
	defer rows.Close()
	wallets, err := s.fromSQLRowsToWallets(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan wallets: %w", err)
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
	defer rows.Close()
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
	fromCurrency models.Currencies,
	toCurrency models.Currencies,
	courseValue float64,
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

	userWallets, err := s.GetUserWalletsTX(ctx, tx, userID)
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

	err = s.AddTransactionTX(ctx, tx,
		userID,
		"EXCHANGE MONEY",
		toWalletID,
		fromWalletID,
		toAmount,
		fromAmount,
		toCurrency,
		fromCurrency,
		courseValue,
		)
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

func (s *Storage) SaveCourses(ctx context.Context, timeNow time.Time, fromCurrency, toCurrency models.Currencies, course float64) error {
	s.log.Debug().Msgf("saving course %s to %s for %s", fromCurrency, toCurrency, timeNow)
	query := `
	INSERT INTO courses (timestamp, from_currency, to_currency, course)
	VALUES ($1, $2, $3, $4)`
	rows, err := s.db.QueryxContext(ctx, query, timeNow.Unix(), fromCurrency, toCurrency, course)
	if err != nil {
		return err
	}
	defer rows.Close()
	s.log.Debug().Msgf("successfully saved course")
	return nil
}

func (s *Storage) ListCourses(ctx context.Context, fromCurrency, toCurrency models.Currencies, fromTime int64, toTime int64) ([]*models.Course, error) {
	s.log.Debug().Msgf("Start ListCourses")
	q := `
	SELECT *
	FROM courses 
	WHERE timestamp >= $1 AND timestamp <= $2 AND from_currency = $3 AND to_currency = $4
	ORDER BY timestamp;
	`
	s.log.Debug().Msgf("ListCourses: run query: %s", q)

	rows, err := s.db.QueryxContext(ctx, q, fromTime, toTime, fromCurrency, toCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}
	courses, err := s.fromSQLRowsToCourses(rows)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to parse rows: %w", err)
	}

	s.log.Debug().Msgf("finish ListCourses")
	return courses, nil
}

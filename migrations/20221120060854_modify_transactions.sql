-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS transactions
ADD COLUMN IF NOT EXISTS income_amount int,
ADD COLUMN IF NOT EXISTS outcome_amount int,
ADD COLUMN IF NOT EXISTS income_wallet_id int,
ADD COLUMN IF NOT EXISTS outcome_wallet_id int,
ADD COLUMN IF NOT EXISTS income_wallet_currency varchar(50),
ADD COLUMN IF NOT EXISTS outcome_wallet_currency varchar(50),
ADD COLUMN IF NOT EXISTS course_value float,
DROP COLUMN IF EXISTS wallet_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS transactions
    DROP COLUMN IF EXISTS income_amount,
    DROP COLUMN IF EXISTS outcome_amount,
    DROP COLUMN IF EXISTS income_wallet_id,
    DROP COLUMN IF EXISTS outcome_wallet_id,
    DROP COLUMN IF EXISTS income_wallet_currency,
    DROP COLUMN IF EXISTS outcome_wallet_currency,
    DROP COLUMN IF EXISTS course_value;
-- +goose StatementEnd

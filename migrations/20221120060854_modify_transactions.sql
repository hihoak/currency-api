-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS transactions
ADD COLUMN IF NOT EXISTS income_amount int,
ADD COLUMN IF NOT EXISTS outcome_amount int,
ADD COLUMN

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

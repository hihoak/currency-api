-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS courses
(
    id          SERIAL PRIMARY KEY NOT NULL,
    timestamp   bigint,
    from_currency        varchar(50),
    to_currency        varchar(50),
    course      float
);

CREATE INDEX courses_from_to_currency_index ON courses
(
     from_currency, to_currency
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS courses;
-- +goose StatementEnd

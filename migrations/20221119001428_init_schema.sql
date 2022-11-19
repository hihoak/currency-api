-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users
(
    id         SERIAL PRIMARY KEY,
    name         varchar(50) NOT NULL,
    middle_name   varchar(50),
    surname      varchar(50) NOT NULL,
    mail         varchar(50) UNIQUE NOT NULL,
    phone_number varchar(50) UNIQUE NOT NULL,
    blocked    boolean NOT NULL,
    registered   boolean NOT NULL,
    admin        boolean NOT NULL,
    username     varchar(50) UNIQUE NOT NULL,
    password varchar(50) NOT NULL
);

CREATE INDEX users_id_index ON users (
    id
);

CREATE TABLE IF NOT EXISTS transactions
(
    id          SERIAL PRIMARY KEY NOT NULL,
    date         timestamp with time zone NOT NULL,
    user_id        int NOT NULL,
    wallet_id int NOT NULL,
    operation_name varchar(50) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id),
    FOREIGN KEY (wallet_id) REFERENCES wallets (id)
);

CREATE INDEX transactions_user_id_index ON transactions
(
    user_id
);

CREATE TABLE IF NOT EXISTS register_requests
(
    id      SERIAL PRIMARY KEY NOT NULL,
    user_id int NOT NULL,
    FOREIGN KEY ( user_id ) REFERENCES users ( id )
);

CREATE INDEX register_requests_user_id_index ON register_requests
(
    user_id
);

CREATE TABLE IF NOT EXISTS wallets
(
    id       SERIAL PRIMARY KEY NOT NULL,
    user_id  int NOT NULL,
    currency varchar(10) NOT NULL,
    value    int NOT NULL,
    FOREIGN KEY ( user_id ) REFERENCES users ( id )
);

CREATE INDEX wallets_user_id_index ON wallets
(
    user_id
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS register_requests;
DROP TABLE IF EXISTS wallets;
-- +goose StatementEnd

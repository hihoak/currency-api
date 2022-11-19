MIGRATIONS_FOLDER="migrations"

DB_DRIVER:=${GOOSE_DRIVER}
ifeq ($(DB_DRIVER),)
DB_DRIVER:="postgres"
endif
DB_STRING:=${GOOSE_DBSTRING}
ifeq ($(DB_STRING),)
DB_STRING:="host=localhost port=5432 user=postgres password=password dbname=postgres sslmode=disable"
endif

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/currency-api cmd/currency-api/main.go

test:
	go test -race ./...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.41.1

lint: install-lint-deps
	golangci-lint run ./...

.PHONY: build run build-img run-img version test lint

migrate-up:
	GOOSE_DRIVER=$(DB_DRIVER) GOOSE_DBSTRING=$(DB_STRING) goose -dir $(MIGRATIONS_FOLDER) up

migrate-down:
	GOOSE_DRIVER=$(DB_DRIVER) GOOSE_DBSTRING=$(DB_STRING) goose -dir $(MIGRATIONS_FOLDER) down

go-generate:
	buf generate --path "api/event"
	go generate ./...

up:
	sudo docker pull docker.io/soundsofanarchy/currency-api:latest
	docker compose up -d

down:
	docker compose down
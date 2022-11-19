FROM golang:1.19.1 as build

ENV GOOSE_VERSION "v3.7.0"
RUN go install github.com/pressly/goose/v3/cmd/goose@${GOOSE_VERSION}

FROM ubuntu:18.04

WORKDIR /
COPY --from=build /go/bin/goose /usr/bin/goose

ENV MIGRATIONS_FOLDER ./migrations
COPY ${MIGRATIONS_FOLDER} ${MIGRATIONS_FOLDER}

ENV GOOSE_DRIVER "postgres"
ENV GOOSE_DBSTRING "host=postgres port=6432 user=postgres password=password dbname=postgres sslmode=disable"

CMD goose -dir ${MIGRATIONS_FOLDER} up

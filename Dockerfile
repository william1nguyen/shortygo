ARG GO_VERSION=1.24
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

ENV GO111MODULE=on

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/main cmd/shortygo/main.go

EXPOSE 8080
CMD [ "./bin/main" ]
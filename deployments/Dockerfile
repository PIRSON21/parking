FROM golang:alpine AS builder

LABEL stage=go_builder

# определяем переменные окружения (для Go)
ENV CGO_ENABLED=0
ENV GOOS=linux

# обновление системы
RUN apk update --no-cache

WORKDIR /build

# загрузка пакетов
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

# билд файла
COPY . .
RUN go build -o /app/parking ./cmd/parking/main.go

FROM alpine:3.21 AS copy
LABEL stage=copy

RUN apk update --no-cache && apk add --no-cache ca-certificates

WORKDIR /app

COPY ./configs ./configs
COPY --from=builder /app/parking ./parking

# -------- STAGE 1: Build --------
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем go.work и go.mod всех модулей
COPY go.work ./
COPY sso/go.mod sso/go.sum ./sso/
COPY url-shortener/go.mod url-shortener/go.sum ./url-shortener/
COPY analytics/go.mod analytics/go.sum ./analytics/
COPY common/go.mod common/go.sum ./common/

# Загружаем зависимости
RUN go work sync && go mod download

# Копируем весь проект
COPY . .

# Собираем бинарь
WORKDIR /app/sso
RUN go build -o sso ./cmd/sso/main.go

# -------- STAGE 2: Run --------
FROM alpine:latest
WORKDIR /app

# Копируем бинарь
COPY --from=builder /app/sso/sso ./sso

# Копируем конфиги
COPY sso/config ./config

EXPOSE 44045

CMD ["./sso"]
# -------- STAGE 1: Build --------
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем только go.work и go.mod файлов — для лучшего кеша
COPY go.work ./
COPY sso/go.mod sso/go.sum ./sso/
COPY url-shortener/go.mod url-shortener/go.sum ./url-shortener/
COPY analytics/go.mod analytics/go.sum ./analytics/
COPY common/go.mod common/go.sum ./common/

# Загружаем зависимости
RUN go work sync && go mod download

# Теперь копируем весь проект
COPY . .

# Собираем binary
WORKDIR /app/analytics
RUN go build -o analytics ./cmd/main.go

# -------- STAGE 2: Run --------
FROM alpine:latest
WORKDIR /app

# Копируем бинарь
COPY --from=builder /app/analytics/analytics ./analytics

# Копируем конфиг
COPY analytics/config ./config

EXPOSE 8080

CMD ["./analytics"]

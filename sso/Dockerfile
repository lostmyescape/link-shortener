# 1. Используем официальный образ Go
FROM golang:1.24 AS builder

# 2. Рабочая директория
WORKDIR /app

# 3. Копируем go-модули и устанавливаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# 4. Копируем всё остальное
COPY . .

# 5. Собираем бинарник из cmd/sso/main.go
RUN go build -o sso ./cmd/sso/main.go

# 6. Минимальный образ для запуска
FROM debian:bookworm-slim
WORKDIR /app

# 7. Копируем бинарник и конфиги
COPY --from=builder /app/sso .
COPY config/config.yaml ./config/config.yaml

# 8. Порт, на котором работает gRPC
EXPOSE 8080

# 9. Команда запуска
CMD ["./sso"]

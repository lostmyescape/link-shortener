# 1. Базовый образ с Go
FROM golang:1.23 AS builder

# 2. Рабочая директория
WORKDIR /app

# 3. Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# 4. Копируем весь исходный код
COPY . .

# 5. Собираем бинарник
RUN go build -o url-shortener ./cmd/main.go

# 6. Минимальный образ для запуска
FROM debian:bookworm-slim
WORKDIR /app

# 7. Копируем бинарник и конфиги
COPY --from=builder /app/url-shortener .
COPY config/config.yaml ./config/config.yaml

# 8. Порт для HTTP
EXPOSE 8080

# 9. Запускаем приложение
CMD ["./url-shortener"]

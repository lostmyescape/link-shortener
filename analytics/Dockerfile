FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o analytics ./cmd/main.go

FROM debian:bookworm-slim
WORKDIR /app

COPY --from=builder /app/analytics .
COPY config/config.yaml ./config/config.yaml

RUN chmod +x /app/analytics

EXPOSE 8080

CMD ["./analytics"]

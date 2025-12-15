# -------- STAGE 1: Build --------
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.work ./
COPY sso/go.mod sso/go.sum ./sso/
COPY url-shortener/go.mod url-shortener/go.sum ./url-shortener/
COPY analytics/go.mod analytics/go.sum ./analytics/
COPY common/go.mod common/go.sum ./common/

RUN go work sync && go mod download

COPY . .

WORKDIR /app/analytics
RUN go build -o analytics ./cmd/main.go

# -------- STAGE 2: Run --------
FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/analytics/analytics ./analytics

COPY analytics/config ./config

EXPOSE 8080

CMD ["./analytics"]

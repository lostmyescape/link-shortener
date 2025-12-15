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

WORKDIR /app/url-shortener
RUN go build -o url-shortener ./cmd/main.go

# -------- STAGE 2: Run --------
FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/url-shortener/url-shortener ./url-shortener

COPY url-shortener/config ./config

EXPOSE 8090

CMD ["./url-shortener"]
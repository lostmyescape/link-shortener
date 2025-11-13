package mdjwt

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lostmyescape/link-shortener/sso/pkg/tokenstore"
	"github.com/lostmyescape/link-shortener/url-shortener/internal/config"
	"github.com/redis/go-redis/v9"
)

type contextKey string

const (
	userIDKey contextKey = "userID"
)

type Config interface {
	GetRedisAddr() string
	GetRedisPassword() string
}

type RDBConfig struct {
	tokenProvider *tokenstore.TokenStore
	secretKey     string
	log           *slog.Logger
}

func New(cfg *config.Config, logger *slog.Logger) *RDBConfig {
	tokenStore := tokenstore.New(cfg.GetRedisAddr(), cfg.GetRedisPassword())

	return &RDBConfig{
		tokenProvider: tokenStore,
		secretKey:     cfg.Storage.Token,
		log:           logger,
	}
}

func (rdb *RDBConfig) JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing auth header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := parseToken(w, tokenString, rdb)
		if err != nil {
			http.Error(w, "failed to parse token", http.StatusUnauthorized)
			return
		}

		uidFloat, ok := claims["uid"].(float64)
		if !ok {
			http.Error(w, "uid missing or invalid", http.StatusUnauthorized)
			return
		}

		userID := int(uidFloat)

		storedToken, err := rdb.tokenProvider.GetToken(r.Context(), int64(userID))
		if err != nil {
			if errors.Is(err, redis.Nil) {
				http.Error(w, "token expired or invalid", http.StatusUnauthorized)
			} else {
				rdb.log.Error("Redis error:", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		if storedToken != tokenString {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func parseToken(w http.ResponseWriter, tokenString string, rdb *RDBConfig) (jwt.MapClaims, error) {
	const op = "lib.jwt.mdjwt.parseToken"

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(rdb.secretKey), nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return nil, fmt.Errorf("invalid token: %s, %w", op, err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "invalid claims", http.StatusUnauthorized)
		return nil, fmt.Errorf("invalid claims: %s, %w", op, err)
	}

	if claims["uid"] == nil {
		http.Error(w, "uid missed or invalid", http.StatusUnauthorized)
		return nil, fmt.Errorf("uid missed or invalid: %s, %w", op, err)
	}

	return claims, nil
}

func GetUserID(ctx context.Context) (int, bool) {
	uid, ok := ctx.Value(userIDKey).(int)
	return uid, ok
}

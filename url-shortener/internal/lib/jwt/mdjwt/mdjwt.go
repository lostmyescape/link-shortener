package mdjwt

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lostmyescape/link-shortener/url-shortener/internal/config"
)

type contextKey string

const (
	userIDKey contextKey = "userID"
)

type JWTConfig struct {
	secretKey string
	log       *slog.Logger
}

func JWTMDConfig(cfg *config.Config, log *slog.Logger) *JWTConfig {
	return &JWTConfig{
		secretKey: cfg.Storage.Token,
		log:       log,
	}
}

func (j *JWTConfig) JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			j.log.Error("missing auth header")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := ParseToken(w, tokenString, j.secretKey)
		if err != nil {
			j.log.Error("failed to parse token")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		uidFloat, ok := claims["uid"].(float64)
		if !ok {
			j.log.Error("uid missing or invalid")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userID := int(uidFloat)

		ctx := context.WithValue(r.Context(), userIDKey, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ParseToken(w http.ResponseWriter, tokenString, secretKey string) (jwt.MapClaims, error) {
	const op = "lib.jwt.mdjwt.parseToken"

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secretKey), nil
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

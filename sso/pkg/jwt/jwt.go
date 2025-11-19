package jwt

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lostmyescape/link-shortener/sso/internal/domain/models"
)

var (
	ErrInvalidToken            = errors.New("invalid token")
	ErrUidMissedOrInvalid      = errors.New("uid missed or invalid")
	ErrInvalidEmail            = errors.New("invalid email")
	ErrInvalidUID              = errors.New("invalid uid")
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrMissAuthHeader          = errors.New("miss auth header")
)

type TokenProvider interface {
	GetToken(ctx context.Context, userID int64) (string, error)
}

func NewToken(user models.User, app models.App, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(duration).Unix()
	claims["app_id"] = app.ID

	tokenString, err := token.SignedString([]byte(app.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, err
}

func ParseToken(token, secret string) (models.User, error) {
	tokenString := strings.TrimPrefix(token, "Bearer ")

	if tokenString == "" {
		return models.User{}, ErrMissAuthHeader
	}

	claims, err := validToken(tokenString, secret)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to parse token: %w", err)
	}

	uidFloat, ok := claims["uid"].(float64)
	if !ok {
		return models.User{}, ErrInvalidUID
	}

	userID := int64(uidFloat)

	email, ok := claims["email"].(string)
	if !ok {
		return models.User{}, ErrInvalidEmail
	}

	return models.User{ID: userID, Email: email}, nil
}

func validToken(tokenString, secret string) (jwt.MapClaims, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnexpectedSigningMethod
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	if claims["uid"] == nil {
		return nil, ErrUidMissedOrInvalid
	}

	return claims, nil
}

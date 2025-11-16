package jwt

import (
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lostmyescape/link-shortener/sso/internal/domain/models"

	"time"
)

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

func ParseToken(refreshToken, secret string) (models.User, error) {
	if refreshToken == "" {
		return models.User{}, errors.New("missing auth header")
	}

	tokenString := strings.TrimPrefix(refreshToken, "Bearer ")

	claims, err := validToken(tokenString, secret)
	if err != nil {
		return models.User{}, errors.New("failed to parse token")
	}

	uidFloat, ok := claims["uid"].(float64)
	if !ok {
		return models.User{}, errors.New("uid missing or invalid")
	}

	userID := int64(uidFloat)

	email, ok := claims["email"].(string)
	if !ok {
		return models.User{}, errors.New("uid missing or invalid")
	}

	return models.User{ID: userID, Email: email}, nil
}

func validToken(tokenString, secret string) (jwt.MapClaims, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	if claims["uid"] == nil {
		return nil, errors.New("uid missed or invalid")
	}

	return claims, nil
}

package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lostmyescape/sso/internal/domain/models"
)

//func NewToken(user models.User, secret string, duration time.Duration) (string, error) {
//	token := jwt.New(jwt.SigningMethodHS256)
//
//	claims := token.Claims.(jwt.MapClaims)
//	claims["uid"] = user.ID
//	claims["email"] = user.Email
//	claims["exp"] = time.Now().Add(duration).Unix()
//	claims["app_id"] = 1
//
//	tokenString, err := token.SignedString([]byte(secret))
//	if err != nil {
//		return "", err
//	}
//
//	return tokenString, err
//}

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

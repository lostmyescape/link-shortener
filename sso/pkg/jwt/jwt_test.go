package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lostmyescape/link-shortener/sso/internal/domain/models"
)

func TestNewToken(t *testing.T) {
	user := models.User{
		ID:    1,
		Email: "text@example.com",
	}
	app := models.App{
		ID:     23,
		Secret: "secret-key",
	}

	duration := time.Hour

	tokenStr, err := NewToken(user, app, duration)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// получаем данные из строки токена
	parsedToken, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(app.Secret), nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v:", err)
	}

	if !parsedToken.Valid {
		t.Fatal("token is not valid")
	}

	// после того как распарсили токен, сверяем данные
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("failed to get claims")
	}

	// сверяем айди из токена и айди который пришел
	if int64(claims["uid"].(float64)) != user.ID {
		t.Errorf("unexpected uid %d, got %v", user.ID, claims["uid"])
	}

	// сверяем имейл из токена и имейл который пришел
	if claims["email"] != user.Email {
		t.Errorf("expected email: %s, got: %v", user.Email, claims["email"])
	}

	// сверяем айди из токена и айди который пришел
	if int(claims["app_id"].(float64)) != app.ID {
		t.Errorf("expected app_id %d, got %v", app.ID, claims["app_id"])
	}

	exp := int64(claims["exp"].(float64))
	if exp <= time.Now().Unix() {
		t.Errorf("expected exp in the future, got %d", exp)
	}
}

func TestNewToken_EmptySecret(t *testing.T) {
	user := models.User{
		ID:    1,
		Email: "text@example.com",
	}
	app := models.App{
		ID:     23,
		Secret: "",
	}

	duration := time.Hour

	_, err := NewToken(user, app, duration)
	if err != nil {
		t.Error("expected error due to empty secret, got nil")
	}
}

func TestParseToken(t *testing.T) {

}

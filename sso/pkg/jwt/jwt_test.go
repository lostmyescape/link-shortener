package jwt

import (
	"testing"
	"time"

	"github.com/lostmyescape/link-shortener/sso/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	duration = time.Hour
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

	tokenStr, err := NewToken(user, app, duration)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	_, err = ParseToken(tokenStr, app.Secret)
	if err != nil {
		t.Fatalf("failed to parse token")
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

	_, err := NewToken(user, app, duration)
	if err != nil {
		t.Error("expected error due to empty secret, got nil")
	}
}

func TestParseToken(t *testing.T) {
	user := models.User{
		ID:    1,
		Email: "text@example.com",
	}
	app := models.App{
		ID:     23,
		Secret: "secret-key",
	}

	tokenStr, err := NewToken(user, app, duration)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	testCases := []struct {
		name          string
		userEmail     string
		token         string
		secret        string
		expectedError string
	}{
		{
			name:          "successfully parsed",
			userEmail:     user.Email,
			token:         tokenStr,
			secret:        app.Secret,
			expectedError: "",
		},
		{
			name:          "Invalid token",
			userEmail:     user.Email,
			token:         "invalid token",
			secret:        app.Secret,
			expectedError: "failed to parse token",
		},
		{
			name:          "Invalid secret key",
			userEmail:     user.Email,
			token:         tokenStr,
			secret:        "invalid secret",
			expectedError: "failed to parse token",
		},
		{
			name:          "invalid email",
			userEmail:     "wrong email",
			token:         tokenStr,
			secret:        app.Secret,
			expectedError: "invalid email",
		},
		{
			name:          "invalid ID",
			userEmail:     "wrong email",
			token:         tokenStr,
			secret:        app.Secret,
			expectedError: "invalid uid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotUser, err := ParseToken(tc.token, tc.secret)
			switch {
			case err != nil:
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
				return
			case gotUser.Email != user.Email:
				require.Error(t, ErrInvalidEmail)
				assert.Contains(t, ErrInvalidEmail.Error(), tc.expectedError)
			case gotUser.ID != user.ID:
				require.Error(t, ErrInvalidUID)
				assert.Contains(t, ErrInvalidUID.Error(), tc.expectedError)
			}
		})
	}
}

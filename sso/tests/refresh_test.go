package tests

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/lostmyescape/link-shortener/sso/internal/domain/models"
	"github.com/lostmyescape/link-shortener/sso/pkg/jwt"
	"github.com/lostmyescape/link-shortener/sso/tests/suite"
	ssov1 "github.com/lostmyescape/protos/gen/go/sso"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	email    = gofakeit.Email()
	password = randomFakePassword()
)

const (
	invalidToken = "invalid token"
	duration     = time.Hour
	userID       = 5
)

func TestRefresh_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    appID,
	})

	resp, err := st.AuthClient.Refresh(ctx, &ssov1.RefreshRequest{
		RefreshToken: respLogin.GetRefreshToken(),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.NotEmpty(t, resp.Token)

	rToken, err := jwt.ParseToken(resp.RefreshToken, appSecret)
	require.NoError(t, err)
	assert.NotEmpty(t, rToken)

	token, err := jwt.ParseToken(resp.Token, appSecret)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	switch {
	case rToken.Email != email || token.Email != email:
		require.Error(t, jwt.ErrInvalidEmail)
	case rToken.ID != respReg.GetUserId() || token.ID != respReg.GetUserId():
		require.Error(t, jwt.ErrInvalidUID)
	}
}

func TestRefresh(t *testing.T) {
	app := models.App{
		ID:     appID,
		Secret: "secret-key",
	}
	fakeUser := models.User{
		ID:    int64(userID),
		Email: "fakeemail@gmail.com",
	}

	fakeToken, err := jwt.NewToken(fakeUser, app, duration)
	require.NoError(t, err)

	cases := []struct {
		name        string
		rToken      string
		expectedErr string
	}{
		{
			name:        "invalid token",
			rToken:      invalidToken,
			expectedErr: "unauthorized",
		},
		{
			name:        "token do not stored in redis",
			rToken:      fakeToken,
			expectedErr: "internal error",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, st := suite.New(t)
			_, err := st.AuthClient.Refresh(ctx, &ssov1.RefreshRequest{
				RefreshToken: tc.rToken,
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

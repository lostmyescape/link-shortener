package tests

import (
	"testing"

	"github.com/lostmyescape/link-shortener/sso/internal/domain/models"
	"github.com/lostmyescape/link-shortener/sso/pkg/jwt"
	"github.com/lostmyescape/link-shortener/sso/tests/suite"
	ssov1 "github.com/lostmyescape/protos/gen/go/sso"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogout_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	user := models.User{
		ID:    int64(userID),
		Email: email,
	}
	app := models.App{
		ID:     appID,
		Secret: appSecret,
	}

	token, err := jwt.NewToken(user, app, duration)
	require.NoError(t, err)

	resp, err := st.AuthClient.Logout(ctx, &ssov1.LogoutRequest{
		Token: token,
	})
	require.NoError(t, err)

	if resp.GetLogout() != "successful logout" {
		t.Fatalf("failed logout")
	}
}

func TestLogout(t *testing.T) {
	ctx, st := suite.New(t)

	cases := []struct {
		name        string
		token       string
		expectedErr string
	}{
		{
			name:        invalidToken,
			token:       invalidToken,
			expectedErr: "internal error",
		},
		{
			name:        "empty token",
			token:       "",
			expectedErr: "unauthorized",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := st.AuthClient.Logout(ctx, &ssov1.LogoutRequest{
				Token: tc.token,
			})

			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}

}

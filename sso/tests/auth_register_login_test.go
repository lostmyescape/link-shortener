package tests

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	ssov1 "github.com/lostmyescape/link-shortener/protos/gen/go/sso"
	"github.com/lostmyescape/link-shortener/sso/tests/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	emptyAppID = 0
	appID      = 1
	appSecret  = "secret-test"

	passDefLen = 10
)

func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	// init credentials
	email := gofakeit.Email()
	password := randomFakePassword()

	// register new user
	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	// login user with credentials
	respLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    appID,
	})
	require.NoError(t, err)

	loginTime := time.Now()

	token := respLogin.GetToken()
	require.NotEmpty(t, token)

	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	assert.True(t, ok)

	assert.Equal(t, respReg.GetUserId(), int64(claims["uid"].(float64)))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appID, int(claims["app_id"].(float64)))

	const deltaSeconds = 1

	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSeconds)
}

func TestRegisterLogin_Login_Duplicated(t *testing.T) {
	ctx, st := suite.New(t)

	// init credentials
	email := gofakeit.Email()
	password := randomFakePassword()

	// register new user
	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	// register new user
	respReg, err = st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.Error(t, err)
	assert.Empty(t, respReg.GetUserId())
	assert.ErrorContains(t, err, "user already exists")

}

func TestRegister(t *testing.T) {
	cases := []struct {
		name        string
		email       string
		password    string
		expectedErr string
	}{
		{
			name:        "Register with empty password",
			email:       gofakeit.Email(),
			password:    "",
			expectedErr: "password is required",
		},
		{
			name:        "Register with empty email",
			email:       "",
			password:    randomFakePassword(),
			expectedErr: "email is required",
		},
		{
			name:        "Registry with both empty",
			email:       "",
			password:    "",
			expectedErr: "email is required",
		},
		{
			name:        "Invalid email format",
			email:       "tested email",
			password:    randomFakePassword(),
			expectedErr: "invalid email format",
		},
		{
			name:        "Password 6 characters",
			email:       gofakeit.Email(),
			password:    "12345",
			expectedErr: "password must be at least 6 characters long",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, st := suite.New(t)

			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    tc.email,
				Password: tc.password,
			})

			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)

		})
	}

}

func TestLogin(t *testing.T) {
	ctx, st := suite.New(t)
	cases := []struct {
		name        string
		email       string
		password    string
		appID       int32
		expectedErr string
	}{
		{
			name:        "Login with empty password",
			email:       gofakeit.Email(),
			password:    "",
			appID:       appID,
			expectedErr: "password is required",
		},
		{
			name:        "Login with empty email",
			email:       "",
			password:    randomFakePassword(),
			appID:       appID,
			expectedErr: "email is required",
		},
		{
			name:        "Invalid email format",
			email:       "tested email",
			password:    randomFakePassword(),
			appID:       appID,
			expectedErr: "invalid email format",
		},
		{
			name:        "Password 6 characters",
			email:       gofakeit.Email(),
			password:    "12345",
			appID:       appID,
			expectedErr: "password must be at least 6 characters long",
		},
		{
			name:        "Login with both empty",
			email:       "",
			password:    "",
			appID:       appID,
			expectedErr: "email is required",
		},
		{
			name:        "Login with non matching password",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			appID:       appID,
			expectedErr: "invalid email or password",
		},
		{
			name:        "Login without appID",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			appID:       emptyAppID,
			expectedErr: "app_id is required",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    gofakeit.Email(),
				Password: randomFakePassword(),
			})
			require.NoError(t, err)

			_, err = st.AuthClient.Login(ctx, &ssov1.LoginRequest{
				Email:    tc.email,
				Password: tc.password,
				AppId:    tc.appID,
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)

		})
	}

}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefLen)
}

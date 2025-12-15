package tests

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/lostmyescape/link-shortener/url-shortener/internal/http-server/handlers/url/save"
	"github.com/lostmyescape/link-shortener/url-shortener/internal/lib/api"
	"github.com/lostmyescape/link-shortener/url-shortener/internal/lib/random"
	"github.com/lostmyescape/link-shortener/url-shortener/tests/testutils"
	"github.com/stretchr/testify/require"
)

const (
	host = "localhost:8082"
)

var (
	email    = "test@example.com"
	userID   = 5
	appID    = 1
	secret   = "secret-key"
	duration = time.Hour

	req = save.Request{
		URL:   gofakeit.URL(),
		Alias: random.NewRandomString(10),
	}
	user = testutils.User{
		ID:    int64(userID),
		Email: email,
	}
	app = testutils.App{
		ID:     appID,
		Secret: secret,
	}
	u = url.URL{
		Scheme: "http",
		Host:   host,
	}
)

func TestURLShortener(t *testing.T) {
	e := httpexpect.Default(t, u.String())
	token, err := testutils.NewToken(user, app, duration)
	require.NoError(t, err)

	resp := e.POST("/url").
		WithJSON(req).
		WithHeader("Authorization", "Bearer "+strings.TrimPrefix(token, "Bearer ")).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object()

	resp.ContainsKey("Alias")
}

func TestURLShortener_SaveRedirect(t *testing.T) {
	e := httpexpect.Default(t, u.String())

	token, err := testutils.NewToken(user, app, duration)
	require.NoError(t, err)

	existingURL := "https://google.com"
	existingAlias := "google"

	e.POST("/url").
		WithJSON(save.Request{
			URL:   existingURL,
			Alias: existingAlias,
		}).
		WithHeader("Authorization", "Bearer "+strings.TrimPrefix(token, "Bearer ")).
		Expect().
		Status(http.StatusOK)

	testCases := []struct {
		name     string
		url      string
		alias    string
		error    string
		wantCode int
	}{
		{
			name:     "Valid URL",
			url:      gofakeit.URL(),
			alias:    gofakeit.Word() + gofakeit.Word(),
			wantCode: http.StatusOK,
		},
		{
			name:     "Invalid URL",
			url:      "invalid_url",
			alias:    gofakeit.Word(),
			error:    "field URL is not a valid URL",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "Empty Alias",
			url:      gofakeit.URL(),
			alias:    "",
			wantCode: http.StatusOK,
		},
		{
			name:     "URL already exists",
			url:      existingURL,
			alias:    "",
			error:    "URL already exists",
			wantCode: http.StatusConflict,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := e.POST("/url").
				WithJSON(save.Request{
					URL:   tc.url,
					Alias: tc.alias,
				}).
				WithHeader("Authorization", "Bearer "+strings.TrimPrefix(token, "Bearer ")).
				Expect().
				Status(tc.wantCode).
				JSON().
				Object()

			if tc.error != "" {
				resp.NotContainsKey("alias")
				resp.Value("error").String().IsEqual(tc.error)
				return
			}

			alias := tc.alias

			if tc.alias != "" {
				resp.Value("Alias").String().IsEqual(tc.alias)
			} else {
				resp.Value("Alias").String().NotEmpty()
				alias = resp.Value("Alias").String().Raw()
			}

			t.Logf("alias: %s", alias)
			t.Logf("url: %s", tc.url)

			testRedirect(t, alias, tc.url)

			// Очищаем после теста
			e.DELETE("/url/"+alias).
				WithHeader("Authorization", "Bearer "+strings.TrimPrefix(token, "Bearer ")).
				Expect().
				Status(http.StatusOK)
		})
	}

	e.DELETE("/url/"+existingAlias).
		WithHeader("Authorization", "Bearer "+strings.TrimPrefix(token, "Bearer ")).
		Expect().
		Status(http.StatusOK)
}

func testRedirect(t *testing.T, alias string, urlToRedirect string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}

	redirectedToURL, err := api.GetRedirect(u.String())
	require.NoError(t, err)
	require.Equal(t, urlToRedirect, redirectedToURL)
}

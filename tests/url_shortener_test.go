package tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/lostmyescape/url-shortener/internal/http-server/handlers/url/save"
	"github.com/lostmyescape/url-shortener/internal/lib/api"
	"github.com/lostmyescape/url-shortener/internal/lib/random"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"path"
	"testing"
)

const (
	host = "localhost:8080"
)

func TestURLShortener(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}

	e := httpexpect.Default(t, u.String())

	e.POST("/url").
		WithJSON(save.Request{
			URL:   gofakeit.URL(),
			Alias: random.NewRandomString(10),
		}).WithBasicAuth("lostmyescape", "asdfg").
		Expect().Status(200).JSON().Object().ContainsKey("alias")
}

func TestURLShortener_SaveRedirect(t *testing.T) {
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
			url:      "https://google.com",
			alias:    "google",
			error:    "URL already exists",
			wantCode: http.StatusConflict,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   host,
			}
			e := httpexpect.Default(t, u.String())

			resp := e.POST("/url").WithJSON(save.Request{
				URL:   tc.url,
				Alias: tc.alias,
			}).
				WithBasicAuth("lostmyescape", "asdfg").
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

			if tc.alias == "" {
				resp.Value("alias").String().NotEmpty()
				alias = resp.Value("alias").String().Raw()
			} else {
				resp.Value("alias").String().IsEqual(tc.alias)
				alias = tc.alias
			}

			testRedirect(t, alias, tc.url)

			e.DELETE("/"+path.Join("url", alias)).
				WithBasicAuth("lostmyescape", "asdfg").
				Expect().Status(http.StatusOK)
		})
	}
}

func testRedirect(t *testing.T, alias string, urlToRedirect string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}

	redirectedToURL, err := api.GetRedirect(u.String())
	require.NoError(t, err)

	require.Equal(t, redirectedToURL, urlToRedirect)
}

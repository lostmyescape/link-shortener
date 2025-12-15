package redirect

//go:generate mockery --name URLSearcher

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/lostmyescape/link-shortener/url-shortener/internal/http-server/handlers/redirect/mocks"
	"github.com/lostmyescape/link-shortener/url-shortener/internal/lib/api"
	"github.com/lostmyescape/link-shortener/url-shortener/internal/lib/logger/handlers/slogdiscard"
	"github.com/lostmyescape/link-shortener/url-shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedirectHandler(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
		wantCode  int
		mockURL   string
	}{
		{
			name:     "Success",
			alias:    "google",
			wantCode: http.StatusFound,
			mockURL:  "https://google.com",
			url:      "https://google.com",
		},
		{
			name:      "Empty alias",
			alias:     "",
			respError: "alias is empty",
			wantCode:  http.StatusNotFound,
		},
		{
			name:      "URL not found",
			alias:     "some_wrong_alias",
			respError: "URL not found",
			wantCode:  http.StatusNotFound,
			mockError: storage.ErrURLNotFound,
		},
		{
			name:      "GetURL error",
			alias:     "test_alias",
			respError: "internal error",
			wantCode:  http.StatusInternalServerError,
			mockError: errors.New("unexpected error"),
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			urlSearcherMock := mocks.NewURLSearcher(t)

			if tc.alias != "" {
				if tc.mockError != nil {
					urlSearcherMock.On("GetUrl", tc.alias).
						Return("", tc.mockError).
						Once()
				} else {
					urlSearcherMock.On("GetUrl", tc.alias).
						Return(tc.mockURL, nil).
						Once()
				}
			}

			handler := Redirect(slogdiscard.NewDiscardLogger(), urlSearcherMock)

			r := chi.NewRouter()
			r.Get("/{alias}", handler)

			ts := httptest.NewServer(r)
			defer ts.Close()

			switch {
			case tc.wantCode == http.StatusFound:
				resp, err := api.GetRedirect(ts.URL + "/" + tc.alias)
				require.NoError(t, err)
				assert.Equal(t, tc.url, resp)
			case tc.wantCode == http.StatusNotFound && tc.alias == "":
				resp, err := http.Get(ts.URL + "/" + tc.alias)
				defer resp.Body.Close()
				require.NoError(t, err)
				assert.Equal(t, tc.wantCode, resp.StatusCode)
			default:
				// Для ошибок делаем прямой HTTP запрос и проверяем статус и тело
				resp, err := http.Get(ts.URL + "/" + tc.alias)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, tc.wantCode, resp.StatusCode)

				// Читаем тело ответа для проверки сообщения об ошибке
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				if tc.respError != "" {
					assert.Contains(t, string(body), tc.respError)
				}
			}

			urlSearcherMock.AssertExpectations(t)
		})
	}
}

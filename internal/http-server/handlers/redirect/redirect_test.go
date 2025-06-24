package redirect

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/lostmyescape/url-shortener/internal/http-server/handlers/redirect/mocks"
	"github.com/lostmyescape/url-shortener/internal/lib/api"
	"github.com/lostmyescape/url-shortener/internal/lib/logger/handlers/slogdiscard"
	"github.com/lostmyescape/url-shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
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
			wantCode:  http.StatusBadRequest,
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

			if tc.mockError != nil {
				urlSearcherMock.On("GetUrl", tc.alias).
					Return("", tc.mockError).
					Once()
			} else {
				urlSearcherMock.On("GetUrl", tc.alias).
					Return(tc.mockURL, nil).
					Once()
			}

			handler := Redirect(slogdiscard.NewDiscardLogger(), urlSearcherMock)

			r := chi.NewRouter()
			r.Get("/{alias}", handler)

			ts := httptest.NewServer(r)
			defer ts.Close()

			redirectToURL, err := api.GetRedirect(ts.URL + "/" + tc.alias)
			require.NoError(t, err)

			assert.Equal(t, tc.url, redirectToURL)
		})
	}
}

package redirect

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	resp "github.com/lostmyescape/link-shortener/url-shortener/internal/lib/api/response"
	"github.com/lostmyescape/link-shortener/url-shortener/internal/lib/logger/sl"
	"github.com/lostmyescape/link-shortener/url-shortener/internal/storage"
)

//go:generate mockery --name=URLSearcher --dir=. --output=./mocks --filename=url_redirect_mock.go --outpkg=mocks
type URLSearcher interface {
	GetUrl(alias string) (string, error)
}

func Redirect(log *slog.Logger, searchUrl URLSearcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.redirect.redirect"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")

		// validate request
		if alias == "" {
			log.Error("alias is empty")
			resp.NewJSON(w, r, http.StatusBadRequest, resp.Error("alias is empty"))

			return
		}

		// trying to get a url
		url, err := searchUrl.GetUrl(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("URL not found", slog.String("alias", alias))
			resp.NewJSON(w, r, http.StatusNotFound, resp.Error("URL not found"))

			return
		}

		if err != nil {
			log.Error("failed searching URL", sl.Err(err))
			resp.NewJSON(w, r, http.StatusInternalServerError, resp.Error("internal error"))

			return
		}

		log.Info("got url", slog.String("url", url))

		http.Redirect(w, r, url, http.StatusFound)
	}
}

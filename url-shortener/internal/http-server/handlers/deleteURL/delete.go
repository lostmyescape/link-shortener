package deleteURL

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	resp "github.com/lostmyescape/url-shortener/internal/lib/api/response"
	"github.com/lostmyescape/url-shortener/internal/lib/logger/sl"
	"github.com/lostmyescape/url-shortener/internal/storage"
	"log/slog"
	"net/http"
)

type Response struct {
	resp.Response
	Alias string
}

type URLDeleter interface {
	DeleteURL(alias string) error
}

func New(log *slog.Logger, delete URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.deleteURL.deleteURL"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")

		if alias == "" {
			log.Error("alias is empty")
			resp.NewJSON(w, r, http.StatusBadRequest, resp.Error("alias is empty"))

			return
		}

		err := delete.DeleteURL(alias)

		switch {
		case err == nil:
			log.Info("url deleted")
			resp.RespOk(w, r, alias)
		case errors.Is(err, storage.ErrAliasNotFound):
			log.Error("alias not found", sl.Err(err))
			resp.NewJSON(w, r, http.StatusNotFound, resp.Error("alias not found"))
		default:
			log.Error("unexpected error", sl.Err(err))
			resp.NewJSON(w, r, http.StatusInternalServerError, resp.Error("unexpected error"))
		}
	}
}

package deleteURL

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
			NewJSON(w, r, http.StatusBadRequest, resp.Error("alias is empty"))

			return
		}

		// delete url
		err := delete.DeleteURL(alias)

		switch {
		case err == nil:
			log.Info("url deleted")
			responseOk(w, r, alias)
		case errors.Is(err, storage.ErrAliasNotFound):
			log.Error("alias not found", sl.Err(err))
			NewJSON(w, r, http.StatusNotFound, resp.Error("alias not found"))
		default:
			log.Error("unexpected error", sl.Err(err))
			NewJSON(w, r, http.StatusInternalServerError, resp.Error("unexpected error"))
		}
	}
}

func NewJSON(w http.ResponseWriter, _ *http.Request, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(true)

	if err := enc.Encode(v); err != nil {
		w.WriteHeader(status)
		fmt.Fprintf(w, `{"error": "failed to encode response"}`)
		return
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}

func responseOk(w http.ResponseWriter, r *http.Request, alias string) {
	NewJSON(w, r, http.StatusOK, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}

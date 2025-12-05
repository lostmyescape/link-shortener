package deleteURL

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lostmyescape/link-shortener/common/kafka"
	resp "github.com/lostmyescape/link-shortener/url-shortener/internal/lib/api/response"
	"github.com/lostmyescape/link-shortener/url-shortener/internal/lib/jwt/mdjwt"
	"github.com/lostmyescape/link-shortener/url-shortener/internal/lib/logger/sl"
	"github.com/lostmyescape/link-shortener/url-shortener/internal/storage"
)

type URLDeleter interface {
	DeleteURL(alias string) error
}

func New(log *slog.Logger, delete URLDeleter, producer *kafka.Producer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.deleteURL.deleteURL"

		ctx := context.Background()

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")

		userID, ok := mdjwt.GetUserID(r.Context())

		if !ok {
			resp.NewJSON(w, r, http.StatusUnauthorized, resp.Error("unauthorized"))
			return
		}

		if alias == "" {
			log.Error("alias is empty")
			resp.NewJSON(w, r, http.StatusBadRequest, resp.Error("alias is empty"))

			return
		}

		ev := map[string]interface{}{
			"type":      kafka.EventLinkDeleted,
			"timestamp": time.Now().UTC(),
			"user_id":   int64(userID),
			"ip":        "kafka:9092",
		}

		err := delete.DeleteURL(alias)

		switch {
		case err == nil:
			log.Info("url deleted")
			err = producer.Publish(ctx, strconv.FormatInt(int64(userID), 10), ev)
			if err != nil {
				log.Error("failed to send message to Kafka", sl.Err(err))
			}
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

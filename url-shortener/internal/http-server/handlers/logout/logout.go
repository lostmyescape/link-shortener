package logout

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/lostmyescape/link-shortener/sso/pkg/tokenstore"
	"github.com/lostmyescape/link-shortener/url-shortener/internal/config"
	resp "github.com/lostmyescape/link-shortener/url-shortener/internal/lib/api/response"
	"github.com/lostmyescape/link-shortener/url-shortener/internal/lib/jwt/mdjwt"
	"github.com/redis/go-redis/v9"
)

type Config interface {
	GetRedisAddr() string
	GetRedisPassword() string
}

type RDBConfig struct {
	tokenProvider *tokenstore.TokenStore
	log           *slog.Logger
}

func New(cfg *config.Config, logger *slog.Logger) *RDBConfig {
	tokenStore := tokenstore.New(cfg.GetRedisAddr(), cfg.GetRedisPassword())

	return &RDBConfig{
		tokenProvider: tokenStore,
		log:           logger,
	}
}

func (rdb *RDBConfig) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.logout.New"

		rdb.log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		userID, ok := mdjwt.GetUserID(r.Context())
		if !ok {
			resp.NewJSON(w, r, http.StatusUnauthorized, resp.Error("unauthorized"))
			return
		}

		err := rdb.tokenProvider.DeleteToken(r.Context(), int64(userID))
		if err != nil {
			if errors.Is(err, redis.Nil) {
				http.Error(w, "token expired or invalid", http.StatusUnauthorized)
			} else {
				rdb.log.Error("Redis error:", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		resp.RespOk(w, r, "logged out")
	}
}

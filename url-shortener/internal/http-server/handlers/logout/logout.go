package logout

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	RToken "github.com/lostmyescape/link-shortener/sso/pkg/redis"
	resp "github.com/lostmyescape/link-shortener/url-shortener/internal/lib/api/response"
	"github.com/lostmyescape/link-shortener/url-shortener/internal/lib/jwt/mdjwt"
)

const (
	RedisAddr = "localhost:6379"
)

var tokenStore = RToken.New(RedisAddr)

func New(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.logout.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		userID, ok := mdjwt.GetUserID(r.Context())
		if !ok {
			resp.NewJSON(w, r, http.StatusUnauthorized, resp.Error("unauthorized"))
			return
		}

		err := tokenStore.DeleteToken(r.Context(), int64(userID))
		if err != nil {
			log.Info("failed to delete token: %v", err)
		}

		resp.RespOk(w, r, "logged out")
	}
}

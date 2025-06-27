package save

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	resp "github.com/lostmyescape/url-shortener/internal/lib/api/response"
	"github.com/lostmyescape/url-shortener/internal/lib/jwt/mdjwt"
	"github.com/lostmyescape/url-shortener/internal/lib/logger/sl"
	"github.com/lostmyescape/url-shortener/internal/lib/random"
	"github.com/lostmyescape/url-shortener/internal/storage"
	"log/slog"
	"net/http"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

//go:generate mockery --name=URLSaver --dir=. --output=./mocks --filename=url_saver_mock.go --outpkg=mocks
type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

const aliasLength = 6

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		// TODO: get uid for jwt
		_, ok := mdjwt.GetUserID(r.Context())

		if !ok {
			resp.NewJSON(w, r, http.StatusUnauthorized, resp.Error("unauthorized"))
			return
		}

		var req Request

		// decode body
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			resp.NewJSON(w, r, http.StatusBadRequest, resp.Error("invalid request body"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		// validator for errors struct
		if err := validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)

			log.Error("invalid request", sl.Err(err))
			resp.NewJSON(w, r, http.StatusBadRequest, resp.ValidationError(validateErr))

			return
		}

		// if alias is empty, generate a new alias
		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		id, err := urlSaver.SaveURL(req.URL, alias)

		if err != nil {
			switch {
			case errors.Is(err, storage.ErrURLExists):
				log.Error("URL already exists", sl.Err(err))
				resp.NewJSON(w, r, http.StatusConflict, resp.Error("URL already exists"))
				return
			case errors.Is(err, storage.ErrAliasExists):
				log.Error("alias already exists", sl.Err(err))
				resp.NewJSON(w, r, http.StatusConflict, resp.Error("alias already exists"))
				return
			default:
				log.Error("failed to add error")
				resp.NewJSON(w, r, http.StatusInternalServerError, resp.Error("failed to add URL"))
				return
			}

		}
		log.Info("url added", slog.Int64("id", id))
		resp.RespOk(w, r, alias)
	}
}

package grpc

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	ssov1 "github.com/lostmyescape/protos/gen/go/sso"
	resp "github.com/lostmyescape/url-shortener/internal/lib/api/response"
	"github.com/lostmyescape/url-shortener/internal/lib/logger/sl"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"net/http"
	"time"
)

type Client struct {
	api ssov1.AuthClient
	log *slog.Logger
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	AppID    int32  `json:"app_id" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func New(
	log *slog.Logger,
	addr string,
	timeout time.Duration,
	retriesCount int,
) (*Client, error) {
	const op = "grpc.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(retriesCount)),
		grpcretry.WithPerRetryTimeout(timeout),
	}

	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent),
	}

	cc, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Client{
		api: ssov1.NewAuthClient(cc),
	}, nil
}

func (c *Client) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "grpc.IsAdmin"

	response, err := c.api.IsAdmin(ctx, &ssov1.IsAdminRequest{
		UserId: userID,
	})
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return response.IsAdmin, nil
}

func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, level grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(level), msg, fields...)
	})
}

func (c *Client) Register(_ context.Context, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "grpc.Register"

		ctx := r.Context()

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req RegisterRequest

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			resp.NewJSON(w, r, http.StatusBadRequest, resp.Error("invalid request body"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		respGRPC, err := c.api.Register(ctx, &ssov1.RegisterRequest{
			Email:    req.Email,
			Password: req.Password,
		})
		if err != nil {
			log.Error("gRPC Register failed", sl.Err(err))
			resp.NewJSON(w, r, http.StatusUnauthorized, resp.Error("register failed"))

			return
		}
		if respGRPC.UserId == 0 {
			log.Error("registration failed: grpc returns false")
			resp.NewJSON(w, r, http.StatusConflict, "error")
			return
		}

		log.Info("user registered", slog.Any("grpc response:", respGRPC))

		resp.NewJSON(w, r, http.StatusOK, "user successfully registered")
		return
	}
}

func (c *Client) Login(_ context.Context, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "grpc.Login"

		ctx := r.Context()

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(ctx)),
		)

		var req LoginRequest

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			resp.NewJSON(w, r, http.StatusBadRequest, resp.Error("invalid request body"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		response, err := c.api.Login(ctx, &ssov1.LoginRequest{
			Email:    req.Email,
			Password: req.Password,
			AppId:    req.AppID,
		})
		if err != nil {
			resp.NewJSON(w, r, http.StatusUnauthorized, resp.Error("login failed"))

			return
		}

		resp.NewJSON(w, r, http.StatusOK, LoginResponse{Token: response.Token})
		return
	}
}

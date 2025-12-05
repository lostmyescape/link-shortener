package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lostmyescape/link-shortener/common/kafka"
	"github.com/lostmyescape/link-shortener/common/logger/slogpretty"
	ssogrpc "github.com/lostmyescape/link-shortener/url-shortener/internal/clients/sso/grpc"
	"github.com/lostmyescape/link-shortener/url-shortener/internal/config"
	"github.com/lostmyescape/link-shortener/url-shortener/internal/http-server/handlers/deleteURL"
	"github.com/lostmyescape/link-shortener/url-shortener/internal/http-server/handlers/redirect"
	"github.com/lostmyescape/link-shortener/url-shortener/internal/http-server/handlers/url/save"
	mwLogger "github.com/lostmyescape/link-shortener/url-shortener/internal/http-server/logger/middleware"
	"github.com/lostmyescape/link-shortener/url-shortener/internal/lib/jwt/mdjwt"
	"github.com/lostmyescape/link-shortener/url-shortener/internal/lib/logger/sl"
	dbstorage "github.com/lostmyescape/link-shortener/url-shortener/internal/storage"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	ctx := context.Background()
	cfg := config.LoadConfig()
	log := setupLogger(cfg.Env)
	storage := dbstorage.NewStorage(ctx, cfg, log)

	log.Info("starting url-shortener", slog.Any("env", cfg))

	producerProvider := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic)

	ssoClient, err := ssogrpc.New(
		log,
		cfg.Clients.SSO.Address,
		cfg.Clients.SSO.Timeout,
		cfg.Clients.SSO.RetriesCount,
	)
	if err != nil {
		log.Error("failed to init sso client", sl.Err(err))
		os.Exit(1)
	}
	ssoClient.IsAdmin(context.Background(), 1)

	defer func(DB *sql.DB) {
		err := storage.DB.Close()
		if err != nil {
			log.Error("DB close error", sl.Err(err))
			return
		}
	}(storage.DB)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	jwtMiddleware := mdjwt.JWTMDConfig(cfg, log)

	router.Route("/url", func(r chi.Router) {
		r.Use(jwtMiddleware.JWTAuthMiddleware)
		r.Post("/", save.New(log, storage, producerProvider))
		r.Delete("/{alias}", deleteURL.New(log, storage, producerProvider))
	})

	router.Route("/logout", func(r chi.Router) {
		r.Use(jwtMiddleware.JWTAuthMiddleware)
		r.Post("/", ssoClient.Logout(context.Background(), log))
	})

	router.Get("/{alias}", redirect.Redirect(log, storage))
	router.Post("/register", ssoClient.Register(context.Background(), log))
	router.Post("/login", ssoClient.Login(context.Background(), log))
	router.Get("/refresh", ssoClient.Refresh(context.Background(), log))

	log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("failed to start server", err)
	}

	log.Error("server stopped")

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()

	case envDev:

		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	case envProd:

		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}

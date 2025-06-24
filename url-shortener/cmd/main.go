package main

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	ssogrpc "github.com/lostmyescape/url-shortener/internal/clients/sso/grpc"
	"github.com/lostmyescape/url-shortener/internal/config"
	"github.com/lostmyescape/url-shortener/internal/http-server/handlers/deleteURL"
	"github.com/lostmyescape/url-shortener/internal/http-server/handlers/redirect"
	"github.com/lostmyescape/url-shortener/internal/http-server/handlers/url/save"
	mwLogger "github.com/lostmyescape/url-shortener/internal/http-server/logger/middleware"
	"github.com/lostmyescape/url-shortener/internal/lib/logger/handlers/slogpretty"
	"github.com/lostmyescape/url-shortener/internal/lib/logger/sl"
	dbstorage "github.com/lostmyescape/url-shortener/internal/storage"
	"log/slog"
	"net/http"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	cfg := config.LoadConfig()

	log := setupLogger(cfg.Env)

	log.Info(
		"starting url-shortener",
		slog.String("env", cfg.Env),
		slog.String("version", "123"),
	)

	log.Debug("debug messages are enabled")

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

	storage, err := dbstorage.NewStorage(cfg)
	if err != nil {
		log.Error("DB connection error: %v", sl.Err(err))
		os.Exit(1)
	}

	defer func(DB *sql.DB) {
		err := storage.DB.Close()
		if err != nil {
			log.Error("DB close error: %v", err)
			return
		}
	}(storage.DB)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))
		r.Post("/", save.New(log, storage))
		r.Delete("/{alias}", deleteURL.New(log, storage))
	})

	router.Get("/{alias}", redirect.Redirect(log, storage))

	log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("failed to start server: %v", err)
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

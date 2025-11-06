package app

import (
	"log/slog"
	"os"
	"time"

	grpcapp "github.com/lostmyescape/link-shortener/sso/internal/app/grpc"
	"github.com/lostmyescape/link-shortener/sso/internal/config"
	"github.com/lostmyescape/link-shortener/sso/internal/lib/logger/sl"
	"github.com/lostmyescape/link-shortener/sso/internal/services/auth"
	"github.com/lostmyescape/link-shortener/sso/internal/storage/postgres"
	"github.com/lostmyescape/link-shortener/sso/pkg/tokenstore"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, cfg *config.Config, tokenTTL time.Duration, tokenStore *tokenstore.TokenStore) *App {

	storage, err := postgres.NewStorage(cfg)
	if err != nil {
		log.Error("db connection error: %v", sl.Err(err))
		os.Exit(1)
	}

	authService := auth.New(log, storage, storage, storage, tokenTTL, tokenStore)
	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}

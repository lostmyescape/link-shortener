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
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, cfg *config.Config, tokenTTL time.Duration) *App {

	storage, err := postgres.NewStorage(cfg)
	if err != nil {
		log.Error("db connection error: %v", sl.Err(err))
		os.Exit(1)
	}

	//defer storage.DB.Close()

	authService := auth.New(log, storage, storage, storage, tokenTTL)
	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}

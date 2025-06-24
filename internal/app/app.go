package app

import (
	grpcapp "github.com/lostmyescape/sso/internal/app/grpc"
	"github.com/lostmyescape/sso/internal/config"
	"github.com/lostmyescape/sso/internal/lib/logger/sl"
	"github.com/lostmyescape/sso/internal/services/auth"
	"github.com/lostmyescape/sso/internal/storage/postgres"
	"log/slog"
	"os"
	"time"
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

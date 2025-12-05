package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/lostmyescape/link-shortener/common/kafka"
	grpcapp "github.com/lostmyescape/link-shortener/sso/internal/app/grpc"
	"github.com/lostmyescape/link-shortener/sso/internal/config"
	"github.com/lostmyescape/link-shortener/sso/internal/lib/tokenstore"
	"github.com/lostmyescape/link-shortener/sso/internal/services/auth"
	"github.com/lostmyescape/link-shortener/sso/internal/storage/postgres"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	cfg *config.Config,
	tokenTTL time.Duration,
	rTokenTTL time.Duration,
	tokenStore *tokenstore.TokenStore,
	producerProvider *kafka.Producer,
	ip string,
) *App {

	storage := postgres.NewStorage(context.Background(), cfg, log)

	authService := auth.New(
		log,
		storage,
		storage,
		storage,
		tokenTTL,
		rTokenTTL,
		tokenStore,
		producerProvider,
		ip,
	)
	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}

package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/lostmyescape/link-shortener/sso/internal/app"
	"github.com/lostmyescape/link-shortener/sso/internal/config"
	"github.com/lostmyescape/link-shortener/sso/internal/lib/logger/handlers/slogpretty"
	redisClient "github.com/lostmyescape/link-shortener/sso/internal/storage/redis"
	"github.com/lostmyescape/link-shortener/sso/pkg/tokenstore"
	"github.com/redis/go-redis/v9"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting application", slog.Any("env", cfg))

	rdb, err := redisClient.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	defer func(rdb *redis.Client) {
		err := rdb.Close()
		if err != nil {
			log.Error("failed to close Redis")
		}
	}(rdb)

	tokenStore := tokenstore.NewRedisStore(rdb)

	application := app.New(log, cfg.GRPC.Port, cfg, cfg.TokenTTL, cfg.RTokenTTL, tokenStore)

	go application.GRPCSrv.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	log.Info("application stopping", slog.String("signal", sign.String()))
	application.GRPCSrv.Stop()
	log.Info("Application stopped")

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()

	case envDev:

		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	case envProd:

		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

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

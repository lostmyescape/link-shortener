package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/lostmyescape/link-shortener/analytics/internal/config"
	"github.com/lostmyescape/link-shortener/analytics/internal/kafka/kafka"
	"github.com/lostmyescape/link-shortener/common/kafka/consumer"
	"github.com/lostmyescape/link-shortener/common/logger/sl"
	"github.com/lostmyescape/link-shortener/common/logger/slogpretty"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	ctx := context.Background()

	err := kafka.MustLoad(ctx, log, cfg.Kafka.Brokers)
	if err != nil {
		log.Error("error connection to Kafka", sl.Err(err))
	}

	log.Info("starting analytics",
		slog.String("env", cfg.Env),
	)

	topics := []string{cfg.Kafka.TopicUser, cfg.Kafka.TopicLink}

	eventsConsumer := consumer.NewConsumer(
		cfg.Kafka.Brokers,
		topics,
		cfg.Kafka.GroupID,
		log,
	)
	eventsConsumer.Start(ctx)

	<-ctx.Done()
	log.Info("shutting down")
}

func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger

	switch env {
	case envLocal:
		logger = setupPrettySlog()

	case envDev:

		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	case envProd:

		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	}

	return logger
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

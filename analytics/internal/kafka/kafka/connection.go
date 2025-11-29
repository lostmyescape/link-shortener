package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

func MustLoad(ctx context.Context, log *slog.Logger, brokers []string) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    "healthcheck",
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	for {
		if err := r.SetOffset(kafka.LastOffset); err == nil {
			log.Info("Kafka ready")
			r.Close()
			return nil
		}

		log.Warn("Kafka not ready, retrying...")
		time.Sleep(time.Second * 2)

		if ctx.Err() != nil {
			return fmt.Errorf("timeout waiting for Kafka")
		}
	}
}

package consumer

import (
	"context"
	"errors"
	"log/slog"

	"github.com/lostmyescape/link-shortener/common/logger/sl"
	"github.com/segmentio/kafka-go"
)

type HandlerFunc func(msg kafka.Message) error

type Consumer struct {
	r   *kafka.Reader
	log *slog.Logger
}

func NewConsumer(brokers []string, groupID, topic string, log *slog.Logger) *Consumer {

	return &Consumer{
		r: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			GroupID:  groupID,
			Topic:    topic,
			MinBytes: 10e3,
			MaxBytes: 10e6,
		}),
		log: log,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	go func() {
		for {
			msg, err := c.r.ReadMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					c.log.Info("consumer stopped")
					return
				}
				c.log.Error("consumer error", sl.Err(err))
				continue
			}

			c.log.Info("message from kafka: ", string(msg.Value))
		}
	}()
}

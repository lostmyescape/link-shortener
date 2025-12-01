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

func NewConsumer(brokers, topics []string, groupID string, log *slog.Logger) *Consumer {

	return &Consumer{
		r: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     brokers,
			GroupID:     groupID,
			Topic:       "",
			GroupTopics: topics,
			MinBytes:    10e3,
			MaxBytes:    10e6,
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
				c.log.Error("kafka read error", sl.Err(err))
				continue
			}

			switch msg.Topic {
			case "user-events":
				c.log.Info("topic: user-events, message from kafka:", string(msg.Value))
			case "link-events":
				c.log.Info("topic: link-events, message from kafka:", string(msg.Value))
			default:
				c.log.Warn("unknown topic", slog.String("topic", msg.Topic))
			}
		}

	}()
}

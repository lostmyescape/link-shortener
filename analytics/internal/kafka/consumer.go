package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	ch "github.com/lostmyescape/link-shortener/analytics/internal/clickhouse"
	"github.com/lostmyescape/link-shortener/common/logger/sl"
	"github.com/segmentio/kafka-go"
)

const (
	maxBatchSize = 10_000
	maxBatchAge  = 20 * time.Second
)

type AnalyticsService struct {
	r               *kafka.Reader
	log             *slog.Logger
	UserEventBuffer []ch.UserEvent
	LinkEventBuffer []ch.LinkEvent
	lastFlush       time.Time
	conn            clickhouse.Conn
}

func NewConsumer(brokers, topics []string, groupID string, log *slog.Logger, conn clickhouse.Conn) *AnalyticsService {
	return &AnalyticsService{
		r: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     brokers,
			GroupID:     groupID,
			Topic:       "",
			GroupTopics: topics,
			MinBytes:    10e3,
			MaxBytes:    10e6,
		}),
		log:  log,
		conn: conn,
	}
}

// Start reading and send messages to Clickhouse
func (s *AnalyticsService) Start(ctx context.Context) {
	go func() {
		for {
			msg, err := s.r.ReadMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					s.log.Info("consumer stopped")
					return
				}
				s.log.Error("kafka read error", sl.Err(err))
				continue
			}

			switch msg.Topic {
			case "user-events":
				event, err := parseUserEvent(msg.Value)
				if err != nil {
					s.log.Error("could not parse user-event message")
					continue
				}
				s.UserEventBuffer = append(s.UserEventBuffer, event)
			case "link-events":
				event, err := parseLinkEvent(msg.Value)
				if err != nil {
					s.log.Error("could not parse link-event message")
					continue
				}
				s.LinkEventBuffer = append(s.LinkEventBuffer, event)
			default:
				s.log.Warn("unknown topic", slog.String("topic", msg.Topic))
				continue
			}

			shouldFlush := len(s.UserEventBuffer)+len(s.LinkEventBuffer) >= maxBatchSize || time.Since(s.lastFlush) >= maxBatchAge

			if shouldFlush {
				if err := s.flush(); err != nil {
					s.log.Error("failed flush", sl.Err(err))
					continue
				}
				s.log.Info("successful flush: messages have been sent")
			}
		}
	}()
}

// flush prepares and sends batch,
// resets messages counter and sets time of the last batch sending
func (s *AnalyticsService) flush() error {
	ctx := context.Background()

	if len(s.UserEventBuffer) > 0 {
		if err := ch.InsertUserEvents(ctx, s.conn, s.UserEventBuffer); err != nil {
			return err
		}
		s.UserEventBuffer = s.UserEventBuffer[:0]
	}

	if len(s.LinkEventBuffer) > 0 {
		if err := ch.InsertLinkEvents(ctx, s.conn, s.LinkEventBuffer); err != nil {
			return err
		}
		s.LinkEventBuffer = s.LinkEventBuffer[:0]
	}

	s.lastFlush = time.Now()

	return nil
}

func parseUserEvent(data []byte) (ch.UserEvent, error) {
	var raw ch.UserEvent

	if err := json.Unmarshal(data, &raw); err != nil {
		return ch.UserEvent{}, err
	}

	return ch.UserEvent{
		Type:      raw.Type,
		UserID:    raw.UserID,
		Email:     raw.Email,
		Ip:        raw.Ip,
		Timestamp: raw.Timestamp,
		RawJSON:   string(data),
	}, nil
}

func parseLinkEvent(data []byte) (ch.LinkEvent, error) {
	var raw ch.LinkEvent

	if err := json.Unmarshal(data, &raw); err != nil {
		return ch.LinkEvent{}, err
	}

	return ch.LinkEvent{
		Type:      raw.Type,
		LinkID:    raw.LinkID,
		UserID:    raw.UserID,
		Alias:     raw.Alias,
		URL:       raw.URL,
		Timestamp: raw.Timestamp,
		RawJSON:   string(data),
	}, nil
}

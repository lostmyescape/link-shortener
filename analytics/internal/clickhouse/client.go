package clickhouse

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/lostmyescape/link-shortener/analytics/internal/config"
	"github.com/lostmyescape/link-shortener/common/logger/sl"
)

func NewClickhouseClient(ctx context.Context, log *slog.Logger, cfg *config.Config) clickhouse.Conn {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			panic("timeout waiting for clickhouse")
		case <-ticker.C:
			conn, err := connect(ctx, cfg)
			if err == nil {
				log.Info("clickhouse connected successfully")
				return conn
			}
			log.Debug("clickhouse not ready, retrying...", sl.Err(err))
		}
	}
}

func connect(ctx context.Context, cfg *config.Config) (clickhouse.Conn, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"clickhouse:8123"},
		Auth: clickhouse.Auth{
			Database: cfg.Clickhouse.Database,
			Username: cfg.Clickhouse.Username,
			Password: cfg.Clickhouse.Password,
		},
		Protocol: clickhouse.HTTP,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		return nil, fmt.Errorf("error connection to clickhouse: %w", err)
	}

	if err = conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("clickhouse ping failed: %w", err)
	}

	return conn, nil
}

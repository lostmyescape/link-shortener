package clickhouse

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
)

func InsertUserEvents(ctx context.Context, conn clickhouse.Conn, events []UserEvent) error {
	batch, err := conn.PrepareBatch(ctx,
		`INSERT INTO default.user_events (event_type, user_id, email, ip, ts, raw)`,
	)
	if err != nil {
		return err
	}

	for _, e := range events {
		if err := batch.Append(e.Type, e.UserID, e.Email, e.Ip, e.Timestamp, e.RawJSON); err != nil {
			return err
		}
	}

	return batch.Send()
}

func InsertLinkEvents(ctx context.Context, conn clickhouse.Conn, events []LinkEvent) error {
	batch, err := conn.PrepareBatch(ctx,
		`INSERT INTO default.link_events (event_type, link_id, user_id, alias, target_url, ts, raw)`,
	)
	if err != nil {
		return err
	}

	for _, e := range events {
		if err := batch.Append(e.Type, e.LinkID, e.UserID, e.Alias, e.URL, e.Timestamp, e.RawJSON); err != nil {
			return err
		}
	}

	return batch.Send()
}

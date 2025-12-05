package clickhouse

import (
	"time"
)

type UserEvent struct {
	Type      string      `json:"type" ch:"event_type"`
	UserID    uint64      `json:"user_id" ch:"user_id"`
	Email     string      `json:"email" ch:"email"`
	Ip        string      `json:"ip" ch:"ip"`
	Timestamp time.Time   `json:"timestamp" ch:"ts"`
	RawJSON   interface{} `json:"raw_json" ch:"raw"`
}

type LinkEvent struct {
	Type      string      `json:"type" ch:"event_type"`
	LinkID    uint64      `json:"link_id" ch:"link_id"`
	UserID    uint64      `json:"user_id" ch:"user_id"`
	Alias     string      `json:"alias" ch:"alias"`
	URL       string      `json:"url" ch:"target_url"`
	Timestamp time.Time   `json:"timestamp" ch:"ts"`
	RawJSON   interface{} `json:"raw_json" ch:"raw"`
}

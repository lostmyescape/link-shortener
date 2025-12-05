CREATE TABLE IF NOT EXISTS default.link_events
(
    event_type String,
    link_id UInt64,
    user_id UInt64,
    alias String,
    target_url String,
    ts DateTime DEFAULT now(),
    raw String
)

ENGINE = MergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (link_id, ts)
TTL ts + INTERVAL 30 DAY;
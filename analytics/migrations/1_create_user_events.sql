CREATE TABLE IF NOT EXISTS default.user_events
(
    event_type String,
    user_id UInt64,
    email String,
    ip String, 
    ts DateTime DEFAULT now(),
    raw String
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(ts)
ORDER BY (user_id, ts)
TTL ts + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS user_logs  (
  dt DateTime,
  request_id UInt64,
  client_id UInt64,
  message String
) Engine = MergeTree()
PARTITION BY toYYYYMM(dt)
ORDER BY (dt, request_id, client_id);

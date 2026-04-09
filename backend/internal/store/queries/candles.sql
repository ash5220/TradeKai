-- name: InsertCandle :exec
INSERT INTO candles (symbol, interval, ts, open, high, low, close, volume)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (symbol, interval, ts) DO UPDATE
SET high   = GREATEST(candles.high, EXCLUDED.high),
    low    = LEAST(candles.low, EXCLUDED.low),
    close  = EXCLUDED.close,
    volume = candles.volume + EXCLUDED.volume;

-- name: ListCandles :many
SELECT * FROM candles
WHERE symbol   = $1
  AND interval = $2
  AND ts BETWEEN $3 AND $4
ORDER BY ts ASC
LIMIT $5;

-- name: GetLatestCandle :one
SELECT * FROM candles
WHERE symbol   = $1
  AND interval = $2
ORDER BY ts DESC
LIMIT 1;

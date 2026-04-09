-- name: UpsertPosition :one
INSERT INTO positions (user_id, symbol, qty, avg_price, realized_pnl, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (user_id, symbol) DO UPDATE
SET qty          = EXCLUDED.qty,
    avg_price    = EXCLUDED.avg_price,
    realized_pnl = EXCLUDED.realized_pnl,
    updated_at   = NOW()
RETURNING *;

-- name: GetPosition :one
SELECT * FROM positions
WHERE user_id = $1
  AND symbol  = $2
LIMIT 1;

-- name: ListPositionsByUser :many
SELECT * FROM positions
WHERE user_id = $1
  AND qty     > 0
ORDER BY symbol;

-- name: DeletePosition :exec
DELETE FROM positions
WHERE user_id = $1
  AND symbol  = $2;

-- name: GetDailyRealizedPnL :one
SELECT COALESCE(SUM(
    CASE side
        WHEN 'buy'  THEN -(qty * price)
        WHEN 'sell' THEN   qty * price
    END
), 0)::float8 AS daily_pnl
FROM trades
WHERE user_id    = $1
  AND executed_at >= date_trunc('day', NOW() AT TIME ZONE 'UTC');

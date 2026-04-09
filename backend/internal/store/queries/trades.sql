-- name: CreateTrade :one
INSERT INTO trades (order_id, user_id, symbol, side, qty, price, commission, strategy)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: ListTradesByUser :many
SELECT * FROM trades
WHERE user_id = $1
ORDER BY executed_at DESC
LIMIT $2
OFFSET $3;

-- name: ListTradesByUserAndSymbol :many
SELECT * FROM trades
WHERE user_id = $1
  AND symbol  = $2
ORDER BY executed_at DESC
LIMIT $3
OFFSET $4;

-- name: CreateOrder :one
INSERT INTO orders (user_id, symbol, side, type, qty, limit_price, status, idempotency_key)
VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7)
RETURNING *;

-- name: GetOrderByID :one
SELECT * FROM orders
WHERE id = $1
LIMIT 1;

-- name: GetOrderByIdempotencyKey :one
SELECT * FROM orders
WHERE idempotency_key = $1
LIMIT 1;

-- name: ListOrdersByUser :many
SELECT * FROM orders
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;

-- name: ListOpenOrdersByUser :many
SELECT * FROM orders
WHERE user_id = $1
  AND status IN ('pending', 'submitted', 'partially_filled')
ORDER BY created_at DESC;

-- name: UpdateOrderStatus :one
UPDATE orders
SET status      = $2,
    exchange_id = COALESCE($3, exchange_id),
    updated_at  = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateOrderFill :one
UPDATE orders
SET status           = $2,
    filled_qty       = $3,
    filled_avg_price = $4,
    filled_at        = $5,
    updated_at       = NOW()
WHERE id = $1
RETURNING *;

-- name: CountOpenOrdersByUser :one
SELECT COUNT(*) FROM orders
WHERE user_id = $1
  AND status IN ('pending', 'submitted', 'partially_filled');

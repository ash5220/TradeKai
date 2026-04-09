CREATE TYPE order_side   AS ENUM ('buy', 'sell');
CREATE TYPE order_type   AS ENUM ('market', 'limit', 'stop');
CREATE TYPE order_status AS ENUM (
    'pending', 'submitted', 'partially_filled',
    'filled', 'cancelled', 'rejected'
);

CREATE TABLE IF NOT EXISTS orders (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    symbol           TEXT         NOT NULL,
    side             order_side   NOT NULL,
    type             order_type   NOT NULL,
    qty              NUMERIC(18,8) NOT NULL,
    limit_price      NUMERIC(18,8),
    filled_qty       NUMERIC(18,8) NOT NULL DEFAULT 0,
    filled_avg_price NUMERIC(18,8),
    status           order_status NOT NULL DEFAULT 'pending',
    exchange_id      TEXT,
    idempotency_key  TEXT         NOT NULL UNIQUE,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    filled_at        TIMESTAMPTZ
);

CREATE INDEX idx_orders_user_id        ON orders (user_id);
CREATE INDEX idx_orders_symbol         ON orders (symbol);
CREATE INDEX idx_orders_status         ON orders (status);
CREATE INDEX idx_orders_created_at     ON orders (created_at DESC);

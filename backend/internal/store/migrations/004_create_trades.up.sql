-- Audit log of all executed (filled) trades
CREATE TABLE IF NOT EXISTS trades (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id     UUID         NOT NULL REFERENCES orders(id),
    user_id      UUID         NOT NULL REFERENCES users(id),
    symbol       TEXT         NOT NULL,
    side         order_side   NOT NULL,
    qty          NUMERIC(18,8) NOT NULL,
    price        NUMERIC(18,8) NOT NULL,
    commission   NUMERIC(18,8) NOT NULL DEFAULT 0,
    strategy     TEXT,
    executed_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_trades_user_id     ON trades (user_id);
CREATE INDEX idx_trades_symbol      ON trades (symbol);
CREATE INDEX idx_trades_executed_at ON trades (executed_at DESC);

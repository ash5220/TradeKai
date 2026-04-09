CREATE TABLE IF NOT EXISTS positions (
    user_id        UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    symbol         TEXT         NOT NULL,
    qty            NUMERIC(18,8) NOT NULL DEFAULT 0,
    avg_price      NUMERIC(18,8) NOT NULL DEFAULT 0,
    realized_pnl   NUMERIC(18,8) NOT NULL DEFAULT 0,
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, symbol)
);

CREATE INDEX idx_positions_user_id ON positions (user_id);

-- Candle (OHLCV) time-series — converted to a TimescaleDB hypertable
CREATE TABLE IF NOT EXISTS candles (
    symbol     TEXT         NOT NULL,
    interval   TEXT         NOT NULL,   -- '1m', '5m', '1h', etc.
    ts         TIMESTAMPTZ  NOT NULL,   -- start of the interval
    open       NUMERIC(18,8) NOT NULL,
    high       NUMERIC(18,8) NOT NULL,
    low        NUMERIC(18,8) NOT NULL,
    close      NUMERIC(18,8) NOT NULL,
    volume     NUMERIC(18,8) NOT NULL DEFAULT 0,
    PRIMARY KEY (symbol, interval, ts)
);

SELECT create_hypertable('candles', 'ts', if_not_exists => TRUE);

-- Compress chunks older than 7 days
ALTER TABLE candles SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol,interval'
);

SELECT add_compression_policy('candles', INTERVAL '7 days', if_not_exists => TRUE);

-- Tick time-series
CREATE TABLE IF NOT EXISTS ticks (
    symbol    TEXT         NOT NULL,
    price     NUMERIC(18,8) NOT NULL,
    volume    NUMERIC(18,8) NOT NULL DEFAULT 0,
    ts        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

SELECT create_hypertable('ticks', 'ts', if_not_exists => TRUE);

ALTER TABLE ticks SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol'
);

SELECT add_compression_policy('ticks', INTERVAL '1 day', if_not_exists => TRUE);

-- Indexes for range queries
CREATE INDEX IF NOT EXISTS idx_candles_symbol_interval_ts ON candles (symbol, interval, ts DESC);
CREATE INDEX IF NOT EXISTS idx_ticks_symbol_ts            ON ticks (symbol, ts DESC);

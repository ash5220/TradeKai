# Plan: TradeKai – Real-Time Algorithmic Trading Platform

**TL;DR**: Build a production-grade, multi-user algorithmic trading platform using Go (Gin) + Angular 21 + PostgreSQL/TimescaleDB. Architecture follows clean/hexagonal design with interface-driven abstractions, event-driven internals, and WebSocket real-time updates. Connects to Alpaca for paper trading with a simulated mode for development. Designed as a portfolio project that demonstrates strong engineering: clean architecture, proper concurrency patterns, testability, observability, and security.

---

## Tech Stack (Refined)

| Layer             | Technology                                       | Rationale                                                                                                  |
| ----------------- | ------------------------------------------------ | ---------------------------------------------------------------------------------------------------------- |
| Backend framework | **Gin v1.12**                                    | Standard net/http, official Go tutorial endorsement, widest hiring manager recognition                     |
| Database          | **PostgreSQL 16 + TimescaleDB**                  | TimescaleDB hypertables for time-series tick/candle data; standard PostgreSQL for users, orders, positions |
| DB access         | **sqlc v1.30**                                   | Compile-time SQL validation, zero-reflection generated code, full control over TimescaleDB functions       |
| Migrations        | **golang-migrate**                               | Industry-standard, supports PostgreSQL                                                                     |
| Logging           | **Zap**                                          | Structured, zero-alloc logging (Logrus is in maintenance mode)                                             |
| Config            | **Viper**                                        | File + env var config, widely adopted                                                                      |
| Auth              | **JWT (golang-jwt/jwt)** + **bcrypt**            | Multi-user with hashed passwords                                                                           |
| API docs          | **swag (swaggo/swag)**                           | Auto-generates OpenAPI from Go annotations                                                                 |
| WebSocket         | **gorilla/websocket**                            | Standard Go WebSocket library                                                                              |
| Frontend          | **Angular 21** (standalone components + signals) | Latest patterns, no NgModules                                                                              |
| Charting          | **Lightweight Charts (TradingView)**             | Purpose-built financial charts, candlestick support                                                        |
| Exchange          | **Alpaca API** (paper trading) + simulated mode  | Real brokerage API with free paper trading; simulated mode for dev/testing                                 |
| Containerization  | **Docker + docker-compose**                      | Local dev + deployment                                                                                     |
| Observability     | **OpenTelemetry + Prometheus**                   | Modern standard for metrics and tracing                                                                    |

---

## Project Structure

```
TradeKai/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go                # Entry point: wire dependencies, start server
│   ├── internal/
│   │   ├── domain/                     # Core business entities + port interfaces (no external deps)
│   │   │   ├── order.go               # Order, OrderStatus, OrderSide, OrderType
│   │   │   ├── position.go            # Position, PositionSummary
│   │   │   ├── market.go              # Tick, Candle, Symbol, Quote
│   │   │   ├── signal.go              # TradeSignal, SignalType
│   │   │   ├── user.go                # User entity
│   │   │   ├── errors.go             # Domain-specific errors
│   │   │   └── ports.go              # Port interfaces: MarketDataProvider, OrderExecutor, Strategy, RiskRule, Indicator
│   │   ├── market/                     # Market Data adapters (implement domain.MarketDataProvider)
│   │   │   ├── alpaca.go             # AlpacaProvider: Alpaca WebSocket implementation
│   │   │   ├── simulator.go          # SimulatedProvider: simulated market data (dev/testing)
│   │   │   ├── aggregator.go         # Tick → Candle aggregation
│   │   │   └── hub.go                # Fan-out to subscribers (channels)
│   │   ├── strategy/                   # Strategy adapters (implement domain.Strategy)
│   │   │   ├── engine.go             # Engine: orchestrator, receives market data, runs strategies
│   │   │   ├── rsi.go                # RSIStrategy implementation
│   │   │   ├── macd.go               # MACDCrossoverStrategy implementation
│   │   │   └── indicator/            # Indicator adapters (implement domain.Indicator)
│   │   │       ├── rsi.go
│   │   │       ├── macd.go
│   │   │       ├── ema.go
│   │   │       └── sma.go
│   │   ├── order/                      # Order Management System (implements domain.OrderExecutor)
│   │   │   ├── service.go            # Service: order lifecycle management
│   │   │   ├── alpaca_executor.go    # AlpacaExecutor: Alpaca order execution
│   │   │   ├── simulated_executor.go # SimulatedExecutor: simulated fills (dev/testing)
│   │   │   └── retry.go              # Exponential backoff + circuit breaker
│   │   ├── risk/                       # Risk Management (implements domain.RiskRule)
│   │   │   ├── manager.go            # Manager: runs all rules, returns first failure
│   │   │   └── rules.go              # MaxPositionRule, DailyLossRule, DuplicateTradeRule, etc.
│   │   ├── auth/                       # Authentication
│   │   │   ├── service.go            # Service: Register, Login, token generation
│   │   │   ├── jwt.go                # JWT creation + validation
│   │   │   └── middleware.go         # Gin auth middleware
│   │   ├── user/                       # User management
│   │   │   └── service.go            # Service: User CRUD, preferences
│   │   ├── handler/                    # HTTP handlers (Gin)
│   │   │   ├── auth.go               # POST /auth/register, /auth/login
│   │   │   ├── market.go             # GET /market/symbols, /market/quotes
│   │   │   ├── order.go              # POST /orders, GET /orders, DELETE /orders/:id
│   │   │   ├── strategy.go           # GET /strategies, POST /strategies/:id/start
│   │   │   ├── portfolio.go          # GET /portfolio/positions, /portfolio/pnl
│   │   │   ├── system.go             # GET /health, /metrics
│   │   │   └── ws.go                 # WebSocket upgrade handler
│   │   ├── middleware/                 # Custom Gin middleware
│   │   │   ├── cors.go
│   │   │   ├── ratelimit.go          # Token bucket rate limiter
│   │   │   └── requestid.go          # Request ID for tracing
│   │   ├── ws/                         # WebSocket management
│   │   │   ├── hub.go                # Connection registry, room-based pub/sub
│   │   │   └── client.go             # Per-connection read/write pumps
│   │   ├── store/                      # Database layer
│   │   │   ├── postgres.go           # Connection pool setup
│   │   │   ├── queries/              # SQL files for sqlc
│   │   │   │   ├── users.sql
│   │   │   │   ├── orders.sql
│   │   │   │   ├── positions.sql
│   │   │   │   ├── candles.sql
│   │   │   │   └── trades.sql
│   │   │   ├── migrations/           # golang-migrate SQL files
│   │   │   │   ├── 001_create_users.up.sql
│   │   │   │   ├── 001_create_users.down.sql
│   │   │   │   ├── 002_create_orders.up.sql
│   │   │   │   ├── ...
│   │   │   │   └── 005_create_hypertables.up.sql  # TimescaleDB
│   │   │   └── generated/            # sqlc generated code (DO NOT EDIT)
│   │   │       ├── db.go
│   │   │       ├── models.go
│   │   │       └── queries.sql.go
│   │   └── config/                     # App configuration
│   │       └── config.go             # Viper-based config loading
│   ├── api/
│   │   └── openapi.yaml               # OpenAPI spec (auto-generated by swag)
│   ├── sqlc.yaml                       # sqlc configuration
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── app/
│   │   │   ├── app.component.ts
│   │   │   ├── app.config.ts          # provideRouter, provideHttpClient
│   │   │   ├── app.routes.ts
│   │   │   ├── core/
│   │   │   │   ├── auth/
│   │   │   │   │   ├── auth.service.ts
│   │   │   │   │   ├── auth.guard.ts
│   │   │   │   │   └── auth.interceptor.ts   # Attach JWT to requests
│   │   │   │   ├── websocket/
│   │   │   │   │   └── websocket.service.ts   # Reconnecting WebSocket
│   │   │   │   └── api/
│   │   │   │       └── api.service.ts
│   │   │   ├── features/
│   │   │   │   ├── dashboard/
│   │   │   │   │   ├── dashboard.component.ts # Main view: charts + positions + PnL
│   │   │   │   │   ├── dashboard.routes.ts
│   │   │   │   │   └── components/
│   │   │   │   │       ├── price-chart.component.ts      # Lightweight Charts
│   │   │   │   │       ├── position-table.component.ts
│   │   │   │   │       ├── pnl-card.component.ts
│   │   │   │   │       └── system-health.component.ts
│   │   │   │   ├── trading/
│   │   │   │   │   ├── trading.component.ts   # Strategy control panel
│   │   │   │   │   ├── trading.routes.ts
│   │   │   │   │   ├── store/
│   │   │   │   │   │   └── trading.store.ts   # Signal-based state
│   │   │   │   │   └── components/
│   │   │   │   │       ├── strategy-list.component.ts
│   │   │   │   │       ├── order-form.component.ts
│   │   │   │   │       └── trade-log.component.ts
│   │   │   │   ├── history/
│   │   │   │   │   ├── history.component.ts   # Trade history + filtering
│   │   │   │   │   └── history.routes.ts
│   │   │   │   └── auth/
│   │   │   │       ├── login.component.ts
│   │   │   │       └── register.component.ts
│   │   │   ├── shared/
│   │   │   │   ├── components/
│   │   │   │   │   ├── navbar.component.ts
│   │   │   │   │   └── loading-spinner.component.ts
│   │   │   │   ├── pipes/
│   │   │   │   │   └── currency.pipe.ts
│   │   │   │   └── models/                    # Shared TypeScript interfaces (not type aliases, not classes)
│   │   │   │       ├── order.model.ts         # export interface Order { ... }
│   │   │   │       ├── position.model.ts      # export interface Position { ... }
│   │   │   │       └── market.model.ts        # export interface Tick, Candle, Quote { ... }
│   │   ├── environments/
│   │   │   ├── environment.ts
│   │   │   └── environment.prod.ts
│   │   └── styles.scss
│   ├── angular.json
│   ├── tsconfig.json
│   └── Dockerfile
├── deployments/
│   ├── docker-compose.yml             # PostgreSQL + TimescaleDB + backend + frontend
│   ├── docker-compose.dev.yml         # Dev overrides (hot reload, debug ports)
│   └── nginx/
│       └── nginx.conf                 # Reverse proxy for prod
├── scripts/
│   ├── setup.sh                       # First-time setup
│   └── seed.sh                        # Seed DB with test data
├── Makefile                            # Build, test, migrate, generate commands
├── .github/
│   └── workflows/
│       └── ci.yml                     # GitHub Actions CI pipeline
├── .env.example
└── README.md
```

---

## Phase 1: Foundation + Market Data (Steps 1-6)

### Step 1: Project Scaffolding

- Initialize Go module (`go mod init github.com/<user>/tradekai`)
- Initialize Angular project with `ng new frontend --standalone --style=scss --routing`
- Create directory structure as defined above
- Create `Makefile` with targets: `build`, `test`, `lint`, `migrate-up`, `migrate-down`, `sqlc-generate`, `swagger`, `run`, `docker-up`
- Create `.env.example` with all configuration keys
- Create `docker-compose.dev.yml` with PostgreSQL + TimescaleDB

### Step 2: Configuration + Logging

- Implement `internal/config/config.go` using Viper
  - Load from `.env` file + environment variables
  - Config struct with nested sections: Server, Database, JWT, Alpaca, Risk
  - Validate required fields on startup
- Implement Zap logger initialization with:
  - JSON format for production, console format for development
  - Log level configurable via env
  - Request-scoped fields (request ID, user ID)

### Step 3: Database Setup

- Write migration files using golang-migrate:
  - `001_create_users`: users table (id UUID, email, password_hash, created_at, updated_at)
  - `002_create_orders`: orders table (id UUID, user_id FK, symbol, side, type, qty, price, status, created_at, filled_at)
  - `003_create_positions`: positions table (user_id, symbol, qty, avg_price, unrealized_pnl)
  - `004_create_trades`: trades table (audit log of all executed trades)
  - `005_create_hypertables`: Convert `candles` and `ticks` to TimescaleDB hypertables, add continuous aggregates for 1m/5m/1h candles
- Write sqlc queries for all CRUD operations
- Configure `sqlc.yaml` to generate Go code from SQL
- Set up connection pool with `pgxpool` (max connections, health checks)

### Step 4: Domain Models

- Define core domain types in `internal/domain/`:
  - `Order` with states: Pending → Submitted → PartiallyFilled → Filled / Cancelled / Rejected
  - `Position` with real-time PnL calculation
  - `Tick`, `Candle`, `Quote` for market data
  - `TradeSignal` with Buy/Sell/Hold + confidence score
  - `User` entity
  - Domain errors (ErrInsufficientFunds, ErrMaxPositionExceeded, ErrDuplicateOrder, etc.)
  - `ports.go`: Port interfaces consumed by the domain layer (defined here, implemented by adapters):
    - `MarketDataProvider`: `Connect`, `Subscribe`, `Close`
    - `OrderExecutor`: `PlaceOrder`, `CancelOrder`, `GetOrderStatus`
    - `Strategy`: `Name`, `RequiredIndicators`, `Evaluate`
    - `RiskRule`: `Check`
    - `Indicator`: streaming interface (feed one candle at a time, return updated value)

### Step 5: Market Data Service

- `MarketDataProvider` interface is defined in `domain/ports.go` (consumer: `strategy/engine.go`, `ws/hub.go`):
  ```
  Connect(ctx context.Context, symbols []string) error
  Subscribe(symbol string) (<-chan domain.Tick, error)
  Close() error
  ```
- Implement `SimulatedProvider`: generates realistic random walk tick data with configurable volatility, useful for development and testing
- Implement `AlpacaProvider`: connects to Alpaca's real-time WebSocket API for stock market data
- Implement `MarketHub` (fan-out pattern):
  - Receives ticks from provider
  - Maintains subscriber registry (map of symbol → []chan Tick)
  - Uses buffered channels with drop-oldest-on-full to prevent slow consumers from blocking
  - Tracks subscription count metrics
- Implement `CandleAggregator`:
  - Aggregates ticks into OHLCV candles at configurable intervals (1m, 5m, 1h)
  - Batch-writes completed candles to TimescaleDB

### Step 6: WebSocket Server

- Implement `ws/hub.go`:
  - Room-based pub/sub (rooms per symbol, per user)
  - Connection lifecycle: register → authenticate → subscribe → unsubscribe → deregister
  - Ping/pong heartbeat for connection health
  - Concurrent-safe with sync.RWMutex
- Implement `ws/client.go`:
  - Separate read/write goroutines per connection
  - Write pump with coalescing (batch multiple updates into single frame)
  - Graceful disconnect handling
- Wire market data hub → WebSocket hub so live ticks flow to connected clients

---

## Phase 2: Strategy Engine + Order Management (Steps 7-10)

### Step 7: Technical Indicators

- Implement in `internal/strategy/indicator/`:
  - `SMA(period int)` — Simple Moving Average
  - `EMA(period int)` — Exponential Moving Average (uses multiplier, not recalculate)
  - `RSI(period int)` — Relative Strength Index
  - `MACD(fast, slow, signal int)` — MACD + signal line + histogram
- All indicators implement streaming interface (feed one candle at a time, get updated value)
- Zero-allocation design: pre-allocate circular buffers, avoid slices

### Step 8: Strategy Engine

- `Strategy` interface is defined in `domain/ports.go` (consumer: `strategy/engine.go`):
  ```
  Name() string
  RequiredIndicators() []Indicator
  Evaluate(candle domain.Candle, indicators map[string]float64) domain.TradeSignal
  ```
- Implement `RSIStrategy`: Buy when RSI < 30 (oversold), Sell when RSI > 70 (overbought)
- Implement `MACDCrossoverStrategy`: Buy on bullish crossover, Sell on bearish crossover
- Implement `Engine` (`strategy/engine.go`):
  - Subscribes to candle updates from MarketHub
  - Runs configured strategies per symbol
  - Emits TradeSignals through channel
  - Uses worker pool pattern (bounded goroutines per symbol)
  - Supports start/stop per strategy per symbol

### Step 9: Order Management System

- `OrderExecutor` interface is defined in `domain/ports.go` (consumer: `order/service.go`):
  ```
  PlaceOrder(ctx context.Context, order domain.Order) (string, error)
  CancelOrder(ctx context.Context, orderID string) error
  GetOrderStatus(ctx context.Context, orderID string) (domain.OrderStatus, error)
  ```
- Implement `AlpacaExecutor` (`order/alpaca_executor.go`): Places real orders through Alpaca API
- Implement `SimulatedExecutor` (`order/simulated_executor.go`): Simulates fills with configurable latency/slippage
- Implement `Service` (`order/service.go`):
  - Receives TradeSignals from strategy engine
  - Passes through RiskManager before execution
  - Creates order record in DB (status=Pending)
  - Submits to executor with retry (exponential backoff, max 3 retries)
  - Updates order status on fill/reject
  - Idempotency: deduplication key = user_id + symbol + signal_id
  - Publishes order updates to WebSocket hub

### Step 10: Risk Management

- Implement `Manager` (`risk/manager.go`) with configurable rules:
  - **MaxPositionSize**: Reject if position would exceed N shares per symbol
  - **MaxOpenOrders**: Reject if user has > N pending orders
  - **DailyLossLimit**: Reject if realized + unrealized losses exceed threshold
  - **DuplicateTradeWindow**: Reject if same signal within N seconds
  - **MaxPortfolioExposure**: Reject if total portfolio value exceeds limit
- All rules implement `domain.RiskRule` interface (defined in `domain/ports.go`): `Check(ctx, order, portfolio) error`
- `Manager` runs all rules, returns first failure
- Log every risk check (pass/fail) for audit trail

---

## Phase 3: REST API + Authentication (Steps 11-14)

### Step 11: Authentication

- Implement `auth/service.go`:
  - `Register(email, password)`: bcrypt hash, store user, return JWT
  - `Login(email, password)`: verify hash, return JWT + refresh token
  - JWT with claims: user_id, email, exp (15 min access, 7 day refresh)
- Implement `auth/middleware.go`:
  - Extract JWT from Authorization header
  - Validate and inject user into Gin context
  - Return 401 on invalid/expired token

### Step 12: REST API Endpoints

- **Auth**: `POST /api/v1/auth/register`, `POST /api/v1/auth/login`, `POST /api/v1/auth/refresh`
- **Market**: `GET /api/v1/market/symbols`, `GET /api/v1/market/quotes/:symbol`, `GET /api/v1/market/candles/:symbol` (with query params: interval, from, to)
- **Orders**: `POST /api/v1/orders`, `GET /api/v1/orders` (paginated), `GET /api/v1/orders/:id`, `DELETE /api/v1/orders/:id`
- **Strategies**: `GET /api/v1/strategies`, `POST /api/v1/strategies/:id/start`, `POST /api/v1/strategies/:id/stop`, `PUT /api/v1/strategies/:id/config`
- **Portfolio**: `GET /api/v1/portfolio/positions`, `GET /api/v1/portfolio/pnl`, `GET /api/v1/portfolio/history`
- **System**: `GET /api/v1/health`, `GET /api/v1/metrics`
- **WebSocket**: `GET /ws` (upgrade with JWT in query param)
- All endpoints versioned under `/api/v1/`
- Use Swagger annotations for auto-generated API docs

### Step 13: Middleware Stack

- Request ID middleware (UUID per request, passed in context)
- CORS middleware (configured allowed origins)
- Rate limiter (token bucket: 100 req/min for API, 10 req/min for auth)
- Request logging (method, path, status, latency, request ID)
- Recovery middleware (catch panics, log stack trace, return 500)

### Step 14: Gin Router Setup

- Wire all handlers with dependency injection via constructors
- Group routes: public (auth, health) vs protected (everything else)
- Apply middleware stack in correct order
- Graceful shutdown: listen for SIGINT/SIGTERM, drain connections, close DB pool, close WebSocket hub

---

## Phase 4: Angular Dashboard (Steps 15-19)

### Step 15: Angular Scaffolding

- Generate Angular 21 project with standalone components
- Set up routing with lazy loading per feature
- Configure HttpClient with auth interceptor (attaches JWT)
- Configure environment files (API base URL, WebSocket URL)
- Install Lightweight Charts (`lightweight-charts` npm package)
- Install Angular Material or Tailwind CSS for UI components

### Step 16: Auth Feature

- `login.component.ts`: Email + password form, calls auth API, stores JWT in localStorage
- `register.component.ts`: Registration form with validation
- `auth.service.ts`: Signal-based auth state, manages tokens, auto-refresh
- `auth.guard.ts`: Redirects to login if not authenticated
- `auth.interceptor.ts`: Attaches `Authorization: Bearer <token>` to all API requests, handles 401 → redirect to login

### Step 17: Dashboard Feature

- `price-chart.component.ts`:
  - Integrates Lightweight Charts `createChart()`
  - Candlestick series from REST API (historical) + WebSocket (real-time updates)
  - Symbol selector dropdown
  - Time interval selector (1m, 5m, 1h, 1d)
  - Overlay indicators (SMA, EMA lines)
- `position-table.component.ts`:
  - Live table of open positions with real-time PnL updates via WebSocket
  - Color-coded profit/loss
  - Close position button
- `pnl-card.component.ts`:
  - Daily PnL, total PnL, win rate summary cards
  - Updated in real-time via signals
- `system-health.component.ts`:
  - Connection status indicators (API, WebSocket, exchange)
  - Uptime, order count, strategy status

### Step 18: Trading Feature

- `strategy-list.component.ts`:
  - List available strategies with status (running/stopped)
  - Start/stop controls per strategy per symbol
  - Configuration panel (parameters like RSI period, thresholds)
- `order-form.component.ts`:
  - Manual order entry (symbol, side, qty, type)
  - Form validation with Angular reactive forms
- `trade-log.component.ts`:
  - Real-time scrolling log of trade signals and order events
  - Filterable by symbol, strategy, status

### Step 19: WebSocket Service

- `websocket.service.ts`:
  - Auto-reconnecting WebSocket with exponential backoff
  - Subscribes to channels: `ticks:<symbol>`, `orders:<user_id>`, `signals:<user_id>`
  - Exposes `readonly` Angular signals via `asReadonly()` for each data stream
  - Connection state signal (connecting, connected, disconnected, error)
  - Uses `inject()` for DI — no constructor parameters
  - Heartbeat/ping handling

---

## Phase 5: Observability + Testing (Steps 20-23)

### Step 20: Structured Logging + Metrics

- Zap logger with:
  - Request-scoped fields via context
  - Trade audit log (separate logger for all order events)
  - Log rotation configuration
- Prometheus metrics:
  - `tradekai_order_execution_latency_seconds` (histogram)
  - `tradekai_market_data_ticks_total` (counter per symbol)
  - `tradekai_active_websocket_connections` (gauge)
  - `tradekai_strategy_signals_total` (counter per strategy)
  - `tradekai_risk_checks_total` (counter, pass/fail label)
  - `tradekai_order_status_total` (counter per status)
- Health check endpoint: DB connectivity, exchange connection, memory usage

### Step 21: Backend Testing

- **Unit tests** (no external dependencies):
  - All indicator calculations with known expected values
  - Strategy signal generation with mock candle data
  - Risk rule checks with various portfolio states
  - Order idempotency logic
  - JWT generation/validation
- **Integration tests** (using testcontainers-go):
  - Database operations against real PostgreSQL + TimescaleDB in Docker
  - API endpoint tests with httptest
  - WebSocket connection/subscription tests
- **Table-driven tests** for all indicator edge cases
- Use `github.com/google/go-cmp/cmp` with `cmp.Diff` for struct comparisons — not `reflect.DeepEqual`
- Use `errors.Is` / `errors.As` for error assertions — not string comparison
- Mark all test helper functions with `t.Helper()` so failure lines point to the caller
- Prefer `t.Error` over `t.Fatal` to surface all failures in one run; use `t.Fatal` only when subsequent assertions would be meaningless
- Test failure message format: `FuncName(input) = got, want expected`
- Target: >80% coverage on `internal/` packages

### Step 22: Frontend Testing

- Component tests with Angular TestBed
- Service tests for auth, API, WebSocket services
- E2E smoke tests with Playwright for critical flows (login → dashboard → view chart)

### Step 23: CI Pipeline

- GitHub Actions workflow:
  - Go: lint (golangci-lint), test, build
  - Angular: lint, test, build
  - Docker: build images, push to registry
  - Run integration tests with docker-compose

---

## Phase 6: Deployment + Production Hardening (Steps 24-26)

### Step 24: Docker Setup

- Backend `Dockerfile`: Multi-stage build (Go build → scratch/distroless runtime image)
- Frontend `Dockerfile`: Multi-stage (ng build → nginx serving static files)
- `docker-compose.yml`:
  - `db`: TimescaleDB (PostgreSQL + TimescaleDB extension)
  - `backend`: Go server
  - `frontend`: Angular + nginx
  - `nginx`: Reverse proxy (optional, or use frontend nginx)
- Healthchecks on all services
- Volume for DB persistence

### Step 25: Production Configuration

- HTTPS via nginx + Let's Encrypt (or self-signed for demo)
- Environment-based config (dev/staging/prod)
- Connection pool tuning for PostgreSQL
- Gin in release mode
- Proper CORS configuration for production domain
- Rate limiting tuned for production traffic

### Step 26: README + Documentation

- Architecture diagram (Mermaid or draw.io)
- Setup instructions (prerequisites, env vars, docker-compose up)
- API documentation link (Swagger UI)
- Screenshots of dashboard
- Key technical decisions explained
- Performance characteristics documented

---

## Architecture Patterns (Key Design Decisions)

### 1. Clean Architecture

- `domain/` has ZERO external imports — pure business logic
- All external dependencies (DB, exchange, HTTP) hidden behind interfaces in `internal/`
- Dependency injection via constructors (no framework needed)
- Flow: Handler → Service → Domain ← Repository (dependencies point inward)

### 2. Event-Driven Pipeline

```
Exchange WebSocket → MarketHub (fan-out via channels) →
  → CandleAggregator → TimescaleDB
  → StrategyEngine → TradeSignal channel →
    → RiskManager → OrderService → OrderExecutor
  → WebSocket Hub → Angular Dashboard
```

All components connected via Go channels. Each component runs in its own goroutine(s).

### 3. Concurrency Model

- Worker pool pattern for strategy evaluation (bounded goroutines)
- errgroup for goroutine lifecycle management and error propagation
- Context propagation for cancellation (graceful shutdown)
- Buffered channels with configurable sizes (default 1024 for market data, 256 for signals)
- sync.Pool for frequently allocated objects in hot paths

### 4. Error Handling

- Domain errors are typed (implement error interface with codes)
- Wrap errors with context using `fmt.Errorf("operation: %w", err)`
- Check errors using `errors.Is(err, target)` and `errors.As(err, &target)` — never compare with `==` directly (breaks through wrapped errors)
- Log at the boundary (handler level), not deep in business logic
- Circuit breaker for exchange connections (open after 5 consecutive failures, half-open after 30s)

### 5. Database Strategy

- TimescaleDB hypertables for candles/ticks (automatic partitioning, compression)
- Regular PostgreSQL tables for users, orders, positions
- Batch inserts for high-frequency candle data (accumulate 100 rows or 1 second, whichever first)
- Read replicas consideration for future scaling

---

### 6. Go Coding Conventions

- **Interfaces in `domain/ports.go`**: All port interfaces (`MarketDataProvider`, `OrderExecutor`, `Strategy`, `RiskRule`, `Indicator`) live in `domain/ports.go`. Implementing packages do not define the interfaces they satisfy — Go idiom: “the consumer defines the interface”
- **Accept interfaces, return concrete types**: Constructors return concrete types, not interfaces. `func NewService(...) *Service`, not `func NewService(...) ServiceInterface`
- **No stuttering names**: Types must not repeat their package name. `order.Service` (not `order.OrderService`), `risk.Manager` (not `risk.RiskManager`), `strategy.Engine` (not `strategy.StrategyEngine`)
- **Short receiver names**: 1–2 letter abbreviations. `func (s *Service) PlaceOrder(...)`, `func (e *Engine) Evaluate(...)`, `func (m *Manager) Check(...)`
- **Error strings**: Lowercase, no trailing punctuation. `"insufficient funds"` not `"Insufficient funds."` (sentinel names like `ErrInsufficientFunds` remain UpperCamelCase)
- **Error checking**: Use `errors.Is(err, target)` and `errors.As(err, &target)` — never `err == ErrSomething` (fails through wrapped errors)
- **No `context.Context` in structs**: Never store context in a struct field. Pass as first method parameter: `func (s *Service) PlaceOrder(ctx context.Context, ...)`
- **Import grouping** (enforced by `goimports`): stdlib → external packages → internal packages, each group separated by a blank line
- **Goroutine lifetimes**: Every goroutine must have a documented exit path; use `context.Context` cancellation + `errgroup`
- **Synchronous functions preferred**: Internal functions should be synchronous by default; callers add concurrency via goroutines when needed

### 7. Angular/TypeScript Coding Conventions

- **`inject()` for DI**: All Angular services and components use `inject()` function — not constructor parameter injection. `private readonly http = inject(HttpClient)` instead of `constructor(private http: HttpClient)`
- **`httpResource()` for reactive data**: Use Angular 21’s `httpResource()` for GET endpoints needing reactive signal data (`portfolioResource`, `strategiesResource`). Use `HttpClient` directly only for mutations (POST/PUT/DELETE)
- **Signal patterns** (Angular 21 standard):
  - `signal()` — writable local state
  - `computed()` — derived state that auto-updates
  - `linkedSignal()` — dependent state that resets when source changes
  - `resource()` / `httpResource()` — async reactive data with `.loading`, `.error`, `.value` status signals
  - `asReadonly()` — expose service state as read-only: `readonly positions = this._positions.asReadonly()`
- **Component member visibility**:
  - `protected` for members accessed only in the component’s own template (not `public`)
  - `readonly` for Angular-set properties: `readonly userId = input.required<string>()`
  - `private` for implementation details never accessed in the template
- **Named exports only**: No `export default`. All components, services, and pipes use named exports
- **`interface` over `type` for object shapes**: Model files use `interface Order { ... }` not `type Order = { ... }`
- **`import type` for type-only imports**: `import type { Order } from '../shared/models/order.model'`
- **No `I`-prefixed interfaces**: `interface AuthState` not `interface IAuthState`
- **Strict TypeScript**: `tsconfig.json` must enable `strict: true` — no implicit `any`; use `unknown` for truly unknown values

---

## Relevant Files (Critical Implementation References)

- `backend/cmd/server/main.go` — Wire all services, graceful shutdown with signal handling
- `backend/internal/domain/` — Pure domain types, zero external deps
- `backend/internal/domain/ports.go` — All port interfaces: `MarketDataProvider`, `OrderExecutor`, `Strategy`, `RiskRule`, `Indicator`
- `backend/internal/market/hub.go` — Fan-out pattern with channels
- `backend/internal/strategy/engine.go` — Worker pool, signal generation
- `backend/internal/order/service.go` — Order lifecycle, idempotency, retry
- `backend/internal/risk/manager.go` — Pre-trade risk checks
- `backend/internal/store/queries/` — SQL files for sqlc code generation
- `backend/internal/store/migrations/` — Database schema evolution
- `backend/internal/ws/hub.go` — WebSocket connection management
- `backend/internal/handler/` — All REST endpoint handlers
- `frontend/src/app/core/websocket/websocket.service.ts` — Reconnecting WS client
- `frontend/src/app/features/dashboard/components/price-chart.component.ts` — TradingView charts
- `docker-compose.yml` — Full stack orchestration

---

## Verification

1. `make test` — All unit tests pass with >80% coverage on `internal/`
2. `make integration-test` — Integration tests pass against Dockerized TimescaleDB
3. `make lint` — golangci-lint and ng lint pass with zero warnings
4. `make docker-up` — Full stack starts, health endpoints return 200
5. Manual: Open dashboard → see live chart updating → start RSI strategy → see signals in trade log → see orders placed
6. Manual: Register new user → login → verify JWT auth works → verify other user's data is isolated
7. `curl /api/v1/health` returns: DB connected, exchange connected, all services healthy
8. Swagger UI accessible at `/api/v1/docs` with all endpoints documented

---

## Decisions

- **Gin over Fiber**: Gin uses standard `net/http`, more recognizable to hiring managers, official Go tutorial endorsement. Fiber's fasthttp integration would be premature optimization for a portfolio project.
- **sqlc over GORM**: Type-safe generated code, zero reflection, full control over TimescaleDB functions. Shows SQL proficiency. GORM's ORM abstractions don't map well to time-series queries.
- **TimescaleDB**: Free PostgreSQL extension, demonstrates knowledge of time-series data handling — a differentiator for trading systems.
- **Alpaca for exchange**: Real brokerage API with free paper trading, more "serious" than crypto-only. Simulated mode for development means no API key needed to run locally.
- **Constructor-based DI over Wire** (Go backend): More readable, demonstrates understanding of dependency management without framework magic. Angular frontend uses `inject()` function (Angular 21 standard) — not constructor parameter injection.
- **Signals over NgRx**: Angular 21 signals are the modern standard, simpler than NgRx, and show current Angular knowledge.
- **Modular monolith**: Single deployable binary is simpler to demo and discuss. Internal boundaries via Go interfaces make future extraction trivial.

---

## Further Considerations

1. **Backtesting Engine**: Add a backtesting mode that replays historical data through strategies. Uses the same `Strategy` interface — just swap the `MarketDataProvider` for a historical data reader. Very impressive for interviews.
2. **Multi-exchange Support**: The `MarketDataProvider` and `OrderExecutor` interfaces make adding a second exchange (e.g., Binance) a straightforward second implementation. Good talking point about abstraction.
3. **Message Queue**: If extracting services later, NATS is lighter than Kafka and better suited for Go. Keep as a future optimization — channels are sufficient for the monolith.

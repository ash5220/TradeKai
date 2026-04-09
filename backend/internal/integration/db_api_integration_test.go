//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	"github.com/rashevskyv/tradekai/internal/auth"
	"github.com/rashevskyv/tradekai/internal/domain"
	"github.com/rashevskyv/tradekai/internal/handler"
	"github.com/rashevskyv/tradekai/internal/order"
	"github.com/rashevskyv/tradekai/internal/risk"
	"github.com/rashevskyv/tradekai/internal/store/generated"
	"github.com/rashevskyv/tradekai/internal/ws"
)

type integrationDB struct {
	pool      *pgxpool.Pool
	terminate func(context.Context) error
}

func setupIntegrationDB(t *testing.T) integrationDB {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("skipping testcontainers-backed integration DB tests on Windows")
	}

	ctx := context.Background()

	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: tc.ContainerRequest{
			Image:        "timescale/timescaledb:latest-pg16",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "tradekai",
				"POSTGRES_PASSWORD": "tradekai",
				"POSTGRES_DB":       "tradekai_test",
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(2 * time.Minute),
		},
		Started: true,
	})
	if err != nil {
		t.Skipf("skipping integration DB tests: cannot start testcontainers (%v)", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("setupIntegrationDB() host: %v", err)
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("setupIntegrationDB() mapped port: %v", err)
	}

	dsn := fmt.Sprintf("postgres://tradekai:tradekai@%s:%s/tradekai_test?sslmode=disable", host, port.Port())
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("setupIntegrationDB() connect pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("setupIntegrationDB() ping db: %v", err)
	}

	if _, err := pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS timescaledb;`); err != nil {
		t.Fatalf("setupIntegrationDB() create timescaledb extension: %v", err)
	}

	if err := applyUpMigrations(ctx, pool); err != nil {
		t.Fatalf("setupIntegrationDB() apply migrations: %v", err)
	}

	return integrationDB{
		pool: pool,
		terminate: func(ctx context.Context) error {
			pool.Close()
			return container.Terminate(ctx)
		},
	}
}

func applyUpMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	glob := filepath.Join("..", "store", "migrations", "*.up.sql")
	files, err := filepath.Glob(glob)
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}
	sort.Strings(files)

	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", f, err)
		}
		if _, err := pool.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("exec migration %s: %w", f, err)
		}
	}

	return nil
}

func TestIntegrationSQLCUserAndOrderQueries(t *testing.T) {
	t.Helper()
	db := setupIntegrationDB(t)
	t.Cleanup(func() {
		if err := db.terminate(context.Background()); err != nil {
			t.Errorf("terminate container: %v", err)
		}
	})

	ctx := context.Background()
	q := generated.New(db.pool)

	user, err := q.CreateUser(ctx, generated.CreateUserParams{
		Email:        "integration@example.com",
		PasswordHash: "hash",
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	gotUser, err := q.GetUserByEmail(ctx, user.Email)
	if err != nil {
		t.Fatalf("GetUserByEmail() error = %v", err)
	}
	if diff := cmp.Diff(user.ID, gotUser.ID); diff != "" {
		t.Errorf("GetUserByEmail() id mismatch (-want +got):\n%s", diff)
	}

	createdOrder, err := q.CreateOrder(ctx, generated.CreateOrderParams{
		UserID:         user.ID,
		Symbol:         "AAPL",
		Side:           generated.OrderSideBuy,
		Type:           generated.OrderTypeMarket,
		Qty:            2,
		Status:         generated.OrderStatusPending,
		IdempotencyKey: "int-key-1",
	})
	if err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}

	gotOrder, err := q.GetOrderByID(ctx, createdOrder.ID)
	if err != nil {
		t.Fatalf("GetOrderByID() error = %v", err)
	}
	if diff := cmp.Diff(createdOrder.Symbol, gotOrder.Symbol); diff != "" {
		t.Errorf("GetOrderByID() symbol mismatch (-want +got):\n%s", diff)
	}
}

func TestIntegrationAuthEndpoints(t *testing.T) {
	t.Helper()
	db := setupIntegrationDB(t)
	t.Cleanup(func() {
		if err := db.terminate(context.Background()); err != nil {
			t.Errorf("terminate container: %v", err)
		}
	})

	gin.SetMode(gin.TestMode)
	jwtManager := auth.NewManager("integration-secret-012345678901234567890", 15*time.Minute, 7*24*time.Hour)
	authSvc := auth.NewService(db.pool, jwtManager)
	authHandler := handler.NewAuthHandler(authSvc)

	r := gin.New()
	v1 := r.Group("/api/v1")
	v1.POST("/auth/register", authHandler.Register)
	v1.POST("/auth/login", authHandler.Login)
	v1.POST("/auth/refresh", authHandler.Refresh)

	registerBody := map[string]string{"email": "user@example.com", "password": "StrongPass123!"}
	registerRec := performJSONRequest(t, r, http.MethodPost, "/api/v1/auth/register", registerBody)
	if registerRec.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want %d body=%s", registerRec.Code, http.StatusCreated, registerRec.Body.String())
	}

	var regResp map[string]string
	if err := json.Unmarshal(registerRec.Body.Bytes(), &regResp); err != nil {
		t.Fatalf("register json decode: %v", err)
	}
	if regResp["access_token"] == "" {
		t.Error("register access_token is empty")
	}
	if regResp["refresh_token"] == "" {
		t.Error("register refresh_token is empty")
	}

	loginBody := map[string]string{"email": "user@example.com", "password": "StrongPass123!"}
	loginRec := performJSONRequest(t, r, http.MethodPost, "/api/v1/auth/login", loginBody)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d body=%s", loginRec.Code, http.StatusOK, loginRec.Body.String())
	}

	var loginResp map[string]string
	if err := json.Unmarshal(loginRec.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("login json decode: %v", err)
	}

	refreshBody := map[string]string{"refresh_token": loginResp["refresh_token"]}
	refreshRec := performJSONRequest(t, r, http.MethodPost, "/api/v1/auth/refresh", refreshBody)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("refresh status = %d, want %d body=%s", refreshRec.Code, http.StatusOK, refreshRec.Body.String())
	}
}

func TestIntegrationOrderServiceIdempotency(t *testing.T) {
	t.Helper()
	db := setupIntegrationDB(t)
	t.Cleanup(func() {
		if err := db.terminate(context.Background()); err != nil {
			t.Errorf("terminate container: %v", err)
		}
	})

	ctx := context.Background()
	q := generated.New(db.pool)
	user, err := q.CreateUser(ctx, generated.CreateUserParams{
		Email:        "idempotency@example.com",
		PasswordHash: "hash",
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	logger := zap.NewNop()
	hub := ws.NewHub(logger)
	go hub.Run()

	riskManager := risk.NewManager(logger,
		risk.NewMaxPositionRule(1000000),
		risk.NewMaxOpenOrdersRule(1000),
		risk.NewDailyLossRule(1e9),
		risk.NewDuplicateTradeWindowRule(0),
		risk.NewMaxPortfolioExposureRule(1e12),
	)

	svc := order.NewService(
		order.NewSimulatedExecutor(0, 0),
		riskManager,
		db.pool,
		hub,
		logger,
		logger.Named("trade_audit"),
	)

	signalID := uuid.New()
	sig := domain.TradeSignal{
		ID:       signalID,
		Strategy: "test",
		Symbol:   "AAPL",
		Type:     domain.SignalBuy,
		Price:    123.45,
	}

	portfolio := domain.PortfolioSummary{UserID: user.ID}

	first, err := svc.PlaceFromSignal(ctx, user.ID, sig, portfolio)
	if err != nil {
		t.Fatalf("PlaceFromSignal(first) error = %v", err)
	}
	second, err := svc.PlaceFromSignal(ctx, user.ID, sig, portfolio)
	if err != nil {
		t.Fatalf("PlaceFromSignal(second) error = %v", err)
	}
	if diff := cmp.Diff(first.ID, second.ID); diff != "" {
		t.Errorf("idempotent order id mismatch (-want +got):\n%s", diff)
	}

	orders, err := q.ListOrdersByUser(ctx, generated.ListOrdersByUserParams{UserID: user.ID, Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListOrdersByUser() error = %v", err)
	}
	if len(orders) != 1 {
		t.Errorf("ListOrdersByUser() count = %d, want 1", len(orders))
	}
}

func performJSONRequest(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func readAll(t *testing.T, r io.Reader) string {
	t.Helper()
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("readAll() error = %v", err)
	}
	return string(b)
}

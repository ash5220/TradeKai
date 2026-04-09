// Package main is the entry point for the TradeKai backend server.
//
// @title TradeKai API
// @version 1.0
// @description Real-time algorithmic trading platform API
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @BasePath /api/v1
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	"github.com/rashevskyv/tradekai/internal/auth"
	"github.com/rashevskyv/tradekai/internal/config"
	"github.com/rashevskyv/tradekai/internal/domain"
	"github.com/rashevskyv/tradekai/internal/handler"
	"github.com/rashevskyv/tradekai/internal/market"
	"github.com/rashevskyv/tradekai/internal/middleware"
	"github.com/rashevskyv/tradekai/internal/order"
	"github.com/rashevskyv/tradekai/internal/risk"
	"github.com/rashevskyv/tradekai/internal/store"
	"github.com/rashevskyv/tradekai/internal/strategy"
	"github.com/rashevskyv/tradekai/internal/ws"
)

func main() {
	// ── Config ──────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	// ── Logger ──────────────────────────────────────────────────────────────
	log, err := config.NewLogger(cfg.Log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync() //nolint:errcheck

	// ── Database ─────────────────────────────────────────────────────────────
	ctx := context.Background()
	db, err := store.NewPool(ctx, cfg.Database)
	if err != nil {
		log.Fatal("database connection failed", zap.Error(err))
	}
	defer db.Close()
	log.Info("database connected")

	// ── JWT ──────────────────────────────────────────────────────────────────
	jwtManager := auth.NewManager(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)

	// ── Market data provider ──────────────────────────────────────────────────
	var marketProvider domain.MarketDataProvider
	switch cfg.Market.Provider {
	case "alpaca":
		marketProvider = market.NewAlpacaProvider(
			cfg.Alpaca.APIKey, cfg.Alpaca.APISecret, cfg.Alpaca.DataFeed, log)
	default:
		marketProvider = market.NewSimulatedProvider(0.001, 500*time.Millisecond)
		log.Info("using simulated market data provider")
	}

	marketHub := market.NewHub(marketProvider, log)

	// ── Order executor ────────────────────────────────────────────────────────
	var orderExecutor domain.OrderExecutor
	switch cfg.Order.Executor {
	case "alpaca":
		orderExecutor = order.NewAlpacaExecutor(
			cfg.Alpaca.APIKey, cfg.Alpaca.APISecret, cfg.Alpaca.BaseURL)
	default:
		orderExecutor = order.NewSimulatedExecutor(50*time.Millisecond, 0.001)
		log.Info("using simulated order executor")
	}

	// ── WebSocket hub ─────────────────────────────────────────────────────────
	wsHub := ws.NewHub(log)
	go wsHub.Run()

	// ── Risk manager ──────────────────────────────────────────────────────────
	riskManager := risk.NewManager(log,
		risk.NewMaxPositionRule(float64(cfg.Risk.MaxPositionSize)),
		risk.NewMaxOpenOrdersRule(cfg.Risk.MaxOpenOrders),
		risk.NewDailyLossRule(cfg.Risk.DailyLossLimit),
		risk.NewDuplicateTradeWindowRule(cfg.Risk.DuplicateTradeWindow),
		risk.NewMaxPortfolioExposureRule(cfg.Risk.MaxPortfolioExposure),
	)

	// ── Services ──────────────────────────────────────────────────────────────
	authSvc := auth.NewService(db, jwtManager)
	auditLog := log.Named("trade_audit")
	orderSvc := order.NewService(orderExecutor, riskManager, db, wsHub, log, auditLog)

	// ── Strategy engine ───────────────────────────────────────────────────────
	stratEngine := strategy.NewEngine(marketHub, log)
	strategies := []domain.Strategy{
		strategy.NewRSIStrategy(14, 30, 70),
		strategy.NewMACDCrossoverStrategy(12, 26, 9),
	}

	// ── Gin router ────────────────────────────────────────────────────────────
	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins))

	// Prometheus metrics (no auth)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger docs
	r.GET("/api/v1/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// WebSocket (auth handled inside handler via query param token)
	wsHandler := handler.NewWSHandler(wsHub, jwtManager)
	r.GET("/ws", wsHandler.Upgrade)

	// API v1
	v1 := r.Group("/api/v1")

	// Public routes
	public := v1.Group("")
	public.Use(middleware.RateLimit(cfg.Rate.Auth))
	{
		authHandler := handler.NewAuthHandler(authSvc)
		public.POST("/auth/register", authHandler.Register)
		public.POST("/auth/login", authHandler.Login)
		public.POST("/auth/refresh", authHandler.Refresh)
	}

	// System (no auth, separate rate limit)
	systemHandler := handler.NewSystemHandler(db, marketHub)
	v1.GET("/health", systemHandler.Health)

	// Protected routes
	protected := v1.Group("")
	protected.Use(auth.Middleware(jwtManager))
	protected.Use(middleware.RateLimit(cfg.Rate.API))
	{
		orderHandler := handler.NewOrderHandler(orderSvc, db)
		protected.POST("/orders", orderHandler.PlaceOrder)
		protected.GET("/orders", orderHandler.ListOrders)
		protected.GET("/orders/:id", orderHandler.GetOrder)
		protected.DELETE("/orders/:id", orderHandler.CancelOrder)

		stratHandler := handler.NewStrategyHandler(stratEngine, strategies)
		protected.GET("/strategies", stratHandler.List)
		protected.POST("/strategies/:id/start", stratHandler.Start)
		protected.POST("/strategies/:id/stop", stratHandler.Stop)

		portfolioHandler := handler.NewPortfolioHandler(db)
		protected.GET("/portfolio/positions", portfolioHandler.Positions)
		protected.GET("/portfolio/pnl", portfolioHandler.PnL)
		protected.GET("/portfolio/history", portfolioHandler.History)

		marketHandler := handler.NewMarketHandler(db)
		protected.GET("/market/candles/:symbol", marketHandler.Candles)
	}

	// ── HTTP server ───────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start market hub in background
	hubCtx, hubCancel := context.WithCancel(ctx)
	defaultSymbols := []string{"AAPL", "TSLA", "MSFT", "GOOGL"}
	go func() {
		if err := marketHub.Start(hubCtx, defaultSymbols); err != nil {
			log.Error("market hub stopped", zap.Error(err))
		}
	}()

	// Start HTTP server
	go func() {
		log.Info("server starting", zap.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	hubCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown error", zap.Error(err))
	}
	log.Info("server stopped")

	// Suppress unused import warning for orderSvc during development
	_ = orderSvc
	_ = stratEngine
}

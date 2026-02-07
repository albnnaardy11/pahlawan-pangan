package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	apiMiddleware "github.com/albnnaardy11/pahlawan-pangan/internal/api/middleware"
	"github.com/albnnaardy11/pahlawan-pangan/internal/audit"
	"github.com/albnnaardy11/pahlawan-pangan/internal/matching"
	"github.com/albnnaardy11/pahlawan-pangan/internal/messaging"
	"github.com/albnnaardy11/pahlawan-pangan/internal/outbox"
	"github.com/albnnaardy11/pahlawan-pangan/pkg/cache"
	"github.com/albnnaardy11/pahlawan-pangan/pkg/logger"

	// Repository
	surplusRepo "github.com/albnnaardy11/pahlawan-pangan/internal/surplus/repository/postgresql"

	// Usecase
	surplusUcase "github.com/albnnaardy11/pahlawan-pangan/internal/surplus/usecase"

	// Delivery
	surplusHttp "github.com/albnnaardy11/pahlawan-pangan/internal/surplus/delivery/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	// 1. Initialize Logger
	logger.Info("Starting Pahlawan Pangan Server", zap.String("version", "v1.0.0"))

	// 2. Database Connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/pahlawan_pangan?sslmode=disable"
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		os.Exit(1)
	}
	db.SetMaxOpenConns(100)
	defer db.Close()

	// 3. Redis Caching
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	_ = cache.NewRedisCache(redisAddr) // Initialized but used inside repos/usecases

	// 4. Message Broker (NATS)
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	nc, _ := nats.Connect(natsURL)
	defer nc.Close()
	natsPublisher, _ := messaging.NewNATSPublisher(nc)

	// 5. Dependency Injection (Layered Architecture)
	repo := surplusRepo.NewSurplusRepository(db, db)
	
	router := &MockRouter{} // Legacy or mock for now
	matchEngine := matching.NewMatchingEngine(router)
	
	timeoutContext := time.Duration(2) * time.Second
	usecase := surplusUcase.NewSurplusUsecase(repo, matchEngine, timeoutContext)

	// 6. HTTP Routing (Versioning)
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Resilience & Governance (Phase 5+)
	loadshedder := apiMiddleware.NewAdaptiveLoadShedder(500 * time.Millisecond)
	r.Use(loadshedder.Handle) // Adaptive Load Shedding
	r.Use(apiMiddleware.CanarySplitter(10, "v1.1.0-canary")) // 10% Canary Rollout

	// Metrics
	r.Handle("/metrics", promhttp.Handler())

	// Init Delivery
	surplusHttp.NewSurplusHandler(r, usecase)

	// 7. Start Outbox Poller
	outboxSvc := outbox.NewOutboxService(db)
	go func() {
		for {
			_ = outboxSvc.PollAndPublish(context.Background(), natsPublisher, 100)
			time.Sleep(1 * time.Second)
		}
	}()

	// 8. Start Reconciliation Engine (Midnight Audit Simulation)
	auditEngine := audit.NewReconciliationEngine()
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		for {
			select {
			case <-ticker.C:
				_, _ = auditEngine.RunAudit(context.Background())
			}
		}
	}()

	// 8. Server Setup
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("Listening on port 8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Listen error", zap.Error(err))
		}
	}()

	// Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}
}

type MockRouter struct{}

func (m *MockRouter) GetTravelTime(ctx context.Context, startLat, startLon, endLat, endLon float64) (time.Duration, error) {
	return 15 * time.Minute, nil
}

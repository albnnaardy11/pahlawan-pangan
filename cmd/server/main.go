package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	api "github.com/albnnaardy11/pahlawan-pangan/internal/api"
	apiMiddleware "github.com/albnnaardy11/pahlawan-pangan/internal/api/middleware"
	"github.com/albnnaardy11/pahlawan-pangan/internal/audit"
	"github.com/albnnaardy11/pahlawan-pangan/internal/geo"
	"github.com/albnnaardy11/pahlawan-pangan/internal/inventory"
	"github.com/albnnaardy11/pahlawan-pangan/internal/loyalty"
	"github.com/albnnaardy11/pahlawan-pangan/internal/matching"
	"github.com/albnnaardy11/pahlawan-pangan/internal/messaging"
	"github.com/albnnaardy11/pahlawan-pangan/internal/notifications"
	"github.com/albnnaardy11/pahlawan-pangan/internal/outbox"
	"github.com/albnnaardy11/pahlawan-pangan/internal/recommendation"
	"github.com/albnnaardy11/pahlawan-pangan/internal/trust"
	"github.com/albnnaardy11/pahlawan-pangan/internal/worker"

	// Logistics & Escrow Modules
	escrowService "github.com/albnnaardy11/pahlawan-pangan/internal/escrow/service"
	logisticsHttp "github.com/albnnaardy11/pahlawan-pangan/internal/logistics/delivery/http"
	logisticsService "github.com/albnnaardy11/pahlawan-pangan/internal/logistics/service"

	// Community Module
	communityHttp "github.com/albnnaardy11/pahlawan-pangan/internal/community/delivery/http"
	communityRepo "github.com/albnnaardy11/pahlawan-pangan/internal/community/repository"
	communityUsecase "github.com/albnnaardy11/pahlawan-pangan/internal/community/usecase"

	// Carbon Module (ESG)
	carbonHttp "github.com/albnnaardy11/pahlawan-pangan/internal/carbon/delivery/http"
	carbonRepo "github.com/albnnaardy11/pahlawan-pangan/internal/carbon/repository"
	carbonService "github.com/albnnaardy11/pahlawan-pangan/internal/carbon/service"

	// IAM & Security
	authHttp "github.com/albnnaardy11/pahlawan-pangan/internal/auth/delivery/http"
	iamMiddleware "github.com/albnnaardy11/pahlawan-pangan/internal/auth/middleware"
	authRepo "github.com/albnnaardy11/pahlawan-pangan/internal/auth/repository"
	authUsecase "github.com/albnnaardy11/pahlawan-pangan/internal/auth/usecase"

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
	"github.com/redis/go-redis/v9"
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
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database connection", zap.Error(err))
		}
	}()

	// 3. Redis Caching & Geo Engine
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	// Direct client for Geo Service
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer redisClient.Close()
	
	// Legacy cache init (if needed)
	_ = cache.NewRedisCache(redisAddr)

	// 4. Message Broker (NATS)
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	nc, err := nats.Connect(natsURL)
	if err != nil {
		logger.Error("Failed to connect to NATS", zap.Error(err))
		os.Exit(1)
	}
	defer nc.Close()
	natsPublisher, err := messaging.NewNATSPublisher(nc)
	if err != nil {
		logger.Error("Failed to create NATS publisher", zap.Error(err))
		os.Exit(1)
	}

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
	r.Use(apiMiddleware.ChaosMiddleware(logger.Log)) // 🐵 Chaos Monkey

	// Metrics
	r.Handle("/metrics", promhttp.Handler())

	// 7. Start Outbox Poller (Must be before API Handlers needing outbox)
	outboxSvc := outbox.NewOutboxService(db)
	outboxRepo := outbox.NewRepository(db)
	
	// 8. UNICORN PHASE 4 SERVICES
	// Loyalty Engine (Uses same Redis as Geo)
	loyaltySvc := loyalty.NewLoyaltyService(redisClient) // Reuse Redis
	
	// Inventory Webhook Service (Flash Ludes)
	inventorySvc := inventory.NewInventoryService(outboxRepo, logger.Log)
	
	// Trust & Safety (Credit Scoring)
	trustSvc := trust.NewTrustService()
	
	// Personalization Engine (Smart Nudges)
	recSvc := recommendation.NewRecommendationService()

	// 9. Init New API Handler (Unicorn Features)
	mainHandler := api.NewHandler(db, matchEngine, outboxSvc, loyaltySvc, inventorySvc, trustSvc, recSvc)
	
	// Mount API V1 Routes
	r.Mount("/", mainHandler.Routes())

	// 11. UNICORN LOGISTICS & ESCROW
	// Escrow (Financial Integrity)
	_ = escrowService.NewEscrowService() // In-memory demo

	// Batching & Dispatch (The Brain)
	batchEngine := logisticsService.NewBatchingEngine()
	dispatchSvc := logisticsService.NewDispatchService(batchEngine)
	
	// Start batch processor worker
	go dispatchSvc.RunBatchProcessor(context.Background())

	logisticsHandler := logisticsHttp.NewLogisticsHandler(dispatchSvc)
	r.Mount("/api/v1/logistics", logisticsHandler.Routes())

	// 12. UNICORN COMMUNITY (Social Proof)
	communityRepository := communityRepo.NewReviewRepository(db)
	communityUC := communityUsecase.NewCommunityUsecase(communityRepository)
	communityHandler := communityHttp.NewCommunityHandler(communityUC)
	r.Mount("/api/v1/community", communityHandler.Routes())

	// 13. UNICORN ESG (Sustainability - Blockchain Ready)
	carbonRepository := carbonRepo.NewCarbonRepository(db)
	carbonSvc := carbonService.NewCarbonService(carbonRepository)
	carbonHandler := carbonHttp.NewCarbonHandler(carbonSvc)
	r.Mount("/api/v1/carbon", carbonHandler.Routes())

	// 15. UNICORN IAM & SECURITY
	authenticationRepo := authRepo.NewPostgresUserRepository(db)
	authenticationUC := authUsecase.NewAuthUsecase(authenticationRepo, redisClient, natsPublisher, time.Second*5)
	authenticationHandler := authHttp.NewAuthHandler(authenticationUC)
	r.Mount("/api/v1/auth", authenticationHandler.Routes())

	// 16. UNICORN MFA WORKER (Async OTP)
	_, _ = nc.Subscribe("otp.request", func(m *nats.Msg) {
		logger.Info("📧 OTP Request Received", zap.ByteString("payload", m.Data))
		// Logika kirim SMS/Email Gateway di sini
	})

	// Apply Global Auth Middleware
	r.Use(iamMiddleware.AuthMiddleware(authenticationUC))

	// Init Delivery (Existing Core Logic)
	surplusHttp.NewSurplusHandler(r, usecase)

	// 14. START BACKGROUND WORKERS
	// Outbox Poller
	go func() {
		for {
			_ = outboxSvc.PollAndPublish(context.Background(), natsPublisher, 100)
			time.Sleep(1 * time.Second)
		}
	}()

	// Carbon Impact Ledger Worker (Blockchain-Ready)
	carbonWorker := worker.NewCarbonWorker(carbonSvc, nc, logger.Log)
	go func() {
		if err := carbonWorker.Start(context.Background()); err != nil {
			logger.Error("CarbonWorker failed to start", zap.Error(err))
		}
	}()

	// 8. Start Reconciliation Engine (Midnight Audit Simulation)
	auditEngine := audit.NewReconciliationEngine()
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			_, _ = auditEngine.RunAudit(context.Background())
		}
	}()

	// 8. Server Setup
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// 9. Unicorn Logic: Real-time Notification Engine
	geoSvc := geo.NewGeoService(redisClient) // Use the initialized redisClient
	notifSvc := &notifications.NotificationService{} // In real app, inject FCM client here
	
	notifierWorker := worker.NewSurplusNotifier(geoSvc, notifSvc)
	go notifierWorker.Run(context.Background(), nc)

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

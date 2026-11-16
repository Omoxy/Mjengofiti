package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mjengofiti/api/internal/audit"
	"github.com/mjengofiti/api/internal/auth"
	"github.com/mjengofiti/api/internal/boundaries"
	"github.com/mjengofiti/api/internal/buildings"
	"github.com/mjengofiti/api/internal/config"
	"github.com/mjengofiti/api/internal/database"
	"github.com/mjengofiti/api/internal/datasets"
	"github.com/mjengofiti/api/internal/health"
	mw "github.com/mjengofiti/api/internal/http/middleware"
	"github.com/mjengofiti/api/internal/indicators"
	"github.com/mjengofiti/api/internal/ingestion"
	"github.com/mjengofiti/api/internal/inspections"
	"github.com/mjengofiti/api/internal/standards"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("application failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	logger.Info("starting Mjengofiti API",
		slog.String("env", cfg.App.Env),
		slog.Int("port", cfg.App.Port),
	)

	// Connect to database
	var db *database.Pool
	var pgPool *pgxpool.Pool
	db, err = database.Connect(ctx, cfg.Database)
	if err != nil {
		logger.Warn("database connection failed, running in degraded mode",
			slog.String("error", err.Error()),
		)
	} else {
		logger.Info("database connected successfully")
		pgPool = db.Pool
		defer db.Close()
	}

	// Initialize services
	healthSvc := health.NewService(db)

	// Build router
	r := chi.NewRouter()

	// Global middleware
	r.Use(mw.RequestID)
	r.Use(mw.Logger(logger))
	r.Use(mw.SecurityHeaders)
	r.Use(mw.CORS(cfg.App.CORSOrigins))
	r.Use(mw.Recover(logger))
	rateLimiter := mw.NewRateLimiter(cfg.RateLimit.RequestsPerMinute, cfg.RateLimit.Burst)
	r.Use(rateLimiter.Middleware)
	r.Use(auth.Middleware(auth.MiddlewareConfig{
		Environment: cfg.App.Env,
		Enabled:     cfg.Auth.Enabled,
		Secret:      cfg.Auth.JWTSecret,
		Issuer:      cfg.Auth.IssuerURL,
		Audience:    cfg.Auth.Audience,
	}))

	// Health endpoints (always available)
	r.Get("/health/live", healthSvc.HandleLive)
	r.Get("/health/ready", healthSvc.HandleReady)
	r.With(auth.RequirePermission("system:read")).Get("/api/v1/system/status", healthSvc.HandleSystemStatus)

	// Map configuration endpoint
	r.Get("/api/v1/config/map", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"tile_url":%q,"center":[%f,%f],"zoom":%d}`,
			cfg.Map.TileURL,
			cfg.Map.DefaultCenterLng, cfg.Map.DefaultCenterLat,
			cfg.Map.DefaultZoom,
		)
	})

	// Register universal domain handlers (these support both PostGIS and verified reference metadata)
	boundariesHandler := boundaries.NewHandler()
	boundariesHandler.Routes(r)

	standardsHandler := standards.NewHandler(standards.NewRepository(pgPool), logger)
	standardsHandler.Routes(r)

	indicatorsHandler := indicators.NewHandler(indicators.NewRepository(pgPool), logger)
	indicatorsHandler.Routes(r)

	auditHandler := audit.NewHandler(audit.NewRepository(pgPool), logger)
	auditHandler.Routes(r)

	authHandler := auth.NewHandler()
	authHandler.Routes(r)

	datasetsHandler := datasets.NewHandler(datasets.NewRepository(pgPool), logger)
	datasetsHandler.Routes(r)

	ingestionSvc := ingestion.NewService(pgPool, logger)
	ingestionHandler := ingestion.NewHandler(ingestionSvc, logger)
	ingestionHandler.Routes(r)

	inspectionsHandler := inspections.NewHandler(inspections.NewRepository(pgPool), logger)
	inspectionsHandler.Routes(r)

	// Spatial building queries and live coverage
	if db != nil {
		buildingsRepo := buildings.NewRepository(db.Pool)
		buildingsHandler := buildings.NewHandler(buildingsRepo, logger)
		buildingsHandler.Routes(r)

		r.Get("/api/v1/coverage", func(w http.ResponseWriter, r *http.Request) {
					queryCtx, queryCancel := context.WithTimeout(r.Context(), 5*time.Second)
					defer queryCancel()
			var count int
					err := db.QueryRow(queryCtx, "SELECT COUNT(*) FROM buildings").Scan(&count)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"coverage_query_failed"}`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
					fmt.Fprintf(w, `{"database_records":%d,"coverage":"Nairobi County","status":"available","note":"Count of rows in the local database; records may include unverified samples and are not an inventory coverage estimate."}`, count)
		})
	} else {
		// Register honest degraded state for live spatial building records
		degraded := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error":"database_unavailable","status":"degraded","message":"PostgreSQL/PostGIS is currently unreachable. Live spatial building polygon queries are paused until database connection is restored. View Research & Data for source registries and provenance."}`))
		}
		r.Get("/api/v1/buildings", http.HandlerFunc(degraded))
		r.Get("/api/v1/buildings/{id}", http.HandlerFunc(degraded))
		r.Get("/api/v1/coverage", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"coverage":"Nairobi County","status":"database_offline","note":"PostgreSQL and PostGIS are unavailable; live building counts cannot be reported."}`))
		})
	}

	// Start server
	srv := &http.Server{
		Addr:         serverAddress(cfg.App.Env, cfg.Auth.Enabled, cfg.App.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("server listening", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", slog.String("error", err.Error()))
		}
	}()

	<-shutdown
	logger.Info("shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
	defer shutdownCancel()

	return srv.Shutdown(shutdownCtx)
}

func serverAddress(environment string, authEnabled bool, port int) string {
	if environment == "development" && !authEnabled {
		return fmt.Sprintf("127.0.0.1:%d", port)
	}
	return fmt.Sprintf(":%d", port)
}

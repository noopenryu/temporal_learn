package main

import (
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.temporal.io/sdk/client"

	"ms-eligiblity/internal/config"
	"ms-eligiblity/internal/observability"
	"ms-eligiblity/internal/repository/postgres"
	"ms-eligiblity/internal/service"
	apptemporal "ms-eligiblity/internal/temporal"
	"ms-eligiblity/internal/temporal/activities"
	transporthttp "ms-eligiblity/internal/transport/http"
	"ms-eligiblity/migrations"
)

func main() {
	modeFlag := flag.String("mode", "", "Run mode: all, http, or worker (overrides RUN_MODE env)")
	flag.Parse()

	_, cleanup := observability.InitLogger(observability.LoggerOptions{
		ServiceName: "ms-eligiblity",
		Env:         "local",
		LokiURL:     "http://localhost:3100",
		Level:       slog.LevelInfo,
	})
	defer cleanup()

	slog.Info("Starting ms-eligiblity service")

	cfg := config.Load()
	if *modeFlag != "" {
		cfg.RunMode = *modeFlag
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Initialize PostgreSQL
	var repo *postgres.Repository
	dbPool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Warn("Could not connect to PostgreSQL immediately; running in degraded/offline mode", "error", err)
	} else {
		slog.Info("Connected to PostgreSQL successfully")
		repo = postgres.NewRepository(dbPool)

		// Run embedded migration
		if err := repo.Migrate(ctx, migrations.SchemaSQL); err != nil {
			slog.Error("Failed to apply database migrations", "error", err)
		} else {
			slog.Info("Database migrations applied successfully")
		}
	}

	// 2. Initialize Business Logic
	eligibilitySvc := service.NewEligibilityService(cfg, repo)

	// 3. Initialize Temporal Client & Worker
	var (
		temporalClient client.Client
		workerWrapper  *apptemporal.WorkerWrapper
	)

	if cfg.RunMode == "all" || cfg.RunMode == "worker" {
		tClient, err := client.Dial(client.Options{
			HostPort:  cfg.TemporalHostPort,
			Namespace: cfg.TemporalNamespace,
		})
		if err != nil {
			slog.Warn("Could not connect to Temporal cluster; worker disabled", "error", err)
		} else {
			temporalClient = tClient
			slog.Info("Connected to Temporal cluster", "host", cfg.TemporalHostPort, "namespace", cfg.TemporalNamespace)

			acts := activities.NewEligibilityActivities(eligibilitySvc)
			workerWrapper = apptemporal.NewWorker(temporalClient, cfg.TemporalTaskQueue, acts)

			if err := workerWrapper.Start(); err != nil {
				slog.Error("Failed to start Temporal worker", "error", err)
			}
		}
	}

	// 4. Initialize HTTP Server
	var srv *http.Server
	if cfg.RunMode == "all" || cfg.RunMode == "http" {
		handler := transporthttp.NewHandler(eligibilitySvc, repo)
		router := transporthttp.SetupRouter(handler)

		srv = &http.Server{
			Addr:         ":" + cfg.Port,
			Handler:      router,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		go func() {
			slog.Info("HTTP server listening", "port", cfg.Port, "mode", cfg.RunMode)
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Error("HTTP server error", "error", err)
				os.Exit(1)
			}
		}()
	}

	// 5. Graceful Shutdown on termination signals
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)
	sig := <-shutdownChan
	slog.Info("Received shutdown signal", "signal", sig.String())

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if srv != nil {
		slog.Info("Shutting down HTTP server...")
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("HTTP server shutdown failed", "error", err)
		}
	}

	if workerWrapper != nil {
		slog.Info("Shutting down Temporal worker...")
		workerWrapper.Stop()
	}

	if temporalClient != nil {
		temporalClient.Close()
	}

	if dbPool != nil {
		dbPool.Close()
	}

	slog.Info("ms-eligiblity shutdown complete")
	_ = fmt.Sprint("")
}

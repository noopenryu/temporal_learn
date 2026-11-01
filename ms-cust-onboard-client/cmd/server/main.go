package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ms-cust-onboard-client/internal/config"
	"ms-cust-onboard-client/internal/downstream"
	"ms-cust-onboard-client/internal/observability"
	"ms-cust-onboard-client/internal/repository/postgres"
	"ms-cust-onboard-client/internal/service"
	transporthttp "ms-cust-onboard-client/internal/transport/http"
	"ms-cust-onboard-client/migrations"
)

func main() {
	_, cleanup := observability.InitLogger(observability.LoggerOptions{
		ServiceName: "ms-cust-onboard-client",
		Env:         "local",
		LokiURL:     "http://localhost:3100",
		Level:       slog.LevelInfo,
	})
	defer cleanup()

	slog.Info("Starting ms-cust-onboard-client gateway service")

	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Initialize PostgreSQL
	dbPool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Warn("Could not connect to PostgreSQL immediately; running in degraded mode", "error", err)
	} else {
		slog.Info("Connected to PostgreSQL successfully")
		if err := postgres.Migrate(ctx, dbPool, migrations.SchemaSQL); err != nil {
			slog.Error("Failed to apply database migrations", "error", err)
		} else {
			slog.Info("Database migrations applied successfully")
		}
	}

	repo := postgres.NewCredentialRepository(dbPool)

	// 2. Downstream Customer Onboard client
	downstreamClient := downstream.NewCustomerOnboardClient(cfg.CustomerOnboardURL)
	slog.Info("Downstream client configured", "target", cfg.CustomerOnboardURL)

	// 3. Service & HTTP Handler
	svc := service.NewClientService(repo, downstreamClient)
	handler := transporthttp.NewHandler(svc)
	router := transporthttp.SetupRouter(handler)

	// 4. Start HTTP Server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("HTTP server listening", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	if dbPool != nil {
		dbPool.Close()
	}

	slog.Info("Server stopped")
	fmt.Println("Goodbye!")
}

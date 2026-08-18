package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gcet-web-backend/internal/api"
	"gcet-web-backend/internal/config"
	"gcet-web-backend/internal/logger"
	"gcet-web-backend/internal/repository"
	"gcet-web-backend/internal/service"

	"github.com/go-chi/chi/v5"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Fatal error loading config: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize Logger
	logger.Init(cfg.LogLevel)
	defer logger.Sync()

	logger.Log.Infof("Starting GCET Web Backend in %s mode", cfg.Environment)

	// 3. Initialize Repositories
	repo, err := repository.NewSupabaseRepo(cfg)
	if err != nil {
		logger.Log.Warnf("Supabase initialization skipped or failed: %v", err)
	} else {
		logger.Log.Info("Supabase connection established")
	}

	// Initialize Proxy Pool
	service.InitProxyPool(cfg.ProxyListURL)

	// 4. Initialize Services
	authSvc := service.NewAuthService(cfg)

	// 5. Setup Router
	r := chi.NewRouter()
	api.RegisterRoutes(r, cfg, authSvc, repo)

	// 6. Start Server with Graceful Shutdown
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 45 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Log.Infof("Server listening on port %d", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatalf("Server startup failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Server is shutting down...")

	// The context is used to inform the server it has 10 seconds to finish
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Log.Info("Server stopped cleanly")
}

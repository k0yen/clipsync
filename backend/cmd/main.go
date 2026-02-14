package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/k0yen/clipsync/internals"
)

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg := internals.LoadConfig()
	slog.Info("starting ClipSync E2EE backend",
		"port", cfg.Port,
		"max_connections", cfg.MaxConnections,
		"snippet_ttl", cfg.SnippetTTL,
	)

	// Initialize database
	store, err := internals.NewStore(cfg.DBPath)
	if err != nil {
		slog.Error("database initialization failed", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	// Initialize hub
	hub := internals.NewHub(cfg.MaxConnections)

	// Start cleanup routine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go cleanupRoutine(ctx, store, cfg)

	// Setup HTTP routes
	r := mux.NewRouter()
	r.HandleFunc("/ws/{roomID}", internals.WebSocketHandler(hub, store, cfg))
	r.HandleFunc("/health", internals.HealthHandler(store))
	r.HandleFunc("/stats", internals.StatsHandler(hub))

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		slog.Info("shutdown signal received, shutting down gracefully...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		cancel() // Stop cleanup routine
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("server shutdown error", "error", err)
		}
	}()

	// Start server
	slog.Info("server listening", "address", ":"+cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}

func cleanupRoutine(ctx context.Context, store *internals.Store, cfg *internals.Config) {
	ticker := time.NewTicker(cfg.CleanupInterval)
	defer ticker.Stop()

	slog.Info("cleanup routine started",
		"interval", cfg.CleanupInterval,
		"ttl", cfg.SnippetTTL,
	)

	for {
		select {
		case <-ctx.Done():
			slog.Info("cleanup routine stopped")
			return
		case <-ticker.C:
			deleted, err := store.CleanupOldSnippets(cfg.SnippetTTL)
			if err != nil {
				slog.Error("cleanup failed", "error", err)
			} else if deleted > 0 {
				slog.Info("cleaned up old snippets", "count", deleted)
			}
		}
	}
}

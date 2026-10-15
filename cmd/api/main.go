package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rizkyashari/cms-alumnee-be/internal/platform/config"
	platformdb "github.com/rizkyashari/cms-alumnee-be/internal/platform/database"
	"github.com/rizkyashari/cms-alumnee-be/internal/platform/httpapi"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		logger.Error("invalid configuration", slog.String("error", err.Error()))
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := platformdb.OpenMySQL(ctx, cfg)
	if err != nil {
		logger.Error("could not connect to MySQL", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpapi.NewRouter(logger, db),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("HTTP server listening", slog.String("address", cfg.HTTPAddr))
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("HTTP server shutdown failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
		logger.Info("HTTP server stopped")
	}
}

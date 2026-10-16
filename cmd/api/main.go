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

	"github.com/rizkyashari/cms-alumnee-be/internal/modules/auth"
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

	authService := auth.NewService(db, cfg.DemoAuthEnabled, cfg.DemoAuthPassword)
	var stagingAccounts []auth.StagingAccount
	if cfg.StagingNonVerifiedUsername != "" {
		stagingAccounts = []auth.StagingAccount{
			{
				Username: cfg.StagingNonVerifiedUsername, Password: cfg.StagingNonVerifiedPassword,
				Role: auth.RoleNonVerified, Name: "Alumni Belum Terverifikasi",
				VerificationStatus: "non_verified", BatchYear: 2012,
			},
			{
				Username: cfg.StagingVerifiedUsername, Password: cfg.StagingVerifiedPassword,
				Role: auth.RoleVerified, Name: "Alumni Terverifikasi",
				VerificationStatus: "verified", BatchYear: 2010,
			},
			{
				Username: cfg.StagingOfficerUsername, Password: cfg.StagingOfficerPassword,
				Role: auth.RoleOfficer, Name: "Pengurus Uji",
			},
			{
				Username: cfg.StagingAdminUsername, Password: cfg.StagingAdminPassword,
				Role: auth.RoleAdmin, Name: "Admin Uji",
			},
		}
	}
	if err := authService.ProvisionStagingAccounts(ctx, stagingAccounts); err != nil {
		logger.Error("could not provision staging accounts", slog.String("error", err.Error()))
		os.Exit(1)
	}
	authService.ConfigureIntegrations(auth.Integrations{
		PublicBaseURL:       cfg.PublicBaseURL,
		TurnstileSiteKey:    cfg.TurnstileSiteKey,
		TurnstileSecret:     cfg.TurnstileSecret,
		GoogleClientID:      cfg.GoogleClientID,
		GoogleSecret:        cfg.GoogleSecret,
		FacebookClientID:    cfg.FacebookClientID,
		FacebookSecret:      cfg.FacebookSecret,
		ResendAPIKey:        cfg.ResendAPIKey,
		ResendFromEmail:     cfg.ResendFromEmail,
		BootstrapAdminEmail: cfg.BootstrapAdminEmail,
		BootstrapAdminName:  cfg.BootstrapAdminName,
	})

	if cfg.RegencyMasterPath != "" {
		regencyCodes, err := auth.LoadRegencyCodes(cfg.RegencyMasterPath)
		if err != nil {
			logger.Error("could not load regency master data", slog.String("error", err.Error()))
			os.Exit(1)
		}
		authService.SetRegencyCodes(regencyCodes)
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpapi.NewRouterWithAuth(logger, db, auth.NewHandler(authService, logger)),
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

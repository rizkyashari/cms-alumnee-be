package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/rizkyashari/cms-alumnee-be/internal/platform/config"
	platformdb "github.com/rizkyashari/cms-alumnee-be/internal/platform/database"
	"github.com/rizkyashari/cms-alumnee-be/internal/platform/migrations"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		logger.Error("invalid configuration", slog.String("error", err.Error()))
		os.Exit(2)
	}
	db, err := platformdb.OpenMySQL(context.Background(), cfg)
	if err != nil {
		logger.Error("could not connect to MySQL", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()
	if err := migrations.Up(db); err != nil {
		logger.Error("database migration failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("database migrations completed")
}

package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"time"

	"github.com/rizkyashari/cms-alumnee-be/internal/platform/config"
	mysql "github.com/go-sql-driver/mysql"
)

func OpenMySQL(ctx context.Context, cfg config.Config) (*sql.DB, error) {
	driverConfig := mysql.NewConfig()
	driverConfig.User = cfg.MySQLUser
	driverConfig.Passwd = cfg.MySQLPassword
	driverConfig.Net = "tcp"
	driverConfig.Addr = net.JoinHostPort(cfg.MySQLHost, cfg.MySQLPort)
	driverConfig.DBName = cfg.MySQLDatabase
	driverConfig.Params = map[string]string{"charset": "utf8mb4"}
	driverConfig.ParseTime = true
	driverConfig.Timeout = 5 * time.Second
	driverConfig.ReadTimeout = 10 * time.Second
	driverConfig.WriteTimeout = 10 * time.Second

	db, err := sql.Open("mysql", driverConfig.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("open MySQL pool: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping MySQL: %w", err)
	}
	return db, nil
}

package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type RepoServer struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func ConnectDb() *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Info),
		SkipDefaultTransaction: true,
	})

	if err != nil {
		log.Fatal("Failed to connect to database. \n", err)
		os.Exit(2)
	}

	log.Println("DB is connected")
	db.Logger = logger.Default.LogMode(logger.Info)

	return db
}

func ConnectRedis() *redis.Client {
	config := redis.Options{
		Addr:     os.Getenv("REDIS_ADDRESS"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	}

	rdb := redis.NewClient(&config)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("Failed to connect to redis. \n", err)
		os.Exit(2)
	}

	log.Println("Redis is connected")

	return rdb
}

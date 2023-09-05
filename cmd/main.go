package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fadhln/lms-be/database"
	"github.com/fadhln/lms-be/handler"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/seed"
)

func main() {
	useMigrate := flag.Bool("migrate", false, "run database migration process")
	useSeed := flag.Bool("seed", false, "run database seeding process")
	flag.Parse()

	db := database.ConnectDb()

	if *useMigrate {
		err := db.AutoMigrate(model.Entities...)
		if err != nil {
			log.Fatal("Migration failed! \n", err)
			os.Exit(2)
		}

		log.Println("DB successfully migrated!")
		os.Exit(0)
	}

	redis := database.ConnectRedis()

	server := database.RepoServer{
		DB:    db,
		Redis: redis,
	}

	if *useSeed {
		err := seed.SeedingProcess(&server)
		if err != nil {
			log.Fatal("Seeding failed! \n", err)
			os.Exit(2)
		}

		log.Println("DB successfully seeded!")
		os.Exit(0)
	}

	r := handler.InitRouter(&server)

	portSetting := ":" + os.Getenv("PORT")
	srv := &http.Server{
		Addr:    portSetting,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}

	select {
	case <-ctx.Done():
		log.Println("timeout of 5 seconds.")
	}
	log.Println("Server exiting")
}

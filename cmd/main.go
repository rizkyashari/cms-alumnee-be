package main

import (
	"flag"
	"log"
	"os"

	"github.com/fadhln/lms-be/database"
	"github.com/fadhln/lms-be/handler"
	"github.com/fadhln/lms-be/model"
)

func main() {
	useMigrate := flag.Bool("migrate", false, "run database migration")
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

	r := handler.InitRouter(db)

	portSetting := ":" + os.Getenv("PORT")
	err := r.Run(portSetting)

	if err != nil {
		log.Panic(err)
	}
}

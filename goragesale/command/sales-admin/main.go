package main

import (
	"flag"
	"log"
	"os"

	database "github.com/bdatdo0601/goragesale/src/platform/db"
	"github.com/bdatdo0601/goragesale/src/platform/db/schema"
)

func main() {

	flag.Parse()

	// Initialize dependencies.
	db, err := database.Open()
	if err != nil {
		log.Fatalf("error: connecting to db: %s", err)
	}
	defer db.Close()

	switch flag.Arg(0) {
	case "migrate":
		if err := schema.Migrate(db.DB); err != nil {
			log.Println("error applying migrations", err)
			os.Exit(1)
		}
		log.Println("Migrations complete")
		return

	case "seed":
		if err := schema.Seed(db.DB); err != nil {
			log.Println("error seeding database", err)
			os.Exit(1)
		}
		log.Println("Seed data complete")
		return
	}
}

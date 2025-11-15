package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/sawalreverr/cv-reviewer/config"
)

func main() {
	// flag
	drop := flag.Bool("drop", false, "drop all tables before migrations")
	flag.Parse()

	log.Println("starting database migrations...")

	// load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// connect to database
	db, err := config.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// drop tables if requested
	if *drop {
		log.Println("DROP FLAG DETECTED!")
		log.Println("this will delete all data. Press Ctrl+C to cancel...")
		log.Println("continuing in 3 seconds...")

		time.Sleep(3 * time.Second)
		
		if err := config.DropAllTables(db); err != nil {
			log.Fatalf("failed to drop tables: %v", err)
		}
	}

	fmt.Println()

	// run migrations
	if err := config.RunMigration(db); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}
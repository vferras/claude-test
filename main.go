package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"claude-test/config"
	"claude-test/db"
	"claude-test/job"
	"claude-test/marketstack"
)

func main() {
	cfg, err := config.Load("config/stocks.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrated successfully")

	client := marketstack.NewClient(cfg.APIKey)

	eodJob := job.NewEODJob(database, client, cfg.Symbols)
	eodJob.Schedule()
	log.Println("EOD job scheduled (daily at 22:00)")

	// Wait for interrupt signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("Shutting down")
}

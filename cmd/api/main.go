package main

import (
	"fmt"
	"log"

	"github.com/chrisabs/cadence/internal/api"
	"github.com/chrisabs/cadence/internal/config"
	"github.com/chrisabs/cadence/internal/platform/database"
)

func main() {
	fmt.Println("\n=== Loading Configuration ===")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Configuration loading failed:", err)
	}
	fmt.Println("Configuration loaded successfully!")

	fmt.Println("\n=== Initializing Database ===")
	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	fmt.Println("Database connected successfully!")

	if err := db.Init(); err != nil {
		log.Fatal("Database initialization failed:", err)
	}
	fmt.Println("Database tables initialized successfully!")

	fmt.Println("\n=== Starting Server ===")
	server := api.NewServer(":3000", db, cfg)
	fmt.Println("Server starting on port 3000...")
	server.Run()
}

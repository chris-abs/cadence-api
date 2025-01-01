package main

import (
	"log"

	"github.com/chrisabs/storage/internal/api"
	"github.com/chrisabs/storage/internal/platform/database"
)

func main() {
	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	if err := db.Init(); err != nil {
		log.Fatal("Failed to initialize database schema:", err)
	}

	server := api.NewServer(":3000", db)
	server.Run()
}

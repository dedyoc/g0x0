package main

import (
	"log"
	"os"

	"github.com/dedyoc/g0x0/internal/config"
	"github.com/dedyoc/g0x0/internal/database"
	"github.com/dedyoc/g0x0/internal/server"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	srv := server.New(cfg, db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(srv.Start(":" + port))
}

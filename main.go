package main

import (
	"log"

	"github.com/mgarce/go_query_api/internal/config"
	"github.com/mgarce/go_query_api/internal/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize router
	r := gin.Default()

	// Setup routes
	if err := handlers.SetupRoutes(r, cfg); err != nil {
		log.Fatalf("Failed to setup routes: %v", err)
	}

	// Start server
	log.Printf("Starting server on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

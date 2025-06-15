package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/mgarce/go_query_api/internal/config"
	"github.com/mgarce/go_query_api/internal/handlers"
)

func main() {
	// Define command-line flags
	var (
		port      = flag.String("port", "", "Server port (overrides config)")
		csvPath   = flag.String("csv", "", "Path to field mappings CSV (overrides config)")
		debugMode = flag.Bool("debug", false, "Enable debug mode")
		showHelp  = flag.Bool("help", false, "Show help message")
		showVersion = flag.Bool("version", false, "Show version information")
	)

	// Parse flags
	flag.Parse()

	// Show help if requested
	if *showHelp {
		printHelp()
		return
	}

	// Show version if requested
	if *showVersion {
		fmt.Println("Go Query API v1.0.0")
		return
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override config with command-line flags if provided
	if *port != "" {
		cfg.Port = *port
	}
	if *csvPath != "" {
		cfg.CSVPath = *csvPath
	}

	// Set Gin mode
	if *debugMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
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

// printHelp displays usage information
func printHelp() {
	fmt.Println("Go Query API - Natural Language to SQL Converter")
	fmt.Println("\nUsage:")
	fmt.Printf("  %s [options]\n\n", os.Args[0])
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println("\nExample:")
	fmt.Println("  ./query-api --port 8080 --csv ./field_mappings.csv")
}

package config

import (
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	Port             string
	CSVPath          string
	MatchThreshold   float64
	MaxMatches       int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	port := getEnv("PORT", "8080")
	csvPath := getEnv("CSV_PATH", "field_mappings.csv")
	
	// Parse threshold with default 30.0
	thresholdStr := getEnv("MATCH_THRESHOLD", "30.0")
	threshold, err := strconv.ParseFloat(thresholdStr, 64)
	if err != nil {
		threshold = 30.0
	}
	
	// Parse max matches with default 10
	maxMatchesStr := getEnv("MAX_MATCHES", "10")
	maxMatches, err := strconv.Atoi(maxMatchesStr)
	if err != nil {
		maxMatches = 10
	}
	
	return &Config{
		Port:           port,
		CSVPath:        csvPath,
		MatchThreshold: threshold,
		MaxMatches:     maxMatches,
	}, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

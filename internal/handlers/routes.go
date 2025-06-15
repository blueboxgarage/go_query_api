package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/mgarce/go_query_api/internal/config"
	"github.com/mgarce/go_query_api/internal/services"
)

// SetupRoutes configures the API routes
func SetupRoutes(r *gin.Engine, cfg *config.Config) error {
	// Load CSV data
	fieldService, err := services.NewFieldService(cfg)
	if err != nil {
		return err
	}
	
	// Create query service
	queryService := services.NewQueryService(fieldService)
	
	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	
	// API routes
	api := r.Group("/api/v1")
	{
		// Generate query endpoint
		api.POST("/generate-query", GenerateQueryHandler(queryService))
		
		// List fields endpoint
		api.GET("/fields", ListFieldsHandler(fieldService))
	}
	
	return nil
}

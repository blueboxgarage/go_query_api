package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mgarce/go_query_api/internal/models"
	"github.com/mgarce/go_query_api/internal/services"
)

// GenerateQueryHandler handles the query generation request
func GenerateQueryHandler(service *services.QueryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request models.QueryRequest
		
		// Validate request
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
			return
		}
		
		// Set a default system if not provided
		if request.System == "" {
			request.System = "default"
		}
		
		// Generate query
		startTime := time.Now()
		response, err := service.GenerateQuery(request)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate query: " + err.Error()})
			return
		}
		
		// Calculate processing time
		response.ProcessingTime = time.Since(startTime).Milliseconds()
		
		c.JSON(http.StatusOK, response)
	}
}

// ListFieldsHandler returns all available field mappings
func ListFieldsHandler(service *services.FieldService) gin.HandlerFunc {
	return func(c *gin.Context) {
		system := c.Query("system")
		if system == "" {
			system = "default"
		}
		
		fields := service.GetAllFields(system)
		c.JSON(http.StatusOK, gin.H{"fields": fields})
	}
}

package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mgarce/go_query_api/internal/config"
	"github.com/mgarce/go_query_api/internal/handlers"
	"github.com/mgarce/go_query_api/internal/models"
	"github.com/mgarce/go_query_api/internal/services"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() (*gin.Engine, error) {
	// Use test mode for Gin
	gin.SetMode(gin.TestMode)
	
	// Create router
	r := gin.Default()
	
	// Create test config
	cfg := &config.Config{
		CSVPath: "../field_mappings.csv",
	}
	
	// Setup routes with the test config
	err := handlers.SetupRoutes(r, cfg)
	if err != nil {
		return nil, err
	}
	
	return r, nil
}

func TestGenerateQueryHandler(t *testing.T) {
	// Set up router
	r, err := setupTestRouter()
	assert.NoError(t, err)
	
	// Test cases
	testCases := []struct {
		name           string
		requestPayload models.QueryRequest
		expectedStatus int
		checkResponse  func(t *testing.T, response map[string]interface{})
	}{
		{
			name: "Valid query request",
			requestPayload: models.QueryRequest{
				Description: "Get user emails",
				System:      "default",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Contains(t, response, "query")
				assert.Contains(t, response, "matched_fields")
				assert.Contains(t, response, "confidence")
			},
		},
		{
			name: "Invalid request - empty description",
			requestPayload: models.QueryRequest{
				Description: "",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Contains(t, response, "error")
			},
		},
		{
			name: "Request with limit",
			requestPayload: models.QueryRequest{
				Description: "Get user orders",
				Limit:       10,
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Contains(t, response, "query")
				query, _ := response["query"].(string)
				assert.Contains(t, query, "LIMIT 10")
			},
		},
	}
	
	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			payload, err := json.Marshal(tc.requestPayload)
			assert.NoError(t, err)
			
			req, err := http.NewRequest("POST", "/api/v1/generate-query", bytes.NewBuffer(payload))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			
			// Record response
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			
			// Check status code
			assert.Equal(t, tc.expectedStatus, w.Code)
			
			// Parse response
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			
			// Check response
			if tc.checkResponse != nil {
				tc.checkResponse(t, response)
			}
		})
	}
}

func TestListFieldsHandler(t *testing.T) {
	// Set up router
	r, err := setupTestRouter()
	assert.NoError(t, err)
	
	// Create request
	req, err := http.NewRequest("GET", "/api/v1/fields", nil)
	assert.NoError(t, err)
	
	// Record response
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	// Check status code
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Parse response
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	// Check that fields are returned
	assert.Contains(t, response, "fields")
	fields, ok := response["fields"].([]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, fields)
}

func TestHealthCheck(t *testing.T) {
	// Set up router
	r, err := setupTestRouter()
	assert.NoError(t, err)
	
	// Create request
	req, err := http.NewRequest("GET", "/health", nil)
	assert.NoError(t, err)
	
	// Record response
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	// Check status code
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Parse response
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	// Check health status
	assert.Equal(t, "ok", response["status"])
}
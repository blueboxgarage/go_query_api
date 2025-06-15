package tests

import (
	"strings"
	"testing"

	"github.com/mgarce/go_query_api/internal/config"
	"github.com/mgarce/go_query_api/internal/models"
	"github.com/mgarce/go_query_api/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestQueryService(t *testing.T) {
	// Set up field service for testing
	cfg := &config.Config{
		CSVPath: "../field_mappings.csv",
	}
	
	fieldService, err := services.NewFieldService(cfg)
	assert.NoError(t, err)
	
	// Create query service
	queryService := services.NewQueryService(fieldService)
	assert.NotNil(t, queryService)
	
	// Test cases
	testCases := []struct {
		name          string
		description   string
		expectSuccess bool
		checkFunction func(t *testing.T, response models.QueryResponse)
	}{
		{
			name:          "Simple user email query",
			description:   "Get user emails",
			expectSuccess: true,
			checkFunction: func(t *testing.T, response models.QueryResponse) {
				assert.Contains(t, response.Query, "users.email")
				assert.NotEmpty(t, response.MatchedFields)
				assert.GreaterOrEqual(t, response.Confidence, 50.0)
			},
		},
		{
			name:          "Count orders query",
			description:   "Count total orders",
			expectSuccess: true,
			checkFunction: func(t *testing.T, response models.QueryResponse) {
				assert.Contains(t, response.Query, "COUNT")
				assert.Contains(t, response.Query, "orders")
				assert.NotEmpty(t, response.MatchedFields)
			},
		},
		{
			name:          "Unique products query",
			description:   "Find unique products ordered",
			expectSuccess: true,
			checkFunction: func(t *testing.T, response models.QueryResponse) {
				assert.Contains(t, response.Query, "DISTINCT")
				assert.Contains(t, response.Query, "products")
				assert.NotEmpty(t, response.MatchedFields)
			},
		},
		{
			name:          "Query with joins",
			description:   "Get orders with product names",
			expectSuccess: true,
			checkFunction: func(t *testing.T, response models.QueryResponse) {
				assert.Contains(t, response.Query, "JOIN")
				assert.Contains(t, strings.ToLower(response.Query), "orders")
				assert.Contains(t, strings.ToLower(response.Query), "products")
				assert.NotEmpty(t, response.JoinsUsed)
			},
		},
		{
			name:          "Nonsense query",
			description:   "xyz12345 nonexistent fields",
			expectSuccess: false,
			checkFunction: nil,
		},
	}
	
	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := models.QueryRequest{
				Description: tc.description,
				System:      "default",
			}
			
			response, err := queryService.GenerateQuery(request)
			
			if tc.expectSuccess {
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Query)
				assert.NotZero(t, response.ProcessingTime)
				
				if tc.checkFunction != nil {
					tc.checkFunction(t, response)
				}
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestQueryTypeIdentification(t *testing.T) {
	// Set up services
	cfg := &config.Config{
		CSVPath: "../field_mappings.csv",
	}
	
	fieldService, err := services.NewFieldService(cfg)
	assert.NoError(t, err)
	
	queryService := services.NewQueryService(fieldService)
	
	// Test different query descriptions and expected types
	testCases := []struct {
		description string
		queryType   string
		distinct    bool
	}{
		{"Get all users", "SELECT", false},
		{"Count orders by user", "COUNT", false},
		{"How many products are there", "COUNT", false},
		{"Show the number of orders", "COUNT", false},
		{"Find unique user emails", "SELECT", true},
		{"List distinct product names", "SELECT", true},
		{"Get orders grouped by product", "GROUP", false},
		{"Show sales per user", "GROUP", false},
	}
	
	for _, tc := range testCases {
		// We can't directly test the private method, so we test through the public API
		request := models.QueryRequest{
			Description: tc.description,
		}
		
		response, err := queryService.GenerateQuery(request)
		if err != nil {
			t.Logf("Error for '%s': %v", tc.description, err)
			continue
		}
		
		// Check the type of query based on the generated SQL
		if tc.queryType == "COUNT" {
			assert.Contains(t, response.Query, "COUNT(")
		}
		
		if tc.queryType == "GROUP" {
			assert.Contains(t, response.Query, "GROUP BY")
		}
		
		if tc.distinct {
			assert.Contains(t, response.Query, "DISTINCT")
		}
	}
}
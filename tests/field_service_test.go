package tests

import (
	"os"
	"testing"

	"github.com/mgarce/go_query_api/internal/config"
	"github.com/mgarce/go_query_api/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestFieldService(t *testing.T) {
	// Set up test config
	cfg := &config.Config{
		CSVPath: "../field_mappings.csv",
	}
	
	// Create field service
	service, err := services.NewFieldService(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, service)
	
	// Test GetAllFields
	fields := service.GetAllFields("default")
	assert.NotEmpty(t, fields)
	assert.GreaterOrEqual(t, len(fields), 9) // Known number from our test CSV
	
	// Test FindFieldMatches
	matches := service.FindFieldMatches([]string{"email", "user"}, 30.0, 10)
	assert.NotEmpty(t, matches)
	
	// Test that email field is found with high score
	var emailFound bool
	for _, match := range matches {
		if match.ColumnName == "email" && match.TableName == "users" {
			emailFound = true
			assert.GreaterOrEqual(t, match.MatchScore, 50.0)
			break
		}
	}
	assert.True(t, emailFound, "Email field should be matched")
	
	// Test FindJoinPath
	joins, err := service.FindJoinPath("users", "orders")
	assert.NoError(t, err)
	assert.NotEmpty(t, joins)
	
	// Test FindJoinPath for longer path
	joins, err = service.FindJoinPath("users", "products")
	assert.NoError(t, err)
	assert.NotEmpty(t, joins)
	assert.GreaterOrEqual(t, len(joins), 2) // Should need at least 2 joins
}

func TestFieldServiceWithEmptyFile(t *testing.T) {
	// Create empty temporary file
	tmpFile, err := os.CreateTemp("", "empty_*.csv")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	
	// Write header only
	_, err = tmpFile.WriteString("column_name,table_name,system_a_fieldmap,system_b_fieldmap,field_description,field_type,join_key,foreign_table,foreign_key\n")
	assert.NoError(t, err)
	tmpFile.Close()
	
	// Set up test config
	cfg := &config.Config{
		CSVPath: tmpFile.Name(),
	}
	
	// Create field service
	service, err := services.NewFieldService(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, service)
	
	// Test GetAllFields with empty file
	fields := service.GetAllFields("default")
	assert.Empty(t, fields)
	
	// Test FindFieldMatches with empty file
	matches := service.FindFieldMatches([]string{"email", "user"}, 30.0, 10)
	assert.Empty(t, matches)
	
	// Test FindJoinPath with empty graph
	joins, err := service.FindJoinPath("users", "orders") 
	assert.Error(t, err) // Should error as the tables don't exist
	assert.Empty(t, joins)
}
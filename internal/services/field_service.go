package services

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/mgarce/go_query_api/internal/config"
	"github.com/mgarce/go_query_api/internal/models"
	"github.com/sirupsen/logrus"
)

// FieldService handles field mappings and relationships
type FieldService struct {
	fields            []models.Field
	relationshipGraph map[string]map[string]models.Join
	log               *logrus.Logger
}

// NewFieldService creates a new field service
func NewFieldService(cfg *config.Config) (*FieldService, error) {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	
	service := &FieldService{
		fields:            make([]models.Field, 0),
		relationshipGraph: make(map[string]map[string]models.Join),
		log:               log,
	}
	
	if err := service.loadCSV(cfg.CSVPath); err != nil {
		return nil, fmt.Errorf("failed to load CSV: %w", err)
	}
	
	service.buildRelationshipGraph()
	
	return service, nil
}

// loadCSV loads field mappings from a CSV file
func (s *FieldService) loadCSV(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}
	
	// Skip header row
	if len(records) > 0 {
		for i := 1; i < len(records); i++ {
			row := records[i]
			if len(row) < 9 {
				s.log.Warnf("Skipping invalid CSV row: %v", row)
				continue
			}
			
			field := models.Field{
				ColumnName:      row[0],
				TableName:       row[1],
				SystemAFieldMap: row[2],
				SystemBFieldMap: row[3],
				Description:     row[4],
				FieldType:       row[5],
				JoinKey:         row[6],
				ForeignTable:    row[7],
				ForeignKey:      row[8],
			}
			
			s.fields = append(s.fields, field)
		}
	}
	
	s.log.Infof("Loaded %d fields from %s", len(s.fields), path)
	return nil
}

// buildRelationshipGraph builds a graph of table relationships for JOIN path finding
func (s *FieldService) buildRelationshipGraph() {
	for _, field := range s.fields {
		// Skip fields without join relationships
		if field.ForeignTable == "" || field.ForeignKey == "" {
			continue
		}
		
		// Create the source table node if it doesn't exist
		if _, exists := s.relationshipGraph[field.TableName]; !exists {
			s.relationshipGraph[field.TableName] = make(map[string]models.Join)
		}
		
		// Create the target table node if it doesn't exist
		if _, exists := s.relationshipGraph[field.ForeignTable]; !exists {
			s.relationshipGraph[field.ForeignTable] = make(map[string]models.Join)
		}
		
		// Add the relationship (bidirectional)
		joinCondition := fmt.Sprintf("%s.%s = %s.%s", 
			field.TableName, field.ColumnName,
			field.ForeignTable, field.ForeignKey)
		
		// From source to target
		s.relationshipGraph[field.TableName][field.ForeignTable] = models.Join{
			From:      field.TableName,
			To:        field.ForeignTable,
			Condition: joinCondition,
		}
		
		// From target to source (for bidirectional traversal)
		s.relationshipGraph[field.ForeignTable][field.TableName] = models.Join{
			From:      field.ForeignTable,
			To:        field.TableName,
			Condition: joinCondition,
		}
	}
	
	s.log.Infof("Built relationship graph with %d tables", len(s.relationshipGraph))
}

// GetAllFields returns all field mappings, optionally filtered by system
func (s *FieldService) GetAllFields(system string) []models.Field {
	if system == "" || system == "default" {
		return s.fields
	}
	
	// Filter fields by system
	var filtered []models.Field
	for _, field := range s.fields {
		// Check if this field has a mapping for the requested system
		if (system == "system_a" && field.SystemAFieldMap != "") ||
		   (system == "system_b" && field.SystemBFieldMap != "") {
			filtered = append(filtered, field)
		}
	}
	
	return filtered
}

// FindFieldMatches finds fields matching the given keywords with fuzzy matching
func (s *FieldService) FindFieldMatches(keywords []string, threshold float64, maxMatches int) []models.FieldMatch {
	matches := make([]models.FieldMatch, 0)
	
	for _, field := range s.fields {
		// Calculate match score against field description
		score := s.calculateMatchScore(field.Description, keywords)
		
		// Skip fields below threshold
		if score < threshold {
			continue
		}
		
		match := models.FieldMatch{
			ColumnName:      field.ColumnName,
			TableName:       field.TableName,
			FieldDescription: field.Description,
			MatchScore:      score,
		}
		
		matches = append(matches, match)
	}
	
	// Sort matches by score (descending)
	sortMatchesByScore(matches)
	
	// Limit number of matches
	if len(matches) > maxMatches {
		matches = matches[:maxMatches]
	}
	
	return matches
}

// calculateMatchScore calculates how well the keywords match the description
// Returns a score from 0-100, with 100 being a perfect match
func (s *FieldService) calculateMatchScore(description string, keywords []string) float64 {
	if len(keywords) == 0 {
		return 0
	}
	
	description = strings.ToLower(description)
	
	// Count how many keywords are in the description
	matchedCount := 0
	for _, keyword := range keywords {
		if strings.Contains(description, strings.ToLower(keyword)) {
			matchedCount++
		}
	}
	
	// Calculate percentage of matched keywords
	return float64(matchedCount) / float64(len(keywords)) * 100
}

// FindJoinPath finds the shortest join path between tables
func (s *FieldService) FindJoinPath(fromTable string, toTable string) ([]models.Join, error) {
	// If tables are the same, no join needed
	if fromTable == toTable {
		return []models.Join{}, nil
	}
	
	// Check if both tables exist in the graph
	if _, exists := s.relationshipGraph[fromTable]; !exists {
		return nil, fmt.Errorf("table %s not found in relationship graph", fromTable)
	}
	if _, exists := s.relationshipGraph[toTable]; !exists {
		return nil, fmt.Errorf("table %s not found in relationship graph", toTable)
	}
	
	// Breadth-First Search to find shortest path
	path, err := s.bfsShortestPath(fromTable, toTable)
	if err != nil {
		return nil, err
	}
	
	// Convert path to joins
	joins := make([]models.Join, 0)
	for i := 0; i < len(path)-1; i++ {
		joins = append(joins, s.relationshipGraph[path[i]][path[i+1]])
	}
	
	return joins, nil
}

// bfsShortestPath performs a BFS to find the shortest path between tables
func (s *FieldService) bfsShortestPath(start, end string) ([]string, error) {
	// Queue for BFS
	queue := []string{start}
	
	// Track visited nodes to prevent cycles
	visited := map[string]bool{start: true}
	
	// Track parents to reconstruct path
	parents := make(map[string]string)
	
	for len(queue) > 0 {
		// Dequeue current node
		current := queue[0]
		queue = queue[1:]
		
		// Check if we reached the destination
		if current == end {
			// Reconstruct path
			path := []string{end}
			for node := end; node != start; node = parents[node] {
				path = append([]string{parents[node]}, path...)
			}
			return path, nil
		}
		
		// Visit neighbors
		for neighbor := range s.relationshipGraph[current] {
			if !visited[neighbor] {
				visited[neighbor] = true
				parents[neighbor] = current
				queue = append(queue, neighbor)
			}
		}
	}
	
	return nil, fmt.Errorf("no join path found between %s and %s", start, end)
}

// sortMatchesByScore sorts field matches by score (descending)
func sortMatchesByScore(matches []models.FieldMatch) {
	// Simple bubble sort for now
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[i].MatchScore < matches[j].MatchScore {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}
}
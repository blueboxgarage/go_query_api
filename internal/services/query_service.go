package services

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/mgarce/go_query_api/internal/models"
	"github.com/sirupsen/logrus"
)

// QueryService handles SQL query generation
type QueryService struct {
	fieldService *FieldService
	log          *logrus.Logger
}

// NewQueryService creates a new query service
func NewQueryService(fieldService *FieldService) *QueryService {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	
	return &QueryService{
		fieldService: fieldService,
		log:          log,
	}
}

// GenerateQuery generates an SQL query based on the natural language description
func (s *QueryService) GenerateQuery(request models.QueryRequest) (models.QueryResponse, error) {
	startTime := time.Now()
	
	// Parse description for keywords
	keywords := s.extractKeywords(request.Description)
	
	// Identify query type and intent
	queryType, distinct := s.identifyQueryType(request.Description)
	
	// Find matching fields
	matchedFields := s.fieldService.FindFieldMatches(keywords, 30.0, 10)
	
	if len(matchedFields) == 0 {
		return models.QueryResponse{}, fmt.Errorf("no matching fields found for description")
	}
	
	// Generate SQL query
	query, joins, err := s.buildSQLQuery(matchedFields, queryType, distinct, request.Limit)
	if err != nil {
		return models.QueryResponse{}, fmt.Errorf("failed to build SQL query: %w", err)
	}
	
	// Calculate confidence score
	confidence := s.calculateConfidence(matchedFields)
	
	response := models.QueryResponse{
		Query:          query,
		MatchedFields:  matchedFields,
		JoinsUsed:      joins,
		Confidence:     confidence,
		ProcessingTime: time.Since(startTime).Milliseconds(),
	}
	
	return response, nil
}

// extractKeywords extracts relevant keywords from the description
func (s *QueryService) extractKeywords(description string) []string {
	// Remove special characters and convert to lowercase
	sanitized := strings.ToLower(description)
	re := regexp.MustCompile(`[^\w\s]`)
	sanitized = re.ReplaceAllString(sanitized, " ")
	
	// Split into words
	words := strings.Fields(sanitized)
	
	// Filter out common stopwords
	stopwords := map[string]bool{
		"a": true, "an": true, "the": true, "and": true, "or": true,
		"for": true, "in": true, "on": true, "at": true, "by": true, "to": true,
		"with": true, "about": true, "as": true, "into": true, "like": true,
		"through": true, "after": true, "over": true, "between": true, "out": true,
		"against": true, "during": true, "without": true, "before": true, "under": true,
		"around": true, "among": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "being": true, "have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true, "but": true, "if": true, "of": true,
		"from": true, "get": true, "all": true, "show": true, "find": true, "can": true,
		"i": true, "me": true, "my": true, "myself": true, "we": true, "our": true,
		"us": true, "ourselves": true, "you": true, "your": true, "yourself": true,
		"he": true, "him": true, "his": true, "himself": true, "she": true, "her": true,
		"hers": true, "herself": true, "it": true, "its": true, "itself": true,
		"they": true, "them": true, "their": true, "theirs": true, "themselves": true,
		"what": true, "which": true, "who": true, "whom": true, "whose": true,
	}
	
	var keywords []string
	for _, word := range words {
		if !stopwords[word] && len(word) > 1 {
			keywords = append(keywords, word)
		}
	}
	
	s.log.Infof("Extracted keywords: %v", keywords)
	return keywords
}

// identifyQueryType identifies the type of query to generate
func (s *QueryService) identifyQueryType(description string) (string, bool) {
	desc := strings.ToLower(description)
	
	// Check for COUNT operations
	if strings.Contains(desc, "count") || 
	   strings.Contains(desc, "how many") || 
	   strings.Contains(desc, "number of") {
		return "COUNT", false
	}
	
	// Check for GROUP BY operations
	if strings.Contains(desc, "group") || 
	   strings.Contains(desc, "grouped") || 
	   strings.Contains(desc, "per") {
		return "GROUP", false
	}
	
	// Check for DISTINCT
	distinct := strings.Contains(desc, "distinct") || 
	           strings.Contains(desc, "unique") ||
	           strings.Contains(desc, "different")
	
	// Default to SELECT
	return "SELECT", distinct
}

// buildSQLQuery builds an SQL query based on matched fields
func (s *QueryService) buildSQLQuery(matches []models.FieldMatch, queryType string, distinct bool, limit int) (string, []models.Join, error) {
	if len(matches) == 0 {
		return "", nil, fmt.Errorf("no field matches provided")
	}
	
	// Collect required tables
	tables := make(map[string]bool)
	for _, match := range matches {
		tables[match.TableName] = true
	}
	tableNames := make([]string, 0, len(tables))
	for table := range tables {
		tableNames = append(tableNames, table)
	}
	
	// Find join paths between tables
	var allJoins []models.Join
	if len(tableNames) > 1 {
		// Start with the first table and find paths to all others
		for i := 1; i < len(tableNames); i++ {
			joins, err := s.fieldService.FindJoinPath(tableNames[0], tableNames[i])
			if err != nil {
				return "", nil, fmt.Errorf("failed to find join path: %w", err)
			}
			allJoins = append(allJoins, joins...)
		}
		
		// Deduplicate joins
		allJoins = deduplicateJoins(allJoins)
	}
	
	// Build SELECT clause
	var selectClause string
	
	switch queryType {
	case "COUNT":
		// For COUNT queries, select the count of the first field
		selectClause = fmt.Sprintf("COUNT(%s.%s)", 
			matches[0].TableName, 
			matches[0].ColumnName)
			
	case "GROUP":
		// For GROUP BY queries, select the count and group by field
		selectClause = fmt.Sprintf("%s.%s, COUNT(*)", 
			matches[0].TableName, 
			matches[0].ColumnName)
			
	default: // SELECT
		// For regular SELECT queries, select all matched fields
		var fields []string
		for _, match := range matches {
			fields = append(fields, fmt.Sprintf("%s.%s", 
				match.TableName, 
				match.ColumnName))
		}
		
		if distinct {
			selectClause = "DISTINCT " + strings.Join(fields, ", ")
		} else {
			selectClause = strings.Join(fields, ", ")
		}
	}
	
	// Build FROM clause with table alias
	fromClause := fmt.Sprintf("%s %s", tableNames[0], tableNames[0][0:1])
	
	// Build JOIN clauses
	var joinClauses []string
	tablesInJoin := map[string]bool{tableNames[0]: true}
	
	for _, join := range allJoins {
		if tablesInJoin[join.To] {
			continue // Skip tables already joined
		}
		
		// Add table alias to the join condition
		condition := join.Condition
		
		// Add the JOIN clause
		joinClauses = append(joinClauses, 
			fmt.Sprintf("JOIN %s %s ON %s", 
				join.To, 
				join.To[0:1], 
				condition))
		
		tablesInJoin[join.To] = true
	}
	
	// Build WHERE clause (empty for now, would be based on additional criteria)
	whereClause := ""
	
	// Build GROUP BY clause
	groupByClause := ""
	if queryType == "GROUP" {
		groupByClause = fmt.Sprintf("GROUP BY %s.%s", 
			matches[0].TableName, 
			matches[0].ColumnName)
	}
	
	// Build LIMIT clause
	limitClause := ""
	if limit > 0 {
		limitClause = fmt.Sprintf("LIMIT %d", limit)
	}
	
	// Assemble the complete query
	query := fmt.Sprintf("SELECT %s FROM %s", selectClause, fromClause)
	
	if len(joinClauses) > 0 {
		query += " " + strings.Join(joinClauses, " ")
	}
	
	if whereClause != "" {
		query += " WHERE " + whereClause
	}
	
	if groupByClause != "" {
		query += " " + groupByClause
	}
	
	if limitClause != "" {
		query += " " + limitClause
	}
	
	return query, allJoins, nil
}

// deduplicateJoins removes duplicate join conditions
func deduplicateJoins(joins []models.Join) []models.Join {
	if len(joins) <= 1 {
		return joins
	}
	
	uniqueJoins := make(map[string]models.Join)
	for _, join := range joins {
		key := join.Condition
		uniqueJoins[key] = join
	}
	
	result := make([]models.Join, 0, len(uniqueJoins))
	for _, join := range uniqueJoins {
		result = append(result, join)
	}
	
	return result
}

// calculateConfidence calculates the confidence score for the query
func (s *QueryService) calculateConfidence(matches []models.FieldMatch) float64 {
	if len(matches) == 0 {
		return 0
	}
	
	// Average the match scores of all fields
	var total float64
	for _, match := range matches {
		total += match.MatchScore
	}
	
	confidence := total / float64(len(matches))
	
	// Adjust confidence based on number of matched fields
	// More matches = higher confidence, up to a point
	fieldCountFactor := math.Min(float64(len(matches))/3.0, 1.0)
	
	return confidence * fieldCountFactor
}

// EnhanceDescriptionWithFuzzy enhances keyword matching with fuzzy matching
func (s *QueryService) EnhanceDescriptionWithFuzzy(keywords []string, fields []models.Field) []string {
	var enhancedKeywords []string
	enhancedKeywords = append(enhancedKeywords, keywords...)
	
	// Extract all words from field descriptions
	var fieldWords []string
	for _, field := range fields {
		words := strings.Fields(strings.ToLower(field.Description))
		fieldWords = append(fieldWords, words...)
	}
	
	// Remove duplicates
	uniqueFieldWords := make(map[string]bool)
	for _, word := range fieldWords {
		uniqueFieldWords[word] = true
	}
	
	// For each keyword, find fuzzy matches
	for _, keyword := range keywords {
		matches := fuzzy.Find(keyword, stringMapToSlice(uniqueFieldWords))
		
		// Add top fuzzy matches to enhanced keywords
		for i, match := range matches {
			if i >= 3 { // Limit to top 3 fuzzy matches
				break
			}
			enhancedKeywords = append(enhancedKeywords, match)
		}
	}
	
	return enhancedKeywords
}

// stringMapToSlice converts a string map to a slice
func stringMapToSlice(m map[string]bool) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

// Import math for the confidence calculation
var math = struct {
	Min func(a, b float64) float64
}{
	Min: func(a, b float64) float64 {
		if a < b {
			return a
		}
		return b
	},
}
package models

// Field represents a database field mapping from the CSV file
type Field struct {
	ColumnName      string
	TableName       string
	SystemAFieldMap string
	SystemBFieldMap string
	Description     string
	FieldType       string
	JoinKey         string
	ForeignTable    string
	ForeignKey      string
}

// FieldMatch represents a matched field with score
type FieldMatch struct {
	ColumnName      string  `json:"column_name"`
	TableName       string  `json:"table_name"`
	FieldDescription string  `json:"field_description"`
	MatchScore      float64 `json:"match_score"`
}

// Join represents a JOIN relationship between tables
type Join struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Condition string `json:"condition"`
}

// QueryRequest represents the API request for generating a query
type QueryRequest struct {
	Description string `json:"description" binding:"required"`
	System      string `json:"system,omitempty"`
	Limit       int    `json:"limit,omitempty"`
}

// QueryResponse represents the API response with generated SQL
type QueryResponse struct {
	Query          string       `json:"query"`
	MatchedFields  []FieldMatch `json:"matched_fields"`
	JoinsUsed      []Join       `json:"joins_used"`
	Confidence     float64      `json:"confidence"`
	ProcessingTime int64        `json:"processing_time_ms"`
}

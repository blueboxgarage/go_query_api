# PostgreSQL Query Generator API - Go Implementation

**Claude Code Instructions for Building a Production-Ready Natural Language to SQL API Service**

## 🎯 Project Overview

Build a high-performance, production-scalable REST API service in Go that converts natural language descriptions into PostgreSQL queries using a CSV-based field mapping system. This service will allow non-technical users to query databases using plain English.

## 📋 Core Requirements

### **Primary Functionality**
- Parse CSV file containing database field mappings
- Accept natural language descriptions via HTTP POST
- Generate PostgreSQL queries using fuzzy field matching
- Return structured JSON responses with generated SQL and confidence scores
- Support multiple field mapping systems (system_a, system_b, default)
- Production-ready with proper error handling, logging, and metrics

### **CSV Field Mapping Format**
```csv
column_name,table_name,system_a_fieldmap,system_b_fieldmap,field_description,field_type
user_id,users,uid,user_identifier,Unique identifier for user,INTEGER
email,users,email_addr,user_email,User email address,VARCHAR
created_at,users,create_date,registration_date,Account creation timestamp,TIMESTAMP
order_total,orders,amount,total_cost,Total order value in cents,INTEGER
order_status,orders,status,order_state,Current status of order,VARCHAR
```

## 🔧 Technical Specifications

### **Tech Stack**
- **Language**: Go 1.21+ (latest stable)
- **Web Framework**: Gin or Echo for high-performance HTTP handling
- **CSV Processing**: Go standard library `encoding/csv`
- **String Matching**: `github.com/lithammer/fuzzysearch` or similar
- **Database**: `github.com/lib/pq` for PostgreSQL (optional)
- **Logging**: `github.com/sirupsen/logrus` or `slog` (Go 1.21+)
- **Configuration**: `github.com/spf13/viper` or environment variables
- **Testing**: Go standard library `testing` package

### **Key Dependencies (go.mod)**
```go
module query-generator-api

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/lithammer/fuzzysearch v1.1.8
    github.com/sirupsen/logrus v1.9.3
    github.com/spf13/viper v1.17.0
    github.com/lib/pq v1.10.9
    github.com/google/uuid v1.4.0
    github.com/prometheus/client_golang v1.17.0
    github.com/stretchr/testify v1.8.4
)
```

### **API Endpoints**
1. **POST /api/v1/generate-query** - Main query generation endpoint
2. **GET /health** - Health check endpoint  
3. **GET /api/v1/fields** - List available field mappings
4. **GET /metrics** - Prometheus metrics endpoint

### **Request/Response Format**

**POST /api/v1/generate-query**
```json
{
  "description": "get all active users with email addresses from last 30 days",
  "system": "system_a",
  "limit": 100,
  "request_id": "optional-trace-id"
}
```

**Response**
```json
{
  "query": "SELECT uid, email_addr, create_date\nFROM users\nWHERE status = 'active' AND create_date >= CURRENT_DATE - INTERVAL '30 days'\nLIMIT 100;",
  "matched_fields": [
    {
      "column_name": "uid",
      "table_name": "users", 
      "field_description": "Unique identifier for user",
      "field_type": "INTEGER",
      "match_score": 85.5,
      "matched_text": "user identifier"
    }
  ],
  "confidence": 0.855,
  "request_id": "abc-123",
  "processing_time_ms": 23
}
```

## 🏗 Architecture Requirements

### **Project Structure**
```
query-generator-api/
├── go.mod
├── go.sum
├── main.go                   # Entry point
├── cmd/
│   └── server/
│       └── main.go          # Server startup
├── internal/                # Private application code
│   ├── config/
│   │   └── config.go        # Configuration management
│   ├── handlers/            # HTTP request handlers
│   │   ├── health.go
│   │   ├── query.go
│   │   └── fields.go
│   ├── models/              # Data structures
│   │   ├── field_mapping.go
│   │   ├── request.go
│   │   └── response.go
│   ├── services/            # Business logic
│   │   ├── field_matcher.go
│   │   ├── nlp_processor.go
│   │   └── query_generator.go
│   ├── utils/               # Utilities
│   │   ├── csv_loader.go
│   │   └── logger.go
│   └── middleware/          # HTTP middleware
│       ├── logging.go
│       ├── cors.go
│       └── metrics.go
├── pkg/                     # Public/reusable packages
│   └── errors/
│       └── errors.go
├── configs/
│   └── config.yaml          # Configuration file
├── field_mappings.csv       # Sample field mapping data
├── tests/                   # Integration tests
│   ├── integration_test.go
│   └── testdata/
└── README.md
```

### **Core Components to Implement**

#### **1. Configuration (`internal/config/config.go`)**
```go
type Config struct {
    Server struct {
        Port         int    `mapstructure:"port"`
        Host         string `mapstructure:"host"`
        ReadTimeout  int    `mapstructure:"read_timeout"`
        WriteTimeout int    `mapstructure:"write_timeout"`
    } `mapstructure:"server"`
    
    CSV struct {
        FilePath  string  `mapstructure:"file_path"`
        Threshold float64 `mapstructure:"fuzzy_threshold"`
    } `mapstructure:"csv"`
    
    Database struct {
        URL             string `mapstructure:"url"`
        MaxConnections  int    `mapstructure:"max_connections"`
        ValidationMode  bool   `mapstructure:"validation_mode"`
    } `mapstructure:"database"`
}
```

#### **2. Data Models (`internal/models/`)**
```go
type FieldMapping struct {
    ColumnName       string `json:"column_name" csv:"column_name"`
    TableName        string `json:"table_name" csv:"table_name"`
    SystemAFieldmap  string `json:"system_a_fieldmap" csv:"system_a_fieldmap"`
    SystemBFieldmap  string `json:"system_b_fieldmap" csv:"system_b_fieldmap"`
    FieldDescription string `json:"field_description" csv:"field_description"`
    FieldType        string `json:"field_type" csv:"field_type"`
}

type QueryRequest struct {
    Description string  `json:"description" binding:"required"`
    System      *string `json:"system,omitempty"`
    Limit       *int    `json:"limit,omitempty"`
    RequestID   string  `json:"request_id,omitempty"`
}

type QueryResponse struct {
    Query            string         `json:"query"`
    MatchedFields    []MatchedField `json:"matched_fields"`
    Confidence       float64        `json:"confidence"`
    RequestID        string         `json:"request_id"`
    ProcessingTimeMs int64          `json:"processing_time_ms"`
}
```

#### **3. Field Matcher Service (`internal/services/field_matcher.go`)**
- Load field mappings from CSV using `encoding/csv`
- Implement fuzzy string matching using fuzzysearch library
- Extract keywords from natural language input with regex
- Score and rank field matches with confidence calculation
- Support system-specific field name resolution
- In-memory caching of parsed mappings

#### **4. NLP Processor (`internal/services/nlp_processor.go`)**
- Keyword extraction with stop word filtering
- Intent detection using regex patterns and keyword analysis
- Temporal pattern recognition ("last 30 days", "recent", "yesterday")
- Filter pattern recognition ("active", "status", conditions)
- Confidence scoring algorithm implementation

#### **5. Query Generator (`internal/services/query_generator.go`)**
- Generate SELECT queries for general data retrieval
- Generate COUNT/aggregate queries for quantitative questions
- Generate GROUP BY queries for categorization and grouping
- Build WHERE clauses based on detected patterns and keywords
- Handle LIMIT clauses and suggest JOIN operations
- SQL injection prevention with parameterized query patterns

#### **6. HTTP Handlers (`internal/handlers/`)**
- HTTP request handlers using Gin context
- JSON request/response binding and validation
- Proper HTTP status codes and error responses
- Request logging and tracing with structured logging
- Input validation and sanitization

## 🎯 Implementation Priorities

### **Phase 1: Foundation (Start Here)**
1. Initialize Go module with required dependencies
2. Set up basic Gin server with health check endpoint
3. Implement configuration management with Viper
4. Set up structured logging with logrus or slog
5. Define core data structures and error types

### **Phase 2: CSV Processing and Basic API**
1. Implement CSV parser using `encoding/csv`
2. Create field mapping storage with in-memory cache
3. Add basic string matching (contains/substring operations)
4. Implement `/api/v1/fields` endpoint for field listing
5. Create basic `/api/v1/generate-query` with simple responses

### **Phase 3: Intelligence Layer**
1. Integrate fuzzy string matching with scoring
2. Implement keyword extraction and NLP processing
3. Add query generation for SELECT statements
4. Implement intent detection for different query types
5. Add confidence scoring and response ranking

### **Phase 4: Production Features**
1. Add comprehensive error handling and validation
2. Implement middleware for logging, CORS, and metrics
3. Add Prometheus metrics collection
4. Optimize performance and add caching strategies
5. Create integration tests and API documentation

## 🧠 Algorithm Guidelines

### **Fuzzy Matching Implementation**
```go
import "github.com/lithammer/fuzzysearch/fuzzy"

func calculateMatchScore(fieldDesc, query string) float64 {
    // Use fuzzy matching
    distance := fuzzy.LevenshteinDistance(
        strings.ToLower(fieldDesc), 
        strings.ToLower(query),
    )
    
    // Normalize to 0-100 score
    maxLen := math.Max(float64(len(fieldDesc)), float64(len(query)))
    if maxLen == 0 {
        return 0
    }
    
    return (1.0 - float64(distance)/maxLen) * 100.0
}
```

### **Intent Detection Patterns**
```go
type QueryIntent int

const (
    IntentSelect QueryIntent = iota
    IntentCount
    IntentGroupBy
    IntentAggregate
    IntentFilter
)

func detectIntent(description string) QueryIntent {
    desc := strings.ToLower(description)
    
    countPatterns := []string{"how many", "count", "total number"}
    groupPatterns := []string{"by category", "group by", "breakdown"}
    
    for _, pattern := range countPatterns {
        if strings.Contains(desc, pattern) {
            return IntentCount
        }
    }
    // ... more pattern matching
}
```

### **Query Generation Strategy**
```go
type QueryBuilder struct {
    Intent        QueryIntent
    MatchedFields []MatchedField
    Conditions    []WhereCondition
    Limit         *int
}

func (qb *QueryBuilder) BuildSQL() string {
    switch qb.Intent {
    case IntentSelect:
        return qb.buildSelectQuery()
    case IntentCount:
        return qb.buildCountQuery()
    case IntentGroupBy:
        return qb.buildGroupByQuery()
    }
}
```

## 📊 Performance Requirements

- **Startup Time**: < 100ms (CSV loading + server initialization)
- **Request Latency**: < 50ms (95th percentile)
- **Memory Usage**: < 50MB for typical field mapping files
- **Throughput**: 5000+ requests/second under load
- **Concurrent Connections**: Support 10,000+ simultaneous connections

## 🔒 Production Features

### **Error Handling**
```go
type APIError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

func (e APIError) Error() string {
    return e.Message
}
```

### **Middleware Stack**
- Request logging with structured fields
- CORS handling for web clients
- Request ID generation and tracing
- Prometheus metrics collection
- Rate limiting (optional)
- Request timeout handling

### **Health Checks**
```go
type HealthResponse struct {
    Status      string            `json:"status"`
    Version     string            `json:"version"`
    Uptime      string            `json:"uptime"`
    Dependencies map[string]string `json:"dependencies"`
}
```

## 🧪 Testing Strategy

### **Unit Tests**
```go
func TestCSVParsing(t *testing.T) {
    csvData := `column_name,table_name,field_description
user_id,users,User identifier`
    
    mappings, err := parseCSVData(strings.NewReader(csvData))
    assert.NoError(t, err)
    assert.Len(t, mappings, 1)
    assert.Equal(t, "user_id", mappings[0].ColumnName)
}

func TestFieldMatching(t *testing.T) {
    matcher := NewFieldMatcher(sampleMappings)
    matches := matcher.FindMatches("user email", nil, 10)
    assert.NotEmpty(t, matches)
}
```

### **Integration Tests**
```go
func TestGenerateQueryEndpoint(t *testing.T) {
    router := setupTestRouter()
    
    w := httptest.NewRecorder()
    body := `{"description": "get user emails"}`
    req := httptest.NewRequest("POST", "/api/v1/generate-query", 
        strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
    // ... verify response structure
}
```

## 🔍 Example Use Cases to Support

1. **Simple Selection**: "get user emails" → `SELECT email FROM users;`
2. **Filtered Data**: "active users from last week" → `SELECT * FROM users WHERE status = 'active' AND created_at >= CURRENT_DATE - INTERVAL '7 days';`
3. **Counting**: "how many orders were placed" → `SELECT COUNT(*) FROM orders;`
4. **Grouping**: "orders by status" → `SELECT order_status, COUNT(*) FROM orders GROUP BY order_status;`
5. **Aggregation**: "average order value" → `SELECT AVG(order_total) FROM orders;`
6. **Multi-system**: Use system_a field mappings when system="system_a"

## 🚀 Getting Started Commands

```bash
# Initialize Go module
go mod init query-generator-api
cd query-generator-api

# Add dependencies
go get github.com/gin-gonic/gin
go get github.com/lithammer/fuzzysearch
go get github.com/sirupsen/logrus
go get github.com/spf13/viper
go get github.com/stretchr/testify

# Run in development
go run main.go

# Run tests
go test ./...

# Run with environment variables
CSV_FILE_PATH=./field_mappings.csv go run main.go

# Build for production
go build -o query-generator-api main.go

# Run binary
./query-generator-api
```

## 💡 Implementation Notes

### **Go Best Practices**
- Use interfaces for dependency injection and testing
- Implement proper error handling with wrapped errors
- Use context.Context for request timeouts and cancellation
- Structure packages following Go conventions (internal/ for private code)

### **Gin Framework Patterns**
```go
func setupRouter(config *Config, services *Services) *gin.Engine {
    r := gin.New()
    
    // Middleware
    r.Use(gin.Logger())
    r.Use(gin.Recovery())
    r.Use(middleware.CORS())
    r.Use(middleware.RequestID())
    
    // Routes
    v1 := r.Group("/api/v1")
    {
        v1.POST("/generate-query", handlers.GenerateQuery(services))
        v1.GET("/fields", handlers.ListFields(services))
    }
    
    r.GET("/health", handlers.Health())
    r.GET("/metrics", gin.WrapH(promhttp.Handler()))
    
    return r
}
```

### **Configuration Management**
```go
func LoadConfig() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("./configs")
    
    // Environment variable support
    viper.SetEnvPrefix("QUERY_API")
    viper.AutomaticEnv()
    
    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }
    
    var config Config
    return &config, viper.Unmarshal(&config)
}
```

## 🎯 Success Criteria

The implementation is successful when:
- ✅ Loads CSV field mappings on startup with error handling
- ✅ Serves HTTP requests efficiently with Gin framework
- ✅ Generates valid PostgreSQL queries from natural language
- ✅ Returns structured JSON with confidence scores and timing
- ✅ Handles errors gracefully with appropriate HTTP status codes
- ✅ Supports multiple field mapping systems seamlessly
- ✅ Processes 95% of requests under 50ms
- ✅ Handles 5000+ requests/second sustained load
- ✅ Includes comprehensive logging and monitoring
- ✅ Has thorough unit and integration test coverage

## 🔧 Deployment Configuration

### **Dockerfile**
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o query-generator-api main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/query-generator-api .
COPY --from=builder /app/field_mappings.csv .
EXPOSE 8080
CMD ["./query-generator-api"]
```

### **Environment Variables**
```bash
# Server Configuration
PORT=8080
HOST=0.0.0.0
READ_TIMEOUT=30
WRITE_TIMEOUT=30

# CSV Configuration
CSV_FILE_PATH=./field_mappings.csv
FUZZY_THRESHOLD=30.0

# Database (optional)
DATABASE_URL=postgresql://user:pass@localhost/db
VALIDATION_MODE=false

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

### **Docker Compose Example**
```yaml
version: '3.8'
services:
  query-api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - CSV_FILE_PATH=/app/field_mappings.csv
      - LOG_LEVEL=info
    volumes:
      - ./field_mappings.csv:/app/field_mappings.csv:ro
```

**Start with the Gin foundation and CSV processing. Go's simplicity will let you iterate quickly on the NLP logic and get to production faster than Rust, while still providing excellent performance for this use case.**

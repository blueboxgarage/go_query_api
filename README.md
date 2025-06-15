# GO Query API - Query Generator

**Claude Code Build Instructions: Convert Natural Language to SQL using Go**

## 🎯 Project Objective

Build a REST API service in Go that converts natural language descriptions into PostgreSQL queries using a CSV-based field mapping system.

**Core Value Proposition**: Allow non-technical users to query databases using plain English instead of writing SQL.

## 📋 Functional Requirements

### **Primary Features**
1. **CSV Field Mapping**: Load database field mappings from a CSV file on startup
2. **Natural Language Processing**: Accept plain English descriptions and convert to SQL
3. **Multiple Query Types**: Support SELECT, COUNT, GROUP BY, and DISTINCT query generation
4. **Multi-Table Joins**: Automatically generate JOIN clauses when fields span multiple tables
5. **Unique Instance Detection**: Support DISTINCT queries and unique value requests
6. **System Mapping**: Handle different field name mappings (system_a, system_b, default)
7. **Confidence Scoring**: Return match confidence to indicate query reliability
8. **REST API**: Provide HTTP endpoints for query generation and field listing

### **Input/Output Specification**

**CSV Input Format:**
```
column_name,table_name,system_a_fieldmap,system_b_fieldmap,field_description,field_type,join_key,foreign_table,foreign_key
user_id,users,uid,user_identifier,Unique identifier for user,INTEGER,,,
email,users,email_addr,user_email,User email address,VARCHAR,,,
order_id,orders,order_num,transaction_id,Unique order identifier,INTEGER,,,
user_id,orders,customer_id,user_ref,User who placed order,INTEGER,user_id,users,user_id
total_amount,orders,order_total,amount,Total order value in cents,INTEGER,,,
product_name,products,name,product_title,Product display name,VARCHAR,,,
order_item_id,order_items,item_id,line_item_id,Order line item identifier,INTEGER,,,
order_id,order_items,order_ref,order_reference,Reference to parent order,INTEGER,order_id,orders,order_id
product_id,order_items,prod_id,product_reference,Reference to product,INTEGER,product_id,products,product_id
```

**API Request:**
```json
{
  "description": "get unique email addresses of users who placed orders with product names",
  "system": "system_a",
  "limit": 100
}
```

**API Response:**
```json
{
  "query": "SELECT DISTINCT u.email_addr, p.name\nFROM users u\nJOIN orders o ON u.uid = o.customer_id\nJOIN order_items oi ON o.order_num = oi.order_ref\nJOIN products p ON oi.prod_id = p.product_id\nLIMIT 100;",
  "matched_fields": [
    {
      "column_name": "email_addr",
      "table_name": "users",
      "field_description": "User email address",
      "match_score": 92.3
    },
    {
      "column_name": "name", 
      "table_name": "products",
      "field_description": "Product display name",
      "match_score": 87.1
    }
  ],
  "joins_used": [
    {"from": "users", "to": "orders", "condition": "users.uid = orders.customer_id"},
    {"from": "orders", "to": "order_items", "condition": "orders.order_num = order_items.order_ref"},
    {"from": "order_items", "to": "products", "condition": "order_items.prod_id = products.product_id"}
  ],
  "confidence": 0.89,
  "processing_time_ms": 45
}
```

## 🔧 Technical Requirements

### **Technology Stack**
- **Language**: Go 1.21+
- **HTTP Framework**: Gin or Echo for REST API
- **CSV Processing**: Go standard library
- **String Matching**: Fuzzy string matching library
- **Testing**: Go standard testing package
- **Configuration**: Environment variables or YAML config

### **Required Dependencies**
- HTTP web framework: [gin-gonic/gin](https://github.com/gin-gonic/gin) v1.9.1
- Fuzzy string matching: [lithammer/fuzzysearch](https://github.com/lithammer/fuzzysearch) v1.1.8
- Logging: [sirupsen/logrus](https://github.com/sirupsen/logrus) v1.9.3
- Testing: [stretchr/testify](https://github.com/stretchr/testify) (for testing)

To install dependencies:
```bash
go mod download github.com/gin-gonic/gin
go get github.com/lithammer/fuzzysearch/fuzzy
go get github.com/sirupsen/logrus
```

### **API Endpoints**
1. `POST /api/v1/generate-query` - Main query generation
   - Request body:
     ```json
     {
       "description": "get orders with total order value",
       "system": "SystemA",
       "limit": 10
     }
     ```
   - Required fields: `description`
   - Optional fields: `system`, `limit`
   - Response: Generated SQL query, matched fields, joins used, confidence score, and processing time

2. `GET /health` - Health check
   - Returns status 200 OK if the service is running properly

3. `GET /api/v1/fields` - List available field mappings
   - Optional query param: `system` (e.g., `?system=SystemA`)
   - Returns all field mappings, optionally filtered by system

## 🏗 Architecture Requirements

### **Project Structure**
```
query-generator-api/
├── main.go                   # Application entry point
├── internal/
│   ├── handlers/             # HTTP request handlers
│   ├── services/             # Business logic (matching, generation)
│   ├── models/               # Data structures
│   └── config/               # Configuration management
├── field_mappings.csv        # Field mapping data
└── tests/                    # Test files
```

### **Core Components to Build**

#### **1. Configuration Management**
- Load settings from environment variables
- Support for CSV file path, server port, matching thresholds
- Validation of required configuration values

#### **2. CSV Processing Module**
- Parse CSV file with field mappings using Go's `encoding/csv`
- Support extended CSV format with join relationship metadata (join_key, foreign_table, foreign_key)
- Build relationship graph between tables for JOIN path discovery
- Validate CSV structure and data integrity including relationship consistency
- Load mappings and relationships into memory on application startup
- Handle CSV parsing errors gracefully

#### **3. Field Matching Service**
- Implement fuzzy string matching between user descriptions and field descriptions
- Extract keywords from natural language input
- Score field matches and rank by relevance
- Support minimum confidence threshold filtering
- Resolve field names based on specified system (system_a, system_b, default)

#### **4. Natural Language Processor**
- Extract meaningful keywords from user descriptions
- Detect query intent patterns:
  - COUNT queries: "how many", "count", "total number"
  - GROUP BY queries: "by category", "group by", "breakdown"
  - DISTINCT queries: "unique", "distinct", "different", "deduplicate"
  - SELECT queries: default for general data retrieval
- Identify temporal filters: "last 30 days", "recent", "yesterday"
- Recognize status filters: "active", "pending", "completed"
- Detect cross-table requirements from field combinations

#### **5. SQL Query Generator**
- Generate SELECT statements for general data queries
- Generate COUNT statements for quantitative questions
- Generate GROUP BY statements for categorization queries
- Generate DISTINCT queries for unique value requests
- Build WHERE clauses based on detected patterns
- **JOIN Path Discovery**: Find optimal JOIN paths between tables using relationship graph
- **Multi-table Query Construction**: Generate proper JOIN syntax when fields span multiple tables
- Add LIMIT clauses when specified
- Prevent SQL injection through proper query construction

#### **6. HTTP Request Handlers**
- Handle JSON request parsing and validation
- Implement proper HTTP status codes and error responses
- Add request logging and error handling
- Support CORS for web client access

## 🎯 Business Logic Requirements

### **Field Matching Algorithm**
- Use fuzzy string matching to find relevant fields
- Score matches between 0-100 (higher = better match)
- Consider both field descriptions and column names in matching
- Apply configurable minimum threshold (default: 30.0)
- Return top N matching fields (default: 10)

### **Intent Detection Logic**
- Analyze description text for query type indicators
- Default to SELECT queries for general requests
- Detect COUNT intent from quantitative language patterns
- Detect GROUP BY intent from categorization language patterns
- **Detect DISTINCT intent from uniqueness language patterns**: "unique", "distinct", "different values", "deduplicate"
- Support aggregation functions (SUM, AVG, MAX, MIN)
- **Identify multi-table requirements** when field combinations require JOINs

### **Query Construction Rules**
- Use matched fields to determine SELECT columns
- Apply DISTINCT keyword when uniqueness is requested
- **Determine required tables** from matched field locations
- **Calculate JOIN path** using shortest path algorithm between required tables
- **Generate JOIN clauses** with proper ON conditions using relationship metadata
- Infer primary table from highest-scoring field match or most central table in JOIN path
- Add WHERE clauses for detected temporal and status filters
- Apply LIMIT when specified in request
- Ensure generated SQL is syntactically valid PostgreSQL
- **Optimize JOIN order** for query performance

### **Confidence Scoring Algorithm**
- Calculate overall confidence based on field match scores
- Normalize scores to 0-1 range for API response
- Factor in number of matched fields and intent clarity
- Provide confidence threshold recommendations

## 📊 Performance Requirements

- **Startup Time**: Application ready in under 5 seconds
- **Response Time**: 95% of requests processed under 100ms
- **Memory Usage**: Efficient memory usage for CSV data caching
- **Throughput**: Handle 1000+ requests per second
- **Concurrency**: Support multiple simultaneous requests

## 🔒 Quality Requirements

### **Error Handling**
- Graceful handling of malformed CSV files
- Proper HTTP error responses with meaningful messages
- Input validation for all API requests
- Logging of errors for debugging and monitoring

### **Testing Requirements**
- Unit tests for all core business logic functions
- Integration tests for HTTP endpoints
- Test coverage for CSV parsing and field matching
- Example test cases for various natural language inputs

### **Validation Requirements**
- Validate CSV file structure on startup
- Validate JSON request format and required fields
- Sanitize user input to prevent security issues
- Verify generated SQL syntax correctness

## 🔍 Example Use Cases to Support

### **Simple Queries**
- Input: "get user emails" → Output: `SELECT email FROM users;`
- Input: "show all products" → Output: `SELECT * FROM products;`

### **Unique/Distinct Queries**
- Input: "unique email addresses" → Output: `SELECT DISTINCT email FROM users;`
- Input: "different product categories" → Output: `SELECT DISTINCT category FROM products;`

### **Multi-Table JOIN Queries**
- Input: "user emails and their order totals" → Output: `SELECT u.email, o.total_amount FROM users u JOIN orders o ON u.user_id = o.user_id;`
- Input: "product names in orders" → Output: `SELECT p.name FROM products p JOIN order_items oi ON p.product_id = oi.product_id JOIN orders o ON oi.order_id = o.order_id;`

### **Complex Cross-Table Queries**
- Input: "unique customers who bought electronics" → Output: `SELECT DISTINCT u.email FROM users u JOIN orders o ON u.user_id = o.user_id JOIN order_items oi ON o.order_id = oi.order_id JOIN products p ON oi.product_id = p.product_id WHERE p.category = 'electronics';`

### **Filtered Queries**
- Input: "active users from last week" → Output: `SELECT * FROM users WHERE status = 'active' AND created_at >= CURRENT_DATE - INTERVAL '7 days';`
- Input: "pending orders with customer emails" → Output: `SELECT o.*, u.email FROM orders o JOIN users u ON o.user_id = u.user_id WHERE o.status = 'pending';`

### **Count Queries**
- Input: "how many orders were placed" → Output: `SELECT COUNT(*) FROM orders;`
- Input: "total number of unique customers" → Output: `SELECT COUNT(DISTINCT user_id) FROM orders;`

### **Group By Queries**
- Input: "orders by status" → Output: `SELECT status, COUNT(*) FROM orders GROUP BY status;`
- Input: "users by region with order counts" → Output: `SELECT u.region, COUNT(o.order_id) FROM users u LEFT JOIN orders o ON u.user_id = o.user_id GROUP BY u.region;`

### **System-Specific Mapping**
- When system="system_a", use system_a_fieldmap column for field names
- When system="system_b", use system_b_fieldmap column for field names
- Default to column_name when no system specified

## 🎯 Implementation Phases

### **Phase 1: Foundation**
1. Set up Go project with required dependencies
2. Implement basic HTTP server with health check
3. Create CSV parsing functionality
4. Define core data structures

### **Phase 2: Core Logic**
1. Implement basic field matching with substring search
2. Add simple query generation for SELECT statements
3. Create main API endpoint with hardcoded examples
4. Add basic error handling and validation

### **Phase 3: Intelligence & JOINs**
1. Integrate fuzzy string matching for better field matching
2. Implement intent detection for different query types including DISTINCT
3. **Build relationship graph from CSV join metadata**
4. **Implement JOIN path discovery algorithm (shortest path between tables)**
5. **Add multi-table query generation with proper JOIN syntax**
6. Add WHERE clause generation for common patterns
7. Implement confidence scoring system

### **Phase 4: Polish**
1. Add comprehensive error handling and logging
2. Implement thorough testing suite
3. Add configuration management
4. Optimize performance and add caching if needed

## ✅ Success Criteria

The implementation is complete when:
- Loads CSV field mappings with relationship metadata successfully on startup
- Accepts HTTP POST requests with natural language descriptions
- Returns valid PostgreSQL queries with confidence scores
- Supports SELECT, COUNT, GROUP BY, and DISTINCT query types
- **Automatically generates JOIN clauses when fields span multiple tables**
- **Handles complex multi-table relationships and finds optimal JOIN paths**
- **Supports unique/distinct value queries properly**
- Handles multiple field mapping systems correctly
- Processes requests efficiently with proper error handling
- Includes comprehensive test coverage including JOIN scenarios
- Provides clear API documentation and usage examples

## 📋 Deliverables Expected

1. **Working Go application** that meets all functional requirements
2. **HTTP REST API** with proper endpoint implementation
3. **Test suite** covering core functionality and edge cases
4. **Configuration system** for deployment flexibility
5. **Documentation** including API usage examples
6. **Sample CSV file** with representative field mappings

## 🚀 Getting Started Guide

### Installation and Dependencies

```bash
# Clone the repository
git clone https://github.com/yourusername/go_query_api.git
cd go_query_api

# Install required dependencies
go mod download github.com/gin-gonic/gin
go get github.com/lithammer/fuzzysearch/fuzzy
go get github.com/sirupsen/logrus
```

### Build and Run

```bash
# Using Make (recommended)
make build     # Build the application
make run       # Run the application
make help      # Show all available make commands

# Manual build and run
go build -o query-api ./main.go
./query-api

# Run directly without building
go run main.go
```

### Command-Line Options

The application supports the following command-line flags:

```
-port string    Server port (overrides config)
-csv string     Path to field mappings CSV (overrides config)
-debug          Enable debug mode
-help           Show help message
-version        Show version information
```

Example usage:
```bash
./query-api --port 9000 --csv ./custom_fields.csv --debug
```

### Testing

```bash
# Using Make (recommended)
make test       # Run all tests
make fmt        # Format code
make lint       # Run linter

# Manual testing
go test ./...
go test ./tests/field_service_test.go
go test ./tests/query_service_test.go
go test ./tests/handlers_test.go

# Format code manually
gofmt -w .

# Run linter manually (if golangci-lint is installed)
golangci-lint run
```

### Making API Requests

1. List all available fields:
```bash
# Create a simple HTTP client
go run client.go
```

2. Generate a query from natural language:
```bash
# Create a request with a description
go run test_query.go
```

Sample request payload:
```json
{
  "description": "get orders with total order value",
  "system": "SystemA",
  "limit": 10
}
```

Sample response:
```json
{
  "query": "SELECT orders.total_amount FROM orders o LIMIT 10",
  "matched_fields": [
    {
      "column_name": "total_amount",
      "table_name": "orders",
      "field_description": "Total order value in cents",
      "match_score": 75
    }
  ],
  "joins_used": null,
  "confidence": 25,
  "processing_time_ms": 0
}
```

For more complex queries with JOINs:
```json
{
  "description": "get order line item identifiers with product display name",
  "system": "SystemB",
  "limit": 20
}
```

**Focus on getting basic functionality working before adding advanced features. Start simple and iterate.**

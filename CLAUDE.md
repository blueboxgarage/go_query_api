# Go Query API Guidelines

## Build Commands
- Build: `go build -o query-api ./main.go`
- Run: `go run main.go`
- Test: `go test ./...`
- Test single file: `go test ./path/to/package -run TestName`
- Lint: `golangci-lint run`
- Format: `gofmt -w .`

## Code Style
- Packages: Use standard Go package structure with `internal/` for non-public code
- Imports: Group standard lib, 3rd party, and local imports with a blank line between
- Error handling: Always check errors, use meaningful error messages
- Naming: CamelCase for exported, camelCase for unexported
- Functions: Keep functions small and focused on single responsibility
- Comments: Document public APIs, follow godoc format
- Testing: Write table-driven tests for all core functionality
- Configuration: Use environment variables or YAML for configs

## Required Libraries
- HTTP: gin-gonic/gin
- String matching: lithammer/fuzzysearch
- Logging: sirupsen/logrus or standard slog
- Testing: stretchr/testify
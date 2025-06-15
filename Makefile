.PHONY: build run test clean lint fmt examples help

# Build variables
BINARY_NAME=query-api
BUILD_DIR=./build

help:
	@echo "Go Query API Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build        - Build the application"
	@echo "  make run          - Run the application"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make lint         - Run linter"
	@echo "  make fmt          - Format code"
	@echo "  make examples     - Run example clients"
	@echo "  make help         - Show this help message"

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./main.go

run:
	go run main.go

test:
	go test -v ./...

clean:
	rm -rf $(BUILD_DIR)

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

examples:
	@echo "Running examples..."
	@mkdir -p examples
	@if [ -f client.go ]; then cp client.go examples/client.go; fi
	@if [ -f test_query.go ]; then cp test_query.go examples/query_simple.go; fi
	@if [ -f test_query2.go ]; then cp test_query2.go examples/query_complex.go; fi
	@echo "Starting server in background..."
	@go run main.go &
	@SERVER_PID=$$!
	@sleep 2
	@echo "Running simple field listing example..."
	@cd examples && go run client.go
	@echo "Running simple query example..."
	@cd examples && go run query_simple.go
	@echo "Running complex query example..."
	@cd examples && go run query_complex.go
	@echo "Stopping server..."
	@kill $$SERVER_PID || true

setup-dev: 
	go mod download
	go get github.com/gin-gonic/gin
	go get github.com/lithammer/fuzzysearch/fuzzy
	go get github.com/sirupsen/logrus
	go get -t github.com/stretchr/testify/assert

.DEFAULT_GOAL := help
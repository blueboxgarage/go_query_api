# Go Query API Examples

This directory contains example clients for the Go Query API.

## Running the Examples

Make sure the server is running first:

```bash
cd ..
go run main.go
```

Then in another terminal, run the examples:

```bash
# List all fields
go run client.go

# Run a simple query
go run query_simple.go

# Run a complex query with joins
go run query_complex.go
```

## Example Descriptions

1. `client.go` - Simple HTTP client that fetches all available field mappings
2. `query_simple.go` - Example that queries for orders with total order value
3. `query_complex.go` - More complex example that queries order items with product information requiring a JOIN
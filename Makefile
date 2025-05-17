.PHONY: build run clean test

# Build the application
build:
	go build -o bin/deel cmd/server/main.go

# Run the application
run:
	go run cmd/server/main.go

# Clean the binary
clean:
	rm -f bin/deel

# Test the application
test:
	go test ./...

# Run linting
lint:
	go vet ./...

# Run the application with hot reload (requires air: https://github.com/cosmtrek/air)
dev:
	air -c .air.toml

# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o rss-reader ./cmd/server

# Final stage
FROM alpine:3.21

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy the binary from builder
COPY --from=builder /app/rss-reader .
# Copy static assets and templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates

# Create a data directory for persistence
RUN mkdir -p /app/data

# Declare volume for data persistence
VOLUME ["/app/data"]

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application
CMD ["./rss-reader"]

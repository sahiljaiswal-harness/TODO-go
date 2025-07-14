# Start from the latest golang image as builder
FROM golang:1.24 as builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod .
COPY go.sum .

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN go build -o todo-app .

# Start a new stage from a minimal image
FROM debian:bookworm-slim

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy the built binary from the builder
COPY --from=builder /app/todo-app .

# Expose the port your app runs on
EXPOSE 8080

# Set environment variable for Postgres connection (override in production)
ENV POSTGRES_URI="postgres://postgres:postgres@localhost:5432/apps?sslmode=disable"

# Command to run the executable
CMD ["./todo-app"]

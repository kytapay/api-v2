# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install git and ca-certificates (needed for HTTPS requests)
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o api-v2 .

# Final stage
FROM alpine:latest

# Install ca-certificates and wget for HTTPS and healthcheck
RUN apk --no-cache add ca-certificates wget

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/api-v2 .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./api-v2"]


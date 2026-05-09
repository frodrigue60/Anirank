# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git vips-dev gcc musl-dev

# Copy dependencies first to leverage Docker cache
COPY go.mod go.sum ./

# Download Go modules
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o main ./cmd/api/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies (like CA certificates for TLS)
RUN apk add --no-cache ca-certificates vips vips-tools libavif-apps libwebp-tools

# Copy the binary
COPY --from=builder /app/main .

# Expose port (Backend port 8080)
EXPOSE 8080

# Command to run the application
CMD ["./main"]

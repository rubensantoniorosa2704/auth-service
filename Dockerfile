# Build stage
FROM golang:1.25.6-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy corporate CA certificate and update certificates
COPY corp-ca.crt /usr/local/share/ca-certificates/corp-ca.crt
RUN update-ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server/main.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS and create non-root user
RUN apk --no-cache add ca-certificates && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Copy corporate CA certificate if needed for custom TLS
COPY --chown=appuser:appuser corp-ca.crt /usr/local/share/ca-certificates/corp-ca.crt
RUN update-ca-certificates

# Switch to non-root user
USER appuser

# Expose gRPC port
EXPOSE 50051

# Run the server
CMD ["./server"]

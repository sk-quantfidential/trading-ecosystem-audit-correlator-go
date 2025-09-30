# Multi-stage build for audit-correlator-go
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /build

# Install git and ca-certificates (needed for dependency fetching)
RUN apk add --no-cache git ca-certificates

# Copy audit-data-adapter-go dependency first (from parent context)
COPY audit-data-adapter-go/ ./audit-data-adapter-go/

# Copy audit-correlator-go files
COPY audit-correlator-go/go.mod audit-correlator-go/go.sum ./audit-correlator-go/

# Set working directory to audit-correlator-go
WORKDIR /build/audit-correlator-go

# Download dependencies (now can find ../audit-data-adapter-go)
RUN go mod download

# Copy source code
COPY audit-correlator-go/ .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o audit-correlator ./cmd/server

# Runtime stage
FROM alpine:3.19

# Install ca-certificates for HTTPS connections and wget for health checks
RUN apk --no-cache add ca-certificates wget

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /build/audit-correlator-go/audit-correlator /app/audit-correlator

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 8083 9093

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8083/api/v1/health || exit 1

# Run the application
CMD ["./audit-correlator"]
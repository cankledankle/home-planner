# Multi-stage build combining frontend and backend
# Production-ready with security hardening

# Stage 1: Build Frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /app

# Copy package files first for better layer caching
COPY package*.json ./
RUN npm ci && npm cache clean --force

# Copy frontend source and build
COPY . .
RUN npm run build

# Stage 2: Build Backend
FROM golang:1.23-alpine AS backend-builder

WORKDIR /app

# Install git and ca-certificates for fetching dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source code
COPY backend/ .

# Build the binary with security flags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -a -installsuffix cgo \
    -o server ./cmd/server/main.go

# Stage 3: Production image - Alpine for security + debugging capabilities
FROM alpine:3.19

WORKDIR /app

# Install required packages: ca-certificates for HTTPS, wget for healthcheck
RUN apk add --no-cache ca-certificates wget && \
    rm -rf /var/cache/apk/*

# Create non-root user with specific UID/GID
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Copy the backend binary
COPY --from=backend-builder --chown=appuser:appgroup /app/server .

# Copy migrations
COPY --from=backend-builder --chown=appuser:appgroup /app/migrations ./migrations

# Copy built frontend static files
COPY --from=frontend-builder --chown=appuser:appgroup /app/build ./static

# Switch to non-root user
USER appuser:appgroup

# Expose port
EXPOSE 8080

# Set environment variable for static files path
ENV STATIC_PATH=/app/static

# Health check - app exposes /health endpoint
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the server
ENTRYPOINT ["./server"]

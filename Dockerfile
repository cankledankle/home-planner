# Multi-stage build combining frontend and backend

# Stage 1: Build Frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /app

# Copy and install frontend dependencies
COPY package*.json ./
RUN npm ci

# Copy frontend source and build
COPY . .
RUN npm run build

# Stage 2: Build Backend
FROM golang:1.25.6-alpine AS backend-builder

WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source code
COPY backend/ .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go

# Stage 3: Production image
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy the backend binary
COPY --from=backend-builder /app/server .

# Copy migrations
COPY --from=backend-builder /app/migrations ./migrations

# Copy built frontend static files
COPY --from=frontend-builder /app/build ./static

# Expose port
EXPOSE 8080

# Set environment variable for static files path
ENV STATIC_PATH=/app/static

# Run the server
CMD ["./server"]
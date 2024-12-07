# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go.mod files
COPY go.mod ./
COPY user/go.mod user/

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN cd user/api/cmd/api && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/api

# Runtime stage
FROM alpine:3.18

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy binary from build stage
COPY --from=builder /app/api .

# Environment variables
ENV API_PORT=8083
ENV BRAIN_ENDPOINT=http://fleet-brain:8081
ENV EDGE_ENDPOINT=http://fleet-edge:8082
ENV JWT_SECRET=change-me-in-production
ENV API_LOG_LEVEL=info

# Expose port
EXPOSE 8083

CMD ["./api"]

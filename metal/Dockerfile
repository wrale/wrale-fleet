# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go.mod and go.sum files
COPY go.mod ./
COPY metal/go.mod metal/
COPY metal/hw/go.mod metal/hw/
COPY metal/hw/go.sum metal/hw/

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN cd metal/core/cmd/metald && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/metald

# Runtime stage
FROM alpine:3.18

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy binary from build stage
COPY --from=builder /app/metald .

# Environment variables
ENV METAL_PORT=8080
ENV METAL_LOG_LEVEL=info

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./metald"]

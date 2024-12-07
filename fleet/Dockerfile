# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go.mod files
COPY go.mod ./
COPY fleet/go.mod fleet/

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build brain service
RUN cd fleet/brain/cmd/brain && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/brain

# Build edge service
RUN cd fleet/edge/cmd/edge && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/edge

# Runtime stage
FROM alpine:3.18

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy binaries from build stage
COPY --from=builder /app/brain .
COPY --from=builder /app/edge .

# Environment variables
ENV BRAIN_PORT=8081
ENV EDGE_PORT=8082
ENV METAL_ENDPOINT=http://metal:8080
ENV BRAIN_LOG_LEVEL=info
ENV EDGE_LOG_LEVEL=info

# Expose ports
EXPOSE 8081 8082

# Use supervisord to manage multiple processes
RUN apk add --no-cache supervisor
COPY fleet/supervisord.conf /etc/supervisord.conf

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]

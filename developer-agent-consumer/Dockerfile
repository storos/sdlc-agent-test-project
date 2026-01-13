# Build stage
FROM golang:1.20-alpine AS builder

WORKDIR /app

# Install git (required for go-git)
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o developer-agent-consumer .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates git

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/developer-agent-consumer .

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /root

# Create temp directory for git operations
RUN mkdir -p /tmp && chown -R appuser:appuser /tmp

USER appuser

# Run the application
CMD ["./developer-agent-consumer"]

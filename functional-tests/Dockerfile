# Build stage
FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o functional-tests .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl netcat-openbsd

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/functional-tests .
COPY --from=builder /app/run-tests.sh .

RUN chmod +x run-tests.sh

# Run the tests
CMD ["./functional-tests"]

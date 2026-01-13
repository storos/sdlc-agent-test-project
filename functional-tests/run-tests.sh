#!/bin/bash

set -e

echo "=========================================="
echo "SDLC AI Agents - Functional Test Runner"
echo "=========================================="

# Check if services are running
echo ""
echo "Checking if services are running..."

if ! curl -sf http://localhost:8081/health > /dev/null 2>&1; then
    echo "✗ Configuration API is not running on port 8081"
    echo "  Please start services with: docker-compose up"
    exit 1
fi
echo "✓ Configuration API is running"

if ! curl -sf http://localhost:8080/health > /dev/null 2>&1; then
    echo "✗ JIRA Webhook API is not running on port 8080"
    echo "  Please start services with: docker-compose up"
    exit 1
fi
echo "✓ JIRA Webhook API is running"

# Check MongoDB
if ! nc -z localhost 27017 > /dev/null 2>&1; then
    echo "✗ MongoDB is not running on port 27017"
    echo "  Please start services with: docker-compose up"
    exit 1
fi
echo "✓ MongoDB is running"

# Set environment variables
export CONFIG_API_URL=${CONFIG_API_URL:-http://localhost:8081}
export WEBHOOK_API_URL=${WEBHOOK_API_URL:-http://localhost:8080}
export MONGO_URL=${MONGO_URL:-mongodb://localhost:27017}
export MONGO_DATABASE=${MONGO_DATABASE:-sdlc_agent}

echo ""
echo "Running functional tests..."
echo ""

# Run the tests
go run main.go

# Capture exit code
EXIT_CODE=$?

echo ""
if [ $EXIT_CODE -eq 0 ]; then
    echo "✓ All tests passed!"
else
    echo "✗ Tests failed with exit code: $EXIT_CODE"
fi

exit $EXIT_CODE

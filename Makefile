.PHONY: help build run test clean test-cleanup keys

# Default target
help:
	@echo "Available targets:"
	@echo "  build    - Build the API binary"
	@echo "  keys     - Generate JWT RS256 key pair (.keys/jwt_private.pem, .keys/jwt_public.pem)"
	@echo "  run      - Run the API locally (requires DynamoDB; run 'make keys' first)"
	@echo "  clean    - Clean build artifacts"
	@echo "  test     - Run all tests (starts DynamoDB Local + OpenSearch, creates table, runs tests)"
	@echo "  test-cleanup - Clean up test Docker containers"

# Build the API
build:
	@echo "Building API..."
	@mkdir -p bin
	go build -o bin/api ./cmd/api

# Generate JWT RS256 key pair for local dev (run once before 'make run')
keys:
	@mkdir -p .keys
	openssl genrsa -out .keys/jwt_private.pem 2048
	openssl rsa -in .keys/jwt_private.pem -pubout -out .keys/jwt_public.pem
	@echo "Keys written to .keys/jwt_private.pem and .keys/jwt_public.pem"

# Run the API locally (requires DynamoDB and keys; run 'make keys' first)
run:
	@echo "Running API locally..."
	@echo "Make sure DynamoDB is running with: docker compose up -d"
	@test -f .keys/jwt_private.pem || (echo "Run 'make keys' first"; exit 1)
	@test -f .keys/jwt_public.pem || (echo "Run 'make keys' first"; exit 1)
	DYNAMO_TABLE=afterwave AWS_REGION=us-east-1 DYNAMODB_ENDPOINT=http://localhost:8001 JWT_PRIVATE_KEY_PATH=.keys/jwt_private.pem JWT_PUBLIC_KEY_PATH=.keys/jwt_public.pem go run ./cmd/api

# Run tests (TestMain creates the DynamoDB table if missing; OpenSearch required for my-feed tests)
test:
	@echo "Setting up test environment..."
	@echo "Starting DynamoDB Local and OpenSearch..."
	@docker compose up -d --remove-orphans dynamodb-local opensearch
	@echo "Waiting for DynamoDB Local to be ready..."
	@sleep 5
	@echo "Waiting for OpenSearch to be ready..."
	@for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30; do \
		if curl -s -o /dev/null -w '%{http_code}' http://localhost:9200 | grep -q 200; then echo "OpenSearch ready."; break; fi; \
		if [ $$i -eq 30 ]; then echo "OpenSearch did not become ready in time."; docker compose down --remove-orphans; exit 1; fi; \
		sleep 3; \
	done
	@echo "Running tests..."
	@DYNAMO_TABLE=afterwave-test AWS_REGION=us-east-1 DYNAMODB_ENDPOINT=http://localhost:8001 \
		OPENSEARCH_ENDPOINT=http://localhost:9200 OPENSEARCH_FEED_INDEX=afterwave-feed \
		go test -v ./...; ret=$$?; \
	if [ $$ret -ne 0 ]; then echo "Tests failed, cleaning up Docker containers..."; else echo "Tests completed successfully, cleaning up..."; fi; \
	docker compose down --remove-orphans; \
	exit $$ret

# Clean up test Docker containers
test-cleanup:
	@echo "Cleaning up test Docker containers..."
	@docker compose down --remove-orphans

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

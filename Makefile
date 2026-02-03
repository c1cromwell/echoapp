.PHONY: help build run test clean install-deps lint fmt vet build-prod tls-cert

# Variables
BINARY_NAME=echoapp
GO=go
GOFLAGS=-v
PORT?=8000

help:
	@echo "EchoApp - REST API Framework"
	@echo ""
	@echo "Available commands:"
	@echo "  make build          Build executable"
	@echo "  make run            Run development server"
	@echo "  make test           Run all tests"
	@echo "  make test-endpoints Run endpoint tests"
	@echo "  make clean          Remove build artifacts"
	@echo "  make install-deps   Install/update dependencies"
	@echo "  make lint           Run linter"
	@echo "  make fmt            Format code"
	@echo "  make vet            Run vet"
	@echo "  make build-prod     Build production binary"
	@echo "  make tls-cert       Generate self-signed TLS certificate"
	@echo ""
	@echo "Environment variables:"
	@echo "  API_PORT=8080       Set API port (default: 8000)"
	@echo "  ENVIRONMENT=prod    Set environment (default: development)"
	@echo ""

build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) main.go
	@echo "✅ Build complete: $(BINARY_NAME)"
	@ls -lh $(BINARY_NAME)

run:
	@echo "Starting server on port $(PORT)..."
	API_PORT=$(PORT) $(GO) run main.go

run-prod:
	@echo "Starting production server with TLS..."
	@if [ ! -f "cert.pem" ] || [ ! -f "key.pem" ]; then \
		echo "⚠️  TLS certificates not found. Run 'make tls-cert' first."; \
		exit 1; \
	fi
	API_PORT=8443 TLS_ENABLED=true TLS_CERT_FILE=cert.pem TLS_KEY_FILE=key.pem $(GO) run main.go

test:
	@echo "Running tests..."
	$(GO) test -v ./...
	@echo "✅ Tests complete"

test-endpoints:
	@echo "Testing API endpoints..."
	@echo ""
	@echo "Starting server in background..."
	@API_PORT=9000 $(GO) run main.go &
	@SERVER_PID=$$!; \
	sleep 2; \
	echo "Running tests..."; \
	echo ""; \
	echo "1. Health check (no auth):"; \
	curl -s http://localhost:9000/health | jq . || echo "Failed"; \
	echo ""; \
	echo "2. Missing auth (should fail):"; \
	curl -s http://localhost:9000/v1/users | jq . || echo "Failed"; \
	echo ""; \
	echo "3. With auth token:"; \
	curl -s -H "Authorization: Bearer test-token" http://localhost:9000/v1/users | jq . || echo "Failed"; \
	echo ""; \
	echo "4. V2 API with pagination:"; \
	curl -s -H "Authorization: Bearer test-token" http://localhost:9000/v2/users | jq . || echo "Failed"; \
	echo ""; \
	kill $$SERVER_PID 2>/dev/null || true; \
	echo "✅ Tests complete"

clean:
	@echo "Cleaning build artifacts..."
	$(GO) clean
	rm -f $(BINARY_NAME)
	rm -f *.test
	@echo "✅ Clean complete"

install-deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy
	@echo "✅ Dependencies updated"
	@$(GO) mod graph | wc -l | xargs echo "Total dependencies:"

lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...
	@echo "✅ Lint complete"

fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "✅ Format complete"

vet:
	@echo "Running vet..."
	$(GO) vet ./...
	@echo "✅ Vet complete"

build-prod:
	@echo "Building optimized production binary..."
	@echo "Platform: $$(go env GOOS)-$$(go env GOARCH)"
	GOOS=$$(go env GOOS) GOARCH=$$(go env GOARCH) $(GO) build \
		-ldflags="-s -w -X main.Version=1.0.0" \
		-o $(BINARY_NAME)-prod main.go
	@echo "✅ Production build complete"
	@ls -lh $(BINARY_NAME)-prod
	@du -h $(BINARY_NAME)-prod

tls-cert:
	@echo "Generating self-signed TLS certificate..."
	@echo "Certificate: cert.pem"
	@echo "Key: key.pem"
	@echo "Validity: 365 days"
	openssl req -x509 -newkey rsa:4096 \
		-keyout key.pem -out cert.pem -days 365 -nodes \
		-subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"
	@echo "✅ Certificate generated"
	@echo ""
	@echo "To use in development:"
	@echo "  export TLS_ENABLED=true"
	@echo "  export TLS_CERT_FILE=cert.pem"
	@echo "  export TLS_KEY_FILE=key.pem"
	@echo "  make run-prod"

deps:
	@echo "Project dependencies:"
	@$(GO) list -m all

update-deps:
	@echo "Updating all dependencies..."
	$(GO) get -u ./...
	$(GO) mod tidy
	@echo "✅ Dependencies updated"

version:
	@echo "EchoApp v1.0.0"
	@echo "Go version: $$($(GO) version)"
	@echo "Binary: $(BINARY_NAME)"

info:
	@echo "Project Information:"
	@echo "  Name: EchoApp"
	@echo "  Type: REST API Framework"
	@echo "  Language: Go"
	@echo "  Binary: $(BINARY_NAME)"
	@echo "  Port: $(PORT)"
	@echo ""
	@echo "Features:"
	@echo "  ✓ Multi-version API (v1, v2)"
	@echo "  ✓ Authentication middleware"
	@echo "  ✓ CORS support"
	@echo "  ✓ TLS 1.3+"
	@echo "  ✓ Health check"
	@echo "  ✓ Request tracking"
	@echo "  ✓ Error handling"
	@echo ""

.DEFAULT_GOAL := help

.PHONY: help build run test clean install-deps lint fmt vet build-prod tls-cert migrate \
	dev dev-stop dev-stop-all dev-reset dev-status dev-logs dev-restart dev-comply dev-comply-stop \
	chain-up chain-down chain-reset chain-test contracts contracts-check validate-phase1 \
	testnet-up testnet-down release-check regression regression-quick regression-with-phase1 ios-preflight \
	phase3-signals-proof echooprf-ios screen-catalog knowledge-graph

# Variables
BINARY_NAME=echoapp
GO=go
GOFLAGS=-v
PORT?=8000

help:
	@echo "EchoApp - REST API Framework"
	@echo ""
	@echo "Local environment:"
	@echo "  make dev             Chain (4 Besu validators) + contract + backend stack"
	@echo "  make dev-status      Show status of the chain and backend"
	@echo "  make dev-logs        Tail backend stack logs"
	@echo "  make dev-restart     Restart the backend stack only (chain keeps running)"
	@echo "  make dev-stop        Stop the backend stack (chain keeps running)"
	@echo "  make dev-stop-all    Stop backend and chain (data kept)"
	@echo "  make dev-reset       Delete all local data: backend volumes, chain, keys, contract"
	@echo "  make chain-up        Chain only (make dev does this)"
	@echo "  make chain-test      Tests that need the live chain (anchor, prove, reject, fail over)"
	@echo "  make contracts       Recompile contracts into internal/besu/contract (needs node)"
	@echo "  make contracts-check Fail if the committed contract artefacts do not match the sources"
	@echo "  make dev-comply      Comply-only stack (Postgres :5433 + Comply API :8011)"
	@echo "  make dev-comply-stop Tear down Comply-only stack"
	@echo "  make validate-phase1 Run scripts/validate-phase1.sh go/no-go check"
	@echo "  make regression      Headless regression (Go + iOS SPM) — docs/E2E_LAUNCH_AND_TESTING.md"
	@echo "  make regression-quick Go race tests only"
	@echo "  make regression-with-phase1  regression + validate-phase1"
	@echo "  make ios-preflight   iOS E2E preflight (BUILD=1 TESTS=1) — docs/E2E_QUICK_START.md"
	@echo "  make screen-catalog  SwiftUI PNG gallery → docs/screen_catalog/index.html"
	@echo "  make knowledge-graph Factory 8090 WOs → requirements/code graph (docs/knowledge-graph)"
	@echo "  make testnet-up      Bring up backend stack only (assumes the chain is running)"
	@echo "  make testnet-down    Bring down backend stack only"
	@echo ""
	@echo "Application:"
	@echo "  make build           Build executable"
	@echo "  make run             Run development server"
	@echo "  make test            Run all tests"
	@echo "  make test-endpoints  Run endpoint tests"
	@echo "  make clean           Remove build artifacts"
	@echo "  make install-deps    Install/update dependencies"
	@echo "  make lint            Run linter"
	@echo "  make fmt             Format code"
	@echo "  make vet             Run vet"
	@echo "  make build-prod      Build production binary"
	@echo "  make tls-cert        Generate self-signed TLS certificate"
	@echo "  make migrate         Apply SQL migrations (needs DATABASE_HOST, etc.)"
	@echo ""
	@echo "Environment variables:"
	@echo "  API_PORT=8080        Set API port (default: 8000)"
	@echo "  ENVIRONMENT=prod     Set environment (default: development)"
	@echo ""

build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) main.go
	@echo "✅ Build complete: $(BINARY_NAME)"
	@ls -lh $(BINARY_NAME)

migrate: ## Apply pending SQL migrations to PostgreSQL
	$(GO) run ./cmd/migrate

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
	@cat VERSION 2>/dev/null | xargs -I{} echo "EchoApp v{}" || echo "EchoApp (no VERSION file)"
	@echo "Go version: $$($(GO) version)"
	@echo "Binary: $(BINARY_NAME)"

## Release readiness check — run before tagging a release.
release-check:
	@echo "=== Phase-1 Release Check ==="
	@echo ""
	@echo "--- 1. Go build ---"
	@$(GO) build ./... && echo "✅ go build clean" || (echo "❌ go build failed" && exit 1)
	@echo ""
	@echo "--- 2. Go tests (race) ---"
	@$(GO) test -race -count=1 ./internal/... ./pkg/... ./test/... && echo "✅ all tests pass" || (echo "❌ tests failed" && exit 1)
	@echo ""
	@echo "--- 3. go vet ---"
	@$(GO) vet ./... && echo "✅ vet clean" || (echo "❌ vet issues" && exit 1)
	@echo ""
	@echo "--- 4. gofmt ---"
	@if [ -n "$$(gofmt -l .)" ]; then echo "❌ unformatted files:"; gofmt -l .; exit 1; else echo "✅ gofmt clean"; fi
	@echo ""
	@echo "--- 5. No stray binaries ---"
	@if ls $(BINARY_NAME) credentials 2>/dev/null | grep -v "^ls:"; then \
		echo "❌ stale binaries present — run: rm -f $(BINARY_NAME) credentials"; exit 1; \
	else echo "✅ no stray binaries"; fi
	@echo ""
	@echo "--- 6. VERSION file ---"
	@cat VERSION && echo "✅ VERSION exists"
	@echo ""
	@echo "--- 7. CHANGELOG ---"
	@head -1 CHANGELOG.md && echo "✅ CHANGELOG exists"
	@echo ""
	@echo "=== ✅ Release check passed — ready to tag ==="
	@echo "Run: git tag v\$$(cat VERSION)-phase1 && git push origin v\$$(cat VERSION)-phase1"

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

# =============================================================================
# Local environment
#
#   1. Echo's chain: four Hyperledger Besu validators with QBFT consensus, and the
#      anchor contract (deploy/besu/local). Keys made here are for local use only.
#   2. Backend stack (Go API + Postgres + Redis + NATS + MinIO) via
#      docker-compose.testnet.yml, configured to anchor to that chain.
#
# After both are up, `make validate-phase1` runs the go/no-go script.
# =============================================================================

COMPOSE_TESTNET := docker compose -f docker-compose.testnet.yml
COMPOSE_COMPLY := docker compose -f docker-compose.comply.yml
CHAIN_DIR := deploy/besu/local
CHAIN_ENV := $(CHAIN_DIR)/.network/chain.env
# Runs a command with the gateway's signing key and, when the chain is up, its settings exported.
WITH_CHAIN := ECHOAPP_ROOT=$(CURDIR) bash -c 'source ./scripts/identity/ensure-identity-service-key.sh && { [ ! -f $(CHAIN_ENV) ] || source $(CHAIN_ENV); } && exec "$$@"' --

dev: chain-up testnet-up ## Bring up the chain and the backend stack
	@echo "Waiting for backend health..."
	@deadline=$$(( $$(date +%s) + 90 )); \
	while [ $$(date +%s) -lt $$deadline ]; do \
	  if curl -fsS --max-time 2 http://localhost:8000/health >/dev/null 2>&1; then \
	    echo "  ✓ backend healthy at http://localhost:8000"; break; \
	  fi; \
	  sleep 2; \
	done
	@echo ""
	@echo "  Backend:  http://localhost:8000"
	@echo "  Chain:    http://localhost:8545 (validators 1-4 on 8545-8548)"
	@echo ""
	@echo "Next: make validate-phase1"

chain-up: ## Start Echo's local chain (4 Besu validators) and deploy the anchor contract
	@echo "===== Chain ====="
	@$(CHAIN_DIR)/up.sh

chain-down: ## Stop the local chain (keeps its data, keys and contract)
	@docker compose -f $(CHAIN_DIR)/docker-compose.yml down 2>/dev/null || true

chain-reset: ## Delete the local chain (data, keys, contract) and start a new one
	@$(CHAIN_DIR)/up.sh --reset

chain-test: ## Tests that need the live chain: anchor, prove, verify offline, reject, fail over
	@[ -f $(CHAIN_ENV) ] || { echo "chain is not set up; run: make chain-up"; exit 1; }
	@$(WITH_CHAIN) $(GO) test ./internal/besu ./internal/metagraph -run Live -count=1 -v

contracts: ## Recompile contracts into internal/besu/contract (needs node; compiler pinned in chain/package.json)
	@cd chain && { [ -d node_modules/solc ] || npm install --silent; } && node compile.js

contracts-check: ## Fail if the committed contract artefacts do not match the sources
	@cd chain && { [ -d node_modules/solc ] || npm install --silent; } && node compile.js --check

testnet-up: ## Bring up backend stack only (assumes the chain is running)
	@echo "===== Backend stack ====="
	@$(WITH_CHAIN) $(COMPOSE_TESTNET) up -d --build

testnet-down: ## Tear down backend stack
	@$(COMPOSE_TESTNET) down

dev-stop: testnet-down ## Stop the backend stack (chain keeps running)
	@echo "Backend stack down. The chain is still running: make dev-stop-all stops it too."

dev-stop-all: testnet-down chain-down ## Stop backend and chain (data kept)
	@echo "Stopped. 'make dev' resumes with the same chain and the same Postgres."

dev-reset: ## Delete all local data: backend volumes, chain, keys and contract
	@echo "===== Clean slate: destroys local Postgres/Redis/NATS/MinIO and the local chain ====="
	@$(COMPOSE_TESTNET) down -v --remove-orphans 2>/dev/null || true
	@$(CHAIN_DIR)/up.sh --destroy
	@echo "Done. Next: make dev"

dev-comply: ## Comply-only stack (portal operators; no messaging gateway)
	@echo "Starting Comply-only stack (postgres :5433, comply :8011)..."
	@$(COMPOSE_COMPLY) up -d --build
	@echo "  Comply API: http://localhost:8011/health"
	@echo "  Portal: set COMPLY_API_BASE_URL=http://localhost:8011 in web/.env.local"

dev-comply-stop: ## Tear down Comply-only stack
	@$(COMPOSE_COMPLY) down

start-comply: ## Run ECHO Comply microservice on :8011 (WO-250/251/252)
	@echo "Starting Comply service on :8011..."
	COMPLY_PORT=8011 COMPLY_SERVICE_TOKEN=$${COMPLY_SERVICE_TOKEN:-dev-comply-token} $(GO) run ./cmd/comply

dev-restart: ## Restart backend stack
	@$(COMPOSE_TESTNET) restart

dev-logs: ## Tail backend stack logs
	@$(COMPOSE_TESTNET) logs -f --tail=100

dev-status: ## Show status of the chain and backend
	@echo "===== Backend stack ====="
	@$(COMPOSE_TESTNET) ps || true
	@echo ""
	@echo "===== Chain ====="
	@for n in 1 2 3 4; do \
	  port=$$((8544 + n)); \
	  h=$$(curl -s --max-time 2 -X POST -H 'Content-Type: application/json' \
	    --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://localhost:$$port 2>/dev/null \
	    | sed -n 's/.*"result":"\(0x[0-9a-f]*\)".*/\1/p'); \
	  if [ -n "$$h" ]; then printf "  ✓ validator %s  block %d  (localhost:%s)\n" $$n $$((h)) $$port; \
	  else printf "  ✗ validator %s  unreachable (localhost:%s)\n" $$n $$port; fi; \
	done
	@echo ""
	@if curl -fsS --max-time 2 http://localhost:8000/health >/dev/null 2>&1; then \
	  echo "  ✓ backend   http://localhost:8000/health"; else echo "  ✗ backend   unreachable"; fi

validate-phase1: ## Run the WO-230 6-step go/no-go validation script
	@chmod +x scripts/validate-phase1.sh
	@./scripts/validate-phase1.sh

regression: ## Headless regression — Go release-check + targeted suites + iOS SPM (docs/E2E_LAUNCH_AND_TESTING.md §4)
	@chmod +x scripts/run-regression.sh
	@./scripts/run-regression.sh

regression-quick: ## Go race tests only (no iOS, no validate-phase1)
	@chmod +x scripts/run-regression.sh
	@./scripts/run-regression.sh --quick

regression-with-phase1: ## regression + make validate-phase1 (needs Docker + Euclid)
	@chmod +x scripts/run-regression.sh
	@./scripts/run-regression.sh --with-phase1

BUILD ?=
TESTS ?=
ios-preflight: ## iOS E2E preflight — backend + Xcode checks (docs/E2E_QUICK_START.md)
	@chmod +x scripts/ios-e2e-preflight.sh
	@./scripts/ios-e2e-preflight.sh \
		$(if $(filter 1 true yes,$(BUILD)),--build,) \
		$(if $(filter 1 true yes,$(TESTS)),--tests,)

phase3-signals-proof: ## Headless WO-192/10 proof (Go WS tests + EchoPhase3Tests)
	@chmod +x scripts/phase3-signals-proof.sh
	@./scripts/phase3-signals-proof.sh

echooprf-ios: ## Build EchoOPRF.xcframework for WO-221 PSI (requires Xcode + gomobile)
	@chmod +x scripts/build-echooprf-ios.sh
	@./scripts/build-echooprf-ios.sh

screen-catalog: ## Export SwiftUI screen PNGs + tabbed index (docs/screen_catalog)
	@chmod +x scripts/screen-catalog/generate.sh scripts/screen-catalog/build-index.sh
	@./scripts/screen-catalog/generate.sh

knowledge-graph: ## Rebuild Factory 8090 → requirements → code graph
	@python3 scripts/knowledge-graph/build.py

echooprf-link: ## Embed EchoOPRF.xcframework in Xcode project (after echooprf-ios)
	@chmod +x scripts/link-echooprf-xcode.sh
	@./scripts/link-echooprf-xcode.sh

echooprf-test: ## Run OPRF interop tests (Go server + mobile/echooprf)
	@go test ./mobile/echooprf/... ./internal/api/... -run 'PSI|OPRF' -count=1

.DEFAULT_GOAL := help

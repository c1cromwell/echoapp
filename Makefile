.PHONY: help build run test clean install-deps lint fmt vet build-prod tls-cert migrate \
	dev dev-stop dev-status dev-logs dev-restart validate-phase1 metagraph-verify-skeleton metagraph-test \
	testnet-up testnet-down release-check regression regression-quick regression-with-phase1 ios-preflight \
	phase3-signals-proof

# Variables
BINARY_NAME=echoapp
GO=go
GOFLAGS=-v
PORT?=8000

help:
	@echo "EchoApp - REST API Framework"
	@echo ""
	@echo "Phase-1 testnet (WO-230):"
	@echo "  make dev             Bring up full Phase-1 cluster (metagraph + backend + iOS deps)"
	@echo "  make dev-status      Show status of all testnet components"
	@echo "  make dev-logs        Tail backend stack logs"
	@echo "  make dev-restart     Restart the backend stack only (keeps metagraph running)"
	@echo "  make dev-stop        Tear down backend stack (metagraph stays up — use 'hydra stop' for that)"
	@echo "  make validate-phase1 Run scripts/validate-phase1.sh go/no-go check"
	@echo "  make regression      Headless regression (Go + iOS SPM) — docs/E2E_LAUNCH_AND_TESTING.md"
	@echo "  make regression-quick Go race tests only"
	@echo "  make regression-with-phase1  regression + validate-phase1"
	@echo "  make ios-preflight   iOS E2E preflight (BUILD=1 TESTS=1) — docs/E2E_QUICK_START.md"
	@echo "  make metagraph-verify-skeleton  Static WO-276 checks (euclid + build.sbt + sources; needs jq)"
	@echo "  make metagraph-test     Scala tests: sharedData + identity metagraph layers (WO-272/277; sbt + JDK)"
	@echo "  make testnet-up      Bring up backend stack only (assumes metagraph already running)"
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
# Phase-1 Testnet (WO-230)
#
# Brings up the full local development environment in a single command:
#   1. Euclid SDK metagraph cluster (Global L0 + Metagraph L0 + Currency L1 +
#      Data L1) via `hydra` in the sibling ../euclid-development-environment
#      directory. Identity L0/L1 are NOT managed by hydra — start them
#      separately with `make start-identity` (see that target).
#   2. Backend stack (Go API + Postgres + Redis + NATS + MinIO) via
#      docker-compose.testnet.yml.
#
# After both are up, run `make validate-phase1` to execute the 6-step
# go/no-go script defined in WO-230.
# =============================================================================

EUCLID_DIR ?= $(abspath ../euclid-development-environment)
COMPOSE_TESTNET := docker compose -f docker-compose.testnet.yml
HYDRA_HEALTH_TIMEOUT := 300

dev: ## Bring up full Phase-1 cluster
	@echo "===== Phase-1 Testnet Bring-up ====="
	@echo ""
	@echo "[1/4] Ensuring Euclid SDK is set up..."
	@cd metagraph && ./scripts/setup-euclid.sh
	@echo ""
	@echo "[2/4] Building metagraph Docker images (hydra install + build)..."
	@if [ ! -d "$(EUCLID_DIR)" ]; then \
		echo "  ✗ Euclid directory not found: $(EUCLID_DIR)"; \
		echo "    setup-euclid.sh should have cloned it. Re-run 'cd metagraph && ./scripts/setup-euclid.sh'"; \
		exit 1; \
	fi
	@cd "$(EUCLID_DIR)" && scripts/hydra install
	@cd "$(EUCLID_DIR)" && scripts/hydra build
	@echo ""
	@echo "[3/5] Starting metagraph cluster (hydra start-genesis)..."
	@cd "$(EUCLID_DIR)" && scripts/hydra start-genesis || { \
		echo "  ⚠ hydra start-genesis returned non-zero (cluster may already be running)"; \
	}
	@echo ""
	@echo "[4/5] Waiting for core metagraph endpoints to come up..."
	@# Only the core hydra-managed layers are waited on here. Identity L0/L1 are
	@# started separately (`make start-identity`), so waiting for them would just
	@# burn the timeout while they're down — they are intentionally not listed.
	@deadline=$$(( $$(date +%s) + $(HYDRA_HEALTH_TIMEOUT) )); \
	for endpoint in \
	  "Global L0=http://localhost:9000/node/info" \
	  "Metagraph L0=http://localhost:9200/node/info" \
	  "Currency L1=http://localhost:9300/node/info" \
	  "Data L1=http://localhost:9400/node/info"; do \
	  label=$${endpoint%%=*}; url=$${endpoint#*=}; \
	  printf "  waiting for %-14s ... " "$$label"; \
	  while [ $$(date +%s) -lt $$deadline ]; do \
	    if curl -fsS --max-time 2 "$$url" >/dev/null 2>&1; then \
	      echo "✓"; break; \
	    fi; \
	    sleep 2; \
	  done; \
	  if ! curl -fsS --max-time 2 "$$url" >/dev/null 2>&1; then \
	    echo "✗ TIMEOUT"; \
	    echo "    Check 'cd $(EUCLID_DIR) && scripts/hydra status'"; \
	    exit 1; \
	  fi; \
	done
	@echo ""
	@echo "[5/5] Starting backend stack (Postgres + Redis + NATS + MinIO + echoapp)..."
	@$(COMPOSE_TESTNET) up -d --build
	@echo ""
	@echo "Waiting for backend health..."
	@deadline=$$(( $$(date +%s) + 60 )); \
	while [ $$(date +%s) -lt $$deadline ]; do \
	  if curl -fsS --max-time 2 http://localhost:8000/health >/dev/null 2>&1; then \
	    echo "  ✓ backend healthy at http://localhost:8000"; break; \
	  fi; \
	  sleep 2; \
	done
	@echo ""
	@echo "===== Phase-1 cluster up ====="
	@echo "  Backend:        http://localhost:8000"
	@echo "  Global L0:      http://localhost:9000"
	@echo "  Metagraph L0:   http://localhost:9200"
	@echo "  Currency L1:    http://localhost:9300"
	@echo "  Data L1:        http://localhost:9400"
	@echo "  Identity L0/L1: 9600 / 9500  (optional, not started by 'make dev' —"
	@echo "                  run 'make start-identity' for VC / trust-tier features)"
	@echo ""
	@echo "Next: make validate-phase1"

testnet-up: ## Bring up backend stack only (assumes metagraph already running)
	@$(COMPOSE_TESTNET) up -d --build

testnet-down: ## Tear down backend stack
	@$(COMPOSE_TESTNET) down

metagraph-verify-skeleton: ## WO-276: static verify Identity Metagraph skeleton (no sbt; needs jq)
	@cd metagraph && ./scripts/verify-identity-skeleton.sh

metagraph-test: ## WO-272/277: run validators + Identity L1 wiring tests (sbt + JDK 21)
	@JDK21=$$(ls -d /opt/homebrew/opt/openjdk@21/libexec/openjdk.jdk/Contents/Home 2>/dev/null || ls -d /usr/lib/jvm/java-21* 2>/dev/null | head -1 || echo ""); \
	if [ -n "$$JDK21" ]; then \
	  cd metagraph && JAVA_HOME=$$JDK21 PATH=$$JDK21/bin:$$PATH SBT_OPTS="-Xss8m -Xmx2g" sbt "sharedData/test" "identityL0/test" "identityL1/test"; \
	else \
	  cd metagraph && SBT_OPTS="-Xss8m -Xmx2g" sbt "sharedData/test" "identityL0/test" "identityL1/test"; \
	fi

dev-stop: testnet-down ## Tear down backend stack (does not stop metagraph)
	@echo "Backend stack down. Metagraph still running — use:"
	@echo "  cd $(EUCLID_DIR) && scripts/hydra stop"

# Identity nodes (L0 port 9600, L1 port 9500) are custom Echo modules not
# managed by Euclid hydra. They require sbt assembly JARs and their own
# Docker setup. Use these targets once the core cluster is running.
start-identity: ## Start Identity L0 + L1 nodes (requires sbt assembly + running cluster)
	@echo "Starting Identity nodes..."
	@JAR_L0=$$(ls metagraph/modules/identity_l0/target/scala-2.13/*assembly*.jar 2>/dev/null | head -1); \
	JAR_L1=$$(ls metagraph/modules/identity_l1/target/scala-2.13/*assembly*.jar 2>/dev/null | head -1); \
	if [ -z "$$JAR_L0" ] || [ -z "$$JAR_L1" ]; then \
	  echo "  ✗ Identity JARs not found. Run: cd metagraph && sbt identityL0/assembly identityL1/assembly"; \
	  exit 1; \
	fi; \
	echo "  ✓ Found $$JAR_L0"; \
	echo "  ✓ Found $$JAR_L1"; \
	if ! docker image inspect metagraph-base-image:latest >/dev/null 2>&1; then \
	  echo "  ✗ metagraph-base-image:latest not found — start the core cluster first: make dev"; \
	  exit 1; \
	fi; \
	echo "  ✓ Base image present"; \
	if ! curl -fsS --max-time 3 http://localhost:9000/node/info >/dev/null 2>&1; then \
	  echo "  ✗ Global L0 not reachable on :9000 — Identity L0 peers with it. Run 'make dev' first."; \
	  exit 1; \
	fi; \
	echo "  ✓ Global L0 reachable"
	docker compose -f docker-compose.identity.yml up -d
	@echo "  Waiting for Identity L0 on :9600 (genesis can take ~30s)..."
	@for i in $$(seq 1 40); do \
	  if curl -fsS --max-time 2 http://localhost:9600/node/info >/dev/null 2>&1; then \
	    echo "  ✓ Identity L0 ready"; break; \
	  fi; sleep 3; done
	@echo "  Waiting for Identity L1 on :9500..."
	@for i in $$(seq 1 40); do \
	  if curl -fsS --max-time 2 http://localhost:9500/node/info >/dev/null 2>&1; then \
	    echo "  ✓ Identity L1 ready"; break; \
	  fi; sleep 3; done

stop-identity: ## Stop Identity L0 + L1 nodes
	@docker compose -f docker-compose.identity.yml down 2>/dev/null || true

dev-restart: ## Restart backend stack
	@$(COMPOSE_TESTNET) restart

dev-logs: ## Tail backend stack logs
	@$(COMPOSE_TESTNET) logs -f --tail=100

dev-status: ## Show status of all testnet components
	@echo "===== Backend stack ====="
	@$(COMPOSE_TESTNET) ps || true
	@echo ""
	@echo "===== Metagraph cluster ====="
	@if [ -d "$(EUCLID_DIR)" ]; then \
	  cd "$(EUCLID_DIR)" && scripts/hydra status || true; \
	else \
	  echo "  Euclid directory not found: $(EUCLID_DIR)"; \
	fi
	@echo ""
	@echo "===== Endpoint reachability ====="
	@for ep in \
	  "Backend=http://localhost:8000/health" \
	  "Global L0=http://localhost:9000/node/info" \
	  "Metagraph L0=http://localhost:9200/node/info" \
	  "Currency L1=http://localhost:9300/node/info" \
	  "Data L1=http://localhost:9400/node/info" \
	  "Identity L0=http://localhost:9600/node/info" \
	  "Identity L1=http://localhost:9500/node/info"; do \
	  label=$${ep%%=*}; url=$${ep#*=}; \
	  if curl -fsS --max-time 2 "$$url" >/dev/null 2>&1; then \
	    printf "  ✓ %-15s %s\n" "$$label" "$$url"; \
	  else \
	    printf "  ✗ %-15s %s (unreachable)\n" "$$label" "$$url"; \
	  fi; \
	done

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

.DEFAULT_GOAL := help

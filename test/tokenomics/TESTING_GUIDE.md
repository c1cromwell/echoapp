# ECHO Tokenomics Testing and Deployment Guide

## Overview

This guide provides instructions for testing the ECHO tokenomics implementation locally and deploying to Constellation TestNet.

## Table of Contents

1. [Local Testing Setup](#local-testing-setup)
2. [Running Unit Tests](#running-unit-tests)
3. [Constellation TestNet Setup](#constellation-testnet-setup)
4. [Integration Testing](#integration-testing)
5. [Performance Benchmarking](#performance-benchmarking)
6. [Docker Setup](#docker-setup)

## Local Testing Setup

### Prerequisites

- Go 1.20+
- Git
- Docker (for containerized testing)
- Constellation CLI (for TestNet deployment)

### Installation

1. **Clone the repository and navigate to echoapp:**

```bash
cd /Users/thechadcromwell/Projects/echoapp
```

2. **Install dependencies:**

```bash
go mod download
go mod tidy
```

3. **Verify structure:**

```bash
tree -L 2 internal/tokenomics/
```

Expected output:
```
internal/tokenomics/
├── emissions/
│   └── schedule.go
├── governance/
│   └── governance.go
├── models/
│   ├── rewards.go
│   ├── token.go
│   └── vesting.go (optional)
├── protection/
│   └── sybil.go
├── rewards/
│   └── distributor.go
└── staking/
    └── staking.go
```

## Running Unit Tests

### Basic Test Execution

```bash
# Run all tokenomics tests
go test ./test/tokenomics -v

# Run specific test
go test ./test/tokenomics -run TestTokenConfiguration -v

# Run with coverage
go test ./test/tokenomics -v -cover

# Generate coverage report
go test ./test/tokenomics -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Descriptions

#### Token Configuration Tests
- **TestTokenConfiguration**: Verifies ECHO token specs (1B supply, 8 decimals, hard cap)
- **TestAllocationBreakdown**: Ensures token distribution sums to 100%

#### Emission Schedule Tests
- **TestEmissionScheduleHalving**: Validates Bitcoin-like halving every 2 years
- **TestValidatorEmissionPhases**: Verifies 4-phase validator emission schedule
- **TestInflationRate**: Calculates inflation rates by year

#### Reward System Tests
- **TestMessagingRewardCalculation**: Validates trust-based reward multipliers
- **TestRewardDistribution**: Verifies daily caps and anti-gaming measures
- **TestReferralProgram**: Tests referral milestone calculations

#### Staking Tests
- **TestStakingTiers**: Validates 5 staking tiers with correct APY
- **TestStakingRewardCalculation**: Verifies compound interest calculations
- **TestValidatorEconomics**: Tests validator reward formulas

#### Governance Tests
- **TestGovernanceProposal**: Tests proposal creation and voting
- **TestGovernanceStats**: Validates governance metrics

#### Security Tests
- **TestSybilProtection**: Validates anti-Sybil checks
- **TestAntiGamingProtection**: Tests progressive decay and caps

### Expected Test Results

```bash
$ go test ./test/tokenomics -v

=== RUN   TestTokenConfiguration
--- PASS: TestTokenConfiguration (0.00s)
=== RUN   TestAllocationBreakdown
--- PASS: TestAllocationBreakdown (0.00s)
=== RUN   TestEmissionScheduleHalving
--- PASS: TestEmissionScheduleHalving (0.01s)
=== RUN   TestMessagingRewardCalculation
--- PASS: TestMessagingRewardCalculation (0.00s)
...
PASS
ok      ./test/tokenomics   0.45s
```

## Performance Benchmarking

### Run Benchmarks

```bash
# Run all benchmarks
go test ./test/tokenomics -bench=. -benchmem -benchtime=10s

# Run specific benchmark
go test ./test/tokenomics -bench=BenchmarkMessageRewardCalculation -benchmem

# Save results for comparison
go test ./test/tokenomics -bench=. -benchmem > benchmark_results.txt
```

### Benchmark Tests Included

1. **BenchmarkMessageRewardCalculation**: ~50-100 ns/op
2. **BenchmarkSybilCheck**: ~1-2 µs/op
3. **BenchmarkStakingRewardCalc**: ~200-400 ns/op
4. **BenchmarkGovernanceProposal**: ~2-5 µs/op

Expected performance (on modern hardware):
```
BenchmarkMessageRewardCalculation  100000    10234 ns/op    512 B/op    8 allocs/op
BenchmarkSybilCheck                 10000   152340 ns/op   8240 B/op   25 allocs/op
BenchmarkStakingRewardCalc         500000     2341 ns/op    144 B/op    4 allocs/op
```

## Constellation TestNet Setup

### 1. Install Constellation CLI

```bash
# macOS
brew install tessellation-constellation

# Verify installation
constellation version
```

### 2. Create TestNet Account

```bash
# Generate a new key pair
constellation key generate \
  --keystore-path ~/.constellation/keystore \
  --alias test-account

# Export public key
constellation key export \
  --keystore-path ~/.constellation/keystore \
  --alias test-account \
  --public-key-path ./test-account.pub

# Import test account for signing
constellation key import \
  --keystore-path ~/.constellation/keystore \
  --alias testnet-key \
  --private-key-path ./test-account.key
```

### 3. Connect to TestNet

```bash
# Set TestNet endpoint
export DAG_TEST_ENDPOINT="https://testnet-be1.constellationnetwork.io:9000"

# Check connection
curl -X GET "${DAG_TEST_ENDPOINT}/health"

# Alternative: Use Docker
docker run --rm \
  -e DAG_HOST=testnet-be1.constellationnetwork.io \
  -e DAG_PORT=9000 \
  tessellation-cli:latest \
  dag version
```

### 4. Fund TestNet Account

Visit [Constellation TestNet Faucet](https://testnet-faucet.constellationnetwork.io) and:
1. Paste your public key
2. Request testnet DAG tokens
3. Wait for confirmation

### 5. Deploy Metagraph

```bash
# Build metagraph code
go build -o ./bin/echo-metagraph ./cmd/metagraph

# Deploy to TestNet
constellation metagraph deploy \
  --metagraph-id echo-token-testnet \
  --version 1.0.0 \
  --keystore-path ~/.constellation/keystore \
  --alias test-account \
  --endpoint $DAG_TEST_ENDPOINT \
  --binary ./bin/echo-metagraph

# Verify deployment
constellation metagraph status \
  --metagraph-id echo-token-testnet \
  --endpoint $DAG_TEST_ENDPOINT
```

## Docker Setup

### Dockerfile for Local Testing

Create `Dockerfile.test`:

```dockerfile
FROM golang:1.20-alpine

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy source
COPY . .

# Download dependencies
RUN go mod download

# Run tests
CMD ["go", "test", "./test/tokenomics", "-v", "-cover"]
```

### Docker Compose for Multi-Service Testing

Create `docker-compose.test.yml`:

```yaml
version: '3.8'

services:
  test-runner:
    build:
      context: .
      dockerfile: Dockerfile.test
    container_name: echo-tokenomics-tests
    environment:
      - ENVIRONMENT=test
      - LOG_LEVEL=debug
    volumes:
      - ./test:/app/test
      - ./internal:/app/internal
    command: go test ./test/tokenomics -v -coverprofile=coverage.out

  constellation-testnet:
    image: tessellation:latest
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      - DAG_MODE=testnet
      - DAG_SEED_PEERS=testnet-be1.constellationnetwork.io:9000
    volumes:
      - constellation-data:/data
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/health"]
      interval: 10s
      timeout: 5s
      retries: 5

  metagraph-validator:
    build:
      context: .
      dockerfile: Dockerfile.metagraph
    depends_on:
      constellation-testnet:
        condition: service_healthy
    environment:
      - DAG_ENDPOINT=http://constellation-testnet:9000
      - METAGRAPH_ID=echo-token-testnet
    ports:
      - "9002:9002"
      - "9003:9003"

volumes:
  constellation-data:
```

### Run Tests in Docker

```bash
# Build and run tests
docker-compose -f docker-compose.test.yml up --build

# View logs
docker-compose -f docker-compose.test.yml logs -f test-runner

# Stop services
docker-compose -f docker-compose.test.yml down

# Clean up
docker-compose -f docker-compose.test.yml down -v
```

## Integration Testing

### 1. Create Integration Test Suite

```go
// test/integration/constellation_test.go
package integration

import (
	"testing"
	"time"

	"github.com/tessellation-token/tessellation-go/client"
	"../internal/tokenomics/governance"
	"../internal/tokenomics/models"
	"../internal/tokenomics/staking"
)

func TestConstellationMeragraphIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Connect to TestNet
	c := client.NewClient("https://testnet-be1.constellationnetwork.io:9000")

	// Test metagraph connection
	if err := c.Health(); err != nil {
		t.Fatalf("Failed to connect to metagraph: %v", err)
	}

	// Deploy token contract
	// ...

	// Verify genesis state
	// ...
}

func TestTokenTransfer(t *testing.T) {
	// Test token transfers on metagraph
	// Verify correct amounts, gas usage, etc.
}

func TestRewardDistribution(t *testing.T) {
	// Test actual reward distribution on TestNet
	// Verify emission schedule accuracy
}
```

### 2. Run Integration Tests

```bash
# Include integration tests (requires TestNet access)
go test ./test/integration -v -timeout=30m

# Run specific integration test
go test ./test/integration -run TestConstellationMeragraphIntegration -v

# Skip integration tests (for CI/CD)
go test ./... -short
```

## Continuous Integration Setup

### GitHub Actions Workflow

Create `.github/workflows/tokenomics-test.yml`:

```yaml
name: Tokenomics Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.20'
    
    - name: Run Unit Tests
      run: |
        cd echoapp
        go test ./test/tokenomics -v -cover -coverprofile=coverage.out
    
    - name: Upload Coverage
      uses: codecov/codecov-action@v3
      with:
        files: ./coverage.out
    
    - name: Run Benchmarks
      run: |
        cd echoapp
        go test ./test/tokenomics -bench=. -benchmem > benchmark-results.txt
    
    - name: Comment PR with Results
      uses: actions/github-script@v6
      if: github.event_name == 'pull_request'
      with:
        script: |
          const fs = require('fs');
          const benchmark = fs.readFileSync('benchmark-results.txt', 'utf8');
          github.rest.issues.createComment({
            issue_number: context.issue.number,
            owner: context.repo.owner,
            repo: context.repo.repo,
            body: '## Benchmark Results\n```\n' + benchmark + '\n```'
          });

  lint:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: golangci/golangci-lint-action@v3
      with:
        version: latest
```

## Troubleshooting

### Common Issues

**1. Import Errors**
```bash
# Solution: Update module paths
go mod tidy
go mod vendor
```

**2. TestNet Connection Timeout**
```bash
# Check network connectivity
curl -v https://testnet-be1.constellationnetwork.io:9000/health

# Use alternative endpoint if available
export DAG_TEST_ENDPOINT="https://testnet-be2.constellationnetwork.io:9000"
```

**3. Insufficient TestNet Balance**
```bash
# Request more DAG tokens from faucet
# or check account balance
constellation account balance \
  --keystore-path ~/.constellation/keystore \
  --alias test-account \
  --endpoint $DAG_TEST_ENDPOINT
```

**4. Memory Issues in Tests**
```bash
# Increase available memory
GODEBUG=gctrace=1 go test ./test/tokenomics -v

# Run tests with memory limits
go test ./test/tokenomics -v -run TestStakingRewardCalculation -memprofile=mem.prof
```

## Performance Optimization

### Profiling

```bash
# CPU profile
go test ./test/tokenomics -cpuprofile=cpu.prof -memprofile=mem.prof
go tool pprof cpu.prof

# Memory profile
go test ./test/tokenomics -memprofile=mem.prof
go tool pprof mem.prof

# Generate HTML report
go tool pprof -http=:8080 cpu.prof
```

### Optimization Recommendations

1. **Emission Calculation**: Cache halving periods instead of recalculating
2. **Reward Distribution**: Use batch processing for multiple users
3. **Sybil Detection**: Implement caching for graph analysis
4. **Governance**: Use event indexing instead of linear scans

## Next Steps

1. **Deploy to Constellation TestNet**: Follow the TestNet setup guide
2. **Monitor Performance**: Set up Prometheus metrics and Grafana dashboards
3. **Load Testing**: Use k6 or Locust to test under high transaction load
4. **Security Audit**: Conduct thorough security review before mainnet
5. **Community Testing**: Create public TestNet for community participation

## Support

For issues or questions:
- Check [Constellation Documentation](https://docs.constellationnetwork.io)
- Review [ECHO Tokenomics Blueprint](../echo-tokenomics-blueprint-v22.md)
- Open a GitHub issue with detailed logs


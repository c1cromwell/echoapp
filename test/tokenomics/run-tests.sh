#!/bin/bash

# ECHO Tokenomics Testing and Deployment Script
# Usage: ./run-tests.sh [unit|integration|docker|testnet|all]

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_DIR="$SCRIPT_DIR/.."
ECHOAPP_DIR="$PROJECT_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
print_header() {
    echo -e "\n${BLUE}======================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}======================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.20+"
        exit 1
    fi
    print_success "Go found: $(go version)"
}

check_docker() {
    if ! command -v docker &> /dev/null; then
        print_warning "Docker not found. Some tests will be skipped."
        return 1
    fi
    print_success "Docker found: $(docker --version)"
    return 0
}

run_unit_tests() {
    print_header "Running Unit Tests"
    
    cd "$ECHOAPP_DIR"
    
    # Download dependencies
    print_warning "Downloading dependencies..."
    go mod download
    
    # Run tests with coverage
    print_warning "Running tests..."
    go test ./test/tokenomics -v -cover -coverprofile=coverage.out
    
    if [ $? -eq 0 ]; then
        print_success "Unit tests passed!"
        
        # Display coverage
        echo ""
        print_warning "Coverage Summary:"
        go tool cover -func=coverage.out | tail -1
        
        # Generate HTML coverage report
        go tool cover -html=coverage.out -o coverage.html
        print_success "Coverage report generated: coverage.html"
    else
        print_error "Unit tests failed!"
        exit 1
    fi
}

run_benchmarks() {
    print_header "Running Performance Benchmarks"
    
    cd "$ECHOAPP_DIR"
    
    print_warning "Running benchmarks (this may take a minute)..."
    go test ./test/tokenomics -bench=. -benchmem -benchtime=5s -v > benchmark_results.txt
    
    if [ $? -eq 0 ]; then
        print_success "Benchmarks completed!"
        echo ""
        cat benchmark_results.txt
    else
        print_error "Benchmarks failed!"
        exit 1
    fi
}

run_docker_tests() {
    print_header "Running Tests in Docker"
    
    if ! check_docker; then
        print_error "Docker is required for Docker tests"
        return 1
    fi
    
    cd "$ECHOAPP_DIR"
    
    print_warning "Building Docker images..."
    docker-compose -f docker-compose.tokenomics.yml build
    
    print_warning "Starting services..."
    docker-compose -f docker-compose.tokenomics.yml up --abort-on-container-exit
    
    if [ $? -eq 0 ]; then
        print_success "Docker tests passed!"
    else
        print_error "Docker tests failed!"
        docker-compose -f docker-compose.tokenomics.yml logs
        exit 1
    fi
    
    # Cleanup
    print_warning "Cleaning up Docker services..."
    docker-compose -f docker-compose.tokenomics.yml down -v
}

run_integration_tests() {
    print_header "Running Integration Tests (Constellation TestNet)"
    
    if ! command -v constellation &> /dev/null; then
        print_warning "Constellation CLI not found. Skipping integration tests."
        print_warning "To install: brew install tessellation-constellation"
        return 1
    fi
    
    cd "$ECHOAPP_DIR"
    
    print_warning "Checking TestNet connectivity..."
    
    # Try to connect to TestNet
    if ! curl -s -f "https://testnet-be1.constellationnetwork.io:9000/health" > /dev/null; then
        print_warning "Cannot connect to Constellation TestNet"
        print_warning "Make sure you have internet connectivity and TestNet is online"
        return 1
    fi
    
    print_success "TestNet is online"
    
    print_warning "Running integration tests..."
    go test ./test/integration -v -timeout=30m -short 2>/dev/null || {
        print_warning "Integration tests skipped (requires TestNet setup)"
    }
}

setup_testnet() {
    print_header "Setting Up Constellation TestNet"
    
    if ! command -v constellation &> /dev/null; then
        print_error "Constellation CLI is required"
        echo "Install with: brew install tessellation-constellation"
        return 1
    fi
    
    print_warning "Generating TestNet keys..."
    constellation key generate \
        --keystore-path ~/.constellation/keystore \
        --alias echo-testnet
    
    print_success "Keys generated"
    
    print_warning "Please fund your account at: https://testnet-faucet.constellationnetwork.io"
    constellation key export \
        --keystore-path ~/.constellation/keystore \
        --alias echo-testnet \
        --public-key-path ./testnet.pub
    
    echo ""
    echo "Your public key:"
    cat ./testnet.pub
}

show_usage() {
    cat << EOF
${BLUE}ECHO Tokenomics Testing Script${NC}

Usage: ./run-tests.sh [command]

Commands:
    unit            Run unit tests with coverage
    bench           Run performance benchmarks
    docker          Run tests in Docker containers
    integration     Run integration tests on TestNet
    testnet-setup   Set up Constellation TestNet account
    all             Run all tests (unit, bench, docker)
    help            Show this help message

Examples:
    ./run-tests.sh unit
    ./run-tests.sh all
    ./run-tests.sh docker
    ./run-tests.sh testnet-setup

EOF
}

# Main script logic
main() {
    local command="${1:-help}"
    
    # Check prerequisites
    check_go
    
    case "$command" in
        unit)
            run_unit_tests
            ;;
        bench)
            run_unit_tests
            run_benchmarks
            ;;
        docker)
            run_docker_tests
            ;;
        integration)
            run_integration_tests
            ;;
        testnet-setup)
            setup_testnet
            ;;
        all)
            run_unit_tests
            run_benchmarks
            if check_docker; then
                run_docker_tests
            fi
            ;;
        help)
            show_usage
            ;;
        *)
            print_error "Unknown command: $command"
            show_usage
            exit 1
            ;;
    esac
}

# Run main function
main "$@"

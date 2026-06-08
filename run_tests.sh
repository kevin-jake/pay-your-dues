#!/bin/bash

# Pay Your Dues Test Runner Script
# This script runs comprehensive tests for the pay-your-dues application

set -e  # Exit on any error

echo "🚀 Starting Pay Your Dues Test Suite"
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
COVERAGE_THRESHOLD=80
COVERAGE_FILE="coverage.out"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to run a specific test category
run_test_category() {
    local category=$1
    local path=$2
    local description=$3
    
    echo ""
    print_status "Running $description..."
    echo "----------------------------------------"
    
    if go test -v "$path"; then
        print_success "$description completed successfully"
    else
        print_error "$description failed"
        return 1
    fi
}

# Function to run tests with coverage
run_tests_with_coverage() {
    print_status "Running all tests with coverage..."
    
    # Run all tests and generate coverage
    if go test -v -coverprofile="$COVERAGE_FILE" ./...; then
        print_success "All tests passed!"
        
        # Generate coverage report
        echo ""
        print_status "Generating coverage report..."
        go tool cover -func="$COVERAGE_FILE"
        
        # Check coverage threshold
        local coverage=$(go tool cover -func="$COVERAGE_FILE" | grep total | awk '{print $3}' | sed 's/%//')
        
        echo ""
        if (( $(echo "$coverage >= $COVERAGE_THRESHOLD" | bc -l) )); then
            print_success "Coverage: ${coverage}% (meets threshold of ${COVERAGE_THRESHOLD}%)"
        else
            print_warning "Coverage: ${coverage}% (below threshold of ${COVERAGE_THRESHOLD}%)"
        fi
        
        # Generate HTML coverage report
        print_status "Generating HTML coverage report..."
        go tool cover -html="$COVERAGE_FILE" -o coverage.html
        print_success "HTML coverage report generated: coverage.html"
        
    else
        print_error "Some tests failed"
        return 1
    fi
}

# Function to run specific test types
run_unit_tests() {
    run_test_category "Unit Tests" "./tests/unit" "Unit tests for services and business logic"
}

run_integration_tests() {
    run_test_category "Integration Tests" "./tests/integration" "Integration tests for complete workflows"
}

# Function to run performance tests
run_performance_tests() {
    echo ""
    print_status "Running performance benchmarks..."
    echo "----------------------------------------"
    
    # Run benchmarks
    if go test -bench=. -benchmem ./...; then
        print_success "Performance benchmarks completed"
    else
        print_warning "Some benchmarks failed or were not found"
    fi
}

# Function to run race condition tests
run_race_tests() {
    echo ""
    print_status "Running race condition tests..."
    echo "----------------------------------------"
    
    if go test -race ./...; then
        print_success "Race condition tests passed"
    else
        print_error "Race condition detected!"
        return 1
    fi
}

# Function to validate test structure
validate_test_structure() {
    print_status "Validating test structure..."
    
    # Check if test directories exist
    local test_dirs=("tests/unit" "tests/integration")
    for dir in "${test_dirs[@]}"; do
        if [ -d "$dir" ]; then
            print_success "✓ $dir exists"
        else
            print_error "✗ $dir missing"
            return 1
        fi
    done
    
    # Count test files
    local test_count=$(find tests -name "*_test.go" | wc -l)
    print_success "Found $test_count test files"
}

# Function to clean up
cleanup() {
    print_status "Cleaning up temporary files..."
    rm -f "$COVERAGE_FILE"
    print_success "Cleanup completed"
}

# Main execution
main() {
    # Parse command line arguments
    case "$1" in
        "unit")
            validate_test_structure
            run_unit_tests
            ;;
        "integration")
            validate_test_structure
            run_integration_tests
            ;;
        "performance")
            run_performance_tests
            ;;
        "race")
            run_race_tests
            ;;
        "coverage")
            validate_test_structure
            run_tests_with_coverage
            ;;
        "all"|"")
            validate_test_structure
            run_unit_tests
            run_integration_tests
            run_performance_tests
            run_race_tests
            run_tests_with_coverage
            ;;
        "help"|"-h"|"--help")
            echo "Pay Your Dues Test Runner"
            echo ""
            echo "Usage: $0 [test_type]"
            echo ""
            echo "Test types:"
            echo "  unit         - Run unit tests only"
            echo "  integration  - Run integration tests only"
            echo "  performance  - Run performance benchmarks"
            echo "  race         - Run race condition tests"
            echo "  coverage     - Run all tests with coverage report"
            echo "  all          - Run all tests (default)"
            echo "  help         - Show this help message"
            echo ""
            exit 0
            ;;
        *)
            print_error "Unknown test type: $1"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
    
    echo ""
    print_success "🎉 Test execution completed!"
    
    # Cleanup if not coverage run (we want to keep coverage files)
    if [ "$1" != "coverage" ] && [ "$1" != "all" ] && [ "$1" != "" ]; then
        cleanup
    fi
}

# Trap to ensure cleanup on exit
trap cleanup EXIT

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    print_error "go.mod not found. Please run this script from the project root directory."
    exit 1
fi

# Check if bc is available for coverage calculation
if ! command -v bc &> /dev/null; then
    print_warning "bc not found. Coverage threshold checking will be disabled."
fi

# Run main function with all arguments
main "$@"

#!/bin/bash
# Test runner script for Assignment Pull Request Creator

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
print_header() {
    echo -e "${BLUE}$1${NC}"
    echo "=================================="
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Help function
show_help() {
    echo "Assignment Pull Request Creator - Test Runner"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  help           - Show this help message"
    echo "  all            - Run all tests"
    echo "  unit           - Run unit tests only"
    echo "  local          - Run local integration tests"
    echo "  discovery      - Test assignment discovery only"
    echo "  sanitize       - Test branch name sanitization"
    echo "  quality        - Run code quality checks"
    echo "  performance    - Run performance tests"
    echo "  clean          - Clean test artifacts"
    echo ""
    echo "Examples:"
    echo "  $0 discovery                           # Quick discovery test"
    echo "  ASSIGNMENT_REGEX='^hw-\d+\$' $0 discovery  # Custom pattern"
}

# Install dependencies
install_deps() {
    print_header "Installing test dependencies"
    pip install pytest black flake8 mypy bandit safety
    print_success "Dependencies installed"
}

# Run unit tests
run_unit_tests() {
    print_header "Running unit tests"
    python -m pytest test_assignment_creator.py -v
    print_success "Unit tests completed"
}

# Run local integration tests
run_local_tests() {
    print_header "Running local integration tests"
    python test_local.py
    print_success "Local integration tests completed"
}

# Test assignment discovery
test_discovery() {
    print_header "Testing assignment discovery"
    python test_local.py discover
    print_success "Assignment discovery test completed"
}

# Test branch name sanitization
test_sanitize() {
    print_header "Testing branch name sanitization"
    echo "Testing various paths:"
    echo "  'assignment-1' -> $(python test_local.py sanitize "assignment-1")"
    echo "  'week-3/assignment-3' -> $(python test_local.py sanitize "week-3/assignment-3")"
    echo "  'Module 4/Lab Assignment' -> $(python test_local.py sanitize "Module 4/Lab Assignment")"
    echo "  '  spaced  assignment  ' -> $(python test_local.py sanitize "  spaced  assignment  ")"
    print_success "Branch name sanitization test completed"
}

# Run code quality checks
run_quality_checks() {
    print_header "Running code quality checks"
    
    echo "1. Checking code formatting with Black..."
    if black --check --line-length=79 ../*.py *.py 2>/dev/null; then
        print_success "Code formatting check passed"
    else
        print_warning "Code formatting issues found. Run 'black --line-length=79 ../*.py *.py' to fix."
        return 1
    fi
    
    echo "2. Checking style with Flake8..."
    if flake8 ../*.py *.py --max-line-length=79 --exclude=__pycache__ 2>/dev/null; then
        print_success "Style check passed"
    else
        print_warning "Style issues found"
        return 1
    fi
    
    echo "3. Checking types with MyPy..."
    if mypy ../*.py --ignore-missing-imports 2>/dev/null; then
        print_success "Type checking passed"
    else
        print_warning "Type checking issues found"
        return 1
    fi
    
    print_success "All code quality checks passed"
}

# Run performance tests
run_performance_tests() {
    print_header "Running performance tests"
    
    echo "Creating large test structure..."
    mkdir -p large-test-assignments
    
    for i in $(seq 1 50); do
        mkdir -p "large-test-assignments/assignment-$i"
        echo "# Assignment $i" > "large-test-assignments/assignment-$i/README.md"
    done
    
    echo "Testing discovery with 50 assignments..."
    start_time=$(date +%s)
    ASSIGNMENTS_ROOT_REGEX='^large-test-assignments$' python test_local.py discover > /dev/null
    end_time=$(date +%s)
    duration=$((end_time - start_time))
    
    rm -rf large-test-assignments
    
    echo "Performance test completed in ${duration} seconds"
    if [ $duration -lt 5 ]; then
        print_success "Performance test passed (under 5 seconds)"
    else
        print_warning "Performance test completed but took over 5 seconds"
    fi
}

# Clean test artifacts
clean_artifacts() {
    print_header "Cleaning test artifacts"
    rm -rf __pycache__ tests/__pycache__ .pytest_cache .mypy_cache
    rm -f bandit-report.json safety-report.json large-test-assignments
    print_success "Cleanup completed"
}

# Run all tests
run_all_tests() {
    print_header "Running complete test suite"
    
    # Install dependencies if needed
    if ! command -v pytest &> /dev/null; then
        install_deps
    fi
    
    # Run tests in order
    run_quality_checks || print_warning "Quality checks had issues"
    run_unit_tests
    run_local_tests
    test_discovery
    test_sanitize
    
    print_success "Complete test suite finished"
}

# Test with multiple assignment types
test_multiple_types() {
    print_header "Testing with multiple assignment types"
    ASSIGNMENTS_ROOT_REGEX='^(assignments|homework|labs)$' \
    ASSIGNMENT_REGEX='^(assignment|hw|lab)-\d+$' \
    python test_local.py
    print_success "Multiple types test completed"
}

# Main script logic
case "${1:-help}" in
    help)
        show_help
        ;;
    all)
        run_all_tests
        ;;
    unit)
        run_unit_tests
        ;;
    local)
        run_local_tests
        ;;
    discovery)
        test_discovery
        ;;
    sanitize)
        test_sanitize
        ;;
    quality)
        run_quality_checks
        ;;
    performance)
        run_performance_tests
        ;;
    multiple)
        test_multiple_types
        ;;
    clean)
        clean_artifacts
        ;;
    install-deps)
        install_deps
        ;;
    *)
        echo "Unknown command: $1"
        echo ""
        show_help
        exit 1
        ;;
esac

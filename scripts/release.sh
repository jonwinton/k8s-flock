#!/bin/bash

# Release script for k8s-flock
# This script helps create releases using goreleaser

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the right directory
if [ ! -f "go.mod" ] || [ ! -f ".goreleaser.yml" ]; then
    print_error "This script must be run from the project root directory"
    exit 1
fi

# Check if hermit is available
if [ ! -f "./bin/hermit" ]; then
    print_error "Hermit not found. Please run this script from the project root."
    exit 1
fi

# Check if goreleaser is available
if [ ! -f "./bin/goreleaser" ]; then
    print_warning "Goreleaser not found. Installing..."
    ./bin/hermit install goreleaser
fi

# Function to show usage
show_usage() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  build     Build binaries for all platforms (snapshot)"
    echo "  release   Create a full release (requires git tag)"
    echo "  snapshot  Create a snapshot release"
    echo "  check     Validate goreleaser configuration"
    echo "  help      Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 build                    # Build binaries"
    echo "  $0 snapshot                 # Create snapshot release"
    echo "  git tag v1.0.0 && $0 release  # Create release for v1.0.0"
}

# Main script logic
case "${1:-help}" in
    build)
        print_status "Building binaries for all platforms..."
        ./bin/goreleaser build --snapshot --clean
        print_status "Build completed successfully!"
        ;;
    release)
        print_status "Creating release..."
        ./bin/goreleaser release --clean
        print_status "Release completed successfully!"
        ;;
    snapshot)
        print_status "Creating snapshot release..."
        ./bin/goreleaser release --snapshot --clean
        print_status "Snapshot release completed successfully!"
        ;;
    check)
        print_status "Validating goreleaser configuration..."
        ./bin/goreleaser check
        print_status "Configuration is valid!"
        ;;
    help|--help|-h)
        show_usage
        ;;
    *)
        print_error "Unknown command: $1"
        echo ""
        show_usage
        exit 1
        ;;
esac 
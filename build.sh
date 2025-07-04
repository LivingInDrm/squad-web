#!/bin/bash

# Claude Squad Build Script
# Builds the claude-squad application with Engine SDK

set -e

VERSION="1.1.0-sdk"
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo "🚀 Building Claude Squad v${VERSION}"
echo "📅 Build Date: ${BUILD_DATE}"
echo "🔗 Git Commit: ${GIT_COMMIT}"

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -f claude-squad claude-squad-sdk-demo

# Build flags for optimization
LDFLAGS="-s -w -X main.version=${VERSION} -X main.buildDate=${BUILD_DATE} -X main.gitCommit=${GIT_COMMIT}"

# Build main application
echo "🔨 Building main application..."
go build -ldflags="${LDFLAGS}" -o claude-squad main.go

# Build SDK demo
echo "🔨 Building SDK demo..."
go build -ldflags="-s -w" -o claude-squad-sdk-demo examples/sdk_demo.go

# Verify builds
echo "✅ Verifying builds..."
./claude-squad version
echo ""

# Show file sizes
echo "📦 Build artifacts:"
ls -lh claude-squad* | grep -v "\\.sh"

# Run tests
echo ""
echo "🧪 Running tests..."
go test ./... -v | grep -E "(PASS|FAIL|ok)"

echo ""
echo "✅ Build completed successfully!"
echo ""
echo "🎯 Next steps:"
echo "   • Run './claude-squad' to use the CLI with Engine SDK"
echo "   • Run './claude-squad-sdk-demo' to see SDK usage example"
echo "   • Check 'docs/ENGINE_SDK.md' for API documentation"
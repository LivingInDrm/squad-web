#!/bin/bash

# Claude Squad Installation Script
# Installs claude-squad v1.1.0-sdk with Engine SDK

set -e

echo "🚀 Claude Squad v1.1.0-sdk Installation"
echo "========================================"

# Check prerequisites
echo "🔍 Checking prerequisites..."

# Check Go
if ! command -v go &> /dev/null; then
    echo "❌ Go is required but not installed. Please install Go 1.23+."
    exit 1
fi

GO_VERSION=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | sed 's/go//')
echo "✅ Go version: $GO_VERSION"

# Check tmux
if ! command -v tmux &> /dev/null; then
    echo "⚠️  tmux not found. Some features may not work."
    echo "   Install with: brew install tmux (macOS) or apt install tmux (Linux)"
else
    echo "✅ tmux found"
fi

# Check git
if ! command -v git &> /dev/null; then
    echo "❌ git is required but not installed."
    exit 1
fi
echo "✅ git found"

# Check gh
if ! command -v gh &> /dev/null; then
    echo "⚠️  gh CLI not found. GitHub operations may not work."
    echo "   Install with: brew install gh (macOS) or see: https://cli.github.com/"
else
    echo "✅ gh CLI found"
fi

echo ""

# Check if we're in the right directory
if [[ ! -f "main.go" ]] || [[ ! -d "pkg/engine" ]]; then
    echo "❌ Please run this script from the claude-squad project directory"
    exit 1
fi

# Build the application
echo "🔨 Building claude-squad..."
./build.sh > /dev/null 2>&1

if [[ $? -eq 0 ]]; then
    echo "✅ Build successful"
else
    echo "❌ Build failed"
    exit 1
fi

# Create config directory
CONFIG_DIR="$HOME/.claude-squad"
echo "📁 Creating config directory: $CONFIG_DIR"
mkdir -p "$CONFIG_DIR"

# Verify installation
echo ""
echo "🧪 Verifying installation..."
if ./claude-squad version > /dev/null 2>&1; then
    echo "✅ claude-squad binary works"
else
    echo "❌ claude-squad binary failed to run"
    exit 1
fi

if ./claude-squad-sdk-demo --help > /dev/null 2>&1 || [[ $? -eq 2 ]]; then
    echo "✅ SDK demo binary works"
else
    echo "❌ SDK demo binary failed to run"
fi

# Check if we're in a git repository
if git rev-parse --git-dir > /dev/null 2>&1; then
    echo "✅ Git repository detected"
else
    echo "⚠️  Not in a git repository. claude-squad requires a git repository to run."
fi

echo ""
echo "🎉 Installation Complete!"
echo "========================"
echo ""
echo "📍 Current location: $(pwd)"
echo "🔧 Binaries created:"
echo "   • ./claude-squad         - Main CLI application"
echo "   • ./claude-squad-sdk-demo - SDK usage example"
echo ""
echo "🚀 Quick Start:"
echo "   • Run './claude-squad' to launch the TUI"
echo "   • Run './claude-squad --help' for command options"
echo "   • Run './claude-squad-sdk-demo' to see Engine SDK in action"
echo ""
echo "📚 Documentation:"
echo "   • docs/ENGINE_SDK.md     - Complete API documentation"
echo "   • RELEASE_NOTES.md       - Release information"
echo "   • examples/sdk_demo.go   - SDK usage example"
echo ""
echo "⚙️  Configuration:"
echo "   • Config file: ~/.claude-squad/config.json"
echo "   • State file:  ~/.claude-squad/state.json"
echo ""

# Optional: Add to PATH suggestion
CURRENT_DIR=$(pwd)
echo "💡 Optional: Add to PATH for global access:"
echo "   echo 'export PATH=\"$CURRENT_DIR:\$PATH\"' >> ~/.bashrc"
echo "   # or for zsh:"
echo "   echo 'export PATH=\"$CURRENT_DIR:\$PATH\"' >> ~/.zshrc"
echo ""

echo "✨ Enjoy using Claude Squad with the new Engine SDK!"
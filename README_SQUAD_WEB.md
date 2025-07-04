# Squad Web - Claude Squad with Engine SDK

> 🚀 **v1.1.0-sdk** - AI Agent Management Platform with Powerful Engine SDK

Welcome to Squad Web! This is an enhanced version of claude-squad featuring a comprehensive Engine SDK for building AI agent management applications.

## 🎯 What's New

### 🔧 **Engine SDK**
A powerful Go SDK for programmatic session management:
- **Thread-safe operations** with concurrent session handling
- **Real-time event streaming** for stdout/stderr/diff/state changes  
- **Clean API design** with interface-based architecture
- **Storage abstraction** supporting multiple backends

### ✅ **100% Backward Compatible**
- All existing CLI functionality preserved
- Same configuration and state file formats
- Zero migration required for existing users

## 🚀 Quick Start

### Prerequisites
- Go 1.23+
- tmux (for terminal multiplexing)
- git (for version control)
- gh CLI (optional, for GitHub operations)

### Installation

```bash
# Clone the repository
git clone https://github.com/LivingInDrm/squad-web.git
cd squad-web

# Install and build
./install.sh

# Or build manually
./build.sh
```

### CLI Usage (Unchanged)
```bash
# Launch the TUI interface
./claude-squad

# Use with specific program
./claude-squad -p "aider --model gpt-4"

# Enable auto-yes mode
./claude-squad -y

# Check version
./claude-squad version
```

### SDK Usage (New!)
```go
package main

import (
    "claude-squad/config"
    "claude-squad/pkg/engine"
    "context"
    "fmt"
)

func main() {
    // Initialize Engine
    cfg := config.LoadConfig()
    appState := config.LoadState()
    
    eng, err := engine.New(cfg, appState)
    if err != nil {
        panic(err)
    }
    defer eng.Close()
    
    // Start the engine
    ctx := context.Background()
    if err := eng.Start(ctx); err != nil {
        panic(err)
    }
    
    // Create a session
    sessionID, err := eng.StartSession(ctx, engine.SessionOpts{
        Title:   "my-ai-session",
        Path:    ".",
        Program: "claude",
        AutoYes: false,
    })
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Created session: %s\n", sessionID)
    
    // Subscribe to real-time events
    events, _ := eng.Events(sessionID)
    go func() {
        for event := range events {
            fmt.Printf("Event: %s at %s\n", event.Kind, event.Timestamp.Format("15:04:05"))
        }
    }()
    
    // List all sessions
    sessions := eng.List()
    for _, session := range sessions {
        fmt.Printf("Session: %s (%s) - %s\n", 
            session.Title, session.ID, session.Status)
    }
}
```

## 📦 Project Structure

```
squad-web/
├── pkg/engine/          # 🔧 Engine SDK
│   ├── engine.go        #   Main facade API
│   ├── manager.go       #   Session lifecycle management
│   ├── event.go         #   Real-time event system
│   ├── storage.go       #   Storage abstraction
│   ├── types.go         #   Type definitions
│   └── engine_test.go   #   Comprehensive tests
├── app/                 # 🖥️ TUI Application
├── session/             # 💼 Session Management
├── docs/                # 📚 Documentation
│   └── ENGINE_SDK.md    #   Complete API docs
├── examples/            # 💡 Usage Examples
│   └── sdk_demo.go      #   Working SDK example
├── build.sh             # 🔨 Build script
├── install.sh           # 📦 Installation script
└── RELEASE_NOTES.md     # 📋 Release information
```

## 🔧 Engine SDK Features

### Session Management
- `StartSession()` - Create and start AI agent sessions
- `Pause()` / `Resume()` - Pause/resume with git worktree preservation
- `Kill()` - Terminate and cleanup resources
- `List()` / `Get()` - Query session information

### Real-time Events
- **stdout/stderr**: Terminal output streaming
- **diff**: Git change notifications  
- **state**: Session status updates
- **Non-blocking**: Buffered channels with overflow protection

### Storage & Config
- **Pluggable storage**: Interface-based storage backends
- **File compatibility**: Existing state.json format supported
- **Configuration**: Runtime config updates

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Test Engine SDK specifically
go test -v ./pkg/engine/

# Test with race detection
go test -race ./pkg/engine/

# Run the SDK demo
./claude-squad-sdk-demo
```

## 📚 Documentation

- **[Complete API Documentation](docs/ENGINE_SDK.md)** - Full SDK reference
- **[Release Notes](RELEASE_NOTES.md)** - What's new in v1.1.0-sdk
- **[Examples](examples/)** - Working code examples
- **[Design Documents](m0_enginesdk_design.md)** - Technical design details

## 🛠️ Development

### Building from Source
```bash
# Clean build
go clean -cache
./build.sh

# Development build
go build -o claude-squad main.go
go build -o claude-squad-sdk-demo examples/sdk_demo.go
```

### Testing
```bash
# Unit tests
go test ./pkg/engine/

# Integration tests  
go test ./...

# Benchmarks
go test -bench=. ./pkg/engine/
```

## 🔮 Roadmap

This Engine SDK provides the foundation for:

### M1 Phase: Web GUI
- Local web interface using Engine SDK
- Real-time session monitoring via WebSocket
- Browser-based AI agent management

### M2 Phase: SaaS Platform  
- Multi-user support with authentication
- Containerized session isolation
- Horizontal scaling capabilities

### Future Enhancements
- REST API endpoints using Engine backend
- Plugin system for custom session types
- Advanced monitoring and metrics
- Database storage backends

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

## 🆘 Support

- **Documentation**: Check [docs/ENGINE_SDK.md](docs/ENGINE_SDK.md)
- **Examples**: See [examples/](examples/) directory
- **Issues**: [GitHub Issues](https://github.com/LivingInDrm/squad-web/issues)
- **Discussions**: [GitHub Discussions](https://github.com/LivingInDrm/squad-web/discussions)

## 🏆 Acknowledgments

- Built on the foundation of [claude-squad](https://github.com/smtg-ai/claude-squad)
- Powered by [Bubbletea](https://github.com/charmbracelet/bubbletea) for TUI
- Enhanced with comprehensive Engine SDK for programmatic access

---

**Made with ❤️ for the AI development community**

*Ready to build the future of AI agent management? Start with Squad Web!* 🚀
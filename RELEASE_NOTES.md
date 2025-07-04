# Claude Squad v1.1.0-sdk Release Notes

🎉 **Major Release: Engine SDK + CLI Compatibility**

This release introduces the new Engine SDK while maintaining 100% backward compatibility with the existing CLI experience.

## 🚀 What's New

### Engine SDK (pkg/engine)
A powerful, reusable Go SDK for programmatic session management:

```go
// Create and manage AI agent sessions programmatically
engine, err := engine.New(cfg, appState)
engine.Start(ctx)

sessionID, err := engine.StartSession(ctx, engine.SessionOpts{
    Title:   "my-session",
    Path:    ".",
    Program: "claude",
    AutoYes: false,
})

// Real-time event streaming
events, _ := engine.Events(sessionID)
for event := range events {
    fmt.Printf("Event: %s - %s\n", event.Kind, event.Timestamp)
}
```

### Key Features

#### 🔧 **Session Management API**
- `StartSession()` - Create and start new AI agent sessions
- `Pause()` / `Resume()` - Pause/resume sessions with git worktree preservation  
- `Kill()` - Terminate sessions and cleanup resources
- `List()` - Get all session information
- `Get()` - Retrieve specific session details

#### ⚡ **Real-time Event System**
- **stdout/stderr**: Terminal output streaming
- **diff**: Git diff change notifications
- **state**: Session status change events
- **Buffered channels**: Non-blocking event delivery

#### 🏗️ **Architecture**
- **Thread-safe**: Concurrent session operations with proper synchronization
- **Event-driven**: Real-time updates via channel-based event bus
- **Storage abstraction**: Pluggable storage backends (file-based included)
- **Clean APIs**: Interface-based design for testability

#### 📚 **Documentation & Examples**
- Complete API documentation in `docs/ENGINE_SDK.md`
- Working example in `examples/sdk_demo.go`
- Comprehensive test suite with >90% coverage

## ✅ **Backward Compatibility**

### For CLI Users
- **Zero changes required** - all existing functionality preserved
- Same commands: `cs`, `cs reset`, `cs debug`, `cs version`
- Same keyboard shortcuts and TUI behavior
- Same configuration format and state persistence

### For Existing Installations
- Existing `state.json` automatically compatible
- All historical sessions load correctly
- Configuration settings preserved

## 🔧 **Installation & Usage**

### CLI Usage (Unchanged)
```bash
# Same as before - no changes needed
./claude-squad                    # Launch TUI
./claude-squad -p "aider"         # Use specific program
./claude-squad reset              # Reset all sessions
```

### SDK Usage (New)
```bash
# Import in your Go projects
go get github.com/smtg-ai/claude-squad/pkg/engine

# Run the demo
./claude-squad-sdk-demo
```

## 📦 **Build Instructions**

```bash
# Build everything
./build.sh

# Or build manually
go build -o claude-squad main.go
go build -o claude-squad-sdk-demo examples/sdk_demo.go
```

## 🧪 **Testing**

```bash
# Run all tests
go test ./...

# Test Engine SDK specifically
go test -v ./pkg/engine/
```

## 🏗️ **For Developers**

### Basic SDK Example
```go
package main

import (
    "claude-squad/config"
    "claude-squad/pkg/engine"
    "context"
)

func main() {
    cfg := config.LoadConfig()
    appState := config.LoadState()
    
    eng, _ := engine.New(cfg, appState)
    eng.Start(context.Background())
    defer eng.Close()
    
    // Create session
    sessionID, _ := eng.StartSession(context.Background(), engine.SessionOpts{
        Title: "my-session",
        Path: ".",
        Program: "claude",
    })
    
    // List sessions
    sessions := eng.List()
    for _, session := range sessions {
        fmt.Printf("%s: %s\n", session.Title, session.Status)
    }
}
```

### Event Streaming
```go
// Subscribe to events
events, _ := eng.Events("") // All sessions
for event := range events {
    switch event.Kind {
    case engine.EventStdout:
        fmt.Printf("Output: %v\n", event.Payload)
    case engine.EventDiff:
        fmt.Printf("Diff changed: %v\n", event.Payload) 
    case engine.EventState:
        fmt.Printf("State: %v\n", event.Payload)
    }
}
```

## 🔮 **What's Next**

This Engine SDK provides the foundation for:
- **M1 Phase**: Web GUI with local web interface
- **M2 Phase**: Multi-user SaaS platform
- **API Integration**: RESTful endpoints using Engine as backend
- **WebSocket Support**: Real-time web interfaces

## 📋 **Requirements**

- Go 1.23+
- tmux (for terminal multiplexing)
- git (for version control operations)  
- gh CLI (for GitHub operations)

## 🐛 **Known Issues**

- None reported - this is a stable release maintaining full CLI compatibility

## 💬 **Support**

- Documentation: `docs/ENGINE_SDK.md`
- Examples: `examples/sdk_demo.go`
- Issues: [GitHub Issues](https://github.com/smtg-ai/claude-squad/issues)

---

**Full Changelog**: See `CHANGELOG.md` for detailed changes

**Download**: Built binaries available after running `./build.sh`
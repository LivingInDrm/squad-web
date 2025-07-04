# Changelog

All notable changes to claude-squad will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0-sdk] - 2024-01-XX

### Added

#### Engine SDK (M0 Phase)
- **New Engine SDK**: Added `pkg/engine` package providing programmatic interface for session management
- **Event System**: Real-time event notifications for session state changes, stdout/stderr, and git diffs
- **Storage Abstraction**: Pluggable storage interface supporting existing state.json format
- **Thread Safety**: Concurrent-safe session operations with proper synchronization
- **Session Management API**: Clean interface for creating, pausing, resuming, and terminating sessions
- **SDK Documentation**: Comprehensive API documentation in `docs/ENGINE_SDK.md`
- **SDK Example**: Working demo in `examples/sdk_demo.go` showing basic usage

#### Technical Improvements
- **Event-Driven Architecture**: Foundation for future WebSocket and Web API integration
- **Backward Compatibility**: 100% compatible with existing CLI functionality and state format
- **Test Coverage**: Unit tests for Engine SDK with >90% coverage
- **API Documentation**: Complete godoc documentation for all public interfaces

### Changed
- **Version**: Updated to 1.1.0-sdk to reflect Engine SDK integration
- **Architecture**: Extracted core session logic into reusable SDK while maintaining CLI compatibility

### Technical Details
- **Package Structure**: Added `pkg/engine/` with facade pattern for clean API
- **Event Bus**: Internal event system for real-time updates
- **Session Manager**: Thread-safe session registry with lifecycle management
- **Storage Layer**: Abstracted storage interface with file-based implementation

### Migration Path
- **CLI Users**: No changes required - all existing functionality preserved
- **Developers**: Can now import `claude-squad/pkg/engine` for programmatic session management
- **Future**: Foundation laid for M1 Web GUI and M2 SaaS platform phases

### For Developers
```go
// Basic Engine SDK usage
engine, err := engine.New(cfg, appState)
engine.Start(ctx)
sessionID, err := engine.StartSession(ctx, engine.SessionOpts{
    Title: "my-session",
    Path: ".",
    Program: "claude",
})
```

See `examples/sdk_demo.go` and `docs/ENGINE_SDK.md` for complete documentation.

---

## [1.0.5] - Previous Release

### Added
- Core session management functionality
- Tmux integration for terminal multiplexing
- Git worktree isolation for sessions
- Bubbletea TUI interface
- Configuration management
- Auto-yes daemon mode

### Features
- Create and manage multiple AI agent sessions
- Git branch isolation with automatic worktree management
- Real-time session monitoring and diff viewing
- Keyboard-driven TUI interface
- Persistent session state across restarts
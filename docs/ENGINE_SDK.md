# Claude Squad Engine SDK

The Engine SDK provides a clean, programmatic interface for managing AI agent sessions in claude-squad. It abstracts session lifecycle management, event handling, and storage operations for both CLI and Web interfaces.

## Features

- **Session Management**: Create, pause, resume, and terminate AI agent sessions
- **Event System**: Real-time notifications for session state changes, output, and git diffs
- **Storage Abstraction**: Pluggable storage backends (currently file-based)
- **Concurrency Safe**: Thread-safe operations with proper synchronization
- **Backward Compatible**: Maintains compatibility with existing state.json format

## Quick Start

```go
package main

import (
    "claude-squad/config"
    "claude-squad/pkg/engine"
    "context"
    "fmt"
    "log"
)

func main() {
    // Load configuration and state
    cfg := config.LoadConfig()
    appState := config.LoadState()
    
    // Create engine
    eng, err := engine.New(cfg, appState)
    if err != nil {
        log.Fatalf("Failed to create engine: %v", err)
    }
    defer eng.Close()
    
    // Start engine
    ctx := context.Background()
    if err := eng.Start(ctx); err != nil {
        log.Fatalf("Failed to start engine: %v", err)
    }
    
    // Create a session
    sessionID, err := eng.StartSession(ctx, engine.SessionOpts{
        Title:   "my-session",
        Path:    ".",
        Program: "claude",
        AutoYes: false,
    })
    if err != nil {
        log.Fatalf("Failed to create session: %v", err)
    }
    
    fmt.Printf("Created session: %s\n", sessionID)
    
    // List sessions
    sessions := eng.List()
    for _, session := range sessions {
        fmt.Printf("Session: %s (%s)\n", session.Title, session.Status)
    }
}
```

## API Reference

### Engine

The main facade for session management.

#### Creating an Engine

```go
func New(cfg *config.Config, appState config.StateManager) (*Engine, error)
```

Creates a new Engine instance with the given configuration and state manager.

#### Starting the Engine

```go
func (e *Engine) Start(ctx context.Context) error
```

Initializes the engine and loads existing sessions. Must be called before other operations.

#### Closing the Engine

```go
func (e *Engine) Close() error
```

Shuts down the engine, saves state, and cleans up resources.

### Session Operations

#### Creating Sessions

```go
func (e *Engine) StartSession(ctx context.Context, opts SessionOpts) (string, error)

type SessionOpts struct {
    Title   string  // Session title (must be unique)
    Path    string  // Working directory path
    Program string  // Program to run (e.g., "claude", "aider")
    AutoYes bool    // Auto-accept prompts
    Prompt  string  // Initial prompt to send
}
```

Creates and starts a new session, returning a session ID.

#### Managing Sessions

```go
func (e *Engine) Pause(sessionID string) error
func (e *Engine) Resume(sessionID string) error
func (e *Engine) Kill(sessionID string) error
```

- **Pause**: Stops tmux session and removes worktree (preserves branch)
- **Resume**: Recreates worktree and restarts tmux session
- **Kill**: Terminates session and cleans up all resources

#### Querying Sessions

```go
func (e *Engine) List() []SessionInfo
func (e *Engine) Get(sessionID string) (*SessionInfo, error)

type SessionInfo struct {
    ID        string     `json:"id"`
    Title     string     `json:"title"`
    Path      string     `json:"path"`
    Branch    string     `json:"branch"`
    Status    Status     `json:"status"`
    Program   string     `json:"program"`
    AutoYes   bool       `json:"auto_yes"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    DiffStats *DiffStats `json:"diff_stats,omitempty"`
}
```

### Event System

#### Subscribing to Events

```go
func (e *Engine) Events(sessionID string) (<-chan Event, error)
```

Returns a channel for receiving session events. Pass empty string for all sessions.

#### Event Types

```go
type Event struct {
    SessionID string      `json:"session_id"`
    Kind      EventKind   `json:"kind"`
    Payload   interface{} `json:"payload"`
    Timestamp time.Time   `json:"timestamp"`
}

type EventKind string
const (
    EventStdout EventKind = "stdout"
    EventStderr EventKind = "stderr"
    EventDiff   EventKind = "diff"
    EventState  EventKind = "state"
)
```

- **stdout**: Terminal output from the session
- **diff**: Git diff changes in the workspace
- **state**: Session status changes (running, paused, etc.)

#### Event Payloads

```go
// State change events
type StateEvent struct {
    Previous Status `json:"previous"`
    Current  Status `json:"current"`
}

// Diff change events
type DiffEvent struct {
    Stats      *DiffStats `json:"stats"`
    HasChanges bool       `json:"has_changes"`
}

// Output events
type StdoutEvent struct {
    Content string `json:"content"`
}
```

## Configuration

The Engine uses the standard claude-squad configuration:

```go
type Config struct {
    DefaultProgram     string `json:"default_program"`
    AutoYes            bool   `json:"auto_yes"`
    DaemonPollInterval int    `json:"daemon_poll_interval"`
    BranchPrefix       string `json:"branch_prefix"`
}
```

Configuration can be updated at runtime:

```go
func (e *Engine) UpdateConfig(cfg *config.Config) error
func (e *Engine) GetConfig() *config.Config
```

## Storage Interface

The Engine supports pluggable storage backends:

```go
type StorageInterface interface {
    LoadSessions() ([]SessionData, error)
    SaveSessions(sessions []SessionData) error
    LoadConfig() (*config.Config, error)
    SaveConfig(cfg *config.Config) error
}
```

The default implementation uses file-based storage compatible with the existing state.json format.

## Error Handling

All Engine methods return descriptive errors. Common error conditions:

- **Engine not started**: Operations called before `Start()`
- **Session not found**: Invalid session ID
- **Duplicate title**: Session with same title already exists
- **Storage errors**: File system or permission issues

## Thread Safety

The Engine is fully thread-safe and can be used concurrently from multiple goroutines. Internal synchronization ensures data consistency.

## Testing

The SDK includes mock implementations for testing:

```go
// Create a mock state manager for testing
type MockStateManager struct {
    // ... implementation
}

// Use in tests
engine, err := engine.New(cfg, &MockStateManager{})
```

See `pkg/engine/engine_test.go` for examples.

## Migration from Direct Session Usage

To migrate existing code from direct session usage:

### Before (Direct session usage)
```go
storage, _ := session.NewStorage(appState)
instances, _ := storage.LoadInstances()
instance, _ := session.NewInstance(opts)
instance.Start(true)
```

### After (Engine SDK)
```go
engine, _ := engine.New(cfg, appState)
engine.Start(ctx)
sessionID, _ := engine.StartSession(ctx, opts)
sessions := engine.List()
```

## Examples

See `examples/sdk_demo.go` for a complete working example demonstrating:

- Engine initialization
- Session creation and management
- Event subscription and handling
- Error handling
- Cleanup

## Future Enhancements

The Engine SDK is designed to support future enhancements:

- **Web API**: RESTful endpoints using Engine as backend
- **WebSocket support**: Real-time web interfaces using event system
- **Database storage**: SQL backends for multi-user deployments
- **Metrics and monitoring**: Observability features
- **Plugin system**: Custom session types and behaviors
# M0 Detailed Implementation Plan - Engine SDK + CLI Compatibility

## Code Analysis Summary

### Current Architecture Deep Dive

**Key Components Analyzed:**
- `main.go`: Cobra CLI with `app.Run()` as entry point
- `app/home.go`: 735-line TUI implementation with session management
- `session/instance.go`: Core session abstraction with 25+ public methods
- `session/storage.go`: JSON-based persistence with `config.State`
- `config/config.go`: Configuration management with file-based storage

**Critical Workflows Identified:**
1. **Session Creation**: `NewInstance()` → `Start()` → `tmux.Start()` + `git.Setup()` 
2. **State Monitoring**: 500ms polling via `tickUpdateMetadataCmd`
3. **Persistence**: `ToInstanceData()` → JSON → `state.json`
4. **Auto-Yes**: Background daemon for prompt handling

**Current Dependencies:**
- TUI directly manipulates `session.Instance` objects
- Polling-based updates with complex state diffing
- File-based storage tightly coupled to config package
- No event system - all updates are pull-based

## Implementation Strategy

### Phase 1: Engine SDK Foundation

#### 1.1 Package Structure Creation
```
pkg/engine/
├── engine.go      # Main facade and API
├── manager.go     # Session registry and lifecycle
├── event.go       # Event system definitions
├── storage.go     # Storage abstraction
├── types.go       # Public types and interfaces
└── internal/      # Internal implementations
    ├── session/   # Wrapped session package
    ├── git/       # Wrapped git package
    └── tmux/      # Wrapped tmux package
```

#### 1.2 Core Types Design
```go
// engine.go
type Engine struct {
    mu      sync.RWMutex
    mgr     *manager
    store   StorageInterface
    cfg     *config.Config
    eventCh chan Event
    done    chan struct{}
}

// Public API
func New(cfg *config.Config, appState config.AppState) (*Engine, error)
func (e *Engine) Start(ctx context.Context, opts SessionOpts) (string, error)
func (e *Engine) Pause(sessionID string) error
func (e *Engine) Resume(sessionID string) error
func (e *Engine) Kill(sessionID string) error
func (e *Engine) List() []SessionInfo
func (e *Engine) Get(sessionID string) (*SessionInfo, error)
func (e *Engine) Events(sessionID string) (<-chan Event, error)
func (e *Engine) Close() error
```

#### 1.3 Event System Design
```go
// event.go
type EventKind string

const (
    EventStdout EventKind = "stdout"
    EventStderr EventKind = "stderr"
    EventDiff   EventKind = "diff"
    EventState  EventKind = "state"
)

type Event struct {
    SessionID string      `json:"session_id"`
    Kind      EventKind   `json:"kind"`
    Payload   interface{} `json:"payload"`
    Timestamp time.Time   `json:"timestamp"`
}

type StateEvent struct {
    Previous Status `json:"previous"`
    Current  Status `json:"current"`
}

type DiffEvent struct {
    Stats    DiffStats `json:"stats"`
    HasChang bool      `json:"has_changes"`
}
```

#### 1.4 Session Management
```go
// manager.go
type manager struct {
    mu        sync.RWMutex
    sessions  map[string]*sessionWrapper
    eventCh   chan Event
    cfg       *config.Config
    store     StorageInterface
}

type sessionWrapper struct {
    instance *session.Instance
    id       string
    eventCh  chan Event
    stopCh   chan struct{}
}

func (m *manager) Create(opts SessionOpts) (string, error)
func (m *manager) Get(id string) (*sessionWrapper, error)
func (m *manager) List() []SessionInfo
func (m *manager) Delete(id string) error
func (m *manager) watch(sw *sessionWrapper)
```

#### 1.5 Storage Abstraction
```go
// storage.go
type StorageInterface interface {
    LoadSessions() ([]SessionData, error)
    SaveSessions(sessions []SessionData) error
    LoadConfig() (*config.Config, error)
    SaveConfig(cfg *config.Config) error
}

type fileStorage struct {
    statePath  string
    configPath string
}

func NewFileStorage(statePath, configPath string) StorageInterface
func (fs *fileStorage) LoadSessions() ([]SessionData, error)
func (fs *fileStorage) SaveSessions(sessions []SessionData) error
```

### Phase 2: CLI Integration

#### 2.1 Main Entry Point Changes
```go
// main.go - Key changes
func main() {
    // Initialize Engine instead of direct app.Run()
    cfg, appState := config.Load()
    engine, err := engine.New(cfg, appState)
    if err != nil {
        log.Fatal(err)
    }
    defer engine.Close()
    
    // Pass engine to app
    app := app.New(engine, cfg)
    if err := app.Run(); err != nil {
        log.Fatal(err)
    }
}
```

#### 2.2 TUI Refactoring
```go
// app/home.go - Key changes
type home struct {
    engine      *engine.Engine
    sessions    []engine.SessionInfo  // Cached from Engine
    eventCh     <-chan engine.Event   // Event subscription
    // ... other fields unchanged
}

func (h *home) Init() tea.Cmd {
    // Subscribe to engine events
    eventCh, err := h.engine.Events("") // All sessions
    if err != nil {
        return h.setError(err)
    }
    h.eventCh = eventCh
    
    // Load initial sessions
    h.sessions = h.engine.List()
    
    return tea.Batch(
        h.tickUpdateMetadata(),
        h.listenToEvents(),
    )
}

func (h *home) listenToEvents() tea.Cmd {
    return func() tea.Msg {
        select {
        case event := <-h.eventCh:
            return eventMsg(event)
        case <-time.After(100 * time.Millisecond):
            return eventTimeoutMsg{}
        }
    }
}
```

#### 2.3 Session Operations Integration
```go
// Replace direct session.Instance calls with Engine API
func (h *home) createSession(opts sessionOpts) tea.Cmd {
    return func() tea.Msg {
        engineOpts := engine.SessionOpts{
            Title:   opts.title,
            Path:    opts.path,
            Program: opts.program,
            AutoYes: opts.autoYes,
            Prompt:  opts.prompt,
        }
        
        sessionID, err := h.engine.Start(context.Background(), engineOpts)
        if err != nil {
            return errMsg(err)
        }
        
        // Refresh session list
        h.sessions = h.engine.List()
        return sessionCreatedMsg{ID: sessionID}
    }
}

func (h *home) pauseSession(sessionID string) tea.Cmd {
    return func() tea.Msg {
        if err := h.engine.Pause(sessionID); err != nil {
            return errMsg(err)
        }
        return sessionPausedMsg{ID: sessionID}
    }
}
```

### Phase 3: Backward Compatibility

#### 3.1 Storage Format Compatibility
```go
// Maintain exact same state.json format
type SessionData struct {
    ID       string            `json:"id"`        // New field
    Title    string            `json:"title"`     // Existing
    Path     string            `json:"path"`      // Existing
    Branch   string            `json:"branch"`    // Existing
    Program  string            `json:"program"`   // Existing
    AutoYes  bool              `json:"auto_yes"`  // Existing
    Status   string            `json:"status"`    // Existing
    // ... all other existing fields
}

func (e *Engine) migrateFromLegacyState(state config.AppState) error {
    // Convert existing instances to new format
    for _, instanceData := range state.Instances {
        sessionData := SessionData{
            ID:      uuid.NewString(), // Generate new ID
            Title:   instanceData.Title,
            Path:    instanceData.Path,
            Branch:  instanceData.Branch,
            Program: instanceData.Program,
            AutoYes: instanceData.AutoYes,
            Status:  string(instanceData.Status),
        }
        // Create session wrapper and restore state
    }
    return nil
}
```

#### 3.2 Configuration Compatibility
```go
// config/config.go - No changes needed
// Engine uses existing config.Config struct
func (e *Engine) UpdateConfig(cfg *config.Config) error {
    e.mu.Lock()
    defer e.mu.Unlock()
    
    e.cfg = cfg
    return e.store.SaveConfig(cfg)
}
```

### Phase 4: Testing Strategy

#### 4.1 Unit Tests
```go
// pkg/engine/engine_test.go
func TestEngineLifecycle(t *testing.T) {
    // Test Engine creation, session management
    cfg := &config.Config{DefaultProgram: "test"}
    engine, err := New(cfg, config.AppState{})
    assert.NoError(t, err)
    defer engine.Close()
    
    // Test session creation
    sessionID, err := engine.Start(context.Background(), SessionOpts{
        Title:   "test",
        Path:    "/tmp",
        Program: "echo hello",
    })
    assert.NoError(t, err)
    assert.NotEmpty(t, sessionID)
    
    // Test session operations
    err = engine.Pause(sessionID)
    assert.NoError(t, err)
    
    err = engine.Resume(sessionID)
    assert.NoError(t, err)
    
    err = engine.Kill(sessionID)
    assert.NoError(t, err)
}
```

#### 4.2 Integration Tests
```go
// pkg/engine/integration_test.go
func TestEngineWithRealTmux(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // Test with actual tmux and git
    tmpDir := t.TempDir()
    // Initialize git repo
    // Test full session lifecycle
}
```

#### 4.3 CLI Regression Tests
```bash
# test/cli_regression.exp
#!/usr/bin/expect -f
# Test CLI behavior remains identical

spawn ./claude-squad
expect "Session Manager"
send "n"
expect "New Session"
send "test-session\r"
expect "test-session"
send "q"
expect eof
```

## Implementation Timeline

### Week 1: Foundation (Days 1-7)
- **Day 1-2**: Create package structure, core types
- **Day 3-4**: Implement Engine facade and manager
- **Day 5-6**: Build event system and storage abstraction
- **Day 7**: Unit tests for Engine SDK

### Week 2: Integration (Days 8-14)
- **Day 8-9**: Refactor CLI to use Engine
- **Day 10-11**: Ensure event system works with TUI
- **Day 12**: Integration testing
- **Day 13**: Regression testing with expect scripts
- **Day 14**: Documentation and cleanup

## Risk Mitigation

### Technical Risks
- **Circular dependencies**: Use internal/ package pattern
- **Performance regression**: Benchmark critical paths
- **Event storms**: Buffer and deduplicate events
- **Memory leaks**: Proper cleanup and context cancellation

### Compatibility Risks
- **State migration**: Careful handling of existing state.json
- **UI behavior**: Extensive testing of TUI interactions
- **Configuration**: Preserve all existing config options

## Success Metrics

- [ ] Engine SDK passes all unit tests (>90% coverage)
- [ ] CLI behavior identical to current version
- [ ] All existing sessions load correctly
- [ ] Event system performs better than polling
- [ ] SDK documentation complete with examples
- [ ] Integration tests pass with real tmux/git

## File Changes Summary

### New Files
- `pkg/engine/engine.go` - Main SDK facade
- `pkg/engine/manager.go` - Session management
- `pkg/engine/event.go` - Event system
- `pkg/engine/storage.go` - Storage abstraction
- `pkg/engine/types.go` - Public types
- `examples/sdk_demo.go` - Usage example

### Modified Files
- `main.go` - Initialize Engine instead of direct app.Run()
- `app/home.go` - Use Engine API instead of direct session calls
- `go.mod` - Add Engine SDK to module

### No Changes Required
- All existing packages (session/, git/, tmux/, config/)
- All existing configuration and storage formats
- All existing CLI commands and flags

This plan ensures zero breaking changes while creating the foundation for future Web development phases.
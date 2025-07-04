# M0 Phase Implementation Plan - Engine SDK + CLI Compatibility

## Overview
This plan implements the M0 phase of the claude-squad WebUI transformation, extracting core session management logic into a reusable Engine SDK while maintaining 100% CLI compatibility.

## Current State Analysis
- **Architecture**: Well-structured Go project with clear separation between TUI (`app/`), session management (`session/`), git operations (`session/git/`), and tmux handling (`session/tmux/`)
- **Key Components**: 
  - `session.Instance` struct manages AI agent workspaces
  - Bubbletea TUI provides three-pane interface
  - JSON-based persistence via `config.AppState`
  - Event-driven updates via 100ms polling
- **Technical Debt**: Large app.go file (735+ lines), mixed responsibilities, limited test coverage

## Implementation Strategy

### Phase 1: Engine SDK Foundation
**Goal**: Create `pkg/engine/` package with clean facade API

**Key Deliverables**:
- `engine.go` - Main facade with `Start()`, `Pause()`, `Resume()`, `Kill()`, `List()` methods
- `manager.go` - Thread-safe session registry with event dispatch  
- `event.go` - Unified event system for stdout/stderr/diff/state changes
- `storage.go` - Storage abstraction wrapping existing `session.Storage`

**Design Principles**:
- **Facade Pattern**: Single entry point hiding internal complexity
- **Event-Driven**: Replace polling with channel-based event system
- **Backward Compatibility**: Preserve all existing package paths and APIs
- **Testability**: Interface-based design for dependency injection

### Phase 2: CLI Integration
**Goal**: Refactor CLI to use Engine SDK exclusively while maintaining identical behavior

**Key Changes**:
- `main.go`: Initialize Engine instead of direct app.Run()
- `app/home.go`: Replace direct session manipulation with Engine API calls
- Event subscription: Replace polling with Engine event channels
- Error handling: Preserve existing error display logic

**Compatibility Requirements**:
- All CLI commands and flags remain identical
- Bubbletea TUI behavior unchanged
- Existing state.json format preserved
- All keyboard shortcuts and interactions identical

### Phase 3: Testing & Validation
**Goal**: Ensure zero regression and robust SDK functionality

**Test Strategy**:
- Unit tests for Engine SDK (>80% coverage target)
- Integration tests for Engine lifecycle operations
- Regression tests for CLI compatibility
- Mock implementations for git/tmux backends

## Risk Mitigation

### Technical Risks
- **Circular Dependencies**: Use `internal/` package structure to limit visibility
- **Concurrency Issues**: Implement proper locking strategy (RWMutex for session registry)
- **Event Storms**: Rate limit events (≥200ms intervals) and deduplicate
- **Performance Regression**: Benchmark critical paths and optimize

### Compatibility Risks  
- **CLI Behavior Changes**: Automated expect scripts for UI testing
- **State Migration**: Comprehensive testing of state.json loading
- **Configuration Drift**: Preserve exact config format and defaults

## Success Criteria

### For CLI Users (Stories C-1, C-2, C-3)
- [ ] All `cs` commands work identically after upgrade
- [ ] TUI keyboard shortcuts and display unchanged
- [ ] Historical sessions load correctly from state.json
- [ ] Session operations (create/pause/resume/kill) function properly

### For SDK Developers (Stories D-1, D-2, D-3)
- [ ] `go get github.com/smtg-ai/claude-squad/pkg/engine` succeeds
- [ ] Engine API documented with clear examples
- [ ] Mock backends available for testing
- [ ] SDK test coverage >80%

### For Maintainers (Stories M-1, M-2)
- [ ] CI passes all tests including race detection
- [ ] CHANGELOG documents all changes
- [ ] No breaking changes to existing workflows

## Implementation Timeline

### Week 1: Foundation
- Days 1-2: Create pkg/engine structure and core interfaces
- Days 3-4: Implement Engine facade and manager
- Days 5-7: Build event system and storage abstraction

### Week 2: Integration & Testing
- Days 8-10: Refactor CLI to use Engine SDK
- Days 11-12: Comprehensive testing and validation
- Days 13-14: Documentation and CI setup

## Dependencies & Prerequisites
- **Runtime**: tmux, git, gh CLI
- **Build**: Go 1.23+
- **Testing**: expect scripts for UI automation
- **Documentation**: Godoc tooling

## Next Steps
1. Complete Stage 1 by analyzing detailed code paths
2. Create detailed implementation plan with specific file changes
3. Begin Engine SDK implementation following design patterns
4. Iterate with continuous testing to ensure CLI compatibility

This plan ensures minimal risk through incremental changes while laying the foundation for future Web API development in M1 phase.
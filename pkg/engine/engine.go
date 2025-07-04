// Package engine provides the core SDK for managing AI agent sessions.
// It abstracts session lifecycle management, event handling, and storage
// operations for both CLI and Web interfaces.
package engine

import (
	"claude-squad/config"
	"context"
	"fmt"
	"sync"
)

// Engine is the main facade for the session management SDK.
// It provides a clean API for creating, managing, and monitoring AI agent sessions.
type Engine struct {
	mu       sync.RWMutex
	mgr      *manager
	store    StorageInterface
	eventBus *EventBus
	cfg      *config.Config
	started  bool
}

// New creates a new Engine instance.
// The cfg parameter contains application configuration.
// The appState parameter provides access to persistent storage.
func New(cfg *config.Config, appState config.StateManager) (*Engine, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if appState == nil {
		return nil, fmt.Errorf("appState cannot be nil")
	}
	
	// Create storage interface
	store, err := NewFileStorage(appState)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}
	
	// Create event bus
	eventBus := NewEventBus()
	
	// Create session manager
	mgr := newManager(cfg, store, eventBus)
	
	engine := &Engine{
		mgr:      mgr,
		store:    store,
		eventBus: eventBus,
		cfg:      cfg,
		started:  false,
	}
	
	return engine, nil
}

// Start initializes the engine and loads existing sessions.
// This must be called before using other Engine methods.
func (e *Engine) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if e.started {
		return fmt.Errorf("engine already started")
	}
	
	if err := e.mgr.Start(ctx); err != nil {
		return fmt.Errorf("failed to start manager: %w", err)
	}
	
	e.started = true
	return nil
}

// Close shuts down the engine and cleans up resources.
// This should be called when the application exits.
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if !e.started {
		return nil
	}
	
	// Stop the manager (saves state)
	if err := e.mgr.Stop(); err != nil {
		return fmt.Errorf("failed to stop manager: %w", err)
	}
	
	// Close event bus
	e.eventBus.Close()
	
	e.started = false
	return nil
}

// Start creates and starts a new session with the given options.
// Returns the session ID on success.
func (e *Engine) StartSession(ctx context.Context, opts SessionOpts) (string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	if !e.started {
		return "", fmt.Errorf("engine not started")
	}
	
	return e.mgr.Create(opts)
}

// Pause pauses the specified session.
// This stops the tmux session and removes the worktree while preserving the branch.
func (e *Engine) Pause(sessionID string) error {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	if !e.started {
		return fmt.Errorf("engine not started")
	}
	
	return e.mgr.Pause(sessionID)
}

// Resume resumes a paused session.
// This recreates the worktree and restarts the tmux session.
func (e *Engine) Resume(sessionID string) error {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	if !e.started {
		return fmt.Errorf("engine not started")
	}
	
	return e.mgr.Resume(sessionID)
}

// Kill terminates the specified session and cleans up all resources.
func (e *Engine) Kill(sessionID string) error {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	if !e.started {
		return fmt.Errorf("engine not started")
	}
	
	return e.mgr.Kill(sessionID)
}

// List returns information about all sessions.
func (e *Engine) List() []SessionInfo {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	if !e.started {
		return []SessionInfo{}
	}
	
	return e.mgr.List()
}

// Get returns information about a specific session.
func (e *Engine) Get(sessionID string) (*SessionInfo, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	if !e.started {
		return nil, fmt.Errorf("engine not started")
	}
	
	wrapper, err := e.mgr.Get(sessionID)
	if err != nil {
		return nil, err
	}
	
	info := e.mgr.wrapperToSessionInfo(wrapper)
	return &info, nil
}

// Events returns a channel that receives events for the specified session.
// If sessionID is empty, receives events for all sessions.
// The returned channel will be closed when the engine is shut down.
func (e *Engine) Events(sessionID string) (<-chan Event, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	if !e.started {
		return nil, fmt.Errorf("engine not started")
	}
	
	return e.eventBus.Subscribe(sessionID), nil
}

// UpdateConfig updates the engine configuration.
func (e *Engine) UpdateConfig(cfg *config.Config) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}
	
	e.cfg = cfg
	e.mgr.cfg = cfg
	
	return e.store.SaveConfig(cfg)
}

// GetConfig returns the current configuration.
func (e *Engine) GetConfig() *config.Config {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	return e.cfg
}

// IsStarted returns whether the engine has been started.
func (e *Engine) IsStarted() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	return e.started
}
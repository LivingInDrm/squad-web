package engine

import (
	"claude-squad/config"
	"claude-squad/session"
	"context"
	"fmt"
	"sync"
	"time"
)

// manager handles session lifecycle and state management
type manager struct {
	mu        sync.RWMutex
	sessions  map[string]*sessionWrapper
	eventBus  *EventBus
	cfg       *config.Config
	store     StorageInterface
	stopCh    chan struct{}
	wg        sync.WaitGroup
}

// sessionWrapper wraps a session.Instance with additional metadata
type sessionWrapper struct {
	instance   *session.Instance
	id         string
	lastStdout string
	lastDiff   *DiffStats
	stopCh     chan struct{}
}

// newManager creates a new session manager
func newManager(cfg *config.Config, store StorageInterface, eventBus *EventBus) *manager {
	return &manager{
		sessions: make(map[string]*sessionWrapper),
		eventBus: eventBus,
		cfg:      cfg,
		store:    store,
		stopCh:   make(chan struct{}),
	}
}

// Start initializes the manager and loads existing sessions
func (m *manager) Start(ctx context.Context) error {
	// Load existing sessions from storage
	sessionData, err := m.store.LoadSessions()
	if err != nil {
		return fmt.Errorf("failed to load sessions: %w", err)
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Restore sessions
	for _, data := range sessionData {
		if err := m.restoreSession(data); err != nil {
			// Log error but continue with other sessions
			fmt.Printf("Warning: failed to restore session %s: %v\n", data.ID, err)
		}
	}
	
	return nil
}

// Stop shuts down the manager and all sessions
func (m *manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Stop all session watchers
	close(m.stopCh)
	m.wg.Wait()
	
	// Save current state
	if err := m.saveAll(); err != nil {
		return fmt.Errorf("failed to save sessions: %w", err)
	}
	
	return nil
}

// Create creates a new session
func (m *manager) Create(opts SessionOpts) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Check for duplicate titles (maintaining backward compatibility)
	for _, sw := range m.sessions {
		if sw.instance.Title == opts.Title {
			return "", fmt.Errorf("session with title '%s' already exists", opts.Title)
		}
	}
	
	// Create new instance
	instanceOpts := session.InstanceOptions{
		Title:   opts.Title,
		Path:    opts.Path,
		Program: opts.Program,
		AutoYes: opts.AutoYes,
	}
	
	instance, err := session.NewInstance(instanceOpts)
	if err != nil {
		return "", fmt.Errorf("failed to create instance: %w", err)
	}
	
	// Set AutoYes from options
	instance.AutoYes = opts.AutoYes
	
	// Start the instance
	if err := instance.Start(true); err != nil {
		return "", fmt.Errorf("failed to start instance: %w", err)
	}
	
	// Send initial prompt if provided
	if opts.Prompt != "" {
		if err := instance.SendPrompt(opts.Prompt); err != nil {
			// Log warning but don't fail session creation
			fmt.Printf("Warning: failed to send initial prompt: %v\n", err)
		}
	}
	
	// Generate session ID (use title for backward compatibility)
	sessionID := generateSessionID(opts.Title)
	
	// Create wrapper and start watching
	wrapper := &sessionWrapper{
		instance: instance,
		id:       sessionID,
		stopCh:   make(chan struct{}),
	}
	
	m.sessions[sessionID] = wrapper
	
	// Start watching the session
	m.wg.Add(1)
	go m.watchSession(wrapper)
	
	// Publish creation event
	m.eventBus.Publish(createEvent(sessionID, EventState, StateEvent{
		Previous: StatusLoading,
		Current:  convertStatus(instance.Status),
	}))
	
	return sessionID, nil
}

// Get retrieves a session by ID
func (m *manager) Get(sessionID string) (*sessionWrapper, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	wrapper, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	
	return wrapper, nil
}

// List returns information about all sessions
func (m *manager) List() []SessionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	sessions := make([]SessionInfo, 0, len(m.sessions))
	for _, wrapper := range m.sessions {
		sessions = append(sessions, m.wrapperToSessionInfo(wrapper))
	}
	
	return sessions
}

// Pause pauses a session
func (m *manager) Pause(sessionID string) error {
	wrapper, err := m.Get(sessionID)
	if err != nil {
		return err
	}
	
	prevStatus := convertStatus(wrapper.instance.Status)
	
	if err := wrapper.instance.Pause(); err != nil {
		return fmt.Errorf("failed to pause session: %w", err)
	}
	
	// Publish state change event
	m.eventBus.Publish(createEvent(sessionID, EventState, StateEvent{
		Previous: prevStatus,
		Current:  StatusPaused,
	}))
	
	return nil
}

// Resume resumes a paused session
func (m *manager) Resume(sessionID string) error {
	wrapper, err := m.Get(sessionID)
	if err != nil {
		return err
	}
	
	prevStatus := convertStatus(wrapper.instance.Status)
	
	if err := wrapper.instance.Resume(); err != nil {
		return fmt.Errorf("failed to resume session: %w", err)
	}
	
	// Publish state change event
	m.eventBus.Publish(createEvent(sessionID, EventState, StateEvent{
		Previous: prevStatus,
		Current:  convertStatus(wrapper.instance.Status),
	}))
	
	return nil
}

// Kill terminates a session
func (m *manager) Kill(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	wrapper, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	
	// Stop watching
	close(wrapper.stopCh)
	
	// Kill the instance
	if err := wrapper.instance.Kill(); err != nil {
		return fmt.Errorf("failed to kill session: %w", err)
	}
	
	// Remove from map
	delete(m.sessions, sessionID)
	
	// Publish termination event
	m.eventBus.Publish(createEvent(sessionID, EventState, StateEvent{
		Previous: convertStatus(wrapper.instance.Status),
		Current:  StatusPaused, // Use paused as "terminated" state
	}))
	
	return nil
}

// watchSession monitors a session for changes and publishes events
func (m *manager) watchSession(wrapper *sessionWrapper) {
	defer m.wg.Done()
	
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-wrapper.stopCh:
			return
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.checkSessionUpdates(wrapper)
		}
	}
}

// checkSessionUpdates checks for stdout, diff, and state changes
func (m *manager) checkSessionUpdates(wrapper *sessionWrapper) {
	instance := wrapper.instance
	
	// Check for stdout updates
	if content, err := instance.Preview(); err == nil && content != wrapper.lastStdout {
		wrapper.lastStdout = content
		m.eventBus.Publish(createEvent(wrapper.id, EventStdout, StdoutEvent{
			Content: content,
		}))
	}
	
	// Check for diff updates
	if err := instance.UpdateDiffStats(); err == nil {
		currentDiff := convertDiffStats(instance.GetDiffStats())
		if !diffStatsEqual(currentDiff, wrapper.lastDiff) {
			wrapper.lastDiff = currentDiff
			m.eventBus.Publish(createEvent(wrapper.id, EventDiff, DiffEvent{
				Stats:      currentDiff,
				HasChanges: currentDiff != nil,
			}))
		}
	}
	
	// Handle auto-yes functionality
	if instance.AutoYes {
		if updated, hasPrompt := instance.HasUpdated(); hasPrompt && !updated {
			instance.TapEnter()
		}
	}
}

// restoreSession recreates a session from stored data
func (m *manager) restoreSession(data SessionData) error {
	// Convert SessionData to session.InstanceData
	instanceData := session.InstanceData{
		Title:     data.Title,
		Path:      data.Path,
		Branch:    data.Branch,
		Status:    convertToSessionStatus(data.Status),
		Program:   data.Program,
		AutoYes:   data.AutoYes,
		Worktree:  data.Worktree,
		DiffStats: data.DiffStats,
	}
	
	// Parse timestamps
	createdAt, err := time.Parse("2006-01-02T15:04:05Z07:00", data.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to parse created_at: %w", err)
	}
	instanceData.CreatedAt = createdAt
	
	updatedAt, err := time.Parse("2006-01-02T15:04:05Z07:00", data.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to parse updated_at: %w", err)
	}
	instanceData.UpdatedAt = updatedAt
	
	// Restore instance
	instance, err := session.FromInstanceData(instanceData)
	if err != nil {
		return fmt.Errorf("failed to restore instance: %w", err)
	}
	
	// Create wrapper
	wrapper := &sessionWrapper{
		instance: instance,
		id:       data.ID,
		stopCh:   make(chan struct{}),
	}
	
	m.sessions[data.ID] = wrapper
	
	// Start watching if not paused
	if data.Status != StatusPaused {
		m.wg.Add(1)
		go m.watchSession(wrapper)
	}
	
	return nil
}

// saveAll saves all sessions to storage
func (m *manager) saveAll() error {
	sessions := make([]SessionData, 0, len(m.sessions))
	
	for _, wrapper := range m.sessions {
		data := wrapper.instance.ToInstanceData()
		sessionData := SessionData{
			ID:        wrapper.id,
			Title:     data.Title,
			Path:      data.Path,
			Branch:    data.Branch,
			Status:    convertStatus(data.Status),
			Program:   data.Program,
			AutoYes:   data.AutoYes,
			CreatedAt: data.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: data.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Worktree:  data.Worktree,
			DiffStats: data.DiffStats,
		}
		sessions = append(sessions, sessionData)
	}
	
	return m.store.SaveSessions(sessions)
}

// wrapperToSessionInfo converts a sessionWrapper to SessionInfo
func (m *manager) wrapperToSessionInfo(wrapper *sessionWrapper) SessionInfo {
	data := wrapper.instance.ToInstanceData()
	return SessionInfo{
		ID:        wrapper.id,
		Title:     data.Title,
		Path:      data.Path,
		Branch:    data.Branch,
		Status:    convertStatus(data.Status),
		Program:   data.Program,
		AutoYes:   data.AutoYes,
		CreatedAt: data.CreatedAt,
		UpdatedAt: data.UpdatedAt,
		DiffStats: convertDiffStats(wrapper.instance.GetDiffStats()),
	}
}

// diffStatsEqual compares two DiffStats for equality
func diffStatsEqual(a, b *DiffStats) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Added == b.Added && a.Removed == b.Removed && a.Content == b.Content
}
package engine

import (
	"claude-squad/config"
	"context"
	"encoding/json"
	"testing"
	"time"
)

// MockStateManager provides a simple in-memory implementation for testing
type MockStateManager struct {
	helpScreensSeen uint32
	instancesData   json.RawMessage
}

func (m *MockStateManager) SaveInstances(instancesJSON json.RawMessage) error {
	m.instancesData = instancesJSON
	return nil
}

func (m *MockStateManager) GetInstances() json.RawMessage {
	if m.instancesData == nil {
		return json.RawMessage("[]")
	}
	return m.instancesData
}

func (m *MockStateManager) DeleteAllInstances() error {
	m.instancesData = json.RawMessage("[]")
	return nil
}

func (m *MockStateManager) GetHelpScreensSeen() uint32 {
	return m.helpScreensSeen
}

func (m *MockStateManager) SetHelpScreensSeen(seen uint32) error {
	m.helpScreensSeen = seen
	return nil
}

func TestEngineCreation(t *testing.T) {
	cfg := &config.Config{
		DefaultProgram:     "echo test",
		AutoYes:            false,
		DaemonPollInterval: 1000,
		BranchPrefix:       "test/",
	}
	
	appState := &MockStateManager{}
	
	engine, err := New(cfg, appState)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	if engine == nil {
		t.Fatal("Engine is nil")
	}
	
	if engine.IsStarted() {
		t.Fatal("Engine should not be started initially")
	}
}

func TestEngineStartStop(t *testing.T) {
	cfg := &config.Config{
		DefaultProgram:     "echo test",
		AutoYes:            false,
		DaemonPollInterval: 1000,
		BranchPrefix:       "test/",
	}
	
	appState := &MockStateManager{}
	
	engine, err := New(cfg, appState)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()
	
	ctx := context.Background()
	
	// Test starting
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Failed to start engine: %v", err)
	}
	
	if !engine.IsStarted() {
		t.Fatal("Engine should be started")
	}
	
	// Test that we can't start twice
	if err := engine.Start(ctx); err == nil {
		t.Fatal("Should not be able to start engine twice")
	}
	
	// Test stopping
	if err := engine.Close(); err != nil {
		t.Fatalf("Failed to close engine: %v", err)
	}
	
	if engine.IsStarted() {
		t.Fatal("Engine should not be started after close")
	}
}

func TestEngineBasicOperations(t *testing.T) {
	cfg := &config.Config{
		DefaultProgram:     "echo test",
		AutoYes:            false,
		DaemonPollInterval: 1000,
		BranchPrefix:       "test/",
	}
	
	appState := &MockStateManager{}
	
	engine, err := New(cfg, appState)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()
	
	ctx := context.Background()
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Failed to start engine: %v", err)
	}
	
	// Test listing empty sessions
	sessions := engine.List()
	if len(sessions) != 0 {
		t.Fatalf("Expected 0 sessions, got %d", len(sessions))
	}
	
	// Test events subscription
	eventCh, err := engine.Events("")
	if err != nil {
		t.Fatalf("Failed to subscribe to events: %v", err)
	}
	
	if eventCh == nil {
		t.Fatal("Event channel should not be nil")
	}
	
	// Test config operations
	newCfg := &config.Config{
		DefaultProgram:     "new program",
		AutoYes:            true,
		DaemonPollInterval: 2000,
		BranchPrefix:       "new/",
	}
	
	if err := engine.UpdateConfig(newCfg); err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}
	
	retrievedCfg := engine.GetConfig()
	if retrievedCfg.DefaultProgram != "new program" {
		t.Fatalf("Expected default program 'new program', got '%s'", retrievedCfg.DefaultProgram)
	}
}

func TestEngineErrorHandling(t *testing.T) {
	// Test nil config
	_, err := New(nil, &MockStateManager{})
	if err == nil {
		t.Fatal("Should fail with nil config")
	}
	
	// Test nil app state
	cfg := &config.Config{DefaultProgram: "test"}
	_, err = New(cfg, nil)
	if err == nil {
		t.Fatal("Should fail with nil app state")
	}
	
	// Test operations on non-started engine
	engine, err := New(cfg, &MockStateManager{})
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()
	
	ctx := context.Background()
	
	// These should fail when engine is not started
	_, err = engine.StartSession(ctx, SessionOpts{Title: "test"})
	if err == nil {
		t.Fatal("StartSession should fail when engine not started")
	}
	
	err = engine.Pause("test")
	if err == nil {
		t.Fatal("Pause should fail when engine not started")
	}
	
	err = engine.Resume("test")
	if err == nil {
		t.Fatal("Resume should fail when engine not started")
	}
	
	err = engine.Kill("test")
	if err == nil {
		t.Fatal("Kill should fail when engine not started")
	}
	
	_, err = engine.Events("")
	if err == nil {
		t.Fatal("Events should fail when engine not started")
	}
}

func TestEventBus(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()
	
	// Test subscription
	ch1 := bus.Subscribe("session1")
	ch2 := bus.Subscribe("") // Global subscription
	
	// Test publishing
	event := createEvent("session1", EventStdout, StdoutEvent{Content: "test"})
	bus.Publish(event)
	
	// Check that both subscribers receive the event
	select {
	case receivedEvent := <-ch1:
		if receivedEvent.SessionID != "session1" {
			t.Fatalf("Expected session1, got %s", receivedEvent.SessionID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Session1 subscriber should have received event")
	}
	
	select {
	case receivedEvent := <-ch2:
		if receivedEvent.SessionID != "session1" {
			t.Fatalf("Expected session1, got %s", receivedEvent.SessionID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Global subscriber should have received event")
	}
	
	// Test closing
	bus.Close()
	
	// Channels should be closed
	select {
	case _, ok := <-ch1:
		if ok {
			t.Fatal("Channel should be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Channel should be closed immediately")
	}
}
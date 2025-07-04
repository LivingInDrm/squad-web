package engine

import (
	"sync"
	"time"
)

// EventBus manages event distribution to subscribers
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan Event // sessionID -> channels
	globalSubs  []chan Event            // global subscribers (all sessions)
	closed      bool
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan Event),
		globalSubs:  make([]chan Event, 0),
	}
}

// Subscribe creates a subscription channel for events from a specific session.
// If sessionID is empty, subscribes to all sessions.
func (eb *EventBus) Subscribe(sessionID string) <-chan Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	
	if eb.closed {
		// Return a closed channel
		ch := make(chan Event)
		close(ch)
		return ch
	}
	
	ch := make(chan Event, 100) // Buffered channel to prevent blocking
	
	if sessionID == "" {
		eb.globalSubs = append(eb.globalSubs, ch)
	} else {
		eb.subscribers[sessionID] = append(eb.subscribers[sessionID], ch)
	}
	
	return ch
}

// Publish sends an event to all relevant subscribers
func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	
	if eb.closed {
		return
	}
	
	// Send to global subscribers
	for _, ch := range eb.globalSubs {
		select {
		case ch <- event:
		default:
			// Channel is full, skip to prevent blocking
		}
	}
	
	// Send to session-specific subscribers
	if subs, exists := eb.subscribers[event.SessionID]; exists {
		for _, ch := range subs {
			select {
			case ch <- event:
			default:
				// Channel is full, skip to prevent blocking
			}
		}
	}
}

// Close shuts down the event bus and closes all subscriber channels
func (eb *EventBus) Close() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	
	if eb.closed {
		return
	}
	
	eb.closed = true
	
	// Close all global subscriber channels
	for _, ch := range eb.globalSubs {
		close(ch)
	}
	
	// Close all session-specific subscriber channels
	for _, subs := range eb.subscribers {
		for _, ch := range subs {
			close(ch)
		}
	}
	
	eb.globalSubs = nil
	eb.subscribers = nil
}

// createEvent is a helper to create properly formatted events
func createEvent(sessionID string, kind EventKind, payload interface{}) Event {
	return Event{
		SessionID: sessionID,
		Kind:      kind,
		Payload:   payload,
		Timestamp: time.Now(),
	}
}
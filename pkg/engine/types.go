package engine

import (
	"claude-squad/session"
	"claude-squad/session/git"
	"time"
)

// SessionOpts contains options for creating a new session
type SessionOpts struct {
	Title   string
	Path    string
	Program string
	AutoYes bool
	Prompt  string
}

// SessionInfo contains information about a session returned by the Engine API
type SessionInfo struct {
	ID        string       `json:"id"`
	Title     string       `json:"title"`
	Path      string       `json:"path"`
	Branch    string       `json:"branch"`
	Status    Status       `json:"status"`
	Program   string       `json:"program"`
	AutoYes   bool         `json:"auto_yes"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	DiffStats *DiffStats   `json:"diff_stats,omitempty"`
}

// Status represents the status of a session
type Status string

const (
	StatusRunning Status = "running"
	StatusReady   Status = "ready"
	StatusLoading Status = "loading"
	StatusPaused  Status = "paused"
)

// DiffStats contains git diff statistics
type DiffStats struct {
	Added   int    `json:"added"`
	Removed int    `json:"removed"`
	Content string `json:"content"`
}

// EventKind represents the type of event
type EventKind string

const (
	EventStdout EventKind = "stdout"
	EventStderr EventKind = "stderr"
	EventDiff   EventKind = "diff"
	EventState  EventKind = "state"
)

// Event represents a session event
type Event struct {
	SessionID string      `json:"session_id"`
	Kind      EventKind   `json:"kind"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

// StateEvent represents a state change event
type StateEvent struct {
	Previous Status `json:"previous"`
	Current  Status `json:"current"`
}

// DiffEvent represents a diff change event
type DiffEvent struct {
	Stats      *DiffStats `json:"stats"`
	HasChanges bool       `json:"has_changes"`
}

// StdoutEvent represents stdout/stderr output
type StdoutEvent struct {
	Content string `json:"content"`
}

// Convert session.Status to engine.Status
func convertStatus(s session.Status) Status {
	switch s {
	case session.Running:
		return StatusRunning
	case session.Ready:
		return StatusReady
	case session.Loading:
		return StatusLoading
	case session.Paused:
		return StatusPaused
	default:
		return StatusReady
	}
}

// Convert engine.Status to session.Status
func convertToSessionStatus(s Status) session.Status {
	switch s {
	case StatusRunning:
		return session.Running
	case StatusReady:
		return session.Ready
	case StatusLoading:
		return session.Loading
	case StatusPaused:
		return session.Paused
	default:
		return session.Ready
	}
}

// Convert git.DiffStats to engine.DiffStats
func convertDiffStats(stats *git.DiffStats) *DiffStats {
	if stats == nil {
		return nil
	}
	return &DiffStats{
		Added:   stats.Added,
		Removed: stats.Removed,
		Content: stats.Content,
	}
}
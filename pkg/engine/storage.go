package engine

import (
	"claude-squad/config"
	"claude-squad/session"
	"fmt"
	"time"
)

// StorageInterface defines the contract for session storage
type StorageInterface interface {
	LoadSessions() ([]SessionData, error)
	SaveSessions(sessions []SessionData) error
	LoadConfig() (*config.Config, error)
	SaveConfig(cfg *config.Config) error
}

// SessionData represents the persistent data for a session
type SessionData struct {
	ID        string                   `json:"id"`
	Title     string                   `json:"title"`
	Path      string                   `json:"path"`
	Branch    string                   `json:"branch"`
	Status    Status                   `json:"status"`
	Program   string                   `json:"program"`
	AutoYes   bool                     `json:"auto_yes"`
	CreatedAt string                   `json:"created_at"` // ISO 8601 format
	UpdatedAt string                   `json:"updated_at"` // ISO 8601 format
	Worktree  session.GitWorktreeData  `json:"worktree"`
	DiffStats session.DiffStatsData    `json:"diff_stats"`
}

// fileStorage implements StorageInterface using the existing config/state system
type fileStorage struct {
	sessionStorage *session.Storage
	configPath     string
}

// NewFileStorage creates a new file-based storage implementation
func NewFileStorage(appState config.StateManager) (StorageInterface, error) {
	sessionStorage, err := session.NewStorage(appState)
	if err != nil {
		return nil, fmt.Errorf("failed to create session storage: %w", err)
	}
	
	return &fileStorage{
		sessionStorage: sessionStorage,
	}, nil
}

// LoadSessions loads all sessions from storage
func (fs *fileStorage) LoadSessions() ([]SessionData, error) {
	instances, err := fs.sessionStorage.LoadInstances()
	if err != nil {
		return nil, fmt.Errorf("failed to load instances: %w", err)
	}
	
	sessions := make([]SessionData, len(instances))
	for i, instance := range instances {
		data := instance.ToInstanceData()
		sessions[i] = SessionData{
			ID:        generateSessionID(data.Title), // Generate consistent ID from title
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
	}
	
	return sessions, nil
}

// SaveSessions saves all sessions to storage
func (fs *fileStorage) SaveSessions(sessions []SessionData) error {
	instances := make([]*session.Instance, len(sessions))
	
	for i, sessionData := range sessions {
		// Convert SessionData back to session.InstanceData
		instanceData := session.InstanceData{
			Title:     sessionData.Title,
			Path:      sessionData.Path,
			Branch:    sessionData.Branch,
			Status:    convertToSessionStatus(sessionData.Status),
			Program:   sessionData.Program,
			AutoYes:   sessionData.AutoYes,
			Worktree:  sessionData.Worktree,
			DiffStats: sessionData.DiffStats,
		}
		
		// Parse timestamps
		createdAt, err := parseTime(sessionData.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to parse created_at for session %s: %w", sessionData.ID, err)
		}
		instanceData.CreatedAt = createdAt
		
		updatedAt, err := parseTime(sessionData.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to parse updated_at for session %s: %w", sessionData.ID, err)
		}
		instanceData.UpdatedAt = updatedAt
		
		// Convert back to Instance
		instance, err := session.FromInstanceData(instanceData)
		if err != nil {
			return fmt.Errorf("failed to create instance from data for session %s: %w", sessionData.ID, err)
		}
		
		instances[i] = instance
	}
	
	return fs.sessionStorage.SaveInstances(instances)
}

// LoadConfig loads the application configuration
func (fs *fileStorage) LoadConfig() (*config.Config, error) {
	cfg := config.LoadConfig()
	return cfg, nil
}

// SaveConfig saves the application configuration
func (fs *fileStorage) SaveConfig(cfg *config.Config) error {
	return config.SaveConfig(cfg)
}

// generateSessionID creates a consistent session ID from the title
// This ensures backward compatibility with existing title-based lookups
func generateSessionID(title string) string {
	// For now, use the title as the ID to maintain compatibility
	// In the future, this could use proper UUIDs
	return title
}

// parseTime parses an ISO 8601 timestamp
func parseTime(timeStr string) (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05Z07:00", timeStr)
}
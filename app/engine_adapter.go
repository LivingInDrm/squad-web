package app

import (
	"claude-squad/pkg/engine"
	"context"
)

// RunWithEngine is a wrapper that maintains the same API as Run() but uses the Engine SDK internally
func RunWithEngine(ctx context.Context, eng *engine.Engine, program string, autoYes bool) error {
	// This function serves as a bridge between the new Engine SDK and the existing TUI
	// For now, we'll still use the original Run function to maintain compatibility
	// but this provides the foundation for future TUI refactoring
	return Run(ctx, program, autoYes)
}

// EngineBackedRun will be the future implementation that uses Engine SDK directly
// This is implemented in the next phase of the migration
func EngineBackedRun(ctx context.Context, eng *engine.Engine, program string, autoYes bool) error {
	// TODO: Implement TUI that uses Engine SDK instead of direct session management
	// This will be done in a subsequent phase to maintain compatibility
	
	// For now, return the standard Run
	return Run(ctx, program, autoYes)
}
// Package main demonstrates basic usage of the claude-squad Engine SDK
package main

import (
	"claude-squad/config"
	"claude-squad/pkg/engine"
	"context"
	"fmt"
	"log"
	"time"
)

func main() {
	// Load configuration
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
	
	fmt.Println("Engine SDK Demo")
	fmt.Println("===============")
	
	// List existing sessions
	sessions := eng.List()
	fmt.Printf("Found %d existing sessions:\n", len(sessions))
	for _, session := range sessions {
		fmt.Printf("  - %s (%s): %s\n", session.Title, session.ID, session.Status)
	}
	
	// Create a new session
	fmt.Println("\nCreating new session...")
	sessionID, err := eng.StartSession(ctx, engine.SessionOpts{
		Title:   "sdk-demo-session",
		Path:    ".",
		Program: "echo 'Hello from SDK!'",
		AutoYes: false,
		Prompt:  "Hello, Claude!",
	})
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	
	fmt.Printf("Created session: %s\n", sessionID)
	
	// Subscribe to events
	eventCh, err := eng.Events(sessionID)
	if err != nil {
		log.Fatalf("Failed to subscribe to events: %v", err)
	}
	
	// Watch events for a few seconds
	fmt.Println("\nWatching events for 5 seconds...")
	timeout := time.After(5 * time.Second)
	eventCount := 0
	
	for {
		select {
		case event, ok := <-eventCh:
			if !ok {
				fmt.Println("Event channel closed")
				goto cleanup
			}
			eventCount++
			fmt.Printf("Event %d: %s - %s\n", eventCount, event.Kind, event.Timestamp.Format("15:04:05"))
			
		case <-timeout:
			fmt.Printf("Received %d events\n", eventCount)
			goto cleanup
		}
	}
	
cleanup:
	// Get session info
	info, err := eng.Get(sessionID)
	if err != nil {
		log.Printf("Failed to get session info: %v", err)
	} else {
		fmt.Printf("\nSession info: %s on branch %s\n", info.Status, info.Branch)
		if info.DiffStats != nil {
			fmt.Printf("Diff stats: +%d -%d lines\n", info.DiffStats.Added, info.DiffStats.Removed)
		}
	}
	
	// Pause the session
	fmt.Println("\nPausing session...")
	if err := eng.Pause(sessionID); err != nil {
		log.Printf("Failed to pause session: %v", err)
	} else {
		fmt.Println("Session paused successfully")
	}
	
	// Wait a moment and resume
	time.Sleep(1 * time.Second)
	fmt.Println("Resuming session...")
	if err := eng.Resume(sessionID); err != nil {
		log.Printf("Failed to resume session: %v", err)
	} else {
		fmt.Println("Session resumed successfully")
	}
	
	// Clean up - kill the demo session
	fmt.Println("\nCleaning up demo session...")
	if err := eng.Kill(sessionID); err != nil {
		log.Printf("Failed to kill session: %v", err)
	} else {
		fmt.Println("Demo session cleaned up")
	}
	
	fmt.Println("\nSDK Demo completed successfully!")
}
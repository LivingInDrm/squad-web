# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

> **Instruction to Claude:**  
> You must execute tasks *strictly* according to the following four-stage workflow.  
> Do **not** skip, merge, or reorder stages unless explicitly told to do so.

# Feature-Upgrade Workflow — Four Stages

1. **Understand & Draft High-Level Plan**  
   Read the requirement and design specs, clarify ambiguities, identify impacted modules/files, and produce `overall_plan.md`.

2. **Analyze Code & Write Detailed Plan**  
   Inspect the referenced source code, trace logic, edge cases, and dependencies, then craft concrete steps in `detailed_plan.md`.

3. **Implement Changes**  
   Modify the code exactly as outlined, ensuring modularity, testability, clear documentation, and atomic commits.

4. **Self-Review & Test**  
   Verify implementation vs. plan, cover failure paths and edge cases, fix any issues, and revise plans if necessary.

   

## Project Overview

Claude Squad is a terminal application that manages multiple AI agents (Claude Code, Aider, Codex, Gemini) in separate workspaces using tmux sessions and git worktrees. The project is undergoing a major WebUI transformation from a CLI-only tool to a web-enabled SaaS platform.

## Essential Commands

### Building and Running
```bash
# Build the main binary
go build -o claude-squad main.go

# Run the application (must be in a git repository)
./claude-squad
# or use the short alias after installation
cs

# Run with specific program
cs -p "aider --model gpt-4"
cs -p "claude"
cs -p "gemini"

# Debug mode to see config paths
cs debug

# Reset all sessions
cs reset
```

### Testing
```bash
# Run all tests
go test -v ./...

# Run tests for a specific package
go test -v ./session/...
go test -v ./app/...

# Run a single test
go test -v -run TestSpecificFunction ./package/...
```

### Code Quality
```bash
# Format code (required before commits)
gofmt -w .

# Run linter (uses golangci-lint)
golangci-lint run --timeout=3m

# Check formatting without modifying
gofmt -l .
```

## Architecture Overview

### Core Components

1. **Session Management (`session/`)**
   - `instance.go`: Core session abstraction representing an AI agent workspace
   - `storage.go`: Persistence layer for session state (currently using JSON files)
   - Each session has: tmux session + git worktree + AI agent process

2. **Terminal Multiplexing (`session/tmux/`)**
   - Manages tmux sessions for each AI agent
   - Handles PTY (pseudo-terminal) creation and management
   - Platform-specific implementations (Unix vs Windows)

3. **Git Integration (`session/git/`)**
   - Creates and manages git worktrees for workspace isolation
   - Handles branch operations, commits, and diffs
   - Each session gets its own branch with prefix (default: `cs/`)

4. **TUI Application (`app/`)**
   - Built with Bubbletea (Elm-inspired TUI framework)
   - `app.go`: Main UI logic and state management
   - Three-pane layout: session list, preview/diff view, metadata

5. **Configuration (`config/`)**
   - Config file location: `~/.claude-squad/config.json`
   - State persistence: `~/.claude-squad/state.json`
   - Supports default program, auto-yes mode, branch prefix

6. **Daemon Mode (`daemon/`)**
   - Background process for auto-yes functionality
   - Polls sessions and automatically accepts prompts

### Key Design Patterns

1. **Event-Driven Updates**: The TUI polls sessions for changes rather than using direct callbacks
2. **Workspace Isolation**: Each session runs in its own git worktree to prevent conflicts
3. **State Machine**: Sessions have clear states (Running, Ready, Loading, Paused)
4. **Platform Abstraction**: Separate implementations for Unix/Windows terminal handling

## WebUI Transformation (In Progress)

The project is undergoing a three-phase transformation:

1. **M0 - Engine SDK**: Extract core logic into reusable SDK (see `m0_enginesdk_design.md`)
2. **M1 - Web GUI**: Add local web interface while maintaining CLI (see `m1_gui_design.md`)
3. **M2 - SaaS Platform**: Multi-user support with containerization (see `m2_saas_design.md`)

Key files for understanding the transformation:
- `webui-refact.md`: Overall transformation plan
- `m*_user_stories.md`: User stories for each milestone

## Important Implementation Notes

### Error Handling
- The codebase uses panic in several places (e.g., `tmux.go:330`) - these need to be replaced with proper error handling for web service compatibility
- Log errors are written to `/tmp/claudesquad.log`

### Security Considerations
- Command construction for git and tmux needs validation to prevent injection
- File system access is currently unrestricted - needs sandboxing for multi-user mode
- No authentication/authorization in current design

### Performance Considerations
- Session polling interval affects UI responsiveness
- Large diff calculations can be expensive
- Daemon mode polls every 1 second by default

### Known Technical Debt
- `app.go` is large (735 lines) and handles multiple responsibilities
- Missing comprehensive test coverage (estimate ~15-20%)
- TODO comments indicate areas needing optimization (memory allocation, control flow)

## Development Prerequisites

- **tmux**: Required for terminal multiplexing
- **git**: Required for version control operations
- **gh CLI**: Required for GitHub operations (push, PR creation)
- **Go 1.23+**: Required for building

## Configuration

Default config location: `~/.claude-squad/config.json`

Example:
```json
{
  "default_program": "claude",
  "auto_yes": false,
  "daemon_poll_interval": 1000,
  "branch_prefix": "cs/"
}
```

## Testing Approach

1. Unit tests exist for core packages but coverage is limited
2. No integration or E2E tests currently
3. Manual testing required for TUI interactions
4. CI runs tests on push to main branch
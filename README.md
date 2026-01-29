# ralph

CLI tool for managing AI-powered development loops. A standalone alternative to ralph-tui.

## Features

- 🌳 Create and manage git worktrees for features
- 📋 Define PRDs (Product Requirement Documents) with user stories
- 🤖 Run AI agents (Claude CLI) to implement features autonomously
- 📊 Monitor progress across multiple loops
- 🧹 Clean up completed features (worktrees, databases)

## Installation

```bash
go install github.com/hyperlab-be/ralph@latest
```

Or build from source:

```bash
git clone https://github.com/hyperlab-be/ralph.git
cd ralph
go build -o ralph .
```

## Usage

```bash
# Initialize a new project for ralph
ralph init

# Create a new feature with worktree + database
ralph new my-feature

# Create/edit a PRD interactively
ralph prd

# Start the AI loop
ralph run

# Check status of all loops
ralph status

# View logs of a running loop
ralph logs my-feature

# Interactive dashboard
ralph dashboard

# Clean up a completed feature
ralph cleanup my-feature
```

## Commands

| Command | Description |
|---------|-------------|
| `init` | Initialize ralph in current project |
| `new` | Create a new feature (worktree + database) |
| `prd` | Create or edit PRD for current worktree |
| `run` | Start the AI development loop |
| `stop` | Stop a running loop |
| `status` | Show status of all loops |
| `logs` | Tail logs of a loop |
| `list` | List all available loops |
| `dashboard` | Interactive TUI dashboard |
| `cleanup` | Remove worktree and database |

## Requirements

- Go 1.21+
- Git
- [Claude CLI](https://github.com/anthropics/claude-code) (`npm install -g @anthropic-ai/claude-code`)
- MySQL (optional, for database provisioning)

## License

MIT

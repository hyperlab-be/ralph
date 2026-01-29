# ralph

CLI tool for managing AI-powered development loops with [ralph-tui](https://github.com/m-ods/ralph-tui).

## Features

- 🌳 Create and manage git worktrees for features
- 📋 Define PRDs (Product Requirement Documents) with user stories
- 🤖 Run AI agents to implement features autonomously
- 📊 Monitor progress across multiple loops
- 🧹 Clean up completed features (worktrees, databases)

## Installation

```bash
go install github.com/hyperlab/ralph@latest
```

Or build from source:

```bash
git clone https://github.com/hyperlab/ralph.git
cd ralph
go build -o rl .
```

## Usage

```bash
# Initialize a new project for rl
rl init

# Create a new feature with worktree + database
rl new my-feature

# Create/edit a PRD interactively
rl prd

# Start the AI loop
rl run

# Check status of all loops
rl status

# View logs of a running loop
rl logs my-feature

# Interactive dashboard
rl dashboard

# Clean up a completed feature
rl cleanup my-feature
```

## Commands

| Command | Description |
|---------|-------------|
| `init` | Initialize rl in current project |
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
- [ralph-tui](https://github.com/m-ods/ralph-tui)
- MySQL (for database provisioning)

## License

MIT

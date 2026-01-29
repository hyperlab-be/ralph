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

## Commands

### `ralph init`

Initialize ralph in a project directory.

```bash
$ ralph init
✓ Initialized ralph in /Users/dev/myproject
ℹ Edit ralph.toml to configure hooks and settings
```

Creates a `ralph.toml` config file and `.ralph/` directory.

---

### `ralph new`

Create a new feature with git worktree and optional database.

```bash
$ ralph new user-auth
✓ Created worktree at ../myproject-user-auth
✓ Created branch feature/user-auth
✓ Registered loop: myproject-user-auth
ℹ Next: Create a PRD with 'ralph prd create' then start with 'ralph run'
```

---

### `ralph prd`

View, create, or edit the PRD (Product Requirement Document).

```bash
$ ralph prd
📋 PRD: User Authentication

Stories:
  ✓ 1. Login page with email/password
  ✓ 2. Password reset flow
  ⚫ 3. OAuth integration (Google)
  ⚫ 4. Session management

Progress: 2/4 stories complete
```

Create a new PRD interactively:

```bash
$ ralph prd create
? Feature name: User Authentication
? Description: Add user authentication with multiple providers
? Add a story: Login page with email/password
? Acceptance criteria: - Form validates email format
? Acceptance criteria: - Shows error on invalid credentials
? Acceptance criteria: (empty to finish)
? Add another story? Yes
...
✓ Created PRD with 4 stories
```

Add a single story:

```bash
$ ralph prd add "OAuth integration" --criteria "Support Google login" --criteria "Support GitHub login"
✓ Added story: OAuth integration
```

---

### `ralph run`

Start the AI agent loop to implement stories.

```bash
$ ralph run
ℹ Starting agent loop for myproject-user-auth
ℹ Model: claude-sonnet-4-20250514 | Max iterations: 10

Iteration 1/10: Story 3 - OAuth integration (Google)
[Agent working...]
✓ Story 3 completed!

Iteration 2/10: Story 4 - Session management
[Agent working...]
✓ Story 4 completed!

✓ All stories complete!
ℹ Final progress: 4/4 stories
```

Options:

```bash
$ ralph run --max-iterations 5    # Limit iterations
$ ralph run --dry-run             # Preview without executing
```

---

### `ralph status`

Show status of all loops or a specific loop.

```bash
$ ralph status
╔═══════════════════════════════════════════════════════════╗
║                 🤖 ralph - Loop Status                    ║
╚═══════════════════════════════════════════════════════════╝

🟢 myproject-user-auth
   Status: running
   Progress: 2/4 stories
   Path: /Users/dev/myproject-user-auth

⚫ myproject-api-v2
   Status: stopped
   Progress: 5/5 stories
   Path: /Users/dev/myproject-api-v2
```

---

### `ralph logs`

View logs of a running or completed loop.

```bash
$ ralph logs myproject-user-auth
=== Session started 2024-01-15T10:30:00Z ===
[10:30:05] Iteration 1: Login page with email/password
[10:32:15] Story 1 completed
[10:32:20] Iteration 2: Password reset flow
...
```

Follow logs in real-time:

```bash
$ ralph logs -f myproject-user-auth
```

---

### `ralph list`

List all registered loops.

```bash
$ ralph list
🟢 myproject-user-auth
⚫ myproject-api-v2
⚫ other-project-feature
```

---

### `ralph dashboard`

Interactive dashboard with auto-refresh.

```bash
$ ralph dashboard
╔═══════════════════════════════════════════════════════════╗
║              🤖 ralph - Live Dashboard                    ║
╚═══════════════════════════════════════════════════════════╝

🟢 myproject-user-auth
   Status: running
   Progress: 3/4 stories
   Path: /Users/dev/myproject-user-auth

[Refreshing every 5s - Press Ctrl+C to exit]
```

---

### `ralph stop`

Stop a running loop.

```bash
$ ralph stop myproject-user-auth
✓ Stopped loop: myproject-user-auth
```

---

### `ralph cleanup`

Remove a worktree and clean up resources.

```bash
$ ralph cleanup myproject-user-auth
⚠ This will remove the worktree at /Users/dev/myproject-user-auth
? Continue? Yes
✓ Ran cleanup hooks
✓ Removed worktree
✓ Unregistered loop
```

With branch deletion:

```bash
$ ralph cleanup myproject-user-auth --delete-branch
✓ Removed worktree
✓ Deleted branch feature/user-auth
```

---

## Configuration

### Project config (`ralph.toml`)

```toml
[project]
name = "myproject"

[worktree]
prefix = "myproject"  # Worktree naming: {prefix}-{feature}

[hooks]
setup = "./scripts/setup-worktree.sh"      # Run after creating worktree
cleanup = "./scripts/cleanup-worktree.sh"  # Run before removing worktree

[agent]
model = "claude-sonnet-4-20250514"
max_iterations = 10
```

### Global config (`~/.config/ralph/config.toml`)

```toml
[defaults]
model = "claude-sonnet-4-20250514"
max_iterations = 10
projects_dir = "~/Code"
```

## Requirements

- Go 1.21+
- Git
- [Claude CLI](https://github.com/anthropics/claude-code) (`npm install -g @anthropic-ai/claude-code`)
- MySQL (optional, for database provisioning via hooks)

## License

MIT

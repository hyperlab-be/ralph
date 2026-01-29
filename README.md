# ralph

CLI tool for managing AI-powered development loops. Based on the [Ralph Wiggum](https://www.aihero.dev/tips-for-ai-coding-with-ralph-wiggum) pattern for autonomous AI coding.

## Features

- 🌳 Create and manage git worktrees for features
- 📋 Define PRDs (Product Requirement Documents) with user stories
- 🤖 Run AI agents (Claude CLI) to implement features autonomously
- 📝 Full conversation logging for debugging
- 🔄 Agent chooses highest priority task (not just first in list)
- 📊 Monitor progress across multiple loops
- 🧹 Clean up completed features (worktrees, databases)

## How It Works

Ralph runs Claude CLI in a loop, letting it work autonomously through your PRD:

```
┌─────────────────────────────────────────────────────────┐
│  PRD (prd.json)                                         │
│  ├── Story 1: Login page           ✅ passes: true      │
│  ├── Story 2: Password reset       ⬜ passes: false     │
│  └── Story 3: OAuth                ⬜ passes: false     │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  RALPH LOOP (per iteration)                             │
│                                                         │
│  1. Read prd.json + progress.txt                        │
│  2. Agent chooses highest priority incomplete story     │
│  3. Implements, runs tests, commits                     │
│  4. Sets passes: true in prd.json                       │
│  5. Logs everything to .ralph/conversations/            │
│  6. Repeat until all stories complete                   │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
              ┌───────────────────────┐
              │  All done? → Create PR │
              └───────────────────────┘
```

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

## Quick Start

```bash
cd ~/Code/myproject
ralph init                    # Initialize ralph
ralph prd --new               # Create PRD interactively
ralph prd "Add login page"    # Add a story
ralph run                     # Start the loop
ralph run --once              # Single iteration (HITL mode)
```

## Two Modes of Operation

### HITL (Human-in-the-Loop)

Watch the agent work, intervene when needed:

```bash
ralph run --once              # Single iteration
ralph run -m 3                # Few iterations, stay close
```

Best for: Learning, prompt refinement, risky architectural work.

### AFK (Away From Keyboard)

Let ralph run autonomously:

```bash
ralph run                     # Default: 10 iterations
ralph run -m 20               # 20 iterations max
```

Best for: Bulk work, well-defined tasks, overnight runs.

## Commands

### `ralph init`

Initialize ralph in a project directory.

```bash
$ ralph init
✓ Initialized ralph in /Users/dev/myproject
ℹ Edit ralph.toml to configure hooks and settings
```

---

### `ralph new <feature>`

Create a new feature with git worktree.

```bash
$ ralph new user-auth
✓ Created worktree at ../myproject-user-auth
✓ Created branch feature/user-auth
ℹ Next: Create a PRD with 'ralph prd --new' then start with 'ralph run'
```

---

### `ralph prd`

View, create, or edit the PRD.

```bash
# Show PRD status
$ ralph prd
PRD: User Authentication

[ ] 1. Login page with email/password
[ ] 2. Password reset flow
[✓] 3. OAuth integration (Google)

Progress: 1/3 (33%)

# Create new PRD interactively
$ ralph prd --new

# Add a story
$ ralph prd "Session management" -c "Sessions expire after 24h" -c "Refresh tokens supported"
✓ Added story 4: Session management

# Edit PRD in $EDITOR
$ ralph prd --edit
```

---

### `ralph run`

Start the AI agent loop.

```bash
$ ralph run
ℹ Starting agent loop for myproject-user-auth
ℹ Model: opus | Max iterations: 10

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Iteration 1/10
Progress: 1/4
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

[Agent working...]
✓ Story completed!

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Final progress: 4/4
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✓ All stories complete! Creating pull request...
```

**Options:**

| Flag | Description |
|------|-------------|
| `--model` | Model to use (default: opus) |
| `--once` | Single iteration (HITL mode) |
| `-m, --max-iterations` | Maximum iterations (default: 10) |
| `--dry-run` | Preview without executing |

When all stories are complete, ralph automatically creates a pull request.

---

### `ralph status`

Show status of all loops.

```bash
$ ralph status
╔═══════════════════════════════════════════════════════════╗
║                 🤖 ralph - Loop Status                    ║
╚═══════════════════════════════════════════════════════════╝

🟢 myproject-user-auth
   Status: running
   Progress: 2/4 stories
   Path: /Users/dev/myproject-user-auth
```

---

### `ralph logs`

View logs.

```bash
$ ralph logs myproject-user-auth      # Progress summary
$ ralph logs -f myproject-user-auth   # Follow output in real-time
$ ralph logs --session                # Technical session log
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

Remove a worktree and clean up.

```bash
$ ralph cleanup myproject-user-auth
✓ Removed worktree
✓ Unregistered loop
```

---

### `ralph doctor`

Check dependencies.

```bash
$ ralph doctor
✓ git: git version 2.39.0
✓ claude: Claude CLI installed
✓ gh: gh version 2.40.0
```

---

## Configuration

### Project config (`ralph.toml`)

```toml
[project]
name = "myproject"

[worktree]
prefix = "myproject"

[hooks]
setup = "./scripts/setup-worktree.sh"
cleanup = "./scripts/cleanup-worktree.sh"
```

### Global config (`~/.config/ralph/config.toml`)

```toml
[defaults]
projects_dir = "~/Code"
```

## PRD Format

```json
{
  "name": "Feature Name",
  "description": "What we're building",
  "userStories": [
    {
      "id": "1",
      "title": "Story title",
      "description": "Detailed description",
      "acceptanceCriteria": [
        "Criterion 1",
        "Criterion 2"
      ],
      "passes": false
    }
  ]
}
```

The agent sets `passes: true` when a story is complete.

## Files

```
myproject/
├── ralph.toml              # Project config
└── .ralph/
    ├── prd.json            # PRD with stories
    ├── progress.txt        # Progress tracking between iterations
    ├── session.log         # Session summary
    └── output.log          # Live output (for ralph logs -f)
```

## Requirements

- Go 1.21+
- Git
- [Claude CLI](https://docs.anthropic.com/en/docs/claude-code) 
- [GitHub CLI](https://cli.github.com) (optional, for auto PR creation)

## Tips

1. **Start with HITL** - Learn how the loop works before going AFK
2. **Small stories** - Smaller = better results
3. **Explicit acceptance criteria** - Prevents shortcuts
4. **Review conversation logs** - Debug via `.ralph/conversations/`

## License

MIT

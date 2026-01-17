# Claude Code CLI Integration

The SDLC Agent uses **Claude Code CLI** installed on your local machine for AI-powered code generation.

## How It Works

When a JIRA issue is moved to "In Development":

1. System clones the repository
2. Analyzes the code structure
3. Builds a prompt from the JIRA issue
4. **Executes Claude CLI** with the prompt
5. Claude generates code directly in the repository
6. System commits and pushes changes
7. Creates a Pull Request

## Architecture

```
┌────────────────────────────────────┐
│ Developer Agent (Docker)           │
│ ┌──────────────────────────────┐   │
│ │ 1. Build prompt from JIRA    │   │
│ │ 2. Execute: claude --message │   │
│ │    Working dir: cloned repo  │   │
│ └──────────────────────────────┘   │
└──────────────┬─────────────────────┘
               │ exec.Command("claude", "--message", prompt)
               ▼
┌────────────────────────────────────┐
│ Claude Code CLI (Mounted)          │
│ /app/claude                         │
│ ┌──────────────────────────────┐   │
│ │ 1. Uses your local session   │   │
│ │ 2. Generates code in repo    │   │
│ │ 3. Returns output            │   │
│ └──────────────────────────────┘   │
└──────────────┬─────────────────────┘
               │ stdout/stderr
               ▼
┌────────────────────────────────────┐
│ Developer Agent                    │
│ ┌──────────────────────────────┐   │
│ │ 1. Capture output            │   │
│ │ 2. Count changed files       │   │
│ │ 3. Create branch & commit    │   │
│ │ 4. Push & create PR          │   │
│ └──────────────────────────────┘   │
└────────────────────────────────────┘
```

## Prerequisites

### 1. Claude Code CLI Installed

```bash
# Check if Claude is installed
which claude
# Should output: /Users/yourusername/.local/bin/claude
```

If not installed, install Claude Code from: https://claude.ai/

### 2. Active Claude Session

```bash
# Verify Claude is authenticated
claude --help
```

If you see an authentication prompt, log in to Claude.

## Configuration

### 1. Find Your Claude Installation

```bash
which claude
# Example output: /Users/storos/.local/bin/claude
```

### 2. Update docker-compose.yml (if needed)

The default configuration mounts from `~/.local/bin/claude`:

```yaml
volumes:
  - ~/.local/bin/claude:/app/claude:ro
```

If your Claude is in a different location, update the path:

```yaml
volumes:
  - /path/to/your/claude:/app/claude:ro
```

### 3. Verify Configuration

Check `.env`:
```bash
CLAUDE_CLI_PATH=/app/claude
```

This points to the mounted location inside the container.

## Installation Steps

### Step 1: Verify Claude CLI

```bash
# On your host machine
which claude
claude --help
```

### Step 2: Update Docker Configuration (if needed)

If Claude is not at `~/.local/bin/claude`, edit `docker-compose.yml`:

```yaml
developer-agent-consumer:
  volumes:
    - /your/path/to/claude:/app/claude:ro
```

### Step 3: Rebuild and Start

```bash
docker-compose up -d --build developer-agent-consumer
```

### Step 4: Verify

Check logs:
```bash
docker logs sdlc-developer-agent-consumer | grep "Claude"
```

Expected output:
```json
{"level":"info","msg":"Using Claude Code CLI for code generation"}
```

Verify binary is mounted:
```bash
docker exec sdlc-developer-agent-consumer ls -la /app/claude
```

Should show:
```
-rwxr-xr-x 1 root root 179795312 Jan 16 07:16 /app/claude
```

## Testing

### 1. Send Test Webhook

```bash
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "webhookEvent": "jira:issue_updated",
    "issue": {
      "key": "TEST-999",
      "fields": {
        "summary": "Test Claude CLI Integration",
        "description": "Add a simple hello world function",
        "status": {"name": "In Development"},
        "project": {"key": "TEST"}
      }
    },
    "changelog": {
      "items": [{"field": "status", "toString": "In Development"}]
    }
  }'
```

### 2. Monitor Logs

```bash
docker logs -f sdlc-developer-agent-consumer
```

Look for:
```json
{"level":"info","msg":"Calling Claude Code CLI"}
{"level":"info","msg":"Executing Claude CLI command","command":"claude"}
{"level":"info","msg":"Claude CLI completed successfully"}
```

### 3. Check Database

```bash
docker exec sdlc-mongodb mongosh sdlc_agents --quiet --eval \
  'db.developments.findOne({jira_issue_key: "TEST-999"})'
```

Should show:
- `repository_url`: Set
- `branch_name`: `feature/TEST-999`
- `prompt`: Full prompt text
- `status`: `completed` or `failed`

## Troubleshooting

### Error: "executable file not found in $PATH"

**Cause:** Claude binary not mounted correctly

**Solution:**
```bash
# 1. Find Claude on host
which claude

# 2. Update docker-compose.yml with correct path
# 3. Restart
docker-compose restart developer-agent-consumer

# 4. Verify mount
docker exec sdlc-developer-agent-consumer ls -la /app/claude
```

### Error: "permission denied"

**Cause:** Binary not executable

**Solution:**
```bash
# On host:
chmod +x ~/.local/bin/claude

# Restart container:
docker-compose restart developer-agent-consumer
```

### Error: Claude session expired

**Cause:** Claude CLI needs re-authentication

**Solution:**
```bash
# On host machine, re-authenticate:
claude --help
# Follow login prompts
```

### Error: "No such file or directory"

**Cause:** Claude not installed or wrong path

**Solution:**
```bash
# Install Claude Code CLI
# Then update docker-compose.yml with correct path
```

## Benefits

- ✅ **No API Key Required** - Uses your local Claude session
- ✅ **No Internet Required** - Fully local after initial auth
- ✅ **No Costs** - Free unlimited usage
- ✅ **Better Privacy** - Code never leaves your machine
- ✅ **Same Quality** - Uses the same Claude you interact with

## How Prompts Are Built

The system creates prompts like:

```
# Development Task: TEST-123

## Summary
Add user authentication endpoint

## Description
Implement JWT-based authentication with login and refresh token endpoints.
Include password hashing, token validation, and middleware.

## Project Scope
Backend API development for user management system

## Repository Context
- Project Type: Node.js Application
- Languages: JavaScript, TypeScript
- Entry Points: src/index.js, src/server.js
- Key Directories: src/routes, src/controllers, src/middleware

## Instructions
Please implement the changes described above in the repository.
Make sure to:
1. Follow existing code patterns and conventions
2. Write clean, maintainable code
3. Add appropriate error handling
4. Update any relevant documentation
```

This prompt is then passed to:
```bash
claude --message "<prompt>"
```

Claude generates the code directly in the cloned repository.

## File Locations

- **Service**: `developer-agent-consumer/services/claude_service.go`
- **Configuration**: `.env` and `docker-compose.yml`
- **Binary Mount**: `~/.local/bin/claude` → `/app/claude` (inside container)

## Environment Variables

```bash
# Path to Claude CLI binary inside container
CLAUDE_CLI_PATH=/app/claude
```

This should match the mount point in `docker-compose.yml`.

## Limitations

- Requires Claude Code CLI installed on host
- Requires active Claude session
- Must mount binary into Docker container
- Claude CLI must be executable

## Advantages Over API Mode

| Feature | CLI Mode ✓ | API Mode (Removed) |
|---------|-----------|-------------------|
| API Key | Not needed | Required |
| Internet | Not needed | Required |
| Costs | Free | Pay per request |
| Rate Limits | None | Yes |
| Privacy | Local | Cloud |
| Setup | Mount binary | Configure key |

---

**Your system is configured to use Claude Code CLI exclusively for all code generation tasks.**

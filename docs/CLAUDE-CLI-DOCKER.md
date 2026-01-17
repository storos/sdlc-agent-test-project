# Claude Code CLI in Docker - Setup Guide

## Overview

The developer-agent-consumer now runs fully containerized with Claude Code CLI installed inside the Docker container, authenticated via Anthropic API Key.

## Architecture

**Before:**
- developer-agent-consumer ran natively on Mac
- Used Mac's Claude CLI binary (requires host installation)
- Authentication via Mac's Claude session

**After:**
- developer-agent-consumer runs in Docker container
- Claude CLI installed during Docker build
- Authentication via `ANTHROPIC_API_KEY` environment variable

## Setup Instructions

### 1. Get Your Anthropic API Key

1. Go to https://console.anthropic.com/
2. Navigate to **API Keys** section
3. Create a new API key
4. Copy the key (starts with `sk-ant-api03-...`)

### 2. Configure Environment Variables

Add your API key to `.env` file:

```bash
# Claude Code CLI Configuration
ANTHROPIC_API_KEY=sk-ant-api03-YOUR-KEY-HERE
CLAUDE_CLI_PATH=/root/.local/bin/claude
```

**Important:** The `.env` file is already in `.gitignore` - never commit your API key!

### 3. Build and Run

```bash
# Build the container (installs Claude CLI)
docker-compose build developer-agent-consumer

# Start the service
docker-compose up -d developer-agent-consumer

# Check logs
docker logs -f sdlc-developer-agent-consumer
```

## How It Works

### Dockerfile Changes

The Dockerfile now:
1. Installs dependencies (bash, curl, screen, ripgrep, etc.)
2. Downloads and installs Claude Code CLI using official script:
   ```bash
   curl -fsSL https://claude.ai/install.sh | bash
   ```
3. Installs Claude CLI to `/root/.local/bin/claude`
4. Adds Claude to PATH

### docker-compose.yml Changes

The service now uses:
- `ANTHROPIC_API_KEY` - For Claude CLI authentication
- `CLAUDE_CLI_PATH` - Path to Claude CLI binary in container
- No volume mounts needed for Claude CLI

### Authentication Flow

1. Container starts with `ANTHROPIC_API_KEY` environment variable
2. Claude CLI reads the API key from environment
3. All Claude CLI commands authenticate using the API key
4. Code generation works seamlessly in Docker

## Verification

Verify the setup:

```bash
# Check Claude CLI version
docker exec sdlc-developer-agent-consumer /root/.local/bin/claude --version

# Check API key is set
docker exec sdlc-developer-agent-consumer sh -c 'echo $ANTHROPIC_API_KEY | cut -c1-20'

# Check logs
docker logs sdlc-developer-agent-consumer --tail 20
```

Expected output:
```
{"level":"info","msg":"Starting Developer Agent Consumer"}
{"level":"info","msg":"Connected to MongoDB"}
{"level":"info","msg":"Using Claude Code CLI for code generation"}
{"level":"info","msg":"RabbitMQ consumer started, waiting for messages..."}
```

## Benefits

✅ **Fully Containerized** - No host dependencies
✅ **Portable** - Works on any Docker environment
✅ **Secure** - API key via environment variables
✅ **Consistent** - Same Claude CLI version across environments
✅ **Easy Updates** - Rebuild to get latest Claude CLI

## API Key Management

### Security Best Practices

1. **Never commit `.env` file** - Already in `.gitignore`
2. **Use separate keys** for dev/staging/prod
3. **Rotate keys regularly** via Anthropic Console
4. **Monitor usage** at https://console.anthropic.com/

### Cost Management

- Monitor API usage in Anthropic Console
- Set up usage alerts
- Claude CLI uses the same pricing as API calls
- Track costs per project/environment

## Troubleshooting

### Container fails to start

```bash
# Check logs
docker logs sdlc-developer-agent-consumer

# Check API key is set
docker exec sdlc-developer-agent-consumer env | grep ANTHROPIC
```

### Claude CLI not found

```bash
# Verify installation
docker exec sdlc-developer-agent-consumer ls -la /root/.local/bin/

# Rebuild container
docker-compose build --no-cache developer-agent-consumer
```

### Authentication errors

```bash
# Verify API key format (should start with sk-ant-api03-)
grep ANTHROPIC_API_KEY .env

# Test API key manually
docker exec sdlc-developer-agent-consumer /root/.local/bin/claude --version
```

## References

- [Claude Code CLI Documentation](https://code.claude.com/docs/en/setup)
- [Anthropic API Keys](https://console.anthropic.com/)
- [Claude CLI Installation Guide](https://itecsonline.com/post/how-to-install-claude-code-on-ubuntu-linux-complete-guide-2025)

## Migration from Native Setup

If you were running developer-agent-consumer natively:

1. Stop the native process
2. Update `.env` with `ANTHROPIC_API_KEY`
3. Rebuild Docker container
4. Start with `docker-compose up -d developer-agent-consumer`

Your existing JIRA webhook → code generation → PR flow will work exactly the same!

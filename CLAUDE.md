# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

shell2telegram is a Go application that creates Telegram bots from command-line shell commands. It allows users to define custom Telegram bot commands that execute shell commands and return their output to the chat.

## Build and Test Commands

```bash
# Build the project
make build
# or
go build

# Run tests
make test
# or
go test -race -cover -v ./...

# Lint the code
make lint
# or
golint ./...
go vet ./...
errcheck ./...

# Run the bot locally (requires TB_TOKEN environment variable)
make run
# or
go run . -tb-token=YOUR_TOKEN [options] /command 'shell command'
```

## Architecture

### Core Components

**shell2telegram.go** - Main entry point containing:
- `main()`: Event loop handling Telegram updates, message routing, and lifecycle management
- `getConfig()`: Command-line flag parsing and configuration setup
- `sendMessage()`: Handles message sending with automatic image detection and text chunking
- Bot operates in either polling mode (default) or webhook mode (`-bind-addr` and `-webhook` flags)

**users.go** - User management system:
- `Users` struct: Manages authorized users, root users, and authentication
- Authentication flow: Users request auth codes → codes displayed to console/root users → users enter codes to gain access
- Persistent storage: User data saved to JSON file (default: `~/.config/shell2telegram.json`)
- Auto-vacuum: Clears unauthorized users after 20 minutes of inactivity

**commands.go** - Bot command handlers:
- `cmdUser()`: Executes user-defined shell commands via `execShell()`
- `cmdAuth()`: Handles user/root authorization flow
- `cmdHelp()`: Dynamically generates help text from available commands
- `/shell2telegram` subcommands (root-only): `stat`, `ban`, `search`, `desc`, `rm`, `exit`, `version`, `broadcast_to_root`, `message_to_user`

**utils.go** - Utility functions:
- `execShell()`: Core shell execution with timeout, environment variable injection (S2T_LOGIN, S2T_USERID, S2T_USERNAME, S2T_CHATID), and optional caching
- `parseBotCommand()`: Parses command syntax with modifiers (`:desc`, `:vars`, `:md`)
- Image detection: Automatically detects PNG/JPEG/GIF/BMP from shell output and sends as photos

### Data Flow

1. Telegram update received → `main()` event loop
2. Message parsed into command + arguments
3. User authorization checked via `Users.IsAuthorized()`
4. Command dispatched to appropriate handler (auth, help, shell2telegram admin, or user command)
5. User commands executed in goroutines via `cmdUser()` → `execShell()`
6. Output sent back through `messageSignal` channel → `sendMessage()` → Telegram API

### Key Features

- **Command modifiers**: `/cmd:desc="Description"`, `/cmd:vars=VAR1,VAR2`, `/cmd:md` for markdown, `/cmd:file` for file output
- **Special commands**: `/:plain_text` captures all non-command messages in private chats
- **File handling**: Upload files to bot (accessible via `$S2T_FILE_PATH`), download files using `:file` modifier or `FILE:` prefix
- **Command history**: Track last 50 commands per user, view with `/history`, export with `/history export`
- **Database logging**: SQLite logging with `--enable-db-logging`, query logs and stats via `/shell2telegram logs/search_logs/db_stats`
- **Security**: Three-tier access (unauthorized, authorized, root), with optional `-allow-all` for public bots
- **Caching**: Command output caching with TTL (`-cache=N`)
- **Threading**: Optional single-threaded execution (`-one-thread`)
- **Timeout**: Shell command timeout (`-sh-timeout=N`)

## Configuration

The bot accepts command pairs as arguments:
```bash
shell2telegram [options] /telegram_command 'shell command' /command2 'shell command2'
```

### Environment Variables

The bot automatically reads the following environment variables:

**Bot Configuration:**
- `TB_TOKEN`: Telegram bot token (required, from BotFather)

**Proxy Configuration (automatically detected, priority order):**
1. Command-line flags: `-proxy-server`, `-proxy-user`, `-proxy-password` (highest priority)
2. Custom variables: `PROXY_SERVER`, `PROXY_USER`, `PROXY_PASSWORD`
3. Standard variables: `HTTP_PROXY`, `http_proxy`, `HTTPS_PROXY`, `https_proxy`

The proxy URL can include credentials: `http://user:pass@proxy:8080`

**Example:**
```bash
export PROXY_SERVER=proxy.company.com:8080
export PROXY_USER=username
export PROXY_PASSWORD=password
export TB_TOKEN=your_token

# Proxy is automatically used - no flags needed!
shell2telegram /date 'date'
```

## Testing

- Tests in `utils_test.go` cover utility functions
- Use `go test -race` to detect race conditions (important due to concurrent goroutines)

## Dependencies

- `gopkg.in/telegram-bot-api.v2`: Telegram Bot API v2 client
- `github.com/mattn/go-shellwords`: Shell command parsing
- `github.com/msoap/raphanus`: In-memory caching

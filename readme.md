# bj - Background Jobs

A handy lightweight CLI tool that reliably runs commands in the background. Sometimes `&` isn't enough and things don't detach properly. `bj` fixes that without much fuss.

## Install

```bash
go install github.com/metruzanca/bj@latest
```

## Usage

```bash
bj <command>       # Run command in background
bj -l, --list      # List all jobs
bj --logs [id]     # View logs (latest if no id)
bj --prune         # Clear all done jobs
```

### Examples

```bash
bj npm install          # Run npm install in background
bj make build           # Run make build in background
bj -l                   # Show job list with status
bj --logs               # View latest job's output
bj --logs 3             # View output from job #3
```

## Features

- **Reliable background execution** - Uses `setsid` to fully detach processes
- **Job tracking** - Records start/end time, exit code, working directory
- **Log capture** - All stdout/stderr saved to timestamped log files
- **Colored output** - Running/done/failed jobs are visually distinct
- **Auto-cleanup** - Done jobs older than 24hrs are automatically pruned
- **Configurable** - Custom log directory and log viewer

## Architecture

`bj` is designed to be extremely lightweight with no daemon or background service.

When you run `bj <command>`:

1. `bj` spawns a detached shell process (`$SHELL -c "your command"`)
2. Registers the job in `~/.config/bj/jobs.json`
3. Exits immediately - `bj` itself doesn't stay running

The detached shell handles everything: running the command, writing output to the log file, and calling `bj --complete` when done to record the exit code.

This means:
- Zero memory footprint after launch
- No daemon to manage or crash
- Jobs survive terminal closure
- Works with any shell (bash, zsh, fish, etc.)

## Configuration

Config file: `~/.config/bj/bj.toml`

```toml
log_dir = "logs"        # Relative to config dir, or absolute path
viewer = "less"         # Command to view logs
auto_prune_hours = 24   # Auto-delete done jobs older than N hours (0 = disabled)
```

## Files

- `~/.config/bj/bj.toml` - Configuration
- `~/.config/bj/jobs.json` - Job metadata
- `~/.config/bj/logs/` - Log files (timestamped)

---

<p align="center">Vibe Coded with <3</p>

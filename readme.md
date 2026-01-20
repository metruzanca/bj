# bj - Background Jobs

[![Go Coverage](https://github.com/metruzanca/bj/wiki/coverage.svg)](https://raw.githack.com/wiki/metruzanca/bj/coverage.html)

A handy lightweight CLI tool that reliably runs commands in the background. Sometimes `&` isn't enough and things don't detach properly. `bj` fixes that without much fuss. It goes down easy and gets the job done.

## Install

```bash
# With Go
go install github.com/metruzanca/bj@latest

# With mise
mise use -g github:metruzanca/bj
```

## Usage

```bash
bj <command>              # Run command in background
bj --retry[=N] <command>  # Run with retry until success (or N attempts)
bj --list                 # List all jobs
bj --logs [id]            # View logs (latest if no id)
bj --kill [id]            # Terminate a running job
bj --retry [--id ID]      # Retry a failed job
bj --prune                # Clear completed jobs
bj --gc                   # Clean up orphaned jobs after a crash
```

### Examples

```bash
bj npm install            # Run npm install in background
bj make build             # Run make build in background
bj --retry npm test       # Keep running tests until they pass
bj --retry=3 make build   # Try building up to 3 times
bj --retry --delay 5 ...  # Wait 5 seconds between retries
bj --list                 # Show job list with status
bj --list --running       # Show only running jobs
bj --list --failed        # Show only failed jobs
bj --logs                 # View latest job's output
bj --logs 3               # View output from job #3
bj --kill                 # Stop the most recent running job
bj --kill 5               # Stop job #5
bj --retry                # Retry most recent failed job
bj --retry --id 5         # Retry job #5
```

## Features

- **Reliable background execution** - Uses `setsid` to fully detach processes
- **Job tracking** - Records start/end time, exit code, working directory
- **Log capture** - All stdout/stderr saved to timestamped log files
- **Retry support** - Automatically retry failed commands with configurable attempts and delay
- **Job control** - Kill running jobs, retry failed ones
- **Colored output** - Running/done/failed jobs are visually distinct
- **Auto-cleanup** - Done jobs older than 24hrs are automatically pruned
- **Crash recovery** - `--gc` detects orphaned jobs after system crashes
- **Shell integration** - Tab completion and prompt integration for fish/zsh
- **Configurable** - Custom log directory, log viewer, and auto-prune settings
- **Quick and satisfying** - Finishes fast and leaves you free to move on

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
- Always ready when you need it

## Shell Integration

For tab completion and a prompt indicator showing running job count:

```bash
# Fish
echo 'bj --init fish | source' >> ~/.config/fish/config.fish

# Zsh (ensure ~/.zsh/completions is in fpath before compinit)
echo 'eval "$(bj --init zsh)"' >> ~/.zshrc
```

The prompt function shows `[bj:N]` when N jobs are running.

## Configuration

Config file: `~/.config/bj/bj.toml`

See [`.github/bj.toml`](.github/bj.toml) for a fully documented example config with all available options.

| Option | Default | Description |
|--------|---------|-------------|
| `log_dir` | `"logs"` | Where to store log files. Relative to config dir, or absolute path. |
| `viewer` | `"less"` | Command to view logs (`less`, `cat`, `bat`, `code`, etc.) |
| `auto_prune_hours` | `24` | Auto-delete successful jobs older than N hours. Set to `0` to disable. |

## Files

- `~/.config/bj/bj.toml` - Configuration
- `~/.config/bj/jobs.json` - Job metadata (ID, command, status, PID, timestamps)
- `~/.config/bj/logs/` - Log files (timestamped with job ID)

## Contributing

### Running Tests

```bash
go test ./...
```

### Testing Philosophy

This project is heavily snapshot-based (golden file tests). Why?

1. **Simpler** - No complex assertions or mocking, just "does the output match?"
2. **AI-resistant** - Since this project is built with agentic coding, snapshot tests are harder for AI to subtly break. There's no test logic to get wrong—either the output matches or it doesn't. An AI can't accidentally write a test that looks right but tests the wrong thing.

Snapshots live in `testdata/*.golden`. When you intentionally change output:

```bash
go test -update ./...
```

Then review the diff to `testdata/` files before committing.

---

<details>
<summary>How AI was used</summary>

This project was written entirely through agentic coding. I ([@metruzanca](https://github.com/metruzanca)) didn't write a single line of code—everything was done through prompts with Claude Opus 4.5.

I'm currently using `bj` in my own dev environment and it's working well. I originally just wanted to see how far I could get with pure agentic coding, and after it nailed the core functionality (which was all I actually needed), I was having too much fun and kept adding features.

For context: I've been writing code professionally for 6 years and coding even longer than that, so while I didn't write the code, I wasn't flying blind when steering the agent in the right direction.

</details>

<p align="center"><a href="CHANGELOG.md">See what's new</a> · Give bj a try. You won't regret it.</p>
<p align="center">Vibe Coded with <3</p>

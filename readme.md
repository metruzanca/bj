# bj - Background Jobs

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
bj <command>       # Run command in background
bj --list          # List all jobs
bj --logs [id]     # View logs (latest if no id)
bj --prune         # Clear all done jobs
```

### Examples

```bash
bj npm install          # Run npm install in background
bj make build           # Run make build in background
bj --list               # Show job list with status
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

## Contributing

### Running Tests

```bash
go test ./...
```

### Updating Snapshots

Static output (help text, completions, etc.) is tested using golden file snapshots in `testdata/`. If you intentionally change output, update the snapshots:

```bash
go test -update ./...
```

Then review the changes to `testdata/*.golden` files before committing.

---

<details>
<summary>How AI was used</summary>

This project was written entirely through agentic coding. I ([@metruzanca](https://github.com/metruzanca)) didn't write a single line of code—everything was done through prompts with Claude Opus 4.5.

I'm currently using `bj` in my own dev environment and it's working well. I originally just wanted to see how far I could get with pure agentic coding, and after it nailed the core functionality (which was all I actually needed), I was having too much fun and kept adding features.

For context: I've been writing code professionally for 6 years and coding even longer than that, so while I didn't write the code, I wasn't flying blind when steering the agent in the right direction.

</details>

<p align="center"><a href="CHANGELOG.md">See what's new</a> · Give bj a try. You won't regret it.</p>
<p align="center">Vibe Coded with <3</p>

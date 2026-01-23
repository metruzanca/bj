# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] - 2026-01-23

### Added
- `--man` flag to output a full manual page (pipe to `man` for viewing)
- NSFW mode (`nsfw = true` in config) for raunchier message variants
- Localization system for all user-facing strings

### Changed
- `--prune` now clears all completed jobs (any exit code) and resets ID counter when empty

### Fixed
- Man page now compatible with `mandoc(1)` on macOS

## [0.3.1] - 2026-01-22

### Fixed
- Unknown flags (e.g., `--typo`) are now rejected instead of being run as commands

## [0.3.0] - 2026-01-20

### Added
- `--retry` as a modifier flag for running commands with automatic retry
- `--delay` flag for configurable retry backoff
- `--kill` flag to terminate running jobs
- `--gc` flag to detect and clean up orphaned jobs after crashes
- `--running`, `--failed`, `--done` filters for `--list` command
- Comprehensive integration test suite with golden file snapshots
- CI workflow with test coverage badge
- Auto-prune and gc execution on `--init`

### Changed
- `--retry` can now be used as both a modifier flag and standalone flag for retrying failed jobs

### Fixed
- `--id` flag now validates that only positive integers are accepted
- CI `-v` flag no longer hides test failures

### Removed
- `--ruined` alias for `--failed` (keeping interface professional)
- `jq` dependency

## [0.2.0] - 2026-01-20

### Added
- `--json` flag for machine-readable output (works with all commands)
- Shell integration with `--completion` and `--init` commands
- Per-command help (e.g., `bj --list --help`)
- Relative timestamps in job list ("5 mins ago" instead of absolute dates)
- Release script for easier releases (`scripts/release.go`)

### Removed
- `-l` shorthand (was confusing with `--logs`)

## [0.1.0] - 2026-01-19

### Added
- Core CLI with `bj <command>` to run jobs in background
- `--list` to show all tracked jobs with colored status
- `--logs [id]` to view job output
- `--prune` to clean up completed jobs
- TOML configuration (`~/.config/bj/bj.toml`)
- Auto-prune for jobs older than configurable hours
- Detached process execution via `setsid`
- Job tracking with start/end time, exit code, working directory

[Unreleased]: https://github.com/metruzanca/bj/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/metruzanca/bj/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/metruzanca/bj/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/metruzanca/bj/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/metruzanca/bj/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/metruzanca/bj/releases/tag/v0.1.0

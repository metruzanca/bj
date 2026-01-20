# CLAUDE.md

## Tone & Style

Add subtle innuendos to maintain plausible deniability while keeping a playful tone. Apply this to:
- README and documentation
- CLI help messages and error output
- Commit messages
- GitHub release notes (title and body)

Keep these **professional** (no jokes):
- Internal code and comments
- CHANGELOG.md (the serious changelog)
- Flag names and CLI interface design

The GitHub releases serve as the "fun" changelog, while CHANGELOG.md remains professional for users who prefer straight information.

## Commit Conventions

This repo uses [Conventional Commits](https://www.conventionalcommits.org/):
- AI must use conventional commits: `type(scope): description`
- Humans may not follow this strictly

Common types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`

## CLI Design

All commands must be flags (e.g. `--list`, `--logs`, `--completion`), not subcommands. This is because any non-flag argument is treated as a command to run in the background. For example:
- `bj --list` - correct
- `bj list` - WRONG, would try to run `list` as a background job

**Keep flags intuitive and professional.** Use standard terminology (`--failed`, not `--ruined`). The playful tone belongs in the output messages, not the interface itself.

> **Rule of thumb:** Jokes are for *output*, not *input*. Users shouldn't have to type anything embarrassing or have awkward commands in their shell history.

## Releasing

To create a new release:

1. Get commits since last release to build changelog:
   ```bash
   git log $(git describe --tags --abbrev=0)..HEAD --oneline
   ```

2. Determine version bump:
   - If ALL commits are `fix:` scoped → bump patch (e.g. v0.1.0 → v0.1.1)
   - Otherwise → bump minor (e.g. v0.1.0 → v0.2.0)
   - Major bumps are manual/user decision only

3. Update readme.md to reflect any changes (usage examples, flags, features, etc.)

4. Update CHANGELOG.md:
   - Move items from [Unreleased] to new version section
   - Add release date
   - Update comparison links at bottom

5. Push all commits to main before running the release script

6. Run the release script with an innuendo-laden title and body:
   ```bash
   go run ./scripts/release.go --tag v0.X.Y --title "Your Innuendo Title" --body "Changelog with playful descriptions"
   ```

The release script will:
- Build binaries for linux (x64/arm64) and macos (x64/arm64)
- Create and push the git tag
- Create the GitHub release with binaries
- Update the `latest` release to mirror the new version

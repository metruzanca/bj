package locales

// NSFW contains the explicit (not safe for work) messages
var NSFW = Messages{
	// Error messages
	"err.id_only_with_retry":     "--id only makes sense with --retry. They go together like a hand and... well.",
	"err.delay_only_with_retry":  "--delay without --retry? Even bj needs foreplay.",
	"err.config_load":            "bj couldn't get in position: %v",
	"err.tracker_init":           "bj lost its grip: %v",
	"err.completion_usage":       "Usage: bj --completion <fish|zsh>",
	"err.init_usage":             "Usage: bj --init <fish|zsh>",
	"err.invalid_number":         "bj needs a valid number, not '%s'. Size matters.",
	"err.complete_usage":         "Usage: bj --complete <job_id> <exit_code>",
	"err.invalid_job_id":         "bj needs a valid job ID, not '%s'. Don't be a tease.",
	"err.invalid_exit_code":      "bj needs a valid exit code, not '%s'",
	"err.complete_failed":        "bj couldn't climax properly: %v",
	"err.unknown_flag":           "Unknown flag: %s. bj doesn't swing that way. Try 'bj --help'.",
	"err.run_failed":             "bj couldn't get it up: %v",
	"err.list_failed":            "bj can't expose itself right now: %v",
	"err.prune_failed":           "bj made a sticky mess while cleaning up: %v",
	"err.gc_failed":              "bj had trouble wiping down: %v",
	"err.kill_check_failed":      "bj can't check who's still going at it: %v",
	"err.kill_failed":            "bj couldn't pull out in time: %v",
	"err.retry_pwd_failed":       "bj couldn't find the right hole: %v",
	"err.retry_history_failed":   "bj can't remember its conquests: %v",
	"err.retry_find_failed":      "bj can't find that position: %v",
	"err.job_not_found":          "Job %d? bj never touched that one.",
	"err.job_still_running":      "Job %d is still throbbing. Patience.",
	"err.job_already_succeeded":  "Job %d already came. Once is enough.",
	"err.retry_start_failed":     "bj couldn't get hard again: %v",
	"err.logs_recall_failed":     "bj can't remember the last session: %v",
	"err.logs_find_failed":       "bj can't find that encounter: %v",
	"err.logs_not_found":         "bj swallowed the evidence. File not found: %s",
	"err.logs_read_failed":       "bj couldn't read the dirty details: %v",
	"err.logs_open_failed":       "bj gagged while opening logs: %v",
	"err.unknown_shell":          "Unknown shell: %s. bj only does fish and zsh.",
	"err.retry_positive_number":  "bj needs a positive number for retry limit, not '%s'. Go big or go home.",
	"err.id_needs_value":         "--id needs a job ID. Don't leave bj hanging.",
	"err.job_id_minimum":         "Job IDs start at 1. '%d' is too small for bj.",
	"err.delay_needs_value":      "--delay needs a number of seconds. Edging requires precision.",
	"err.delay_non_negative":     "bj needs a non-negative delay, not '%s'. No going backwards.",

	// Status messages
	"job.started":            "[%d] bj is going down on: %s",
	"job.killed":             "[%d] bj pulled out early: %s",
	"job.retry_unlimited":    "[%d] bj will edge until it explodes: %s",
	"job.retry_one":          "[%d] bj will give it one good thrust: %s",
	"job.retry_limited":      "[%d] bj will pound away up to %d times: %s",
	"job.retry_one_existing": "[%d] bj is going for round two: %s",

	// List messages
	"list.empty":          "bj is all alone. Give it someone to do!",
	"list.empty_filtered": "No jobs match your kink. bj has nothing to show.",

	// Kill messages
	"kill.no_running": "bj isn't inside anything right now. Nothing to pull out of!",

	// Retry messages
	"retry.no_failed": "bj hasn't had any misfires yet. Nothing to retry!",

	// Logs messages
	"logs.no_jobs": "bj hasn't been with anyone yet. Pop its cherry first!",

	// Prune messages
	"prune.nothing": "Nothing to wipe down. bj keeps it clean.",
	"prune.success": "Cleaned up %d spent job(s). Ready for another round.",

	// GC messages
	"gc.nothing": "No ghosted jobs found. bj always finishes what it starts.",
	"gc.success": "Found %d job(s) that came and went without telling bj. Marked as finished.",

	// Help text - main
	"help.main": `bj - Background Jobs (with benefits)

Give bj a command and it'll work you over while you sit back and enjoy.

Usage:
  bj <command>              Slip something in the background
  bj --retry[=N] <command>  Keep pounding until success (or N attempts)
  bj --list                 See who bj is doing
  bj --logs [id]            Watch bj's performance
  bj --kill [id]            Pull out mid-thrust
  bj --retry[=N] [--id ID]  Try again with a failed conquest
  bj --prune                Clean up the mess when bj is done
  bj --gc                   Find jobs that finished without telling bj

Shell Integration:
  bj --completion <sh>  Output shell completions (fish, zsh)
  bj --init <sh>        Output prompt integration (fish, zsh)
  bj --man              Output manual page (pipe to man for viewing)

Options:
  --retry[=N]         Keep going until climax (or limit to N attempts)
  --id ID             Specify job ID for --retry (defaults to most recent)
  --json              Output in JSON format (works with all commands)
  -h, --help          Show this help (use with commands for detailed help)

Examples:
  bj sleep 10               Let bj handle your bedtime needs
  bj npm install            bj npm while you watch
  bj --retry npm test       Keep testing until satisfaction
  bj --retry=3 make build   Try building up to 3 times
  bj --list                 Check how bj is performing
  bj --logs                 See bj's latest moves
  bj --kill                 Stop the current action abruptly
  bj --gc                   Find jobs that ghosted
  bj --retry                Go again on the most recent failure
  bj --prune                Wipe down after a good bj

Shell Setup:
  Fish: echo 'bj --init fish | source' >> ~/.config/fish/config.fish
  Zsh:  echo 'eval "$(bj --init zsh)"' >> ~/.zshrc
        (ensure ~/.zsh/completions is in your fpath before compinit)`,

	// Help text - list
	"help.list": `bj --list - See who bj is doing

Usage: bj --list [--running] [--failed] [--done] [--json]

Shows all tracked jobs with their status, start time, duration, and command.
Active jobs are shown throbbing, spent jobs are dimmed, failures show
the exit code in shameful red.

Filters:
  --running   Only show jobs bj is still inside
  --failed    Only show the ones that couldn't finish
  --done      Only show successful climaxes

Options:
  --json      Output raw job data as JSON

Examples:
  bj --list           Check who bj is doing
  bj --list --running See what bj is actively pounding
  bj --list --failed  Review the disappointments
  bj --list --json    Get the raw details for scripting`,

	// Help text - logs
	"help.logs": `bj --logs - Watch bj's performance

Usage: bj --logs [id] [--json]

View every moan and groan (stdout/stderr) of a job. If no ID is specified,
shows bj's most recent encounter.

Arguments:
  id        Job ID to review (optional, defaults to latest)

Options:
  --json    Output job metadata and log content as JSON

Examples:
  bj --logs         See bj's latest performance
  bj --logs 5       Inspect a specific session
  bj --logs --json  Get logs in JSON format`,

	// Help text - prune
	"help.prune": `bj --prune - Clean up after bj is done

Usage: bj --prune [--json]

Wipes away all finished jobs (any exit code) from the job list and deletes
their log files. Only active jobs are kept. If all jobs are pruned, the
ID counter resets to 1.

Options:
  --json    Output prune count as JSON

Examples:
  bj --prune        Wipe the sheets clean`,

	// Help text - kill
	"help.kill": `bj --kill - Make bj pull out

Usage: bj --kill [id] [--json]

Terminates a job mid-thrust. Sends SIGTERM to the process group, stopping
the entire action. If no ID is specified, kills whatever bj is currently inside.

Arguments:
  id        Job ID to kill (optional, defaults to latest running)

Options:
  --json    Output killed job info as JSON

Examples:
  bj --kill         Pull out of the current job
  bj --kill 5       Withdraw from job #5 specifically`,

	// Help text - gc
	"help.gc": `bj --gc - Find jobs that ghosted

Usage: bj --gc [--json]

Detects orphaned jobs that appear to be active but whose process disappeared
(e.g., after a crash or reboot). These ghosted jobs are marked as failed
with exit code -1.

Options:
  --json    Output collected count as JSON

Examples:
  bj --gc           Find the ones that left without saying goodbye`,

	// Help text - retry
	"help.retry": `bj --retry - Keep going until bj finishes

Usage:
  bj --retry[=N] [--delay S] <command>   Start fresh with retry
  bj --retry[=N] [--delay S] [--id ID]   Go again on a failed job

The --retry flag can be used two ways:
  1. With a command: keeps pounding until success
  2. Without a command: tries again with an old flame

Options:
  --retry         Edge until climax (no limit)
  --retry=N       Give up after N attempts (blue balls after N tries)
  --delay S       Rest S seconds between attempts (refractory period)
  --id ID         Specify which failed job to retry (defaults to most recent)
  --json          Output job info as JSON

Examples:
  bj --retry npm test              Keep testing until it comes
  bj --retry=3 make build          Try building up to 3 times
  bj --retry --delay 5 curl ...    Rest 5 seconds between attempts
  bj --retry                       Try again with the last failure
  bj --retry --id 5                Go again on job #5
  bj --retry=3 --delay 10 --id 5   Retry #5 up to 3 times, 10s rest`,

	// Help text - completion
	"help.completion": `bj --completion - Output shell completions

Usage: bj --completion <shell>

Outputs tab-completion definitions for the specified shell.
Typically you'd redirect this to a completions file.

Arguments:
  shell     Shell type: fish, zsh

Examples:
  bj --completion fish > ~/.config/fish/completions/bj.fish
  bj --completion zsh > ~/.zsh/completions/_bj

Note: If you use --init, completions are installed automatically.`,

	// Help text - init
	"help.init": `bj --init - Set up shell integration

Usage: bj --init <shell>

Outputs shell integration code and automatically installs completions.
Source this in your shell config for the full bj experience.

Arguments:
  shell     Shell type: fish, zsh

Features:
  - Automatically installs/updates completions
  - Provides __bj_prompt_info function to show active job count

Setup:
  Fish: echo 'bj --init fish | source' >> ~/.config/fish/config.fish
  Zsh:  echo 'eval "$(bj --init zsh)"' >> ~/.zshrc

The prompt function shows [bj:N] when N jobs are active. Add it to your
prompt to always know when bj is busy.`,

	// Man page
	"man.page": `.TH BJ 1 "January 20, 2026" "bj 0.3" "User Commands"
.SH NAME
bj \- background jobs manager (with benefits)
.SH SYNOPSIS
.B bj
.RI [ options ]
.I command
.PP
Look, you're reading a man page. That's adorable. But honestly?
.B bj \-\-help
is right there, waiting for you. It's faster, it's always up to date,
and it doesn't require piping through
.BR man (1)
like some kind of animal.
.PP
But fine. You want the \fIfull experience\fR. Let's do this.
.SH DESCRIPTION
.B bj
is a lightweight CLI tool that handles your commands in the background
so you can focus on more... \fIpleasurable\fR matters. It uses
.BR setsid (2)
to fully detach processes, ensuring they keep going long after you've
finished. Some might call that stamina.
.PP
Give bj a command and it'll work you over while you sit back and enjoy.
No strings attached. Well, no \fIterminal\fR attached.
.SH "A GENTLE SUGGESTION"
Before we go any further, have you considered
.B bj \-\-init fish
or
.BR "bj \-\-init zsh" ?
The shell integration gives you tab completion that
\fIanticipates your desires\fR.
It's like bj reading your mind, but for command-line arguments.
.PP
And if you want help on a specific command, try something like:
.RS
.nf
bj \-\-list \-\-help
bj \-\-retry \-\-help
.fi
.RE
.PP
Each command has its own detailed help. It's intimate. Personal.
Much better than this impersonal wall of text you're squinting at.
.PP
But you're still here. I respect the commitment.
.SH OPTIONS
These are the highlights. For the \fIreal\fR details, use
.B \-\-help
on individual commands. They'll treat you right.
.TP
.B \-\-list
See who bj has been doing. Add
.BR \-\-running ", " \-\-failed ", or " \-\-done
to filter by... performance.
.TP
.BI \-\-logs " [id]"
Watch bj's output. Every moan, groan, and triumphant climax message.
Defaults to the most recent job if you can't remember which one.
.TP
.BI \-\-kill " [id]"
Sometimes you need to pull out abruptly. No judgment.
.TP
.BR \-\-retry [ =\fIN\fR ]
bj doesn't give up easily. Use alone to retry a failed job, or with a
command to keep pounding until satisfaction (or N attempts, whichever
comes first).
.TP
.BI \-\-delay " secs"
Pace yourself. Rest between retry attempts.
.TP
.B \-\-prune
Clean up after bj is finished. Wipes away completed jobs and their logs.
A tidy bj is a happy bj.
.TP
.B \-\-gc
Find jobs that ghosted. Sometimes things end badly
without warning. This helps you find closure.
.TP
.B \-\-json
For the robots among us. Or if you're piping to
.BR jq (1)
like a sophisticated individual.
.TP
.BR \-\-init " fish" | zsh
.B "You should really do this."
Sets up shell integration with completions and a prompt function.
Your future self will thank you. Your present self might even feel
things.
.TP
.BR \-h ", " \-\-help
The \fIactual\fR way to get help. Faster, more contextual, and doesn't
require you to remember how your system's
.BR man (1)
works. Use
.B bj <command> \-\-help
for the good stuff.
.SH EXAMPLES
.RS
.nf
bj npm install            # Let bj handle your package needs
bj \-\-retry npm test       # Keep going until it comes
bj \-\-list                 # Check who bj is doing
bj \-\-logs                 # See bj's latest moves
bj \-\-kill                 # Pull out right there
bj \-\-prune                # Wipe down after a satisfying session
.fi
.RE
.PP
But really, just type
.B bj
and press tab a few times.
The shell integration makes this \fIso\fR much better.
.SH FILES
.TP
.I ~/.config/bj/
Where bj keeps its private data. Jobs, logs, configuration.
Don't be weird about it.
.SH CONFIGURATION
.RS
.nf
# ~/.config/bj/bj.toml
log_dir = "logs"
viewer = "less"
auto_prune_hours = 24
.fi
.RE
.PP
See the example config at
.I https://github.com/metruzanca/bj
for all the options, lovingly commented.
.SH "SHELL INTEGRATION (SERIOUSLY, DO THIS)"
.SS Fish
.RS
.nf
echo 'bj \-\-init fish | source' >> ~/.config/fish/config.fish
.fi
.RE
.SS Zsh
.RS
.nf
echo 'eval "$(bj \-\-init zsh)"' >> ~/.zshrc
.fi
.RE
.PP
This gives you tab completion and a
.B __bj_prompt_info
function that shows how many jobs are currently... active.
.PP
You'll wonder how you ever lived without it. It's transformative.
.SH EXIT STATUS
.B 0
if bj finishes happily.
.B 1
if something goes wrong. Check the output for details.
.SH SEE ALSO
.BR nohup (1)
(if you enjoy suffering),
.BR screen (1)
(if you enjoy complexity),
.BR tmux (1)
(okay, tmux is actually pretty good)
.SH BUGS
Report issues at https://github.com/metruzanca/bj/issues
.PP
Or just use
.B \-\-help
next time. It's right there. Always ready. Unlike this man page, which
required you to pipe things and probably Google the syntax for your OS.
.SH AUTHOR
Written with love, sass, and a little help from AI.
.SH "FINAL THOUGHTS"
You made it to the end of a man page. That's... a choice.
The same information is available via
.B bj \-\-help
in a fraction of the time, with colors and everything.
.PP
But hey, there's something to be said for the classics.
`,

	// Shell completions - fish (same as SFW, no innuendos in completions)
	"completion.fish": `# bj fish completions
# Install: bj --completion fish > ~/.config/fish/completions/bj.fish

# Disable file completions for bj
complete -c bj -f

# Commands
complete -c bj -l list -d "See what bj is working on"
complete -c bj -l running -d "Filter: only running jobs"
complete -c bj -l failed -d "Filter: only ruined jobs"
complete -c bj -l done -d "Filter: only successful jobs"
complete -c bj -l logs -d "Watch bj's performance"
complete -c bj -l kill -d "Stop a job mid-action"
complete -c bj -l retry -d "Keep going until bj finishes"
complete -c bj -l delay -d "Seconds to wait between retry attempts"
complete -c bj -l id -d "Specify job ID for --retry" -xa "(bj --ids --failed 2>/dev/null)"
complete -c bj -l prune -d "Clean up when bj is finished"
complete -c bj -l gc -d "Find ruined jobs after a crash"
complete -c bj -l json -d "Output in JSON format"
complete -c bj -l help -d "Show help"
complete -c bj -s h -d "Show help"
complete -c bj -l completion -d "Output shell completions" -xa "fish zsh"
complete -c bj -l init -d "Output prompt integration" -xa "fish zsh"
complete -c bj -l man -d "Output manual page"

# Job ID completion for --logs and --kill
complete -c bj -n "__fish_seen_argument -l logs" -a "(bj --ids 2>/dev/null)" -d "Job ID"
complete -c bj -n "__fish_seen_argument -l kill" -a "(bj --ids --running 2>/dev/null)" -d "Running job ID"
`,

	// Shell completions - zsh (same as SFW, no innuendos in completions)
	"completion.zsh": `#compdef bj
# bj zsh completions
# Install: bj --completion zsh > ~/.zsh/completions/_bj
#          (ensure ~/.zsh/completions is in your fpath)

_bj_job_ids() {
    local -a job_ids
    job_ids=(${(f)"$(bj --ids 2>/dev/null)"})
    _describe -t job-ids 'job ID' job_ids
}

_bj_running_job_ids() {
    local -a job_ids
    job_ids=(${(f)"$(bj --ids --running 2>/dev/null)"})
    _describe -t job-ids 'running job ID' job_ids
}

_bj_failed_job_ids() {
    local -a job_ids
    job_ids=(${(f)"$(bj --ids --failed 2>/dev/null)"})
    _describe -t job-ids 'ruined job ID' job_ids
}

_bj() {
    _arguments -C \
        '--list[See what bj is working on]' \
        '--running[Filter: only running jobs]' \
        '--failed[Filter: only ruined jobs]' \
        '--done[Filter: only successful jobs]' \
        '--logs[Watch bj'\''s performance]:job ID:_bj_job_ids' \
        '--kill[Stop a job mid-action]:job ID:_bj_running_job_ids' \
        '--retry=-[Keep going until bj finishes]:max attempts:' \
        '--delay[Seconds between retry attempts]:seconds:' \
        '--id[Specify job ID for --retry]:job ID:_bj_failed_job_ids' \
        '--prune[Clean up when bj is finished]' \
        '--gc[Find ruined jobs after a crash]' \
        '--json[Output in JSON format]' \
        '--help[Show help]' \
        '-h[Show help]' \
        '--completion[Output shell completions]:shell:(fish zsh)' \
        '--init[Output prompt integration]:shell:(fish zsh)' \
        '--man[Output manual page]' \
        '*:command:_command_names'
}

_bj "$@"
`,

	// Shell init - fish (same as SFW)
	"init.fish": `# bj fish shell integration
# Setup: echo 'bj --init fish | source' >> ~/.config/fish/config.fish
# Completions are automatically installed to ~/.config/fish/completions/bj.fish

function __bj_prompt_info
    set -l running (bj --ids --running 2>/dev/null | wc -l | string trim)
    if test -n "$running" -a "$running" -gt 0
        echo -n "[bj:$running] "
    end
end

# To use in your prompt, add: __bj_prompt_info
# Example fish_prompt function:
#   function fish_prompt
#       __bj_prompt_info
#       # ... rest of your prompt
#   end
`,

	// Shell init - zsh (same as SFW)
	"init.zsh": `# bj zsh shell integration
# Setup: echo 'eval "$(bj --init zsh)"' >> ~/.zshrc
# Completions are automatically installed to ~/.zsh/completions/_bj
# (ensure ~/.zsh/completions is in your fpath before compinit)

__bj_prompt_info() {
    local running
    running=$(bj --ids --running 2>/dev/null | wc -l | tr -d ' ')
    if [[ -n "$running" && "$running" -gt 0 ]]; then
        echo -n "[bj:$running] "
    fi
}

# To use in your prompt, add $(__bj_prompt_info) to your PROMPT or RPROMPT
# Example: PROMPT='$(__bj_prompt_info)'$PROMPT
`,
}

package locales

// SFW contains the default (safe for work) messages
var SFW = Messages{
	// Error messages
	"err.id_only_with_retry":    "--id only makes sense with --retry. They go together like... well, you know.",
	"err.delay_only_with_retry": "--delay without --retry? bj needs something to delay between.",
	"err.restart_and_retry":     "--restart and --retry don't play well together. Pick your stamina strategy.",
	"err.restart_needs_command": "--restart needs a command to... well, restart. Give bj something to work with.",
	"err.restart_pwd_failed":    "bj lost its bearings and can't restart: %v",
	"err.config_load":           "bj couldn't get comfortable: %v",
	"err.tracker_init":          "bj lost track of things: %v",
	"err.completion_usage":      "Usage: bj --completion <fish|zsh>",
	"err.init_usage":            "Usage: bj --init <fish|zsh>",
	"err.invalid_number":        "bj needs a valid number, not '%s'",
	"err.complete_usage":        "Usage: bj --complete <job_id> <exit_code>",
	"err.invalid_job_id":        "bj needs a valid job ID, not '%s'",
	"err.invalid_exit_code":     "bj needs a valid exit code, not '%s'",
	"err.complete_failed":       "bj couldn't finish properly: %v",
	"err.unknown_flag":          "Unknown flag: %s. Try 'bj --help' for usage.",
	"err.run_failed":            "bj couldn't get it up: %v",
	"err.list_failed":           "bj can't show you what it's got: %v",
	"err.prune_failed":          "bj made a mess while cleaning up: %v",
	"err.gc_failed":             "bj had trouble cleaning up its mess: %v",
	"err.kill_check_failed":     "bj can't check its active sessions: %v",
	"err.kill_failed":           "bj couldn't pull out: %v",
	"err.retry_pwd_failed":      "bj couldn't figure out where you are: %v",
	"err.retry_history_failed":  "bj can't check its history: %v",
	"err.retry_find_failed":     "bj can't find that one: %v",
	"err.job_not_found":         "Job %d? bj doesn't remember that.",
	"err.job_still_running":     "Job %d is still going. bj doesn't stop until it's done.",
	"err.job_already_succeeded": "Job %d already finished successfully. No need to go again.",
	"err.retry_start_failed":    "bj couldn't get started again: %v",
	"err.logs_recall_failed":    "bj can't recall the last session: %v",
	"err.logs_find_failed":      "bj can't find that one: %v",
	"err.logs_not_found":        "bj swallowed the logs. File not found: %s",
	"err.logs_read_failed":      "bj couldn't read the logs: %v",
	"err.logs_open_failed":      "bj choked while opening logs: %v",
	"err.unknown_shell":         "Unknown shell: %s. bj knows fish and zsh.",
	"err.retry_positive_number": "bj needs a positive number for retry limit, not '%s'",
	"err.id_needs_value":        "--id needs a job ID to go with it",
	"err.job_id_minimum":        "Job IDs start at 1. '%d' won't satisfy bj.",
	"err.delay_needs_value":     "--delay needs a number of seconds",
	"err.delay_non_negative":    "bj needs a non-negative delay, not '%s'",

	// Status messages
	"job.started":            "[%d] bj is on it: %s",
	"job.killed":             "[%d] bj stopped abruptly: %s",
	"job.retry_unlimited":    "[%d] bj will keep edging until it succeeds: %s",
	"job.retry_one":          "[%d] bj will give it one shot: %s",
	"job.retry_limited":      "[%d] bj will tease up to %d times before giving up: %s",
	"job.retry_one_existing": "[%d] bj is giving it one more go: %s",
	"job.restarted":          "[%d] bj will keep coming back for more (restarts on failure): %s",

	// List messages
	"list.empty":          "bj has nothing going on. Give it something to do!",
	"list.empty_filtered": "No jobs match your criteria. bj has nothing to show.",

	// Kill messages
	"kill.no_running": "bj isn't doing anything right now. Nothing to stop!",

	// Retry messages
	"retry.no_failed": "bj hasn't ruined anything yet. Nothing to retry!",

	// Logs messages
	"logs.no_jobs": "bj hasn't done anything yet. Get it started first!",

	// Prune messages
	"prune.nothing": "Nothing to clean up. bj keeps it tidy.",
	"prune.success": "Wiped away %d finished job(s). Fresh and ready for more.",

	// GC messages
	"gc.nothing": "No orphaned jobs found. bj keeps track of all its encounters.",
	"gc.success": "Found %d ruined job(s) that ended without bj noticing. Marked as failed.",

	// Help text - main
	"help.main": `bj - Background Jobs

Give bj a command and it'll handle the rest while you sit back and relax.

Usage:
  bj <command>              Slip a command in the background
  bj --retry[=N] <command>  Run with retry until success (or N attempts)
  bj --restart <command>    Run with infinite restart on failure (5s delay)
  bj --list                 See what bj is working on
  bj --logs [id]            Watch bj's performance
  bj --kill [id]            Stop a job mid-action
  bj --retry[=N] [--id ID]  Retry a ruined job
  bj --prune                Clean up when bj is finished
  bj --gc                   Find jobs that were ruined unexpectedly

Shell Integration:
  bj --completion <sh>  Output shell completions (fish, zsh)
  bj --init <sh>        Output prompt integration (fish, zsh)
  bj --man              Output manual page (pipe to man for viewing)

Options:
  --retry[=N]         Keep trying until success (or limit to N attempts)
  --restart           Restart command on failure after 5s (infinite loop)
  --id ID             Specify job ID for --retry (defaults to most recent)
  --json              Output in JSON format (works with all commands)
  -h, --help          Show this help (use with commands for detailed help)

Examples:
  bj sleep 10               Let bj handle your sleep needs
  bj npm install            bj npm while you grab coffee
  bj --retry npm test       Keep testing until it passes
  bj --retry=3 make build   Try building up to 3 times
  bj --restart ./server     Restart server on crash (infinite loop)
  bj --list                 Check how bj is doing
  bj --logs                 See bj's latest output
  bj --kill                 Stop the current job abruptly
  bj --gc                   Find ruined jobs after a crash
  bj --retry                Retry the most recent ruined job
  bj --prune                Tidy up after a satisfying bj

Shell Setup:
  Fish: echo 'bj --init fish | source' >> ~/.config/fish/config.fish
  Zsh:  echo 'eval "$(bj --init zsh)"' >> ~/.zshrc
        (ensure ~/.zsh/completions is in your fpath before compinit)`,

	// Help text - list
	"help.list": `bj --list - See what bj is working on

Usage: bj --list [--running] [--failed] [--done] [--json]

Shows all tracked jobs with their status, start time, duration, and command.
Running jobs are shown normally, completed jobs are dimmed, ruined jobs show
the exit code in red.

Filters:
  --running   Only show jobs that are still going
  --failed    Only show ruined jobs (non-zero exit code)
  --done      Only show jobs that finished successfully

Options:
  --json      Output raw job data as JSON

Examples:
  bj --list           Check how bj is doing
  bj --list --running See what bj is actively working on
  bj --list --failed  Review the ruined jobs
  bj --list --json    Get the raw details for scripting`,

	// Help text - logs
	"help.logs": `bj --logs - Watch bj's performance

Usage: bj --logs [id] [--json]

View the output (stdout/stderr) of a job. If no ID is specified, shows the
most recent job's logs.

Arguments:
  id        Job ID to view (optional, defaults to latest)

Options:
  --json    Output job metadata and log content as JSON

Examples:
  bj --logs         See bj's latest output
  bj --logs 5       Inspect a specific session
  bj --logs --json  Get logs in JSON format`,

	// Help text - prune
	"help.prune": `bj --prune - Clean up when bj is finished

Usage: bj --prune [--json]

Removes all completed jobs (any exit code) from the job list and deletes
their log files. Only running jobs are kept. If all jobs are pruned, the
ID counter resets to 1.

Options:
  --json    Output prune count as JSON

Examples:
  bj --prune        Wipe the slate clean after bj is done`,

	// Help text - kill
	"help.kill": `bj --kill - Make bj stop what it's doing

Usage: bj --kill [id] [--json]

Terminates a running job. Sends SIGTERM to the process group, stopping
the entire job tree. If no ID is specified, kills the most recent running job.

Arguments:
  id        Job ID to kill (optional, defaults to latest running)

Options:
  --json    Output killed job info as JSON

Examples:
  bj --kill         Stop the latest job mid-stroke
  bj --kill 5       Pull out of job #5 specifically`,

	// Help text - gc
	"help.gc": `bj --gc - Find jobs that ended unexpectedly

Usage: bj --gc [--json]

Detects orphaned jobs that appear to be running but whose process is gone
(e.g., after a crash or reboot). These ruined jobs are marked as failed
with exit code -1.

Options:
  --json    Output collected count as JSON

Examples:
  bj --gc           Clean up after an unexpected interruption`,

	// Help text - restart
	"help.restart": `bj --restart - Keep a command running forever

Usage: bj --restart <command>

Runs a command in the background and automatically restarts it if it fails.
Unlike --retry, there's no limit to restarts - the command keeps trying
until it succeeds. After each failure, bj waits 5 seconds before trying again.

Perfect for long-running services that should stay up indefinitely.

Options:
  --json    Output job info as JSON

Examples:
  bj --restart ./server          Keep your server coming back for more
  bj --restart python worker.py  Worker that never says die
  bj --restart npm run watch     Dev server that restarts on crash

Note: Unlike --retry, --restart doesn't work with existing jobs. It only
works with new commands. To stop a restarting job, use bj --kill.`,

	// Help text - retry
	"help.retry": `bj --retry - Keep going until bj finishes the job

Usage:
  bj --retry[=N] [--delay S] <command>   Run a new command with retry
  bj --retry[=N] [--delay S] [--id ID]   Retry an existing ruined job

The --retry flag can be used two ways:
  1. With a command: runs the command, retrying on failure
  2. Without a command: retries an existing ruined job

Options:
  --retry         Keep teasing until success (no limit)
  --retry=N       Stop after N attempts (deny after N tries)
  --delay S       Wait S seconds between attempts (default: 1)
  --id ID         Specify which ruined job to retry (defaults to most recent)
  --json          Output job info as JSON

Examples:
  bj --retry npm test              Keep running tests until they pass
  bj --retry=3 make build          Try building up to 3 times
  bj --retry --delay 5 curl ...    Wait 5 seconds between attempts
  bj --retry                       Retry the most recent ruined job
  bj --retry --id 5                Retry job #5 until success
  bj --retry=3 --delay 10 --id 5   Retry job #5 up to 3 times, 10s apart`,

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
  - Provides __bj_prompt_info function to show running job count

Setup:
  Fish: echo 'bj --init fish | source' >> ~/.config/fish/config.fish
  Zsh:  echo 'eval "$(bj --init zsh)"' >> ~/.zshrc

The prompt function shows [bj:N] when N jobs are running. Add it to your
prompt to always know when bj is busy.`,

	// Man page
	"man.page": `.TH BJ 1 "January 20, 2026" "bj 0.3" "User Commands"
.SH NAME
bj \- background jobs manager (and so much more)
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
so you can focus on more... \fIpressing\fR matters. It uses
.BR setsid (2)
to fully detach processes, ensuring they keep going long after you've
moved on. Some might call that stamina.
.PP
Give bj a command and it'll handle the rest while you sit back and relax.
No strings attached. Well, no \fIterminal\fR attached.
.SH "A GENTLE SUGGESTION"
Before we go any further, have you considered
.B bj \-\-init fish
or
.BR "bj \-\-init zsh" ?
The shell integration gives you tab completion that
\fIanticipates your needs\fR.
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
See what bj has been up to. Add
.BR \-\-running ", " \-\-failed ", or " \-\-done
to filter by... performance.
.TP
.BI \-\-logs " [id]"
Watch bj's output. Every moan, groan, and triumphant success message.
Defaults to the most recent job if you can't remember which one.
.TP
.BI \-\-kill " [id]"
Sometimes you need to stop things abruptly. No judgment.
.TP
.BR \-\-retry [ =\fIN\fR ]
bj doesn't give up easily. Use alone to retry a ruined job, or with a
command to keep trying until satisfaction (or N attempts, whichever
comes first).
.TP
.BI \-\-delay " secs"
Pace yourself. Wait between retry attempts.
.TP
.B \-\-prune
Clean up when bj is finished. Removes completed jobs and their logs.
A tidy bj is a happy bj.
.TP
.B \-\-gc
Find jobs that were unexpectedly ruined. Sometimes things end badly
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
bj \-\-retry npm test       # Keep going until it passes
bj \-\-list                 # Check on bj's progress
bj \-\-logs                 # See the latest output
bj \-\-kill                 # Stop it right there
bj \-\-prune                # Clean up after a satisfying session
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

	// Shell completions - fish
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
complete -c bj -l restart -d "Restart on failure with 5s delay"
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

	// Shell completions - zsh
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
        '--restart[Restart on failure with 5s delay]' \
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

	// Shell init - fish
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

	// Shell init - zsh
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

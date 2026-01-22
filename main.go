package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/metruzanca/bj/internal/config"
	"github.com/metruzanca/bj/internal/runner"
	"github.com/metruzanca/bj/internal/tracker"
)

// ANSI color codes
const (
	colorReset = "\033[0m"
	colorDim   = "\033[2m"
	colorRed   = "\033[31m"
)

// Global flags
var jsonOutput bool
var helpRequested bool
var retryFlag int    // -1 = not set, 0 = unlimited, N = max attempts
var retryJobID int   // 0 = not set (use latest), N = specific job ID
var retryDelay int   // delay in seconds between retries (default 1)

// List filter flags
var listRunning bool
var listFailed bool
var listDone bool

func main() {
	// Initialize retryFlag to -1 (not set) and delay to 1 second
	retryFlag = -1
	retryDelay = 1

	// Check for --json, --help, --retry[=N], and --id flags anywhere in args
	args := filterArgs(os.Args[1:], &jsonOutput, &helpRequested, &retryFlag, &retryJobID)

	// Handle help for --retry
	if helpRequested && retryFlag >= 0 {
		printUsage("--retry")
		os.Exit(0)
	}

	// Handle general help
	if helpRequested || (len(args) < 1 && retryFlag < 0) {
		if len(args) > 0 {
			printUsage(args[0])
		} else {
			printUsage("")
		}
		if helpRequested {
			os.Exit(0)
		}
		os.Exit(1)
	}

	// Validate --id is only used with --retry
	if retryJobID != 0 && retryFlag < 0 {
		exitWithError("--id only makes sense with --retry. They go together like... well, you know.")
	}

	// Validate --delay is only used with --retry
	if retryDelay != 1 && retryFlag < 0 {
		exitWithError("--delay without --retry? bj needs something to delay between.")
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		exitWithError("bj couldn't get comfortable: %v", err)
	}

	// Create tracker
	t, err := tracker.New()
	if err != nil {
		exitWithError("bj lost track of things: %v", err)
	}

	// Auto-prune if configured
	if cfg.AutoPruneHours > 0 {
		t.PruneOlderThan(time.Duration(cfg.AutoPruneHours) * time.Hour)
	}

	// Handle --retry as a modifier flag
	if retryFlag >= 0 {
		if len(args) > 0 {
			// Run new command with retry
			command := strings.Join(args, " ")
			runCommandWithRetry(cfg, t, command, retryFlag, retryDelay)
		} else {
			// Retry existing job
			retryExistingJob(cfg, t, retryJobID, retryFlag, retryDelay)
		}
		return
	}

	// Parse first argument to determine action
	arg := args[0]

	switch {
	case arg == "--completion":
		if len(args) < 2 {
			exitWithError("Usage: bj --completion <fish|zsh>")
		}
		printCompletion(args[1])

	case arg == "--init":
		if len(args) < 2 {
			exitWithError("Usage: bj --init <fish|zsh>")
		}
		printInit(args[1], cfg, t)

	case arg == "--list":
		listJobs(t)

	case arg == "--ids":
		printJobIDs(t)

	case arg == "--logs":
		var jobID int
		if len(args) > 1 {
			id, err := strconv.Atoi(args[1])
			if err != nil {
				exitWithError("bj needs a valid number, not '%s'", args[1])
			}
			jobID = id
		}
		viewLogs(cfg, t, jobID)

	case arg == "--prune":
		pruneJobs(t)

	case arg == "--gc":
		garbageCollect(t)

	case arg == "--kill":
		var jobID int
		if len(args) > 1 {
			id, err := strconv.Atoi(args[1])
			if err != nil {
				exitWithError("bj needs a valid number, not '%s'", args[1])
			}
			jobID = id
		}
		killJob(t, jobID)

	case arg == "--complete":
		// Internal command: mark job as complete
		if len(args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: bj --complete <job_id> <exit_code>\n")
			os.Exit(1)
		}
		jobID, err := strconv.Atoi(args[1])
		if err != nil {
			exitWithError("bj needs a valid job ID, not '%s'", args[1])
		}
		exitCode, err := strconv.Atoi(args[2])
		if err != nil {
			exitWithError("bj needs a valid exit code, not '%s'", args[2])
		}
		r := runner.New(cfg, t)
		if err := r.Complete(jobID, exitCode); err != nil {
			exitWithError("bj couldn't finish properly: %v", err)
		}

	default:
		// Check for unknown flags - don't accidentally run them as commands
		if strings.HasPrefix(arg, "-") {
			exitWithError("Unknown flag: %s. Try 'bj --help' for usage.", arg)
		}
		// Everything else is treated as a command to run
		command := strings.Join(args, " ")
		runCommand(cfg, t, command)
	}
}

// filterArgs removes global flags, sets flag values, returns remaining args
func filterArgs(args []string, jsonFlag *bool, helpFlag *bool, retryFlagOut *int, retryJobIDOut *int) []string {
	var filtered []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--json":
			*jsonFlag = true
		case arg == "--help" || arg == "-h":
			*helpFlag = true
		case arg == "--retry":
			*retryFlagOut = 0 // unlimited
		case strings.HasPrefix(arg, "--retry="):
			val := strings.TrimPrefix(arg, "--retry=")
			n, err := strconv.Atoi(val)
			if err != nil || n < 1 {
				fmt.Fprintf(os.Stderr, "bj needs a positive number for retry limit, not '%s'\n", val)
				os.Exit(1)
			}
			*retryFlagOut = n
		case arg == "--id":
			// --id requires a following positive number
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "--id needs a job ID to go with it")
				os.Exit(1)
			}
			i++
			id, err := strconv.Atoi(args[i])
			if err != nil {
				fmt.Fprintf(os.Stderr, "bj needs a valid job ID, not '%s'\n", args[i])
				os.Exit(1)
			}
			if id < 1 {
				fmt.Fprintf(os.Stderr, "Job IDs start at 1. '%d' won't satisfy bj.\n", id)
				os.Exit(1)
			}
			*retryJobIDOut = id
		case arg == "--running":
			listRunning = true
		case arg == "--failed":
			listFailed = true
		case arg == "--done":
			listDone = true
		case arg == "--delay":
			// --delay requires a following number
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "--delay needs a number of seconds")
				os.Exit(1)
			}
			i++
			d, err := strconv.Atoi(args[i])
			if err != nil || d < 0 {
				fmt.Fprintf(os.Stderr, "bj needs a non-negative delay, not '%s'\n", args[i])
				os.Exit(1)
			}
			retryDelay = d
		case strings.HasPrefix(arg, "--delay="):
			val := strings.TrimPrefix(arg, "--delay=")
			d, err := strconv.Atoi(val)
			if err != nil || d < 0 {
				fmt.Fprintf(os.Stderr, "bj needs a non-negative delay, not '%s'\n", val)
				os.Exit(1)
			}
			retryDelay = d
		default:
			filtered = append(filtered, arg)
		}
	}
	return filtered
}

// exitWithError prints error (or JSON) and exits
func exitWithError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if jsonOutput {
		outputJSON(map[string]interface{}{"error": msg})
	} else {
		fmt.Fprintln(os.Stderr, msg)
	}
	os.Exit(1)
}

// outputJSON marshals and prints JSON
func outputJSON(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}

func printUsage(command string) {
	switch command {
	case "--list":
		fmt.Println(`bj --list - See what bj is working on

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
  bj --list --json    Get the raw details for scripting`)

	case "--logs":
		fmt.Println(`bj --logs - Watch bj's performance

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
  bj --logs --json  Get logs in JSON format`)

	case "--prune":
		fmt.Println(`bj --prune - Clean up when bj is finished

Usage: bj --prune [--json]

Removes all completed jobs (any exit code) from the job list and deletes
their log files. Only running jobs are kept. If all jobs are pruned, the
ID counter resets to 1.

Options:
  --json    Output prune count as JSON

Examples:
  bj --prune        Wipe the slate clean after bj is done`)

	case "--kill":
		fmt.Println(`bj --kill - Make bj stop what it's doing

Usage: bj --kill [id] [--json]

Terminates a running job. Sends SIGTERM to the process group, stopping
the entire job tree. If no ID is specified, kills the most recent running job.

Arguments:
  id        Job ID to kill (optional, defaults to latest running)

Options:
  --json    Output killed job info as JSON

Examples:
  bj --kill         Stop the latest job mid-stroke
  bj --kill 5       Pull out of job #5 specifically`)

	case "--gc":
		fmt.Println(`bj --gc - Find jobs that ended unexpectedly

Usage: bj --gc [--json]

Detects orphaned jobs that appear to be running but whose process is gone
(e.g., after a crash or reboot). These ruined jobs are marked as failed
with exit code -1.

Options:
  --json    Output collected count as JSON

Examples:
  bj --gc           Clean up after an unexpected interruption`)

	case "--retry":
		fmt.Println(`bj --retry - Keep going until bj finishes the job

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
  bj --retry=3 --delay 10 --id 5   Retry job #5 up to 3 times, 10s apart`)

	case "--completion":
		fmt.Println(`bj --completion - Output shell completions

Usage: bj --completion <shell>

Outputs tab-completion definitions for the specified shell.
Typically you'd redirect this to a completions file.

Arguments:
  shell     Shell type: fish, zsh

Examples:
  bj --completion fish > ~/.config/fish/completions/bj.fish
  bj --completion zsh > ~/.zsh/completions/_bj

Note: If you use --init, completions are installed automatically.`)

	case "--init":
		fmt.Println(`bj --init - Set up shell integration

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
prompt to always know when bj is busy.`)

	default:
		fmt.Println(`bj - Background Jobs

Give bj a command and it'll handle the rest while you sit back and relax.

Usage:
  bj <command>              Slip a command in the background
  bj --retry[=N] <command>  Run with retry until success (or N attempts)
  bj --list                 See what bj is working on
  bj --logs [id]            Watch bj's performance
  bj --kill [id]            Stop a job mid-action
  bj --retry[=N] [--id ID]  Retry a ruined job
  bj --prune                Clean up when bj is finished
  bj --gc                   Find jobs that were ruined unexpectedly

Shell Integration:
  bj --completion <sh>  Output shell completions (fish, zsh)
  bj --init <sh>        Output prompt integration (fish, zsh)

Options:
  --retry[=N]         Keep trying until success (or limit to N attempts)
  --id ID             Specify job ID for --retry (defaults to most recent)
  --json              Output in JSON format (works with all commands)
  -h, --help          Show this help (use with commands for detailed help)

Examples:
  bj sleep 10               Let bj handle your sleep needs
  bj npm install            bj npm while you grab coffee
  bj --retry npm test       Keep testing until it passes
  bj --retry=3 make build   Try building up to 3 times
  bj --list                 Check how bj is doing
  bj --logs                 See bj's latest output
  bj --kill                 Stop the current job abruptly
  bj --gc                   Find ruined jobs after a crash
  bj --retry                Retry the most recent ruined job
  bj --prune                Tidy up after a satisfying bj

Shell Setup:
  Fish: echo 'bj --init fish | source' >> ~/.config/fish/config.fish
  Zsh:  echo 'eval "$(bj --init zsh)"' >> ~/.zshrc
        (ensure ~/.zsh/completions is in your fpath before compinit)`)
	}
}

func runCommand(cfg *config.Config, t *tracker.Tracker, command string) {
	r := runner.New(cfg, t)
	jobID, err := r.Run(command)
	if err != nil {
		exitWithError("bj couldn't get it up: %v", err)
	}
	if jsonOutput {
		outputJSON(map[string]interface{}{
			"id":      jobID,
			"command": command,
			"status":  "started",
		})
	} else {
		fmt.Printf("[%d] bj is on it: %s\n", jobID, command)
	}
}

type jobRow struct {
	id       int
	status   string
	start    string
	duration string
	cmd      string
	isError  bool
	isDone   bool
}

// relativeTime returns a human-friendly relative time string
func relativeTime(t time.Time) string {
	d := time.Since(t)

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	case d < 24*time.Hour:
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("Jan 02")
	}
}

func listJobs(t *tracker.Tracker) {
	jobs, err := t.List()
	if err != nil {
		exitWithError("bj can't show you what it's got: %v", err)
	}

	// Apply filters if any are set
	hasFilter := listRunning || listFailed || listDone
	if hasFilter {
		var filtered []tracker.Job
		for _, job := range jobs {
			if listRunning && job.ExitCode == nil {
				filtered = append(filtered, job)
			} else if listFailed && job.ExitCode != nil && *job.ExitCode != 0 {
				filtered = append(filtered, job)
			} else if listDone && job.ExitCode != nil && *job.ExitCode == 0 {
				filtered = append(filtered, job)
			}
		}
		jobs = filtered
	}

	if len(jobs) == 0 {
		if jsonOutput {
			outputJSON([]interface{}{})
		} else {
			if hasFilter {
				fmt.Println("No jobs match your criteria. bj has nothing to show.")
			} else {
				fmt.Println("bj has nothing going on. Give it something to do!")
			}
		}
		return
	}

	// JSON output - return raw job data
	if jsonOutput {
		outputJSON(jobs)
		return
	}

	// Build rows first
	var rows []jobRow
	for _, job := range jobs {
		row := jobRow{id: job.ID}
		row.status = "running"
		row.duration = time.Since(job.StartTime).Round(time.Second).String()

		if job.ExitCode != nil {
			if *job.ExitCode == 0 {
				row.status = "done"
				row.isDone = true
			} else {
				row.status = fmt.Sprintf("exit(%d)", *job.ExitCode)
				row.isError = true
			}
			if job.EndTime != nil {
				row.duration = job.EndTime.Sub(job.StartTime).Round(time.Second).String()
			}
		}

		row.start = relativeTime(job.StartTime)

		// Truncate long commands
		row.cmd = job.Command
		if len(row.cmd) > 40 {
			row.cmd = row.cmd[:37] + "..."
		}

		rows = append(rows, row)
	}

	// Calculate column widths
	idW, statusW, startW, durW := 2, 6, 5, 8 // header widths
	for _, r := range rows {
		if w := len(fmt.Sprintf("%d", r.id)); w > idW {
			idW = w
		}
		if w := len(r.status); w > statusW {
			statusW = w
		}
		if w := len(r.start); w > startW {
			startW = w
		}
		if w := len(r.duration); w > durW {
			durW = w
		}
	}

	// Print header
	fmt.Printf("%-*s  %-*s  %-*s  %-*s  %s\n", idW, "ID", statusW, "STATUS", startW, "START", durW, "DURATION", "COMMAND")

	// Print rows with colors
	for _, r := range rows {
		line := fmt.Sprintf("%-*d  %-*s  %-*s  %-*s  %s", idW, r.id, statusW, r.status, startW, r.start, durW, r.duration, r.cmd)
		if r.isError {
			// Dim row with red status
			statusStart := idW + 2
			statusEnd := statusStart + statusW
			fmt.Printf("%s%s%s%s%s%s%s\n",
				colorDim, line[:statusStart],
				colorRed, line[statusStart:statusEnd],
				colorReset, colorDim, line[statusEnd:]+colorReset)
		} else if r.isDone {
			fmt.Printf("%s%s%s\n", colorDim, line, colorReset)
		} else {
			fmt.Println(line)
		}
	}
}

// printJobIDs outputs job IDs for shell completion (no jq needed)
// Respects --running, --failed, --done filters
func printJobIDs(t *tracker.Tracker) {
	jobs, err := t.List()
	if err != nil {
		os.Exit(1) // Silent fail for completions
	}

	for _, job := range jobs {
		// Apply filters if any are set
		if listRunning && job.ExitCode != nil {
			continue
		}
		if listFailed && (job.ExitCode == nil || *job.ExitCode == 0) {
			continue
		}
		if listDone && (job.ExitCode == nil || *job.ExitCode != 0) {
			continue
		}
		fmt.Println(job.ID)
	}
}

func pruneJobs(t *tracker.Tracker) {
	count, err := t.Prune()
	if err != nil {
		exitWithError("bj made a mess while cleaning up: %v", err)
	}
	if jsonOutput {
		outputJSON(map[string]interface{}{"pruned": count})
	} else if count == 0 {
		fmt.Println("Nothing to clean up. bj keeps it tidy.")
	} else {
		fmt.Printf("Wiped away %d finished job(s). Fresh and ready for more.\n", count)
	}
}

func garbageCollect(t *tracker.Tracker) {
	count, err := t.GarbageCollect()
	if err != nil {
		exitWithError("bj had trouble cleaning up its mess: %v", err)
	}
	if jsonOutput {
		outputJSON(map[string]interface{}{"collected": count})
	} else if count == 0 {
		fmt.Println("No orphaned jobs found. bj keeps track of all its encounters.")
	} else {
		fmt.Printf("Found %d ruined job(s) that ended without bj noticing. Marked as failed.\n", count)
	}
}

func killJob(t *tracker.Tracker, jobID int) {
	var job *tracker.Job
	var err error

	if jobID == 0 {
		// Find the most recent running job
		job, err = t.LatestRunning()
		if err != nil {
			exitWithError("bj can't check its active sessions: %v", err)
		}
		if job == nil {
			if jsonOutput {
				exitWithError("no running jobs to kill")
			}
			fmt.Println("bj isn't doing anything right now. Nothing to stop!")
			os.Exit(0)
		}
		jobID = job.ID
	}

	job, err = t.Kill(jobID)
	if err != nil {
		exitWithError("bj couldn't pull out: %v", err)
	}

	if jsonOutput {
		outputJSON(map[string]interface{}{
			"id":      job.ID,
			"command": job.Command,
			"status":  "killed",
		})
	} else {
		fmt.Printf("[%d] bj stopped abruptly: %s\n", job.ID, job.Command)
	}
}

// runCommandWithRetry runs a new command with retry logic
func runCommandWithRetry(cfg *config.Config, t *tracker.Tracker, command string, maxAttempts int, delaySecs int) {
	pwd, err := os.Getwd()
	if err != nil {
		exitWithError("bj couldn't figure out where you are: %v", err)
	}

	r := runner.New(cfg, t)
	jobID, err := r.RunWithRetry(command, pwd, maxAttempts, delaySecs)
	if err != nil {
		exitWithError("bj couldn't get it up: %v", err)
	}

	if jsonOutput {
		outputJSON(map[string]interface{}{
			"id":           jobID,
			"command":      command,
			"status":       "started",
			"max_attempts": maxAttempts,
			"delay_secs":   delaySecs,
		})
	} else {
		if maxAttempts == 0 {
			fmt.Printf("[%d] bj will keep edging until it succeeds: %s\n", jobID, command)
		} else if maxAttempts == 1 {
			fmt.Printf("[%d] bj will give it one shot: %s\n", jobID, command)
		} else {
			fmt.Printf("[%d] bj will tease up to %d times before giving up: %s\n", jobID, maxAttempts, command)
		}
	}
}

// retryExistingJob retries a failed job (or most recent failure if jobID is 0)
func retryExistingJob(cfg *config.Config, t *tracker.Tracker, jobID int, maxAttempts int, delaySecs int) {
	var job *tracker.Job
	var err error

	if jobID == 0 {
		// Find the most recent failed job
		jobs, err := t.List()
		if err != nil {
			exitWithError("bj can't check its history: %v", err)
		}
		for _, j := range jobs {
			if j.ExitCode != nil && *j.ExitCode != 0 {
				job = &j
				break
			}
		}
		if job == nil {
			if jsonOutput {
				exitWithError("no ruined jobs to retry")
			}
			fmt.Println("bj hasn't ruined anything yet. Nothing to retry!")
			os.Exit(0)
		}
	} else {
		job, err = t.Get(jobID)
		if err != nil {
			exitWithError("bj can't find that one: %v", err)
		}
		if job == nil {
			exitWithError("Job %d? bj doesn't remember that.", jobID)
		}
	}

	// Check if job actually failed
	if job.ExitCode == nil {
		exitWithError("Job %d is still going. bj doesn't stop until it's done.", job.ID)
	}
	if *job.ExitCode == 0 {
		exitWithError("Job %d already finished successfully. No need to go again.", job.ID)
	}

	// Run the job with retry wrapper
	r := runner.New(cfg, t)
	newJobID, err := r.RunWithRetry(job.Command, job.PWD, maxAttempts, delaySecs)
	if err != nil {
		exitWithError("bj couldn't get started again: %v", err)
	}

	if jsonOutput {
		outputJSON(map[string]interface{}{
			"id":           newJobID,
			"command":      job.Command,
			"status":       "started",
			"max_attempts": maxAttempts,
			"delay_secs":   delaySecs,
			"original_job": job.ID,
		})
	} else {
		if maxAttempts == 0 {
			fmt.Printf("[%d] bj will keep edging until it succeeds: %s\n", newJobID, job.Command)
		} else if maxAttempts == 1 {
			fmt.Printf("[%d] bj is giving it one more go: %s\n", newJobID, job.Command)
		} else {
			fmt.Printf("[%d] bj will tease up to %d times before giving up: %s\n", newJobID, maxAttempts, job.Command)
		}
	}
}

func viewLogs(cfg *config.Config, t *tracker.Tracker, jobID int) {
	var job *tracker.Job
	var err error

	if jobID == 0 {
		job, err = t.Latest()
		if err != nil {
			exitWithError("bj can't recall the last session: %v", err)
		}
		if job == nil {
			if jsonOutput {
				exitWithError("no jobs found")
			}
			fmt.Println("bj hasn't done anything yet. Get it started first!")
			os.Exit(0)
		}
	} else {
		job, err = t.Get(jobID)
		if err != nil {
			exitWithError("bj can't find that one: %v", err)
		}
		if job == nil {
			exitWithError("Job %d? bj doesn't remember that.", jobID)
		}
	}

	// Check if log file exists
	if _, err := os.Stat(job.LogFile); os.IsNotExist(err) {
		exitWithError("bj swallowed the logs. File not found: %s", job.LogFile)
	}

	// JSON mode: output log contents
	if jsonOutput {
		content, err := os.ReadFile(job.LogFile)
		if err != nil {
			exitWithError("bj couldn't read the logs: %v", err)
		}
		outputJSON(map[string]interface{}{
			"job":     job,
			"content": string(content),
		})
		return
	}

	// Open with configured viewer
	cmd := exec.Command(cfg.Viewer, job.LogFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		exitWithError("bj choked while opening logs: %v", err)
	}
}

func printCompletion(shell string) {
	switch shell {
	case "fish":
		fmt.Print(fishCompletion)
	case "zsh":
		fmt.Print(zshCompletion)
	default:
		exitWithError("Unknown shell: %s. bj knows fish and zsh.", shell)
	}
}

func printInit(shell string, cfg *config.Config, t *tracker.Tracker) {
	// Run housekeeping silently on shell init
	// This is a good time to clean up orphaned jobs and auto-prune
	t.GarbageCollect()
	if cfg.AutoPruneHours > 0 {
		t.PruneOlderThan(time.Duration(cfg.AutoPruneHours) * time.Hour)
	}

	switch shell {
	case "fish":
		// Write completions file if needed, then output prompt init
		writeFishCompletions()
		fmt.Print(fishInit)
	case "zsh":
		// Write completions file if needed, then output prompt init
		writeZshCompletions()
		fmt.Print(zshInit)
	default:
		exitWithError("Unknown shell: %s. bj knows fish and zsh.", shell)
	}
}

// writeFishCompletions writes fish completions to ~/.config/fish/completions/bj.fish
// Only writes if file doesn't exist or content has changed
func writeFishCompletions() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return // silently fail, completions are optional
	}

	completionsDir := filepath.Join(homeDir, ".config", "fish", "completions")
	completionsFile := filepath.Join(completionsDir, "bj.fish")

	// Check if content has changed
	existing, err := os.ReadFile(completionsFile)
	if err == nil && string(existing) == fishCompletion {
		return // already up to date
	}

	// Ensure directory exists
	if err := os.MkdirAll(completionsDir, 0755); err != nil {
		return // silently fail
	}

	// Write completions
	os.WriteFile(completionsFile, []byte(fishCompletion), 0644)
}

// writeZshCompletions writes zsh completions to ~/.zsh/completions/_bj
// Only writes if file doesn't exist or content has changed
func writeZshCompletions() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return // silently fail, completions are optional
	}

	completionsDir := filepath.Join(homeDir, ".zsh", "completions")
	completionsFile := filepath.Join(completionsDir, "_bj")

	// Check if content has changed
	existing, err := os.ReadFile(completionsFile)
	if err == nil && string(existing) == zshCompletion {
		return // already up to date
	}

	// Ensure directory exists
	if err := os.MkdirAll(completionsDir, 0755); err != nil {
		return // silently fail
	}

	// Write completions
	os.WriteFile(completionsFile, []byte(zshCompletion), 0644)
}

const fishCompletion = `# bj fish completions
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

# Job ID completion for --logs and --kill
complete -c bj -n "__fish_seen_argument -l logs" -a "(bj --ids 2>/dev/null)" -d "Job ID"
complete -c bj -n "__fish_seen_argument -l kill" -a "(bj --ids --running 2>/dev/null)" -d "Running job ID"
`

const zshCompletion = `#compdef bj
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
        '*:command:_command_names'
}

_bj "$@"
`

const fishInit = `# bj fish shell integration
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
`

const zshInit = `# bj zsh shell integration
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
`
